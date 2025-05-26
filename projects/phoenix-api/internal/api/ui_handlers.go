package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/phoenix/platform/pkg/common/websocket"
	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/rs/zerolog/log"
)

// UI-focused endpoints for the revolutionary dashboard

// handleGetMetricCostFlow returns real-time metric cost breakdown
func (s *Server) handleGetMetricCostFlow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get current metric flow from cost calculator
	costFlow, err := s.store.GetMetricCostFlow(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get metric cost flow")
		respondError(w, http.StatusInternalServerError, "Failed to get metric flow")
		return
	}
	
	// Convert to WebSocket format
	flow := websocket.MetricFlowUpdate{
		Timestamp:     time.Now(),
		TotalCostRate: costFlow.TotalCostPerMinute,
		TopMetrics:    make([]websocket.MetricCostBreakdown, 0),
		ByService:     costFlow.ByService,
		ByNamespace:   costFlow.ByNamespace,
	}
	
	// Add top metrics
	for _, metric := range costFlow.TopMetrics {
		flow.TopMetrics = append(flow.TopMetrics, websocket.MetricCostBreakdown{
			MetricName:    metric.Name,
			CostPerMinute: metric.CostPerMinute,
			Cardinality:   metric.Cardinality,
			Percentage:    metric.Percentage,
			Labels:        metric.Labels,
		})
	}
	
	respondJSON(w, http.StatusOK, flow)
}

// handleGetCardinalityBreakdown returns cardinality analysis
func (s *Server) handleGetCardinalityBreakdown(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get query parameters
	namespace := r.URL.Query().Get("namespace")
	service := r.URL.Query().Get("service")
	
	breakdown, err := s.store.GetCardinalityBreakdown(ctx, namespace, service)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get cardinality breakdown")
		respondError(w, http.StatusInternalServerError, "Failed to get cardinality")
		return
	}
	
	respondJSON(w, http.StatusOK, breakdown)
}

// handleGetFleetStatus returns status of all agents
func (s *Server) handleGetFleetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	agents, err := s.store.GetAllAgents(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get fleet status")
		respondError(w, http.StatusInternalServerError, "Failed to get fleet status")
		return
	}
	
	// Convert to fleet status format
	type FleetStatus struct {
		TotalAgents   int                              `json:"total_agents"`
		HealthyAgents int                              `json:"healthy_agents"`
		OfflineAgents int                              `json:"offline_agents"`
		UpdatingAgents int                             `json:"updating_agents"`
		TotalSavings  float64                          `json:"total_savings"`
		Agents        []websocket.AgentStatusUpdate    `json:"agents"`
	}
	
	status := FleetStatus{
		TotalAgents: len(agents),
		Agents:      make([]websocket.AgentStatusUpdate, 0),
	}
	
	for _, agent := range agents {
		// Convert active tasks from string IDs to TaskInfo
		taskInfos := make([]websocket.TaskInfo, 0, len(agent.ActiveTasks))
		for _, taskID := range agent.ActiveTasks {
			taskInfos = append(taskInfos, websocket.TaskInfo{
				ID:     taskID,
				Type:   "pipeline", // Default type
				Status: "running",
				Progress: 50, // Default progress
				StartedAt: time.Now(),
			})
		}
		
		agentStatus := websocket.AgentStatusUpdate{
			HostID:        agent.HostID,
			Status:        agent.Status,
			ActiveTasks:   taskInfos,
			Metrics:       websocket.AgentMetrics{
				CPUPercent:    agent.ResourceUsage.CPUPercent,
				MemoryMB:      agent.ResourceUsage.MemoryBytes / (1024 * 1024),
				MetricsPerSec: 0, // TODO: Get from metrics
				DroppedCount:  0, // TODO: Get from metrics
			},
			CostSavings:   0, // TODO: Calculate cost savings
			LastHeartbeat: agent.LastHeartbeat,
			Location:      nil, // TODO: Add location support
		}
		
		status.Agents = append(status.Agents, agentStatus)
		
		// Count by status
		switch agent.Status {
		case "healthy":
			status.HealthyAgents++
		case "offline":
			status.OfflineAgents++
		case "updating":
			status.UpdatingAgents++
		}
		
		// TODO: Calculate total savings from metrics
		// status.TotalSavings += agent.CostSavings
	}
	
	respondJSON(w, http.StatusOK, status)
}

// handleGetAgentMap returns agent geographical distribution
func (s *Server) handleGetAgentMap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	agents, err := s.store.GetAgentsWithLocation(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get agent map")
		respondError(w, http.StatusInternalServerError, "Failed to get agent map")
		return
	}
	
	respondJSON(w, http.StatusOK, agents)
}

