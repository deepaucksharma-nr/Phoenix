package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// Hub maintains active WebSocket connections and broadcasts events
type Hub struct {
	// Registered clients
	clients map[*Client]bool
	
	// Client management channels
	register   chan *Client
	unregister chan *Client
	
	// Event broadcasting
	broadcast chan Event
	
	// Event-specific channels for different update types
	agentUpdates      chan AgentStatusUpdate
	experimentUpdates chan ExperimentUpdateEvent
	metricFlows       chan MetricFlowUpdate
	taskProgress      chan TaskProgressUpdate
	alerts            chan AlertEvent
	
	// Mutex for thread-safe operations
	mu sync.RWMutex
	
	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		clients:           make(map[*Client]bool),
		register:          make(chan *Client),
		unregister:        make(chan *Client),
		broadcast:         make(chan Event, 256),
		agentUpdates:      make(chan AgentStatusUpdate, 100),
		experimentUpdates: make(chan ExperimentUpdateEvent, 100),
		metricFlows:       make(chan MetricFlowUpdate, 100),
		taskProgress:      make(chan TaskProgressUpdate, 100),
		alerts:            make(chan AlertEvent, 50),
		ctx:               ctx,
		cancel:            cancel,
	}
}

// Run starts the hub's event loop
func (h *Hub) Run() {
	// Start metric flow ticker for periodic updates
	metricTicker := time.NewTicker(1 * time.Second)
	defer metricTicker.Stop()
	
	// Start agent status ticker
	agentTicker := time.NewTicker(5 * time.Second)
	defer agentTicker.Stop()
	
	for {
		select {
		case <-h.ctx.Done():
			return
			
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client registered: %s", client.id)
			
			// Send initial state to new client
			go h.sendInitialState(client)
			
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.mu.Unlock()
				log.Printf("Client unregistered: %s", client.id)
			} else {
				h.mu.Unlock()
			}
			
		case event := <-h.broadcast:
			h.broadcastToClients(event)
			
		case update := <-h.agentUpdates:
			event := Event{
				Type:      EventAgentStatus,
				Timestamp: time.Now(),
				Data:      update,
			}
			h.broadcastToClients(event)
			
		case update := <-h.experimentUpdates:
			event := Event{
				Type:      EventExperimentUpdate,
				Timestamp: time.Now(),
				Data:      update,
			}
			h.broadcastToClients(event)
			
		case update := <-h.metricFlows:
			event := Event{
				Type:      EventMetricFlow,
				Timestamp: time.Now(),
				Data:      update,
			}
			h.broadcastToClients(event)
			
		case update := <-h.taskProgress:
			event := Event{
				Type:      EventTaskProgress,
				Timestamp: time.Now(),
				Data:      update,
			}
			h.broadcastToClients(event)
			
		case alert := <-h.alerts:
			event := Event{
				Type:      EventAlert,
				Timestamp: time.Now(),
				Data:      alert,
			}
			h.broadcastToClients(event)
			
		case <-metricTicker.C:
			// Broadcast periodic metric flow updates
			// This would fetch real data from the metrics service
			h.broadcastMetricFlow()
			
		case <-agentTicker.C:
			// Broadcast periodic agent status updates
			// This would fetch real data from the agent registry
			h.broadcastAgentStatus()
		}
	}
}

// Stop gracefully shuts down the hub
func (h *Hub) Stop() {
	h.cancel()
	
	// Close all client connections
	h.mu.Lock()
	for client := range h.clients {
		close(client.send)
	}
	h.mu.Unlock()
}

// SendAgentUpdate sends an agent status update
func (h *Hub) SendAgentUpdate(update AgentStatusUpdate) {
	select {
	case h.agentUpdates <- update:
	case <-time.After(100 * time.Millisecond):
		log.Println("Agent update channel full, dropping update")
	}
}

// SendExperimentUpdate sends an experiment update
func (h *Hub) SendExperimentUpdate(update ExperimentUpdateEvent) {
	select {
	case h.experimentUpdates <- update:
	case <-time.After(100 * time.Millisecond):
		log.Println("Experiment update channel full, dropping update")
	}
}

// SendMetricFlow sends a metric flow update
func (h *Hub) SendMetricFlow(update MetricFlowUpdate) {
	select {
	case h.metricFlows <- update:
	case <-time.After(100 * time.Millisecond):
		log.Println("Metric flow channel full, dropping update")
	}
}

// SendTaskProgress sends a task progress update
func (h *Hub) SendTaskProgress(update TaskProgressUpdate) {
	select {
	case h.taskProgress <- update:
	case <-time.After(100 * time.Millisecond):
		log.Println("Task progress channel full, dropping update")
	}
}

// SendAlert sends an alert
func (h *Hub) SendAlert(alert AlertEvent) {
	select {
	case h.alerts <- alert:
	case <-time.After(100 * time.Millisecond):
		log.Println("Alert channel full, dropping alert")
	}
}

// broadcastToClients sends an event to all connected clients
func (h *Hub) broadcastToClients(event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}
	
	for client := range h.clients {
		// Check if client is subscribed to this event type
		if client.isSubscribed(event.Type) {
			select {
			case client.send <- data:
			default:
				// Client's send channel is full, close it
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

// sendInitialState sends initial state to a newly connected client
func (h *Hub) sendInitialState(client *Client) {
	// Send current agent statuses
	// This would fetch from the actual data store
	initialAgents := h.getInitialAgentStatus()
	for _, agent := range initialAgents {
		event := Event{
			Type:      EventAgentStatus,
			Timestamp: time.Now(),
			Data:      agent,
		}
		if data, err := json.Marshal(event); err == nil {
			select {
			case client.send <- data:
			case <-time.After(100 * time.Millisecond):
				return
			}
		}
	}
	
	// Send current experiments
	initialExperiments := h.getInitialExperiments()
	for _, exp := range initialExperiments {
		event := Event{
			Type:      EventExperimentUpdate,
			Timestamp: time.Now(),
			Data:      exp,
		}
		if data, err := json.Marshal(event); err == nil {
			select {
			case client.send <- data:
			case <-time.After(100 * time.Millisecond):
				return
			}
		}
	}
}

// Placeholder methods that would integrate with actual services
func (h *Hub) broadcastMetricFlow() {
	// TODO: Fetch real metric flow data from metrics service
}

func (h *Hub) broadcastAgentStatus() {
	// TODO: Fetch real agent status from agent registry
}

func (h *Hub) getInitialAgentStatus() []AgentStatusUpdate {
	// TODO: Fetch initial agent status from data store
	return []AgentStatusUpdate{}
}

func (h *Hub) getInitialExperiments() []ExperimentUpdateEvent {
	// TODO: Fetch initial experiments from data store
	return []ExperimentUpdateEvent{}
}