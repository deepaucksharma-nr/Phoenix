package websocket

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
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

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		ID:           uuid.New().String(),
		conn:         conn,
		send:         make(chan []byte, 256),
		hub:          hub,
		topics:       make(map[string]bool),
		lastActivity: time.Now(),
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.lastActivity = time.Now()
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Error("WebSocket read error", 
					zap.String("clientID", c.ID),
					zap.Error(err))
			}
			break
		}

		c.lastActivity = time.Now()

		// Parse the message
		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			c.sendError("Invalid message format")
			continue
		}

		// Handle different message types
		switch msg.Type {
		case MessageTypeSubscribe:
			var sub struct {
				Topic string `json:"topic"`
			}
			if err := json.Unmarshal(msg.Data, &sub); err == nil && sub.Topic != "" {
				c.hub.Subscribe(c.ID, sub.Topic)
				c.sendSubscriptionConfirmation(sub.Topic, true)
			}

		case MessageTypeUnsubscribe:
			var unsub struct {
				Topic string `json:"topic"`
			}
			if err := json.Unmarshal(msg.Data, &unsub); err == nil && unsub.Topic != "" {
				c.hub.Unsubscribe(c.ID, unsub.Topic)
				c.sendSubscriptionConfirmation(unsub.Topic, false)
			}

		case MessageTypeHeartbeat:
			// Client heartbeat received, update last activity
			c.sendHeartbeatResponse()

		default:
			c.hub.logger.Warn("Unknown message type", 
				zap.String("clientID", c.ID),
				zap.String("type", string(msg.Type)))
		}
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

			// Add queued messages to the current WebSocket frame
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

// sendError sends an error message to the client
func (c *Client) sendError(message string) {
	errMsg := &Message{
		Type: MessageTypeError,
		Data: json.RawMessage(`{"error": "` + message + `"}`),
		Timestamp: time.Now(),
	}

	if data, err := json.Marshal(errMsg); err == nil {
		select {
		case c.send <- data:
		default:
			// Send buffer full
		}
	}
}

// sendSubscriptionConfirmation sends a subscription confirmation
func (c *Client) sendSubscriptionConfirmation(topic string, subscribed bool) {
	action := "subscribed"
	if !subscribed {
		action = "unsubscribed"
	}

	msg := &Message{
		Type:  MessageTypeNotification,
		Topic: topic,
		Data:  json.RawMessage(`{"` + action + `": true, "topic": "` + topic + `"}`),
		Timestamp: time.Now(),
	}

	if data, err := json.Marshal(msg); err == nil {
		select {
		case c.send <- data:
		default:
			// Send buffer full
		}
	}
}

// sendHeartbeatResponse sends a heartbeat response
func (c *Client) sendHeartbeatResponse() {
	msg := &Message{
		Type:      MessageTypeHeartbeat,
		Data:      json.RawMessage(`{"status": "alive", "clientID": "` + c.ID + `"}`),
		Timestamp: time.Now(),
	}

	if data, err := json.Marshal(msg); err == nil {
		select {
		case c.send <- data:
		default:
			// Send buffer full
		}
	}
}