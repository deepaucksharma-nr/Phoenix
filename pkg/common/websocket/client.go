package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second
	
	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second
	
	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10
	
	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

// Client represents a WebSocket client connection
type Client struct {
	id   string
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	
	// Subscription preferences
	subscriptions map[EventType]bool
	filters       Filters
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		id:            uuid.New().String(),
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[EventType]bool),
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		// Process incoming message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}
		
		c.handleMessage(msg)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			
			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}
			
			if err := w.Close(); err != nil {
				return
			}
			
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from the client
func (c *Client) handleMessage(msg Message) {
	switch msg.Type {
	case "subscribe":
		var sub Subscription
		if err := json.Unmarshal(msg.Payload, &sub); err != nil {
			log.Printf("Error unmarshaling subscription: %v", err)
			return
		}
		c.updateSubscriptions(sub)
		
	case "unsubscribe":
		var events []EventType
		if err := json.Unmarshal(msg.Payload, &events); err != nil {
			log.Printf("Error unmarshaling unsubscribe: %v", err)
			return
		}
		c.removeSubscriptions(events)
		
	case "ping":
		// Send pong response
		response := Event{
			Type:      EventType("pong"),
			Timestamp: time.Now(),
			Data:      map[string]string{"client_id": c.id},
		}
		if data, err := json.Marshal(response); err == nil {
			select {
			case c.send <- data:
			default:
			}
		}
		
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// updateSubscriptions updates the client's event subscriptions
func (c *Client) updateSubscriptions(sub Subscription) {
	for _, event := range sub.Events {
		c.subscriptions[event] = true
	}
	c.filters = sub.Filters
	
	log.Printf("Client %s updated subscriptions: %v", c.id, sub.Events)
}

// removeSubscriptions removes event subscriptions
func (c *Client) removeSubscriptions(events []EventType) {
	for _, event := range events {
		delete(c.subscriptions, event)
	}
	
	log.Printf("Client %s removed subscriptions: %v", c.id, events)
}

// isSubscribed checks if the client is subscribed to an event type
func (c *Client) isSubscribed(eventType EventType) bool {
	// Check if subscribed to all events (empty subscriptions means all)
	if len(c.subscriptions) == 0 {
		return true
	}
	
	return c.subscriptions[eventType]
}

// matchesFilters checks if an event matches the client's filters
func (c *Client) matchesFilters(event Event) bool {
	// TODO: Implement filter matching based on event data
	// For now, return true to send all subscribed events
	return true
}