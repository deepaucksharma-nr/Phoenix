package simulator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/phoenix/platform/pkg/interfaces"
	"go.uber.org/zap"
)

// ControlAPI provides HTTP endpoints for controlling the simulator
type ControlAPI struct {
	logger    *zap.Logger
	simulator interfaces.LoadSimulator
	port      int
}

// NewControlAPI creates a new control API
func NewControlAPI(logger *zap.Logger, simulator interfaces.LoadSimulator, port int) *ControlAPI {
	return &ControlAPI{
		logger:    logger,
		simulator: simulator,
		port:      port,
	}
}

// Start starts the control API server
func (api *ControlAPI) Start(ctx context.Context) error {
	router := chi.NewRouter()
	
	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// Routes
	router.Route("/api/v1", func(r chi.Router) {
		// Simulation management
		r.Post("/simulations", api.handleCreateSimulation)
		r.Get("/simulations", api.handleListSimulations)
		r.Get("/simulations/{id}", api.handleGetSimulation)
		r.Post("/simulations/{id}/start", api.handleStartSimulation)
		r.Post("/simulations/{id}/stop", api.handleStopSimulation)
		r.Get("/simulations/{id}/results", api.handleGetResults)
		
		// Chaos engineering
		r.Post("/chaos/cpu-spike", api.handleCPUSpike)
		r.Post("/chaos/memory-leak", api.handleMemoryLeak)
		r.Post("/chaos/process-kill", api.handleProcessKill)
		
		// Health and info
		r.Get("/health", api.handleHealth)
		r.Get("/info", api.handleInfo)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", api.port),
		Handler: router,
	}

	go func() {
		api.logger.Info("starting control API", zap.Int("port", api.port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			api.logger.Error("control API error", zap.Error(err))
		}
	}()

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			api.logger.Error("control API shutdown error", zap.Error(err))
		}
	}()

	return nil
}

// Handlers

func (api *ControlAPI) handleCreateSimulation(w http.ResponseWriter, r *http.Request) {
	var config interfaces.SimulationConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		api.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	simulation, err := api.simulator.CreateSimulation(r.Context(), &config)
	if err != nil {
		api.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.respondJSON(w, http.StatusCreated, simulation)
}

func (api *ControlAPI) handleListSimulations(w http.ResponseWriter, r *http.Request) {
	filter := &interfaces.SimulationFilter{}
	
	// Parse query parameters
	if status := r.URL.Query().Get("status"); status != "" {
		s := interfaces.SimulationStatus(status)
		filter.Status = &s
	}

	simulations, err := api.simulator.ListSimulations(r.Context(), filter)
	if err != nil {
		api.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.respondJSON(w, http.StatusOK, map[string]interface{}{
		"simulations": simulations,
	})
}

func (api *ControlAPI) handleGetSimulation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	simulation, err := api.simulator.GetSimulation(r.Context(), id)
	if err != nil {
		api.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	api.respondJSON(w, http.StatusOK, simulation)
}

func (api *ControlAPI) handleStartSimulation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	if err := api.simulator.StartSimulation(r.Context(), id); err != nil {
		api.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	api.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Simulation started",
		"id":      id,
	})
}

func (api *ControlAPI) handleStopSimulation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	if err := api.simulator.StopSimulation(r.Context(), id); err != nil {
		api.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	api.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Simulation stopped",
		"id":      id,
	})
}

func (api *ControlAPI) handleGetResults(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	results, err := api.simulator.GetSimulationResults(r.Context(), id)
	if err != nil {
		api.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	api.respondJSON(w, http.StatusOK, results)
}

// Chaos engineering handlers

func (api *ControlAPI) handleCPUSpike(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProcessPattern string  `json:"process_pattern"`
		Duration       string  `json:"duration"`
		Intensity      float64 `json:"intensity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// In a real implementation, this would trigger CPU spikes in matching processes
	api.logger.Info("triggering CPU spike",
		zap.String("pattern", req.ProcessPattern),
		zap.String("duration", req.Duration),
		zap.Float64("intensity", req.Intensity),
	)

	api.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "CPU spike triggered",
		"pattern": req.ProcessPattern,
	})
}

func (api *ControlAPI) handleMemoryLeak(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProcessPattern string `json:"process_pattern"`
		LeakRate       string `json:"leak_rate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// In a real implementation, this would trigger memory leaks in matching processes
	api.logger.Info("triggering memory leak",
		zap.String("pattern", req.ProcessPattern),
		zap.String("leak_rate", req.LeakRate),
	)

	api.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Memory leak triggered",
		"pattern":  req.ProcessPattern,
		"leakRate": req.LeakRate,
	})
}

func (api *ControlAPI) handleProcessKill(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProcessPattern string `json:"process_pattern"`
		Count          int    `json:"count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// In a real implementation, this would kill matching processes
	api.logger.Info("killing processes",
		zap.String("pattern", req.ProcessPattern),
		zap.Int("count", req.Count),
	)

	api.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Processes killed",
		"pattern": req.ProcessPattern,
		"count":   req.Count,
	})
}

// Health and info handlers

func (api *ControlAPI) handleHealth(w http.ResponseWriter, r *http.Request) {
	api.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (api *ControlAPI) handleInfo(w http.ResponseWriter, r *http.Request) {
	api.respondJSON(w, http.StatusOK, map[string]interface{}{
		"service":     "phoenix-process-simulator",
		"version":     "0.1.0",
		"description": "Process simulator for Phoenix platform experiments",
		"profiles": []string{
			"realistic",
			"high-cardinality",
			"process-churn",
			"chaos",
		},
	})
}

// Helper methods

func (api *ControlAPI) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		api.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (api *ControlAPI) respondError(w http.ResponseWriter, status int, message string) {
	api.respondJSON(w, status, map[string]interface{}{
		"error": message,
	})
}