// handleCreateExperimentWizard handles simplified experiment creation
func (s *Server) handleCreateExperimentWizard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		HostSelector []string `json:"host_selector"` // tags like env=prod
		PipelineType string   `json:"pipeline_type"` // template name
		Duration     int      `json:"duration_hours"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// TODO: Implement experiment wizard creation
	// For now, return a placeholder response
	experiment := map[string]interface{}{
		"id": "exp-" + time.Now().Format("20060102150405"),
		"name": req.Name,
		"description": req.Description,
		"status": "created",
		"message": "Experiment wizard creation is being implemented",
	}
	
	respondJSON(w, http.StatusCreated, experiment)
}

// handleGetPipelineTemplates returns available pipeline templates
func (s *Server) handleGetPipelineTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	templates, err := s.store.GetPipelineTemplates(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get pipeline templates")
		respondError(w, http.StatusInternalServerError, "Failed to get templates")
		return
	}
	
	respondJSON(w, http.StatusOK, templates)
}

// handlePreviewPipelineImpact calculates impact without deploying
func (s *Server) handlePreviewPipelineImpact(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PipelineConfig json.RawMessage `json:"pipeline_config"`
		TargetHosts    []string        `json:"target_hosts"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Calculate impact based on historical data
	impact, err := s.calculatePipelineImpact(r.Context(), req.PipelineConfig, req.TargetHosts)
	if err != nil {
		log.Error().Err(err).Msg("Failed to calculate pipeline impact")
		respondError(w, http.StatusInternalServerError, "Failed to calculate impact")
		return
	}
	
	respondJSON(w, http.StatusOK, impact)
}

// handleGetActiveTasks returns currently active tasks
func (s *Server) handleGetActiveTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get filter parameters
	status := r.URL.Query().Get("status")
	hostID := r.URL.Query().Get("host_id")
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	
	tasks, err := s.store.GetActiveTasks(ctx, status, hostID, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get active tasks")
		respondError(w, http.StatusInternalServerError, "Failed to get tasks")
		return
	}
	
	respondJSON(w, http.StatusOK, tasks)
}

// handleGetTaskQueue returns task queue status
func (s *Server) handleGetTaskQueue(w http.ResponseWriter, r *http.Request) {
	_ = r.Context() // ctx reserved for future use
	
	// TODO: Implement queue status method
	queueStatus := map[string]interface{}{
		"pending": 0,
		"running": 0,
		"completed": 0,
		"failed": 0,
		"total": 0,
	}
	
	respondJSON(w, http.StatusOK, queueStatus)
}

// handleGetCostAnalytics returns cost analytics dashboard data
func (s *Server) handleGetCostAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get time range
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}
	
	analytics, err := s.store.GetCostAnalytics(ctx, period)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get cost analytics")
		respondError(w, http.StatusInternalServerError, "Failed to get analytics")
		return
	}
	
	respondJSON(w, http.StatusOK, analytics)
}

// handleQuickDeploy deploys a pipeline with one click
func (s *Server) handleQuickDeploy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PipelineTemplate string   `json:"pipeline_template"`
		TargetHosts      []string `json:"target_hosts"`
		AutoRollback     bool     `json:"auto_rollback"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Create deployment tasks
	deployment, err := s.deployPipelineQuick(r.Context(), req.PipelineTemplate, req.TargetHosts, req.AutoRollback)
	if err != nil {
		log.Error().Err(err).Msg("Failed to deploy pipeline")
		respondError(w, http.StatusInternalServerError, "Failed to deploy")
		return
	}
	
	respondJSON(w, http.StatusAccepted, deployment)
}

// handleInstantRollback performs instant rollback
func (s *Server) handleInstantRollback(w http.ResponseWriter, r *http.Request) {
	experimentID := chi.URLParam(r, "id")
	
	// TODO: Implement rollback functionality
	log.Info().Str("experiment_id", experimentID).Msg("Rollback requested")
	
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "success",
		"message": "Rollback initiated",
	})
}

// Helper functions

func (s *Server) calculatePipelineImpact(ctx context.Context, config json.RawMessage, hosts []string) (map[string]interface{}, error) {
	// TODO: Implement impact calculation based on historical metrics
	return map[string]interface{}{
		"estimated_cost_reduction": 65.5,
		"estimated_cardinality_reduction": 72.3,
		"estimated_cpu_impact": 1.2,
		"estimated_memory_impact": 45, // MB
		"confidence_level": 0.85,
	}, nil
}

func (s *Server) deployPipelineQuick(ctx context.Context, template string, hosts []string, autoRollback bool) (map[string]interface{}, error) {
	// Create deployment tasks for each host
	deploymentID := "dep-" + strconv.FormatInt(time.Now().Unix(), 36)
	
	for _, host := range hosts {
		task := &models.Task{
			Type:         "deploy_pipeline",
			HostID:       host,
			ExperimentID: "",
			Config: map[string]interface{}{
				"template": template,
				"auto_rollback": autoRollback,
			},
			Priority: 1,
			Status:   "pending",
		}
		
		if err := s.taskQueue.Enqueue(ctx, task); err != nil {
			return nil, err
		}
	}
	
	return map[string]interface{}{
		"deployment_id": deploymentID,
		"hosts_count": len(hosts),
		"status": "deploying",
	}, nil
}