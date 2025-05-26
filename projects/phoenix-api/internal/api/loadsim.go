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

// LoadSimulation represents a load simulation job
type LoadSimulation struct {
	ID           string                 `json:"id"`
	ExperimentID string                 `json:"experiment_id"`
	Profile      string                 `json:"profile"`
	TargetHosts  []string               `json:"target_hosts"`
	Duration     time.Duration          `json:"duration"`
	ProcessCount int                    `json:"process_count"`
	Status       string                 `json:"status"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// POST /api/v1/loadsimulations - Start a load simulation
func (s *Server) handleStartLoadSimulation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExperimentID string   `json:"experiment_id"`
		Profile      string   `json:"profile"`
		TargetHosts  []string `json:"target_hosts,omitempty"`
		Duration     string   `json:"duration"`
		ProcessCount int      `json:"process_count"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Parse duration
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid duration format")
		return
	}
	
	// Get experiment if specified
	var targetHosts []string
	if req.ExperimentID != "" {
		exp, err := s.store.GetExperiment(r.Context(), req.ExperimentID)
		if err != nil {
			respondError(w, http.StatusNotFound, "Experiment not found")
			return
		}
		targetHosts = exp.Config.TargetHosts
	} else if len(req.TargetHosts) > 0 {
		targetHosts = req.TargetHosts
	} else {
		respondError(w, http.StatusBadRequest, "Either experiment_id or target_hosts must be specified")
		return
	}
	
	// Default values
	if req.Profile == "" {
		req.Profile = "realistic"
	}
	if req.ProcessCount == 0 {
		req.ProcessCount = 10
	}
	
	// Create load simulation tasks for each host
	simID := generateID("sim")
	for _, host := range targetHosts {
		task := &models.Task{
			HostID:       host,
			ExperimentID: req.ExperimentID,
			Type:         "loadsim",
			Action:       "start",
			Priority:     0,
			Config: map[string]interface{}{
				"simulation_id": simID,
				"profile":       req.Profile,
				"duration":      duration.String(),
				"process_count": req.ProcessCount,
			},
		}
		
		if err := s.taskQueue.Enqueue(r.Context(), task); err != nil {
			log.Error().Err(err).Str("host", host).Msg("Failed to enqueue load simulation task")
			respondError(w, http.StatusInternalServerError, "Failed to start load simulation")
			return
		}
	}
	
	// Create response
	now := time.Now()
	sim := &LoadSimulation{
		ID:           simID,
		ExperimentID: req.ExperimentID,
		Profile:      req.Profile,
		TargetHosts:  targetHosts,
		Duration:     duration,
		ProcessCount: req.ProcessCount,
		Status:       "starting",
		CreatedAt:    now,
		UpdatedAt:    now,
		Metadata: map[string]interface{}{
			"requested_by": r.Header.Get("X-User-ID"),
		},
	}
	
	// Broadcast start event
	data, _ := json.Marshal(sim)
	s.hub.Broadcast <- &websocket.Message{
		Type: "loadsim_started",
		Data: data,
	}
	
	respondJSON(w, http.StatusCreated, sim)
}

// GET /api/v1/loadsimulations - List load simulations
func (s *Server) handleListLoadSimulations(w http.ResponseWriter, r *http.Request) {
	experimentID := r.URL.Query().Get("experiment_id")
	
	// Get tasks of type loadsim
	filters := map[string]interface{}{
		"type": "loadsim",
	}
	if experimentID != "" {
		filters["experiment_id"] = experimentID
	}
	
	tasks, err := s.store.ListTasks(r.Context(), filters)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list load simulation tasks")
		respondError(w, http.StatusInternalServerError, "Failed to list load simulations")
		return
	}
	
	// Group tasks by simulation ID
	simMap := make(map[string]*LoadSimulation)
	for _, task := range tasks {
		simID, _ := task.Config["simulation_id"].(string)
		if simID == "" {
			continue
		}
		
		if sim, exists := simMap[simID]; exists {
			// Update simulation based on task status
			if task.Status == "running" && sim.StartedAt == nil {
				sim.StartedAt = task.StartedAt
			}
			if task.Status == "completed" || task.Status == "failed" {
				sim.CompletedAt = task.CompletedAt
			}
		} else {
			// Create new simulation entry
			sim := &LoadSimulation{
				ID:           simID,
				ExperimentID: task.ExperimentID,
				Profile:      getStringFromConfig(task.Config, "profile", "unknown"),
				Duration:     parseDurationFromConfig(task.Config, "duration"),
				ProcessCount: getIntFromConfig(task.Config, "process_count", 0),
				Status:       task.Status,
				TargetHosts:  []string{task.HostID},
				CreatedAt:    task.CreatedAt,
				UpdatedAt:    task.UpdatedAt,
			}
			
			if task.Status == "running" {
				sim.StartedAt = task.StartedAt
			}
			if task.Status == "completed" || task.Status == "failed" {
				sim.CompletedAt = task.CompletedAt
			}
			
			simMap[simID] = sim
		}
	}
	
	// Convert map to slice
	simulations := make([]*LoadSimulation, 0, len(simMap))
	for _, sim := range simMap {
		simulations = append(simulations, sim)
	}
	
	respondJSON(w, http.StatusOK, simulations)
}

