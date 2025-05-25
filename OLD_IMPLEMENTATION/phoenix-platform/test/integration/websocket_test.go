package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/phoenix/platform/pkg/api"
	"github.com/phoenix/platform/pkg/interfaces"
)

type WebSocketTestSuite struct {
	suite.Suite
	server      *httptest.Server
	wsHandler   *api.WebSocketHandler
	eventBus    interfaces.EventBus
	logger      *zap.Logger
	wsURL       string
}

func (suite *WebSocketTestSuite) SetupSuite() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	suite.logger = logger

	// Create event bus
	suite.eventBus = interfaces.NewInMemoryEventBus()

	// Create WebSocket handler
	suite.wsHandler = api.NewWebSocketHandler(suite.logger, suite.eventBus)

	// Create test server
	suite.server = httptest.NewServer(suite.wsHandler)
	suite.wsURL = "ws" + strings.TrimPrefix(suite.server.URL, "http")
}

func (suite *WebSocketTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// Test Cases

func (suite *WebSocketTestSuite) TestWebSocketConnection() {
	// Connect to WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(suite.wsURL, nil)
	require.NoError(suite.T(), err)
	defer ws.Close()

	// Send ping
	pingMsg := api.WebSocketMessage{
		Type:      "ping",
		Timestamp: time.Now(),
	}
	err = ws.WriteJSON(pingMsg)
	require.NoError(suite.T(), err)

	// Read pong response
	var pongMsg api.WebSocketMessage
	err = ws.ReadJSON(&pongMsg)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "pong", pongMsg.Type)
}

func (suite *WebSocketTestSuite) TestSubscribeToEvents() {
	// Connect to WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(suite.wsURL, nil)
	require.NoError(suite.T(), err)
	defer ws.Close()

	// Subscribe to experiment events
	subscribeMsg := api.WebSocketMessage{
		Type:      "subscribe",
		Payload:   json.RawMessage(`{"event": "experiment.created"}`),
		Timestamp: time.Now(),
	}
	err = ws.WriteJSON(subscribeMsg)
	require.NoError(suite.T(), err)

	// Publish event through event bus
	go func() {
		time.Sleep(100 * time.Millisecond)
		suite.eventBus.Publish(interfaces.Event{
			Type: "experiment.created",
			Data: map[string]interface{}{
				"id":   "exp-123",
				"name": "Test Experiment",
			},
			Timestamp: time.Now(),
		})
	}()

	// Read event message
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	var eventMsg api.WebSocketMessage
	err = ws.ReadJSON(&eventMsg)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "experiment.update", eventMsg.Type)
	assert.Equal(suite.T(), "experiment.created", eventMsg.Event)

	// Verify payload
	var payload map[string]interface{}
	err = json.Unmarshal(eventMsg.Payload, &payload)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "exp-123", payload["id"])
}

func (suite *WebSocketTestSuite) TestUnsubscribeFromEvents() {
	// Connect to WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(suite.wsURL, nil)
	require.NoError(suite.T(), err)
	defer ws.Close()

	// Subscribe first
	subscribeMsg := api.WebSocketMessage{
		Type:      "subscribe",
		Payload:   json.RawMessage(`{"event": "experiment.deleted"}`),
		Timestamp: time.Now(),
	}
	err = ws.WriteJSON(subscribeMsg)
	require.NoError(suite.T(), err)

	// Unsubscribe
	unsubscribeMsg := api.WebSocketMessage{
		Type:      "unsubscribe",
		Payload:   json.RawMessage(`{"event": "experiment.deleted"}`),
		Timestamp: time.Now(),
	}
	err = ws.WriteJSON(unsubscribeMsg)
	require.NoError(suite.T(), err)

	// Publish event - should not receive it
	suite.eventBus.Publish(interfaces.Event{
		Type:      "experiment.deleted",
		Data:      map[string]interface{}{"id": "exp-123"},
		Timestamp: time.Now(),
	})

	// Should timeout reading (no message received)
	ws.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	var msg api.WebSocketMessage
	err = ws.ReadJSON(&msg)
	assert.Error(suite.T(), err)
}

func (suite *WebSocketTestSuite) TestBroadcastSystemEvents() {
	// Connect multiple clients
	ws1, _, err := websocket.DefaultDialer.Dial(suite.wsURL, nil)
	require.NoError(suite.T(), err)
	defer ws1.Close()

	ws2, _, err := websocket.DefaultDialer.Dial(suite.wsURL, nil)
	require.NoError(suite.T(), err)
	defer ws2.Close()

	// Publish system alert
	go func() {
		time.Sleep(100 * time.Millisecond)
		suite.eventBus.Publish(interfaces.Event{
			Type: "system.alert",
			Data: map[string]interface{}{
				"level":   "warning",
				"title":   "High CPU Usage",
				"message": "CPU usage exceeded 80%",
			},
			Timestamp: time.Now(),
		})
	}()

	// Both clients should receive the alert
	for i, ws := range []*websocket.Conn{ws1, ws2} {
		ws.SetReadDeadline(time.Now().Add(2 * time.Second))
		var msg api.WebSocketMessage
		err = ws.ReadJSON(&msg)
		require.NoError(suite.T(), err, "Client %d failed to receive message", i+1)
		assert.Equal(suite.T(), "alert", msg.Type)
	}
}

