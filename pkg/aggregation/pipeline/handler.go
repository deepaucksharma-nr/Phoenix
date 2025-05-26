package pipeline

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// StatusHandler handles HTTP requests for pipeline status
type StatusHandler struct {
	aggregator *StatusAggregator
	logger     *zap.Logger
}

// NewStatusHandler creates a new status handler
func NewStatusHandler(aggregator *StatusAggregator, logger *zap.Logger) *StatusHandler {
	return &StatusHandler{
		aggregator: aggregator,
		logger:     logger,
	}
}

// GetStatus handles GET requests for pipeline status
func (h *StatusHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	deploymentID := chi.URLParam(r, "deploymentID")
	if deploymentID == "" {
		h.respondError(w, "deployment ID is required", http.StatusBadRequest)
		return
	}

	// Get aggregated status
	status, err := h.aggregator.GetStatus(r.Context(), deploymentID)
	if err != nil {
		h.logger.Error("failed to get status", zap.String("deployment_id", deploymentID), zap.Error(err))
		h.respondError(w, "failed to retrieve status", http.StatusInternalServerError)
		return
	}

	// Return as JSON
	h.respondJSON(w, status, http.StatusOK)
}

// GetStatusSummary returns a summary of all deployments
func (h *StatusHandler) GetStatusSummary(w http.ResponseWriter, r *http.Request) {
	// This would query all deployments and provide a summary
	// For now, return a simple response
	summary := map[string]interface{}{
		"total_deployments": 0,
		"healthy":          0,
		"degraded":         0,
		"unhealthy":        0,
	}

	h.respondJSON(w, summary, http.StatusOK)
}

// respondJSON sends a JSON response
func (h *StatusHandler) respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// respondError sends an error response
func (h *StatusHandler) respondError(w http.ResponseWriter, message string, status int) {
	h.respondJSON(w, map[string]string{
		"error": message,
	}, status)
}

// RegisterRoutes registers the status handler routes
func (h *StatusHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/pipelines/deployments/{deploymentID}/status", h.GetStatus)
	r.Get("/api/v1/pipelines/status/summary", h.GetStatusSummary)
}