// GET /api/v1/loadsimulations/{id} - Get load simulation status
func (s *Server) handleGetLoadSimulation(w http.ResponseWriter, r *http.Request) {
	simID := chi.URLParam(r, "id")
	
	// Get all tasks for this simulation
	tasks, err := s.store.ListTasks(r.Context(), map[string]interface{}{
		"type": "loadsim",
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to get load simulation tasks")
		respondError(w, http.StatusInternalServerError, "Failed to get load simulation")
		return
	}
	
	// Find tasks for this simulation
	var sim *LoadSimulation
	var targetHosts []string
	
	for _, task := range tasks {
		taskSimID, _ := task.Config["simulation_id"].(string)
		if taskSimID != simID {
			continue
		}
		
		if sim == nil {
			sim = &LoadSimulation{
				ID:           simID,
				ExperimentID: task.ExperimentID,
				Profile:      getStringFromConfig(task.Config, "profile", "unknown"),
				Duration:     parseDurationFromConfig(task.Config, "duration"),
				ProcessCount: getIntFromConfig(task.Config, "process_count", 0),
				Status:       task.Status,
				CreatedAt:    task.CreatedAt,
				UpdatedAt:    task.UpdatedAt,
			}
		}
		
		targetHosts = append(targetHosts, task.HostID)
		
		// Update status based on task states
		if task.Status == "running" && sim.StartedAt == nil {
			sim.StartedAt = task.StartedAt
		}
		if task.Status == "completed" || task.Status == "failed" {
			sim.CompletedAt = task.CompletedAt
		}
	}
	
	if sim == nil {
		respondError(w, http.StatusNotFound, "Load simulation not found")
		return
	}
	
	sim.TargetHosts = targetHosts
	respondJSON(w, http.StatusOK, sim)
}

// DELETE /api/v1/loadsimulations/{id} - Stop a load simulation
func (s *Server) handleStopLoadSimulation(w http.ResponseWriter, r *http.Request) {
	simID := chi.URLParam(r, "id")
	
	// Get all tasks for this simulation
	tasks, err := s.store.ListTasks(r.Context(), map[string]interface{}{
		"type": "loadsim",
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to get load simulation tasks")
		respondError(w, http.StatusInternalServerError, "Failed to stop load simulation")
		return
	}
	
	// Create stop tasks for each host
	stopped := false
	for _, task := range tasks {
		taskSimID, _ := task.Config["simulation_id"].(string)
		if taskSimID != simID {
			continue
		}
		
		// Only stop if running
		if task.Status != "running" && task.Status != "pending" {
			continue
		}
		
		stopTask := &models.Task{
			HostID:       task.HostID,
			ExperimentID: task.ExperimentID,
			Type:         "loadsim",
			Action:       "stop",
			Priority:     2, // High priority
			Config: map[string]interface{}{
				"simulation_id": simID,
			},
		}
		
		if err := s.taskQueue.Enqueue(r.Context(), stopTask); err != nil {
			log.Error().Err(err).Str("host", task.HostID).Msg("Failed to enqueue stop task")
		} else {
			stopped = true
		}
	}
	
	if !stopped {
		respondError(w, http.StatusNotFound, "Load simulation not found or not running")
		return
	}
	
	// Broadcast stop event
	data, _ := json.Marshal(map[string]string{
		"simulation_id": simID,
		"status":        "stopping",
	})
	s.hub.Broadcast <- &websocket.Message{
		Type: "loadsim_stopped",
		Data: data,
	}
	
	w.WriteHeader(http.StatusAccepted)
}

// Helper functions
func generateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func getStringFromConfig(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}

func getIntFromConfig(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key].(float64); ok {
		return int(val)
	}
	if val, ok := config[key].(int); ok {
		return val
	}
	return defaultValue
}

func parseDurationFromConfig(config map[string]interface{}, key string) time.Duration {
	if val, ok := config[key].(string); ok {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return 0
}