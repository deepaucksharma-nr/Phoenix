package api

import (
	"net/http"

	"go.uber.org/zap"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	logger *zap.Logger
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(logger *zap.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		logger: logger,
	}
}

// ServeHTTP handles WebSocket upgrade and connections
func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement WebSocket handling
	h.logger.Info("WebSocket connection attempt")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("WebSocket not implemented"))
}