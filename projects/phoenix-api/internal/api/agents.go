package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/websocket"
	"github.com/rs/zerolog/log"
)

// GET /api/v1/agent/tasks - Long polling endpoint for agents to get tasks
func (s *Server) handleAgentGetTasks(w http.ResponseWriter, r *http.Request) {
	hostID := r.Context().Value("hostID").(string)
	
	// Long polling with 30s timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	// Get pending tasks for this host
	tasks, err := s.taskQueue.GetPendingTasks(ctx, hostID)
	if err != nil {
		log.Error().Err(err).Str("host", hostID).Msg("Failed to get tasks")
		respondError(w, http.StatusInternalServerError, "Failed to get tasks")
		return
	}
	
	// Mark tasks as assigned
	for _, task := range tasks {
		if err := s.taskQueue.UpdateTaskStatus(ctx, task.ID, "assigned"); err != nil {
			log.Error().Err(err).Str("task", task.ID).Msg("Failed to update task status")
		}
	}
	
	respondJSON(w, http.StatusOK, tasks)
}

// POST /api/v1/agent/tasks/{taskId}/status - Update task status
func (s *Server) handleTaskStatusUpdate(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskId")
	hostID := r.Context().Value("hostID").(string)
	
	var update struct {
		Status       string                 `json:"status"`
		Result       map[string]interface{} `json:"result,omitempty"`
		ErrorMessage string                 `json:"error_message,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Validate task belongs to this host
	task, err := s.taskQueue.GetTask(r.Context(), taskID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Task not found")
		return
	}
	
	if task.HostID != hostID {
		respondError(w, http.StatusForbidden, "Task does not belong to this host")
		return
	}
	
	// Update task status
	if err := s.taskQueue.UpdateTaskStatusWithResult(r.Context(), taskID, update.Status, update.Result, update.ErrorMessage); err != nil {
		log.Error().Err(err).Str("task", taskID).Msg("Failed to update task status")
		respondError(w, http.StatusInternalServerError, "Failed to update task status")
		return
	}
	
	// Broadcast update via WebSocket
	data, _ := json.Marshal(map[string]interface{}{
		"task_id": taskID,
		"host_id": hostID,
		"status":  update.Status,
	})
	s.hub.Broadcast <- &websocket.Message{
		Type: "task_update",
		Data: data,
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/v1/agent/heartbeat - Agent heartbeat
func (s *Server) handleAgentHeartbeat(w http.ResponseWriter, r *http.Request) {
	hostID := r.Context().Value("hostID").(string)
	
	var heartbeat models.AgentHeartbeat
	if err := json.NewDecoder(r.Body).Decode(&heartbeat); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	heartbeat.HostID = hostID
	heartbeat.LastHeartbeat = time.Now()
	
	// Update agent status
	if err := s.store.UpdateAgentHeartbeat(r.Context(), &heartbeat); err != nil {
		log.Error().Err(err).Str("host", hostID).Msg("Failed to update agent status")
		respondError(w, http.StatusInternalServerError, "Failed to update agent status")
		return
	}
	
	// Broadcast status update
	data, _ := json.Marshal(heartbeat)
	s.hub.Broadcast <- &websocket.Message{
		Type: "agent_heartbeat",
		Data: data,
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/v1/agent/metrics - Push metrics from agent
func (s *Server) handleAgentMetrics(w http.ResponseWriter, r *http.Request) {
	hostID := r.Context().Value("hostID").(string)
	
	var metrics struct {
		Timestamp time.Time                `json:"timestamp"`
		Metrics   []map[string]interface{} `json:"metrics"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Store metrics in cache for faster queries
	for _, metric := range metrics.Metrics {
		if err := s.store.CacheMetric(r.Context(), hostID, metric); err != nil {
			log.Error().Err(err).Str("host", hostID).Msg("Failed to cache metric")
		}
	}
	
	// Also forward to Pushgateway if configured
	if s.config.Features.UsePushgateway {
		// TODO: Implement Pushgateway client
		log.Debug().Str("host", hostID).Int("count", len(metrics.Metrics)).Msg("Would forward metrics to Pushgateway")
	}
	
	w.WriteHeader(http.StatusAccepted)
}

// POST /api/v1/agent/logs - Stream logs from agent
func (s *Server) handleAgentLogs(w http.ResponseWriter, r *http.Request) {
	hostID := r.Context().Value("hostID").(string)
	
	var logs struct {
		TaskID string `json:"task_id"`
		Logs   []struct {
			Timestamp time.Time `json:"timestamp"`
			Level     string    `json:"level"`
			Message   string    `json:"message"`
		} `json:"logs"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&logs); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Broadcast logs via WebSocket for real-time monitoring
	data, _ := json.Marshal(map[string]interface{}{
		"host_id": hostID,
		"task_id": logs.TaskID,
		"logs":    logs.Logs,
	})
	s.hub.Broadcast <- &websocket.Message{
		Type: "agent_logs",
		Data: data,
	}
	
	w.WriteHeader(http.StatusAccepted)
}