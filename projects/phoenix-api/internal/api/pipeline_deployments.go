package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	
	"github.com/go-chi/chi/v5"
	"github.com/phoenix/platform/pkg/common/models"
	internalModels "github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/services"
	"github.com/phoenix/platform/projects/phoenix-api/internal/websocket"
	"github.com/rs/zerolog/log"
)

// POST /api/v1/deployments - Create a pipeline deployment
func (s *Server) handleCreateDeployment(w http.ResponseWriter, r *http.Request) {
	var req models.CreateDeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Validate request
	if req.DeploymentName == "" {
		respondError(w, http.StatusBadRequest, "Deployment name is required")
		return
	}
	if req.PipelineName == "" {
		respondError(w, http.StatusBadRequest, "Pipeline name is required")
		return
	}
	if len(req.TargetNodes) == 0 {
		respondError(w, http.StatusBadRequest, "At least one target node is required")
		return
	}
	
	// Create deployment
	deployment := &models.PipelineDeployment{
		ID:             fmt.Sprintf("dep-%d", time.Now().UnixNano()),
		DeploymentName: req.DeploymentName,
		PipelineName:   req.PipelineName,
		Namespace:      req.Namespace,
		TargetNodes:    req.TargetNodes,
		Parameters:     req.Parameters,
		Resources:      req.Resources,
		Status:         "pending",
		Phase:          "creating",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      r.Header.Get("X-User-ID"),
	}
	
	if deployment.Namespace == "" {
		deployment.Namespace = "default"
	}
	if deployment.Parameters == nil {
		deployment.Parameters = make(map[string]interface{})
	}
	if deployment.Resources == nil {
		deployment.Resources = &models.ResourceRequirements{}
	}
	
	// Store deployment
	if err := s.store.CreateDeployment(r.Context(), deployment); err != nil {
		log.Error().Err(err).Msg("Failed to create deployment")
		respondError(w, http.StatusInternalServerError, "Failed to create deployment")
		return
	}
	
	// Create deployment tasks for each target node
	for nodeName, nodeSelector := range deployment.TargetNodes {
		// Render pipeline configuration for this deployment
		templateData := services.TemplateData{
			ExperimentID: "", // No experiment ID for direct deployments
			Variant:      deployment.Variant,
			HostID:       nodeSelector,
			Config:       deployment.Parameters,
		}
		
		// Default variant if not specified
		if templateData.Variant == "" {
			templateData.Variant = "candidate"
		}
		
		// Use the specified pipeline template or default to baseline
		templateName := deployment.PipelineName
		if templateName == "" {
			templateName = "baseline"
		}
		
		// Render the pipeline configuration
		pipelineConfig, err := s.templateRenderer.RenderTemplate(r.Context(), templateName, templateData)
		if err != nil {
			log.Error().Err(err).
				Str("deployment_id", deployment.ID).
				Str("template", templateName).
				Msg("Failed to render pipeline template")
			// Fall back to raw config if template rendering fails
			pipelineConfig = ""
		}
		
		task := &internalModels.Task{
			HostID:       nodeSelector,
			Type:         "deployment",
			Action:       "deploy",
			Priority:     1,
			Config: map[string]interface{}{
				"deployment_id":     deployment.ID,
				"deployment_name":   deployment.DeploymentName,
				"pipeline_name":     deployment.PipelineName,
				"node_name":         nodeName,
				"parameters":        deployment.Parameters,
				"resources":         deployment.Resources,
				"pipeline_config":   pipelineConfig,
				"rendered_template": templateName,
			},
		}
		
		if err := s.taskQueue.Enqueue(r.Context(), task); err != nil {
			log.Error().Err(err).
				Str("deployment_id", deployment.ID).
				Str("node", nodeName).
				Msg("Failed to enqueue deployment task")
		}
	}
	
	// Broadcast deployment created event
	data, _ := json.Marshal(deployment)
	s.hub.Broadcast <- &websocket.Message{
		Type: "deployment_created",
		Data: data,
	}
	
	respondJSON(w, http.StatusCreated, deployment)
}

