package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	
	"github.com/phoenix/platform/pkg/clients/controller"
	controllerv1 "github.com/phoenix/platform/api/proto/v1/controller"
)

// ControlHandler handles HTTP requests for control operations
type ControlHandler struct {
	logger *zap.Logger
	client *controller.Client
}

// NewControlHandler creates a new control handler
func NewControlHandler(logger *zap.Logger, client *controller.Client) *ControlHandler {
	return &ControlHandler{
		logger: logger,
		client: client,
	}
}

// HandleControlSignals handles /api/v1/control/signals
func (h *ControlHandler) HandleControlSignals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.executeControlSignal(w, r)
	case http.MethodGet:
		h.listControlSignals(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleControlStatus handles /api/v1/control/status/{experiment_id}
func (h *ControlHandler) HandleControlStatus(w http.ResponseWriter, r *http.Request) {
	// Extract experiment ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/control/status/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Experiment ID required", http.StatusBadRequest)
		return
	}
	experimentID := parts[0]

	switch r.Method {
	case http.MethodGet:
		h.getControlStatus(w, r, experimentID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ControlHandler) executeControlSignal(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var req struct {
		ExperimentID string          `json:"experiment_id"`
		Type         string          `json:"type"`
		Action       json.RawMessage `json:"action"`
		Reason       string          `json:"reason"`
		DryRun       bool            `json:"dry_run"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("executing control signal",
		zap.String("experiment_id", req.ExperimentID),
		zap.String("type", req.Type),
		zap.Bool("dry_run", req.DryRun),
	)

	// Create control signal based on type
	var signal *controllerv1.ControlSignal
	var err error

	switch strings.ToLower(req.Type) {
	case "traffic_split":
		signal, err = h.createTrafficSplitSignal(req.ExperimentID, req.Action, req.Reason)
	case "pipeline_state":
		signal, err = h.createPipelineStateSignal(req.ExperimentID, req.Action, req.Reason)
	case "rollback":
		signal, err = h.createRollbackSignal(req.ExperimentID, req.Action, req.Reason)
	case "config_update":
		signal, err = h.createConfigUpdateSignal(req.ExperimentID, req.Action, req.Reason)
	default:
		http.Error(w, "Invalid signal type", http.StatusBadRequest)
		return
	}

	if err != nil {
		h.logger.Error("failed to create control signal", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to create signal: %v", err), http.StatusBadRequest)
		return
	}

	// Execute the signal
	executedSignal, validationErrors, err := h.client.ExecuteControlSignal(ctx, signal, req.DryRun)
	if err != nil {
		h.logger.Error("failed to execute control signal", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to execute signal: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"signal":            executedSignal,
		"validation_errors": validationErrors,
	}

	statusCode := http.StatusOK
	if len(validationErrors) > 0 {
		statusCode = http.StatusBadRequest
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *ControlHandler) listControlSignals(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	experimentID := r.URL.Query().Get("experiment_id")
	if experimentID == "" {
		http.Error(w, "experiment_id query parameter required", http.StatusBadRequest)
		return
	}

	h.logger.Debug("listing control signals", zap.String("experiment_id", experimentID))

	// Create list request
	listReq := &controller.ListSignalsRequest{
		ExperimentID: experimentID,
		PageSize:     50,
	}

	// Parse status filter
	statusStr := r.URL.Query().Get("statuses")
	if statusStr != "" {
		statuses := strings.Split(statusStr, ",")
		for _, status := range statuses {
			if protoStatus := stringToSignalStatus(strings.TrimSpace(status)); protoStatus != controllerv1.SignalStatus_SIGNAL_STATUS_UNSPECIFIED {
				listReq.Statuses = append(listReq.Statuses, protoStatus)
			}
		}
	}

	signals, _, err := h.client.ListControlSignals(ctx, listReq)
	if err != nil {
		h.logger.Error("failed to list control signals", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to list signals: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to REST response
	response := map[string]interface{}{
		"signals": signals,
		"count":   len(signals),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *ControlHandler) getControlStatus(w http.ResponseWriter, r *http.Request, experimentID string) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	h.logger.Debug("getting control status", zap.String("experiment_id", experimentID))

	status, err := h.client.GetControlLoopStatus(ctx, experimentID)
	if err != nil {
		h.logger.Error("failed to get control status", zap.String("experiment_id", experimentID), zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to get control status: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// Helper methods for creating different types of control signals

func (h *ControlHandler) createTrafficSplitSignal(experimentID string, actionData json.RawMessage, reason string) (*controllerv1.ControlSignal, error) {
	var action struct {
		BaselinePipelineID  string `json:"baseline_pipeline_id"`
		CandidatePipelineID string `json:"candidate_pipeline_id"`
		CandidatePercentage int32  `json:"candidate_percentage"`
	}

	if err := json.Unmarshal(actionData, &action); err != nil {
		return nil, fmt.Errorf("invalid traffic split action: %w", err)
	}

	return controller.NewTrafficSplitSignal(experimentID, action.BaselinePipelineID, action.CandidatePipelineID, action.CandidatePercentage), nil
}

func (h *ControlHandler) createPipelineStateSignal(experimentID string, actionData json.RawMessage, reason string) (*controllerv1.ControlSignal, error) {
	var action struct {
		PipelineID string `json:"pipeline_id"`
		State      string `json:"state"`
	}

	if err := json.Unmarshal(actionData, &action); err != nil {
		return nil, fmt.Errorf("invalid pipeline state action: %w", err)
	}

	state := stringToPipelineState(action.State)
	if state == controllerv1.PipelineState_PIPELINE_STATE_UNSPECIFIED {
		return nil, fmt.Errorf("invalid pipeline state: %s", action.State)
	}

	return controller.NewPipelineStateSignal(experimentID, action.PipelineID, state), nil
}

func (h *ControlHandler) createRollbackSignal(experimentID string, actionData json.RawMessage, reason string) (*controllerv1.ControlSignal, error) {
	var action struct {
		TargetPipelineID     string `json:"target_pipeline_id"`
		RollbackToPipelineID string `json:"rollback_to_pipeline_id"`
		Immediate            bool   `json:"immediate"`
	}

	if err := json.Unmarshal(actionData, &action); err != nil {
		return nil, fmt.Errorf("invalid rollback action: %w", err)
	}

	return controller.NewRollbackSignal(experimentID, action.TargetPipelineID, action.RollbackToPipelineID, action.Immediate), nil
}

func (h *ControlHandler) createConfigUpdateSignal(experimentID string, actionData json.RawMessage, reason string) (*controllerv1.ControlSignal, error) {
	var action struct {
		PipelineID string `json:"pipeline_id"`
		Changes    []struct {
			ComponentID string `json:"component_id"`
			Parameter   string `json:"parameter"`
			OldValue    string `json:"old_value"`
			NewValue    string `json:"new_value"`
		} `json:"changes"`
	}

	if err := json.Unmarshal(actionData, &action); err != nil {
		return nil, fmt.Errorf("invalid config update action: %w", err)
	}

	var changes []*controllerv1.ConfigChange
	for _, change := range action.Changes {
		changes = append(changes, &controllerv1.ConfigChange{
			ComponentId: change.ComponentID,
			Parameter:   change.Parameter,
			OldValue:    change.OldValue,
			NewValue:    change.NewValue,
		})
	}

	return controller.NewConfigUpdateSignal(experimentID, action.PipelineID, changes), nil
}

// Helper functions

func stringToSignalStatus(status string) controllerv1.SignalStatus {
	switch strings.ToLower(status) {
	case "pending":
		return controllerv1.SignalStatus_SIGNAL_STATUS_PENDING
	case "executing":
		return controllerv1.SignalStatus_SIGNAL_STATUS_EXECUTING
	case "completed":
		return controllerv1.SignalStatus_SIGNAL_STATUS_COMPLETED
	case "failed":
		return controllerv1.SignalStatus_SIGNAL_STATUS_FAILED
	case "cancelled":
		return controllerv1.SignalStatus_SIGNAL_STATUS_CANCELLED
	default:
		return controllerv1.SignalStatus_SIGNAL_STATUS_UNSPECIFIED
	}
}

func stringToPipelineState(state string) controllerv1.PipelineState {
	switch strings.ToLower(state) {
	case "active":
		return controllerv1.PipelineState_PIPELINE_STATE_ACTIVE
	case "paused":
		return controllerv1.PipelineState_PIPELINE_STATE_PAUSED
	case "stopped":
		return controllerv1.PipelineState_PIPELINE_STATE_STOPPED
	default:
		return controllerv1.PipelineState_PIPELINE_STATE_UNSPECIFIED
	}
}