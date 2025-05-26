package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/services"
	"github.com/phoenix/platform/projects/phoenix-api/internal/websocket"
	"github.com/rs/zerolog/log"
)

// handleCalculateKPIs triggers KPI calculation for an experiment
func (s *Server) handleCalculateKPIs(w http.ResponseWriter, r *http.Request) {
	experimentID := chi.URLParam(r, "id")
	
	// Duration parameter is available but currently unused by AnalyzeExperiment
	// _ = r.URL.Query().Get("duration")
	
	// Start metrics collection if not already started
	if err := s.metricsCollector.StartCollection(r.Context(), experimentID); err != nil {
		log.Debug().Err(err).Str("experiment_id", experimentID).Msg("Metrics collection already started or failed")
	}
	
	// Analyze experiment (duration parameter is currently unused in AnalyzeExperiment)
	kpis, err := s.analysisService.AnalyzeExperiment(r.Context(), experimentID)
	if err != nil {
		log.Error().Err(err).Str("experiment_id", experimentID).Msg("Failed to analyze experiment")
		http.Error(w, "Failed to analyze experiment", http.StatusInternalServerError)
		return
	}
	
	// Send WebSocket update
	data, _ := json.Marshal(map[string]interface{}{
		"experiment_id": experimentID,
		"kpis":          kpis,
		"timestamp":     time.Now(),
	})
	s.hub.Broadcast <- &websocket.Message{
		Type: "kpis_calculated",
		Data: data,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kpis)
}

// handleGetKPIs returns the latest KPIs for an experiment
func (s *Server) handleGetKPIs(w http.ResponseWriter, r *http.Request) {
	experimentID := chi.URLParam(r, "id")
	
	// Duration parameter is available but currently unused by AnalyzeExperiment
	// _ = r.URL.Query().Get("duration")
	
	// Calculate fresh KPIs
	kpis, err := s.analysisService.AnalyzeExperiment(r.Context(), experimentID)
	if err != nil {
		log.Error().Err(err).Str("experiment_id", experimentID).Msg("Failed to analyze experiment")
		http.Error(w, "Failed to analyze experiment", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kpis)
}

// handleAnalyzeExperiment performs comprehensive analysis of an experiment
func (s *Server) handleAnalyzeExperiment(w http.ResponseWriter, r *http.Request) {
	experimentID := chi.URLParam(r, "id")
	
	// Perform analysis
	analysis, err := s.analysisService.AnalyzeExperiment(r.Context(), experimentID)
	if err != nil {
		log.Error().Err(err).Str("experiment_id", experimentID).Msg("Failed to analyze experiment")
		http.Error(w, "Failed to analyze experiment", http.StatusInternalServerError)
		return
	}
	
	// Send WebSocket update
	data, _ := json.Marshal(map[string]interface{}{
		"experiment_id": experimentID,
		"analysis":      analysis,
		"timestamp":     time.Now(),
	})
	s.hub.Broadcast <- &websocket.Message{
		Type: "experiment_analyzed",
		Data: data,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}

// handleGetMetrics returns metrics for an experiment
func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	experimentID := chi.URLParam(r, "id")
	
	// Get query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	// Check for time range parameters
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	
	var metrics []*models.Metric
	var err error
	
	if startStr != "" && endStr != "" {
		// Parse time range
		start, err1 := time.Parse(time.RFC3339, startStr)
		end, err2 := time.Parse(time.RFC3339, endStr)
		if err1 != nil || err2 != nil {
			http.Error(w, "Invalid time format. Use RFC3339", http.StatusBadRequest)
			return
		}
		
		metrics, err = s.metricsCollector.GetMetricsInRange(r.Context(), experimentID, start, end)
	} else {
		// Get latest metrics
		metrics, err = s.metricsCollector.GetLatestMetrics(r.Context(), experimentID, limit)
	}
	
	if err != nil {
		log.Error().Err(err).Str("experiment_id", experimentID).Msg("Failed to get metrics")
		http.Error(w, "Failed to get metrics", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"experiment_id": experimentID,
		"metrics":       metrics,
		"count":         len(metrics),
	})
}

// handleGetPipelineTemplates returns available pipeline templates
func (s *Server) handleGetPipelineTemplates(w http.ResponseWriter, r *http.Request) {
	templates := s.templateRenderer.GetBuiltinTemplates()
	
	response := make([]map[string]string, 0, len(templates))
	for name := range templates {
		response = append(response, map[string]string{
			"name":        name,
			"description": getPipelineTemplateDescription(name),
		})
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": response,
	})
}

// handleGeneratePipeline generates an optimized pipeline configuration
func (s *Server) handleGeneratePipeline(w http.ResponseWriter, r *http.Request) {
	experimentID := chi.URLParam(r, "id")
	
	// Get experiment
	experiment, err := s.store.GetExperiment(r.Context(), experimentID)
	if err != nil {
		http.Error(w, "Experiment not found", http.StatusNotFound)
		return
	}
	
	// Get latest KPIs
	kpis, err := s.analysisService.AnalyzeExperiment(r.Context(), experimentID)
	if err != nil {
		log.Error().Err(err).Str("experiment_id", experimentID).Msg("Failed to analyze experiment")
		kpis = nil
	}
	
	// Generate optimized pipeline
	pipelineConfig, err := s.templateRenderer.GenerateOptimizedPipeline(r.Context(), experiment, kpis)
	if err != nil {
		log.Error().Err(err).Str("experiment_id", experimentID).Msg("Failed to generate pipeline")
		http.Error(w, "Failed to generate pipeline", http.StatusInternalServerError)
		return
	}
	
	// Validate the pipeline
	if err := s.templateRenderer.ValidatePipelineConfig(pipelineConfig); err != nil {
		log.Error().Err(err).Msg("Generated pipeline is invalid")
		http.Error(w, "Generated pipeline is invalid", http.StatusInternalServerError)
		return
	}
	
	// Convert to YAML
	yamlConfig, err := s.templateRenderer.RenderPipelineYAML(pipelineConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render pipeline YAML")
		http.Error(w, "Failed to render pipeline", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"experiment_id": experimentID,
		"config":        pipelineConfig,
		"yaml":          yamlConfig,
	})
}

// handleRenderPipelineTemplate renders a specific pipeline template
func (s *Server) handleRenderPipelineTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Template string                 `json:"template"`
		Data     map[string]interface{} `json:"data"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Create template data
	data := services.TemplateData{
		ExperimentID: req.Data["experiment_id"].(string),
		Variant:      "candidate",
		HostID:       req.Data["host_id"].(string),
		Config:       req.Data,
	}
	
	// Render template
	rendered, err := s.templateRenderer.RenderTemplate(r.Context(), req.Template, data)
	if err != nil {
		log.Error().Err(err).Str("template", req.Template).Msg("Failed to render template")
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"template": req.Template,
		"rendered": rendered,
	})
}

// Helper function to get template descriptions
func getPipelineTemplateDescription(name string) string {
	descriptions := map[string]string{
		"baseline": "Basic pipeline with no optimization",
		"topk":     "Keeps only top K metrics by value",
		"adaptive": "Dynamically filters metrics based on usage patterns",
		"hybrid":   "Combines multiple optimization strategies",
	}
	
	if desc, ok := descriptions[name]; ok {
		return desc
	}
	return "Custom pipeline template"
}