func (suite *WebSocketTestSuite) TestMultipleSubscriptions() {
	// Connect to WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(suite.wsURL, nil)
	require.NoError(suite.T(), err)
	defer ws.Close()

	// Subscribe to multiple events
	events := []string{"metrics.cpu", "metrics.memory", "metrics.network"}
	for _, event := range events {
		msg := api.WebSocketMessage{
			Type:      "subscribe",
			Payload:   json.RawMessage(fmt.Sprintf(`{"event": "%s"}`, event)),
			Timestamp: time.Now(),
		}
		err = ws.WriteJSON(msg)
		require.NoError(suite.T(), err)
	}

	// Publish different events
	go func() {
		time.Sleep(100 * time.Millisecond)
		for _, event := range events {
			suite.eventBus.Publish(interfaces.Event{
				Type: event,
				Data: map[string]interface{}{
					"value": 42,
					"unit":  "percent",
				},
				Timestamp: time.Now(),
			})
			time.Sleep(50 * time.Millisecond)
		}
	}()

	// Should receive all events
	receivedEvents := make(map[string]bool)
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	
	for i := 0; i < len(events); i++ {
		var msg api.WebSocketMessage
		err = ws.ReadJSON(&msg)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), "metrics.update", msg.Type)
		receivedEvents[msg.Event] = true
	}

	// Verify all events were received
	for _, event := range events {
		assert.True(suite.T(), receivedEvents[event], "Did not receive event: %s", event)
	}
}

func (suite *WebSocketTestSuite) TestConnectionWithAuth() {
	// Connect with auth header
	header := http.Header{}
	header.Add("X-User-ID", "user-123")
	
	ws, _, err := websocket.DefaultDialer.Dial(suite.wsURL, header)
	require.NoError(suite.T(), err)
	defer ws.Close()

	// Connection should be established
	pingMsg := api.WebSocketMessage{
		Type:      "ping",
		Timestamp: time.Now(),
	}
	err = ws.WriteJSON(pingMsg)
	require.NoError(suite.T(), err)

	var pongMsg api.WebSocketMessage
	err = ws.ReadJSON(&pongMsg)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "pong", pongMsg.Type)
}

func (suite *WebSocketTestSuite) TestInvalidMessageHandling() {
	// Connect to WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(suite.wsURL, nil)
	require.NoError(suite.T(), err)
	defer ws.Close()

	// Send invalid message type
	invalidMsg := api.WebSocketMessage{
		Type:      "invalid-type",
		Timestamp: time.Now(),
	}
	err = ws.WriteJSON(invalidMsg)
	require.NoError(suite.T(), err)

	// Should not crash, connection should remain open
	pingMsg := api.WebSocketMessage{
		Type:      "ping",
		Timestamp: time.Now(),
	}
	err = ws.WriteJSON(pingMsg)
	require.NoError(suite.T(), err)

	var pongMsg api.WebSocketMessage
	err = ws.ReadJSON(&pongMsg)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "pong", pongMsg.Type)
}

func (suite *WebSocketTestSuite) TestReconnection() {
	// Connect to WebSocket
	ws1, _, err := websocket.DefaultDialer.Dial(suite.wsURL, nil)
	require.NoError(suite.T(), err)

	// Close connection
	ws1.Close()

	// Should be able to reconnect
	ws2, _, err := websocket.DefaultDialer.Dial(suite.wsURL, nil)
	require.NoError(suite.T(), err)
	defer ws2.Close()

	// New connection should work
	pingMsg := api.WebSocketMessage{
		Type:      "ping",
		Timestamp: time.Now(),
	}
	err = ws2.WriteJSON(pingMsg)
	require.NoError(suite.T(), err)

	var pongMsg api.WebSocketMessage
	err = ws2.ReadJSON(&pongMsg)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "pong", pongMsg.Type)
}

func (suite *WebSocketTestSuite) TestConcurrentConnections() {
	numClients := 10
	clients := make([]*websocket.Conn, numClients)
	
	// Connect multiple clients concurrently
	errChan := make(chan error, numClients)
	for i := 0; i < numClients; i++ {
		go func(index int) {
			ws, _, err := websocket.DefaultDialer.Dial(suite.wsURL, nil)
			if err != nil {
				errChan <- err
				return
			}
			clients[index] = ws
			errChan <- nil
		}(i)
	}

	// Wait for all connections
	for i := 0; i < numClients; i++ {
		err := <-errChan
		require.NoError(suite.T(), err)
	}

	// Cleanup
	for _, ws := range clients {
		if ws != nil {
			ws.Close()
		}
	}
}

// Run the test suite
func TestWebSocketIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(WebSocketTestSuite))
}