package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// MessageType defines the type of WebSocket messages
type MessageType string

const (
	MessageTypeExperimentUpdate  MessageType = "experiment_update"
	MessageTypeMetricUpdate      MessageType = "metric_update"
	MessageTypeStatusChange      MessageType = "status_change"
	MessageTypeNotification      MessageType = "notification"
	MessageTypeHeartbeat         MessageType = "heartbeat"
	MessageTypeSubscribe         MessageType = "subscribe"
	MessageTypeUnsubscribe       MessageType = "unsubscribe"
	MessageTypeError             MessageType = "error"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType     `json:"type"`
	Topic     string          `json:"topic,omitempty"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
}

// Client represents a WebSocket client
type Client struct {
	ID           string
	conn         *websocket.Conn
	send         chan []byte
	hub          *Hub
	topics       map[string]bool
	topicsMutex  sync.RWMutex
	lastActivity time.Time
}

// Hub maintains active WebSocket clients and broadcasts messages
type Hub struct {
	clients      map[string]*Client
	clientsMutex sync.RWMutex
	
	Broadcast    chan *Message
	Register     chan *Client
	Unregister   chan *Client
	
	topics       map[string]map[string]*Client // topic -> clientID -> client
	topicsMutex  sync.RWMutex
	
	logger       *zap.Logger
}

// NewHub creates a new WebSocket hub
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client, 16),
		unregister: make(chan *Client, 16),
		topics:     make(map[string]map[string]*Client),
		logger:     logger,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastMessage(message)

		case <-ticker.C:
			h.sendHeartbeats()
			h.cleanupInactiveClients()
		}
	}
}

// RegisterClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.clientsMutex.Lock()
	h.clients[client.ID] = client
	h.clientsMutex.Unlock()

	h.logger.Info("WebSocket client registered", zap.String("clientID", client.ID))
	
	// Send welcome message
	welcomeMsg := &Message{
		Type:      MessageTypeNotification,
		Data:      json.RawMessage(`{"message": "Connected to Phoenix Platform WebSocket"}`),
		Timestamp: time.Now(),
	}
	if data, err := json.Marshal(welcomeMsg); err == nil {
		client.send <- data
	}
}

// UnregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.clientsMutex.Lock()
	if _, exists := h.clients[client.ID]; exists {
		delete(h.clients, client.ID)
		close(client.send)
	}
	h.clientsMutex.Unlock()

	// Remove from all topics
	h.topicsMutex.Lock()
	for topic := range client.topics {
		if clients, ok := h.topics[topic]; ok {
			delete(clients, client.ID)
			if len(clients) == 0 {
				delete(h.topics, topic)
			}
		}
	}
	h.topicsMutex.Unlock()

	h.logger.Info("WebSocket client unregistered", zap.String("clientID", client.ID))
}

// BroadcastMessage broadcasts a message to relevant clients
func (h *Hub) broadcastMessage(message *Message) {
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal message", zap.Error(err))
		return
	}

	// If message has a topic, only send to subscribed clients
	if message.Topic != "" {
		h.topicsMutex.RLock()
		if clients, ok := h.topics[message.Topic]; ok {
			for _, client := range clients {
				select {
				case client.send <- data:
				default:
					// Client's send channel is full, skip
					h.logger.Warn("Client send buffer full", zap.String("clientID", client.ID))
				}
			}
		}
		h.topicsMutex.RUnlock()
	} else {
		// Broadcast to all clients
		h.clientsMutex.RLock()
		for _, client := range h.clients {
			select {
			case client.send <- data:
			default:
				// Client's send channel is full, skip
				h.logger.Warn("Client send buffer full", zap.String("clientID", client.ID))
			}
		}
		h.clientsMutex.RUnlock()
	}
}

// Subscribe subscribes a client to a topic
func (h *Hub) Subscribe(clientID, topic string) {
	h.clientsMutex.RLock()
	client, exists := h.clients[clientID]
	h.clientsMutex.RUnlock()

	if !exists {
		return
	}

	client.topicsMutex.Lock()
	client.topics[topic] = true
	client.topicsMutex.Unlock()

	h.topicsMutex.Lock()
	if h.topics[topic] == nil {
		h.topics[topic] = make(map[string]*Client)
	}
	h.topics[topic][clientID] = client
	h.topicsMutex.Unlock()

	h.logger.Info("Client subscribed to topic", 
		zap.String("clientID", clientID),
		zap.String("topic", topic))
}

// Unsubscribe unsubscribes a client from a topic
func (h *Hub) Unsubscribe(clientID, topic string) {
	h.clientsMutex.RLock()
	client, exists := h.clients[clientID]
	h.clientsMutex.RUnlock()

	if !exists {
		return
	}

	client.topicsMutex.Lock()
	delete(client.topics, topic)
	client.topicsMutex.Unlock()

	h.topicsMutex.Lock()
	if clients, ok := h.topics[topic]; ok {
		delete(clients, clientID)
		if len(clients) == 0 {
			delete(h.topics, topic)
		}
	}
	h.topicsMutex.Unlock()

	h.logger.Info("Client unsubscribed from topic", 
		zap.String("clientID", clientID),
		zap.String("topic", topic))
}

// BroadcastExperimentUpdate broadcasts an experiment update
func (h *Hub) BroadcastExperimentUpdate(experimentID string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		h.logger.Error("Failed to marshal experiment update", zap.Error(err))
		return
	}

	message := &Message{
		Type:      MessageTypeExperimentUpdate,
		Topic:     "experiment:" + experimentID,
		Data:      jsonData,
		Timestamp: time.Now(),
	}

	h.Broadcast <- message
}

// BroadcastMetricUpdate broadcasts a metric update
func (h *Hub) BroadcastMetricUpdate(experimentID string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		h.logger.Error("Failed to marshal metric update", zap.Error(err))
		return
	}

	message := &Message{
		Type:      MessageTypeMetricUpdate,
		Topic:     "metrics:" + experimentID,
		Data:      jsonData,
		Timestamp: time.Now(),
	}

	h.Broadcast <- message
}

// sendHeartbeats sends heartbeat messages to all clients
func (h *Hub) sendHeartbeats() {
	heartbeat := &Message{
		Type:      MessageTypeHeartbeat,
		Data:      json.RawMessage(`{"status": "alive"}`),
		Timestamp: time.Now(),
	}

	h.Broadcast <- heartbeat
}

// cleanupInactiveClients removes clients that haven't been active
func (h *Hub) cleanupInactiveClients() {
	threshold := time.Now().Add(-5 * time.Minute)
	
	h.clientsMutex.RLock()
	var inactiveClients []*Client
	for _, client := range h.clients {
		if client.lastActivity.Before(threshold) {
			inactiveClients = append(inactiveClients, client)
		}
	}
	h.clientsMutex.RUnlock()

	for _, client := range inactiveClients {
		h.logger.Info("Removing inactive client", zap.String("clientID", client.ID))
		h.Unregister <- client
	}
}