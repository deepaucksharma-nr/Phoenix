package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/websocket"
	"github.com/rs/zerolog/log"
)

// POST /api/v1/experiments - Create a new experiment
func (s *Server) handleCreateExperiment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string                  `json:"name"`
		Description string                  `json:"description"`
		Config      models.ExperimentConfig `json:"config"`
		Namespace   string                  `json:"namespace"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	if len(req.Config.TargetHosts) == 0 {
		respondError(w, http.StatusBadRequest, "At least one target host is required")
		return
	}

	// Deployment mode will be managed at the pipeline level

	// Create experiment
	exp := &models.Experiment{
		Name:        req.Name,
		Description: req.Description,
		Phase:       "created",
		Config:      req.Config,
		Status:      models.ExperimentStatus{},
		Metadata: map[string]interface{}{
			"namespace": req.Namespace,
		},
	}

	if req.Namespace == "" {
		req.Namespace = "default" // Use default namespace if not specified
		exp.Metadata["namespace"] = req.Namespace
	}

	if err := s.store.CreateExperiment(r.Context(), exp); err != nil {
		log.Error().Err(err).Msg("Failed to create experiment")
		respondError(w, http.StatusInternalServerError, "Failed to create experiment")
		return
	}

	// Broadcast creation event
	expData, _ := json.Marshal(exp)
	s.hub.Broadcast <- &websocket.Message{
		Type: "experiment_created",
		Data: json.RawMessage(expData),
	}

	respondJSON(w, http.StatusCreated, exp)
}

// GET /api/v1/experiments - List experiments
func (s *Server) handleListExperiments(w http.ResponseWriter, r *http.Request) {
	experiments, err := s.store.ListExperiments(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to list experiments")
		respondError(w, http.StatusInternalServerError, "Failed to list experiments")
		return
	}

	respondJSON(w, http.StatusOK, experiments)
}

// GET /api/v1/experiments/{id} - Get experiment details
func (s *Server) handleGetExperiment(w http.ResponseWriter, r *http.Request) {
	expID := chi.URLParam(r, "id")

	exp, err := s.store.GetExperiment(r.Context(), expID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Experiment not found")
		return
	}

	respondJSON(w, http.StatusOK, exp)
}

// PUT /api/v1/experiments/{id}/phase - Update experiment phase
func (s *Server) handleUpdateExperimentPhase(w http.ResponseWriter, r *http.Request) {
	expID := chi.URLParam(r, "id")

	var req struct {
		Phase string `json:"phase"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := s.store.UpdateExperimentPhase(r.Context(), expID, req.Phase); err != nil {
		log.Error().Err(err).Msg("Failed to update experiment phase")
		respondError(w, http.StatusInternalServerError, "Failed to update experiment phase")
		return
	}

	// Broadcast phase update
	phaseData, _ := json.Marshal(map[string]string{
		"experiment_id": expID,
		"phase":         req.Phase,
	})
	s.hub.Broadcast <- &websocket.Message{
		Type: "experiment_phase_updated",
		Data: json.RawMessage(phaseData),
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /api/v1/experiments/{id}/start - Start an experiment
func (s *Server) handleStartExperiment(w http.ResponseWriter, r *http.Request) {
	expID := chi.URLParam(r, "id")

	exp, err := s.store.GetExperiment(r.Context(), expID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Experiment not found")
		return
	}

	// Start experiment using agent architecture
	if err := s.expController.StartExperiment(r.Context(), exp); err != nil {
		log.Error().Err(err).Str("experiment_id", expID).Msg("Failed to start experiment")
		respondError(w, http.StatusInternalServerError, "Failed to start experiment")
		return
	}
	
	// Broadcast experiment started event
	startData, _ := json.Marshal(map[string]interface{}{
		"experiment_id": expID,
		"name":          exp.Name,
		"phase":         exp.Phase,
		"config":        exp.Config,
	})
	s.hub.Broadcast <- &websocket.Message{
		Type: "experiment_started",
		Data: startData,
	}

	w.WriteHeader(http.StatusAccepted)
}

// POST /api/v1/experiments/{id}/stop - Stop an experiment
func (s *Server) handleStopExperiment(w http.ResponseWriter, r *http.Request) {
	expID := chi.URLParam(r, "id")

	if err := s.expController.StopExperiment(r.Context(), expID); err != nil {
		log.Error().Err(err).Str("experiment_id", expID).Msg("Failed to stop experiment")
		respondError(w, http.StatusInternalServerError, "Failed to stop experiment")
		return
	}
	
	// Broadcast experiment stopped event
	stopData, _ := json.Marshal(map[string]interface{}{
		"experiment_id": expID,
		"reason":        "user_requested",
	})
	s.hub.Broadcast <- &websocket.Message{
		Type: "experiment_stopped",
		Data: stopData,
	}

	w.WriteHeader(http.StatusAccepted)
}

// GET /api/v1/experiments/{id}/metrics - Get experiment metrics
func (s *Server) handleGetExperimentMetrics(w http.ResponseWriter, r *http.Request) {
	expID := chi.URLParam(r, "id")
	
	// Get experiment to check if it exists
	_, err := s.store.GetExperiment(r.Context(), expID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Experiment not found")
		return
	}
	
	// Get metrics from store
	metrics, err := s.store.GetExperimentMetrics(r.Context(), expID)
	if err != nil {
		log.Error().Err(err).Str("experiment_id", expID).Msg("Failed to get experiment metrics")
		respondError(w, http.StatusInternalServerError, "Failed to get metrics")
		return
	}
	
	respondJSON(w, http.StatusOK, metrics)
}

// GET /api/v1/experiments/{id}/metrics - Get experiment metrics (old implementation)
func (s *Server) handleGetExperimentMetrics_old(w http.ResponseWriter, r *http.Request) {
	expID := chi.URLParam(r, "id")
	
	// Get experiment to check if it exists
	exp, err := s.store.GetExperiment(r.Context(), expID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Experiment not found")
		return
	}
	
	// Build metrics response structure that matches CLI expectations
	metrics := map[string]interface{}{
		"experiment_id": expID,
		"timestamp":     time.Now(),
		"summary": map[string]interface{}{
			"total_metrics":          0,
			"metrics_per_second":     0,
			"cardinality_reduction":  0,
			"cpu_usage":              0,
			"memory_usage":           0,
		},
		"baseline": map[string]interface{}{
			"cardinality":     []interface{}{},
			"cpu_usage":       []interface{}{},
			"memory_usage":    []interface{}{},
			"network_traffic": []interface{}{},
		},
		"candidate": map[string]interface{}{
			"cardinality":     []interface{}{},
			"cpu_usage":       []interface{}{},
			"memory_usage":    []interface{}{},
			"network_traffic": []interface{}{},
		},
	}
	
	// If experiment has KPIs in status, add them to summary
	if exp.Status.KPIs != nil && len(exp.Status.KPIs) > 0 {
		if summary, ok := metrics["summary"].(map[string]interface{}); ok {
			// Extract KPI values if they exist
			if cardReduction, ok := exp.Status.KPIs["cardinality_reduction"]; ok {
				summary["cardinality_reduction"] = cardReduction
			}
			if cpuUsage, ok := exp.Status.KPIs["cpu_usage"]; ok {
				summary["cpu_usage"] = cpuUsage
			}
			if memUsage, ok := exp.Status.KPIs["memory_usage"]; ok {
				summary["memory_usage"] = memUsage
			}
			if totalMetrics, ok := exp.Status.KPIs["total_metrics"]; ok {
				summary["total_metrics"] = totalMetrics
			}
			if metricsPerSec, ok := exp.Status.KPIs["metrics_per_second"]; ok {
				summary["metrics_per_second"] = metricsPerSec
			}
		}
	}
	
	// TODO: In the future, integrate with MetricsCollector service to get time-series data
	// For now, return empty time series arrays which the CLI can handle
	
	respondJSON(w, http.StatusOK, metrics)
}

// POST /api/v1/experiments/{id}/promote - Promote experiment to production
func (s *Server) handlePromoteExperiment(w http.ResponseWriter, r *http.Request) {
	expID := chi.URLParam(r, "id")

	if err := s.expController.PromoteExperiment(r.Context(), expID); err != nil {
		log.Error().Err(err).Str("experiment_id", expID).Msg("Failed to promote experiment")
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