// GET /api/v1/deployments - List pipeline deployments
func (s *Server) handleListDeployments(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	req := &models.ListDeploymentsRequest{
		Namespace:    r.URL.Query().Get("namespace"),
		PipelineName: r.URL.Query().Get("pipeline"),
		Status:       r.URL.Query().Get("status"),
	}
	
	// Parse pagination
	limit := 20
	
	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			limit = ps
			req.PageSize = ps
		}
	}
	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			req.Page = p
		}
	}
	
	// Get deployments
	deployments, total, err := s.store.ListDeployments(r.Context(), req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list deployments")
		respondError(w, http.StatusInternalServerError, "Failed to list deployments")
		return
	}
	
	// Return paginated response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"deployments": deployments,
		"total":       total,
		"page":        req.Page,
		"page_size":   limit,
	})
}

// GET /api/v1/deployments/{id} - Get deployment details
func (s *Server) handleGetDeployment(w http.ResponseWriter, r *http.Request) {
	deploymentID := chi.URLParam(r, "id")
	
	deployment, err := s.store.GetDeployment(r.Context(), deploymentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Deployment not found")
			return
		}
		log.Error().Err(err).Msg("Failed to get deployment")
		respondError(w, http.StatusInternalServerError, "Failed to get deployment")
		return
	}
	
	respondJSON(w, http.StatusOK, deployment)
}

// PUT /api/v1/deployments/{id} - Update deployment
func (s *Server) handleUpdateDeployment(w http.ResponseWriter, r *http.Request) {
	deploymentID := chi.URLParam(r, "id")
	
	var req models.UpdateDeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Update deployment
	if err := s.store.UpdateDeployment(r.Context(), deploymentID, &req); err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Deployment not found")
			return
		}
		log.Error().Err(err).Msg("Failed to update deployment")
		respondError(w, http.StatusInternalServerError, "Failed to update deployment")
		return
	}
	
	// Broadcast update event
	data, _ := json.Marshal(map[string]interface{}{
		"deployment_id": deploymentID,
		"update":        req,
	})
	s.hub.Broadcast <- &websocket.Message{
		Type: "deployment_updated",
		Data: data,
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/v1/deployments/{id} - Delete deployment
func (s *Server) handleDeleteDeployment(w http.ResponseWriter, r *http.Request) {
	deploymentID := chi.URLParam(r, "id")
	
	// Get deployment first to find target nodes
	deployment, err := s.store.GetDeployment(r.Context(), deploymentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Deployment not found")
			return
		}
		log.Error().Err(err).Msg("Failed to get deployment")
		respondError(w, http.StatusInternalServerError, "Failed to delete deployment")
		return
	}
	
	// Create undeploy tasks for each target node
	for nodeName, nodeSelector := range deployment.TargetNodes {
		task := &internalModels.Task{
			HostID:   nodeSelector,
			Type:     "deployment",
			Action:   "undeploy",
			Priority: 2, // Higher priority for cleanup
			Config: map[string]interface{}{
				"deployment_id":   deployment.ID,
				"deployment_name": deployment.DeploymentName,
				"node_name":       nodeName,
			},
		}
		
		if err := s.taskQueue.Enqueue(r.Context(), task); err != nil {
			log.Error().Err(err).
				Str("deployment_id", deployment.ID).
				Str("node", nodeName).
				Msg("Failed to enqueue undeploy task")
		}
	}
	
	// Mark deployment as deleting
	updateReq := &models.UpdateDeploymentRequest{
		Status: "deleting",
		Phase:  "terminating",
	}
	
	if err := s.store.UpdateDeployment(r.Context(), deploymentID, updateReq); err != nil {
		log.Error().Err(err).Msg("Failed to update deployment status")
		respondError(w, http.StatusInternalServerError, "Failed to delete deployment")
		return
	}
	
	// Broadcast delete event
	data, _ := json.Marshal(map[string]string{
		"deployment_id": deploymentID,
		"status":        "deleting",
	})
	s.hub.Broadcast <- &websocket.Message{
		Type: "deployment_deleted",
		Data: data,
	}
	
	w.WriteHeader(http.StatusAccepted)
}

