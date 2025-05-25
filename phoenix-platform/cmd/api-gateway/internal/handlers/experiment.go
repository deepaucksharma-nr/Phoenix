package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	
	"github.com/phoenix/platform/pkg/clients/experiment"
	experimentv1 "github.com/phoenix/platform/api/proto/v1/experiment"
)

// ExperimentHandler handles HTTP requests for experiments
type ExperimentHandler struct {
	logger *zap.Logger
	client *experiment.Client
}

// NewExperimentHandler creates a new experiment handler
func NewExperimentHandler(logger *zap.Logger, client *experiment.Client) *ExperimentHandler {
	return &ExperimentHandler{
		logger: logger,
		client: client,
	}
}

// HandleExperiments handles /api/v1/experiments
func (h *ExperimentHandler) HandleExperiments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listExperiments(w, r)
	case http.MethodPost:
		h.createExperiment(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleExperimentByID handles /api/v1/experiments/{id}
func (h *ExperimentHandler) HandleExperimentByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/experiments/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Experiment ID required", http.StatusBadRequest)
		return
	}
	id := parts[0]

	switch r.Method {
	case http.MethodGet:
		h.getExperiment(w, r, id)
	case http.MethodPut:
		h.updateExperiment(w, r, id)
	case http.MethodDelete:
		h.deleteExperiment(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ExperimentHandler) listExperiments(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	h.logger.Debug("listing experiments")

	// Parse query parameters
	states := parseStates(r.URL.Query().Get("states"))
	labels := parseLabels(r.URL.Query().Get("labels"))

	experiments, _, err := h.client.ListExperiments(ctx, states, labels, 50, "")
	if err != nil {
		h.logger.Error("failed to list experiments", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to list experiments: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to REST response
	response := map[string]interface{}{
		"experiments": experiments,
		"count":       len(experiments),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *ExperimentHandler) createExperiment(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var req struct {
		Name                 string            `json:"name"`
		Description          string            `json:"description"`
		BaselinePipelineID   string            `json:"baseline_pipeline_id"`
		CandidatePipelineID  string            `json:"candidate_pipeline_id"`
		TrafficPercentage    int32             `json:"traffic_percentage"`
		TargetServices       []string          `json:"target_services"`
		Labels               map[string]string `json:"labels"`
		ValidateOnly         bool              `json:"validate_only"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("creating experiment", zap.String("name", req.Name))

	// Convert to proto
	experiment := &experimentv1.Experiment{
		Name:                 req.Name,
		Description:          req.Description,
		BaselinePipelineId:   req.BaselinePipelineID,
		CandidatePipelineId:  req.CandidatePipelineID,
		TrafficPercentage:    req.TrafficPercentage,
		TargetServices:       req.TargetServices,
		Labels:               req.Labels,
	}

	created, err := h.client.CreateExperiment(ctx, experiment, req.ValidateOnly)
	if err != nil {
		h.logger.Error("failed to create experiment", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to create experiment: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(created); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *ExperimentHandler) getExperiment(w http.ResponseWriter, r *http.Request, id string) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	h.logger.Debug("getting experiment", zap.String("id", id))

	experiment, err := h.client.GetExperiment(ctx, id)
	if err != nil {
		h.logger.Error("failed to get experiment", zap.String("id", id), zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to get experiment: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(experiment); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *ExperimentHandler) updateExperiment(w http.ResponseWriter, r *http.Request, id string) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var req struct {
		State  string `json:"state"`
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("updating experiment", zap.String("id", id), zap.String("state", req.State))

	// Convert state string to proto enum
	state := stringToExperimentState(req.State)
	if state == experimentv1.ExperimentState_EXPERIMENT_STATE_UNSPECIFIED {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	updated, err := h.client.UpdateExperimentState(ctx, id, state, req.Reason)
	if err != nil {
		h.logger.Error("failed to update experiment", zap.String("id", id), zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update experiment: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updated); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *ExperimentHandler) deleteExperiment(w http.ResponseWriter, r *http.Request, id string) {
	h.logger.Info("delete experiment requested", zap.String("id", id))
	
	// For now, return not implemented since the gRPC service doesn't support it
	http.Error(w, "Delete experiment not yet implemented", http.StatusNotImplemented)
}

// Helper functions

func parseStates(statesStr string) []experimentv1.ExperimentState {
	if statesStr == "" {
		return nil
	}

	states := strings.Split(statesStr, ",")
	var protoStates []experimentv1.ExperimentState
	
	for _, state := range states {
		protoState := stringToExperimentState(strings.TrimSpace(state))
		if protoState != experimentv1.ExperimentState_EXPERIMENT_STATE_UNSPECIFIED {
			protoStates = append(protoStates, protoState)
		}
	}

	return protoStates
}

func parseLabels(labelsStr string) map[string]string {
	if labelsStr == "" {
		return nil
	}

	labels := make(map[string]string)
	pairs := strings.Split(labelsStr, ",")
	
	for _, pair := range pairs {
		kv := strings.Split(strings.TrimSpace(pair), "=")
		if len(kv) == 2 {
			labels[kv[0]] = kv[1]
		}
	}

	return labels
}

func stringToExperimentState(state string) experimentv1.ExperimentState {
	switch strings.ToLower(state) {
	case "pending":
		return experimentv1.ExperimentState_EXPERIMENT_STATE_PENDING
	case "initializing":
		return experimentv1.ExperimentState_EXPERIMENT_STATE_INITIALIZING
	case "running":
		return experimentv1.ExperimentState_EXPERIMENT_STATE_RUNNING
	case "pausing":
		return experimentv1.ExperimentState_EXPERIMENT_STATE_PAUSING
	case "paused":
		return experimentv1.ExperimentState_EXPERIMENT_STATE_PAUSED
	case "resuming":
		return experimentv1.ExperimentState_EXPERIMENT_STATE_RESUMING
	case "completing":
		return experimentv1.ExperimentState_EXPERIMENT_STATE_COMPLETING
	case "completed":
		return experimentv1.ExperimentState_EXPERIMENT_STATE_COMPLETED
	case "failed":
		return experimentv1.ExperimentState_EXPERIMENT_STATE_FAILED
	case "cancelled":
		return experimentv1.ExperimentState_EXPERIMENT_STATE_CANCELLED
	default:
		return experimentv1.ExperimentState_EXPERIMENT_STATE_UNSPECIFIED
	}
}