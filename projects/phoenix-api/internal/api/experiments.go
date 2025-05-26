package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/phoenix/platform/projects/phoenix-api/internal/analyzer"
	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/websocket"
	"github.com/rs/zerolog/log"
)

// POST /api/v1/experiments - Create a new experiment
func (s *Server) handleCreateExperiment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string                   `json:"name"`
		Description string                   `json:"description"`
		Config      models.ExperimentConfig  `json:"config"`
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
		Metadata:    map[string]interface{}{},
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
	
	w.WriteHeader(http.StatusAccepted)
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

// POST /api/v1/experiments/{id}/kpis - Calculate KPIs for an experiment
func (s *Server) handleCalculateKPIs(w http.ResponseWriter, r *http.Request) {
	expID := chi.URLParam(r, "id")
	
	var req struct {
		Duration string `json:"duration"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Duration = "30m" // Default duration
	}
	
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid duration format")
		return
	}
	
	// Calculate KPIs
	kpiCalculator, err := analyzer.NewKPICalculator(s.config.PrometheusURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create KPI calculator")
		respondError(w, http.StatusInternalServerError, "Failed to initialize KPI calculator")
		return
	}
	
	result, err := kpiCalculator.CalculateExperimentKPIs(r.Context(), expID, duration)
	if err != nil {
		log.Error().Err(err).Str("experiment_id", expID).Msg("Failed to calculate KPIs")
		respondError(w, http.StatusInternalServerError, "Failed to calculate KPIs")
		return
	}
	
	// Store KPI results
	exp, _ := s.store.GetExperiment(r.Context(), expID)
	if exp != nil {
		exp.Status.KPIs = map[string]float64{
			"cardinality_reduction": result.CardinalityReduction,
			"cost_reduction":        result.CostReduction,
			"cpu_reduction":         result.CPUUsage.Reduction,
			"memory_reduction":      result.MemoryUsage.Reduction,
			"data_accuracy":         result.DataAccuracy,
		}
		s.store.UpdateExperiment(r.Context(), exp)
	}
	
	respondJSON(w, http.StatusOK, result)
}

// GET /api/v1/experiments/{id}/kpis - Get cached KPIs
func (s *Server) handleGetKPIs(w http.ResponseWriter, r *http.Request) {
	expID := chi.URLParam(r, "id")
	
	exp, err := s.store.GetExperiment(r.Context(), expID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Experiment not found")
		return
	}
	
	if exp.Status.KPIs == nil || len(exp.Status.KPIs) == 0 {
		respondError(w, http.StatusNotFound, "KPIs not yet calculated")
		return
	}
	
	respondJSON(w, http.StatusOK, exp.Status.KPIs)
}

// POST /api/v1/experiments/{id}/analyze - Run analysis on experiment
func (s *Server) handleAnalyzeExperiment(w http.ResponseWriter, r *http.Request) {
	expID := chi.URLParam(r, "id")
	
	// Get experiment
	exp, err := s.store.GetExperiment(r.Context(), expID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Experiment not found")
		return
	}
	
	// TODO: Implement experiment analysis
	// For now, return a placeholder analysis
	analysis := map[string]interface{}{
		"experiment_id": exp.ID,
		"status": "analysis_pending",
		"message": "Analysis functionality is being implemented",
	}
	
	respondJSON(w, http.StatusOK, analysis)
}