// POST /api/v1/deployments/{id}/rollback - Rollback deployment
func (s *Server) handleRollbackDeployment(w http.ResponseWriter, r *http.Request) {
	deploymentID := chi.URLParam(r, "id")
	
	var req struct {
		Version int `json:"version"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to previous version
		req.Version = -1
	}
	
	// Get deployment
	deployment, err := s.store.GetDeployment(r.Context(), deploymentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Deployment not found")
			return
		}
		log.Error().Err(err).Msg("Failed to get deployment")
		respondError(w, http.StatusInternalServerError, "Failed to rollback deployment")
		return
	}
	
	// Get previous version (if versioning is implemented)
	// For now, we'll just create a rollback task
	for nodeName, nodeSelector := range deployment.TargetNodes {
		task := &internalModels.Task{
			HostID:   nodeSelector,
			Type:     "deployment",
			Action:   "rollback",
			Priority: 2,
			Config: map[string]interface{}{
				"deployment_id":   deployment.ID,
				"deployment_name": deployment.DeploymentName,
				"node_name":       nodeName,
				"target_version":  req.Version,
			},
		}
		
		if err := s.taskQueue.Enqueue(r.Context(), task); err != nil {
			log.Error().Err(err).
				Str("deployment_id", deployment.ID).
				Str("node", nodeName).
				Msg("Failed to enqueue rollback task")
		}
	}
	
	// Update deployment status
	updateReq := &models.UpdateDeploymentRequest{
		Status: "rolling_back",
		Phase:  "updating",
	}
	
	if err := s.store.UpdateDeployment(r.Context(), deploymentID, updateReq); err != nil {
		log.Error().Err(err).Msg("Failed to update deployment status")
	}
	
	// Broadcast rollback event
	data, _ := json.Marshal(map[string]interface{}{
		"deployment_id": deploymentID,
		"action":        "rollback",
		"version":       req.Version,
	})
	s.hub.Broadcast <- &websocket.Message{
		Type: "deployment_rollback",
		Data: data,
	}
	
	w.WriteHeader(http.StatusAccepted)
}

// GET /api/v1/deployments/{id}/status - Get deployment status
func (s *Server) handleGetDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	deploymentID := chi.URLParam(r, "id")
	
	deployment, err := s.store.GetDeployment(r.Context(), deploymentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Deployment not found")
			return
		}
		log.Error().Err(err).Msg("Failed to get deployment")
		respondError(w, http.StatusInternalServerError, "Failed to get deployment status")
		return
	}
	
	// Get deployment tasks to determine actual status
	tasks, err := s.store.ListTasks(r.Context(), map[string]interface{}{
		"type": "deployment",
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to get deployment tasks")
	}
	
	// Count task statuses for this deployment
	var pending, running, completed, failed int
	for _, task := range tasks {
		if depID, ok := task.Config["deployment_id"].(string); ok && depID == deploymentID {
			switch task.Status {
			case "pending":
				pending++
			case "running", "assigned":
				running++
			case "completed":
				completed++
			case "failed":
				failed++
			}
		}
	}
	
	total := pending + running + completed + failed
	
	// Determine overall status
	status := deployment.Status
	if failed > 0 {
		status = "failed"
	} else if completed == total && total > 0 {
		status = "ready"
	} else if running > 0 {
		status = "deploying"
	}
	
	// Build status response
	statusResp := map[string]interface{}{
		"deployment_id": deployment.ID,
		"status":        status,
		"phase":         deployment.Phase,
		"nodes": map[string]interface{}{
			"total":     len(deployment.TargetNodes),
			"ready":     completed,
			"deploying": running,
			"pending":   pending,
			"failed":    failed,
		},
		"created_at": deployment.CreatedAt,
		"updated_at": deployment.UpdatedAt,
	}
	
	// Add metrics if available
	if deployment.Metrics != nil {
		statusResp["metrics"] = deployment.Metrics
	}
	
	respondJSON(w, http.StatusOK, statusResp)
}