package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	internalModels "github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/store"
	phoenixws "github.com/phoenix/platform/projects/phoenix-api/internal/websocket"
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
	
	// Return the cost flow directly
	respondJSON(w, http.StatusOK, costFlow)
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
		TotalAgents    int                            `json:"total_agents"`
		HealthyAgents  int                            `json:"healthy_agents"`
		OfflineAgents  int                            `json:"offline_agents"`
		UpdatingAgents int                            `json:"updating_agents"`
		TotalSavings   float64                        `json:"total_savings"`
		Agents         []map[string]interface{}       `json:"agents"`
	}
	
	status := FleetStatus{
		TotalAgents: len(agents),
		Agents:      make([]map[string]interface{}, 0),
	}
	
	for _, agent := range agents {
		agentData := map[string]interface{}{
			"host_id":        agent.HostID,
			"hostname":       agent.Hostname,
			"status":         agent.Status,
			"active_tasks":   agent.ActiveTasks,
			"cpu_percent":    agent.ResourceUsage.CPUPercent,
			"memory_mb":      agent.ResourceUsage.MemoryBytes / (1024 * 1024),
			"last_heartbeat": agent.LastHeartbeat,
			"agent_version":  agent.AgentVersion,
		}
		
		status.Agents = append(status.Agents, agentData)
		
		// Count by status
		switch agent.Status {
		case "healthy":
			status.HealthyAgents++
		case "offline":
			status.OfflineAgents++
		case "updating":
			status.UpdatingAgents++
		}
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
		Name               string            `json:"name"`
		Description        string            `json:"description"`
		TargetHosts        []string          `json:"target_hosts"`
		BaselineTemplate   string            `json:"baseline_template"`
		CandidateTemplate  string            `json:"candidate_template"`
		TemplateVariables  map[string]string `json:"template_variables"`
		Duration           int               `json:"duration_minutes"`
		WarmupDuration     int               `json:"warmup_duration_minutes"`
		OptimizationGoal   string            `json:"optimization_goal"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Create experiment using the wizard data
	experiment := &internalModels.Experiment{
		ID:          fmt.Sprintf("exp-%s", time.Now().Format("20060102150405")),
		Name:        req.Name,
		Description: req.Description,
		Phase:       internalModels.PhasePending,
		Config: internalModels.ExperimentConfig{
			TargetHosts: req.TargetHosts,
			BaselineTemplate: internalModels.PipelineTemplate{
				Name:      req.BaselineTemplate,
				ConfigURL: fmt.Sprintf("file:///configs/%s.yaml", req.BaselineTemplate),
			},
			CandidateTemplate: internalModels.PipelineTemplate{
				Name:      req.CandidateTemplate,
				ConfigURL: fmt.Sprintf("file:///configs/%s.yaml", req.CandidateTemplate),
				Variables: req.TemplateVariables,
			},
			Duration:       time.Duration(req.Duration) * time.Minute,
			WarmupDuration: time.Duration(req.WarmupDuration) * time.Minute,
		},
		Metadata: map[string]interface{}{
			"wizard_version": "1.0",
			"optimization_goal": req.OptimizationGoal,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Save experiment
	if err := s.store.CreateExperiment(r.Context(), experiment); err != nil {
		log.Error().Err(err).Msg("Failed to create experiment from wizard")
		respondError(w, http.StatusInternalServerError, "Failed to create experiment")
		return
	}
	
	// Create initial event
	event := &internalModels.ExperimentEvent{
		ExperimentID: experiment.ID,
		EventType:    "experiment_created",
		Phase:        "created",
		Message:      "Experiment created via wizard",
		Metadata: map[string]interface{}{
			"wizard_data": req,
		},
	}
	
	if err := s.store.CreateExperimentEvent(r.Context(), event); err != nil {
		log.Error().Err(err).Msg("Failed to create experiment event")
	}
	
	// Broadcast creation event
	s.broadcastExperimentUpdate(experiment.ID, "created", map[string]interface{}{
		"experiment": experiment,
	})
	
	respondJSON(w, http.StatusCreated, experiment)
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
	ctx := r.Context()
	
	queueStatus, err := s.store.GetTaskQueueStatus(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get task queue status")
		respondError(w, http.StatusInternalServerError, "Failed to get queue status")
		return
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
	ctx := r.Context()
	experimentID := chi.URLParam(r, "id")
	
	// Get experiment
	exp, err := s.store.GetExperiment(ctx, experimentID)
	if err != nil {
		log.Error().Err(err).Str("experiment_id", experimentID).Msg("Failed to get experiment")
		respondError(w, http.StatusNotFound, "Experiment not found")
		return
	}
	
	// Check if experiment is in a state that can be rolled back
	if exp.Phase != "running" && exp.Phase != "completed" {
		respondError(w, http.StatusBadRequest, "Experiment must be running or completed to rollback")
		return
	}
	
	// Create rollback tasks for each host
	rollbackTasks := 0
	for _, host := range exp.Config.TargetHosts {
		// Stop candidate pipeline
		task := &internalModels.Task{
			HostID:       host,
			ExperimentID: experimentID,
			Type:         "collector",
			Action:       "stop",
			Priority:     3, // High priority for rollback
			Config: map[string]interface{}{
				"id": fmt.Sprintf("%s-candidate", experimentID),
			},
		}
		
		if err := s.taskQueue.Enqueue(ctx, task); err != nil {
			log.Error().Err(err).Str("host", host).Msg("Failed to enqueue rollback task")
			continue
		}
		rollbackTasks++
	}
	
	// Update experiment status
	if err := s.store.UpdateExperimentPhase(ctx, experimentID, "rollback"); err != nil {
		log.Error().Err(err).Msg("Failed to update experiment phase")
	}
	
	// Create experiment event
	event := &internalModels.ExperimentEvent{
		ExperimentID: experimentID,
		EventType:    "experiment_rollback",
		Phase:        "rollback",
		Message:      fmt.Sprintf("Rollback initiated for %d hosts", rollbackTasks),
		Metadata: map[string]interface{}{
			"hosts_affected": rollbackTasks,
			"reason":         r.URL.Query().Get("reason"),
		},
	}
	
	if err := s.store.CreateExperimentEvent(ctx, event); err != nil {
		log.Error().Err(err).Msg("Failed to create rollback event")
	}
	
	// Broadcast rollback event
	data, _ := json.Marshal(map[string]interface{}{
		"experiment_id": experimentID,
		"action":        "rollback",
		"hosts":         rollbackTasks,
	})
	s.hub.Broadcast <- &phoenixws.Message{
		Type:      phoenixws.MessageType("experiment_rollback"),
		Topic:     "experiments",
		Data:      data,
		Timestamp: time.Now(),
	}
	
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":         "success",
		"message":        "Rollback initiated",
		"experiment_id":  experimentID,
		"hosts_affected": rollbackTasks,
	})
}

// handleGetPipelineTemplates returns available pipeline templates
func (s *Server) handleGetPipelineTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get filter parameters
	category := r.URL.Query().Get("category")
	tag := r.URL.Query().Get("tag")

	// Get templates from database
	templates, err := s.store.GetPipelineTemplates(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get pipeline templates")
		respondError(w, http.StatusInternalServerError, "Failed to fetch pipeline templates")
		return
	}

	// Apply filters
	filtered := templates
	if category != "" || tag != "" {
		filtered = make([]*store.PipelineTemplate, 0)
		for _, template := range templates {
		if category != "" && template.Metadata["category"] != category {
			continue
		}
		if tag != "" {
			hasTag := false
			for _, t := range template.Tags {
				if t == tag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}
		filtered = append(filtered, template)
		}
	}

	_ = ctx // ctx reserved for future use
	
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"templates": filtered,
		"total":     len(filtered),
		"filters": map[string]string{
			"category": category,
			"tag":      tag,
		},
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
		task := &internalModels.Task{
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

// Helper function to broadcast experiment updates via WebSocket
func (s *Server) broadcastExperimentUpdate(experimentID, action string, data map[string]interface{}) {
	msgData, _ := json.Marshal(map[string]interface{}{
		"experiment_id": experimentID,
		"action":        action,
		"data":          data,
		"timestamp":     time.Now(),
	})
	
	s.hub.Broadcast <- &phoenixws.Message{
		Type:      phoenixws.MessageTypeExperimentUpdate,
		Topic:     "experiments",
		Data:      msgData,
		Timestamp: time.Now(),
	}
}