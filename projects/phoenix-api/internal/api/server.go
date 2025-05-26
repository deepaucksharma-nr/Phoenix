package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/phoenix/platform/projects/phoenix-api/internal/config"
	"github.com/phoenix/platform/projects/phoenix-api/internal/controller"
	"github.com/phoenix/platform/projects/phoenix-api/internal/store"
	"github.com/phoenix/platform/projects/phoenix-api/internal/tasks"
	"github.com/phoenix/platform/projects/phoenix-api/internal/websocket"
	"github.com/rs/zerolog/log"
)

type Server struct {
	store         store.Store
	hub           *websocket.Hub
	config        *config.Config
	taskQueue     *tasks.Queue
	expController *controller.ExperimentController
}

func NewServer(store store.Store, hub *websocket.Hub, config *config.Config) *Server {
	taskQueue := tasks.NewQueue(store)
	expController := controller.NewExperimentController(store, taskQueue)
	
	return &Server{
		store:         store,
		hub:           hub,
		config:        config,
		taskQueue:     taskQueue,
		expController: expController,
	}
}

func (s *Server) SetupRoutes(r chi.Router) {
	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Experiment endpoints (from controller service)
		r.Route("/experiments", func(r chi.Router) {
			r.Post("/", s.handleCreateExperiment)
			r.Get("/", s.handleListExperiments)
			r.Get("/{id}", s.handleGetExperiment)
			r.Put("/{id}/phase", s.handleUpdateExperimentPhase)
			r.Post("/{id}/start", s.handleStartExperiment)
			r.Post("/{id}/stop", s.handleStopExperiment)
			r.Post("/{id}/promote", s.handlePromoteExperiment)
			r.Post("/{id}/kpis", s.handleCalculateKPIs)
			r.Get("/{id}/kpis", s.handleGetKPIs)
			r.Post("/{id}/analyze", s.handleAnalyzeExperiment)
		})

		// Pipeline endpoints (existing from platform-api)
		r.Route("/pipelines", func(r chi.Router) {
			r.Get("/", s.handleListPipelines)
			r.Get("/{id}", s.handleGetPipeline)
			r.Get("/status", s.handleGetPipelineStatus)
		})
		
		// Pipeline deployment endpoints
		r.Route("/deployments", func(r chi.Router) {
			r.Post("/", s.handleCreateDeployment)
			r.Get("/", s.handleListDeployments)
			r.Get("/{id}", s.handleGetDeployment)
			r.Put("/{id}", s.handleUpdateDeployment)
			r.Delete("/{id}", s.handleDeleteDeployment)
			r.Post("/{id}/rollback", s.handleRollbackDeployment)
			r.Get("/{id}/status", s.handleGetDeploymentStatus)
		})
		
		// Load simulation endpoints
		r.Route("/loadsimulations", func(r chi.Router) {
			r.Post("/", s.handleStartLoadSimulation)
			r.Get("/", s.handleListLoadSimulations)
			r.Get("/{id}", s.handleGetLoadSimulation)
			r.Delete("/{id}", s.handleStopLoadSimulation)
		})

		// WebSocket endpoint
		r.HandleFunc("/ws", s.handleWebSocket)
		
		// UI-focused endpoints
		r.Route("/metrics", func(r chi.Router) {
			r.Get("/cost-flow", s.handleGetMetricCostFlow)
			r.Get("/cardinality", s.handleGetCardinalityBreakdown)
		})
		
		r.Route("/fleet", func(r chi.Router) {
			r.Get("/status", s.handleGetFleetStatus)
			r.Get("/map", s.handleGetAgentMap)
		})
		
		r.Route("/experiments", func(r chi.Router) {
			r.Post("/wizard", s.handleCreateExperimentWizard)
			r.Post("/{id}/rollback", s.handleInstantRollback)
		})
		
		r.Route("/pipelines", func(r chi.Router) {
			r.Get("/templates", s.handleGetPipelineTemplates)
			r.Post("/preview", s.handlePreviewPipelineImpact)
			r.Post("/quick-deploy", s.handleQuickDeploy)
		})
		
		r.Route("/tasks", func(r chi.Router) {
			r.Get("/active", s.handleGetActiveTasks)
			r.Get("/queue", s.handleGetTaskQueue)
		})
		
		r.Get("/cost-analytics", s.handleGetCostAnalytics)

		// Agent endpoints (new for lean architecture)
		r.Route("/agent", func(r chi.Router) {
			r.Use(s.agentAuthMiddleware)
			
			// Task polling (long-poll with 30s timeout)
			r.Get("/tasks", s.handleAgentGetTasks)
			
			// Task status updates
			r.Post("/tasks/{taskId}/status", s.handleTaskStatusUpdate)
			
			// Agent heartbeat
			r.Post("/heartbeat", s.handleAgentHeartbeat)
			
			// Metrics push (batch)
			r.Post("/metrics", s.handleAgentMetrics)
			
			// Log streaming
			r.Post("/logs", s.handleAgentLogs)
		})
	})
}

// Middleware to authenticate agents
func (s *Server) agentAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple host-based auth for now
		hostID := r.Header.Get("X-Agent-Host-ID")
		if hostID == "" {
			http.Error(w, "Missing X-Agent-Host-ID header", http.StatusUnauthorized)
			return
		}
		
		// Add host ID to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "hostID", hostID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Failed to encode response")
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	client, err := s.hub.HandleConnection(w, r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to handle WebSocket connection")
		return
	}
	
	// Handle client messages
	go client.ReadMessages()
	go client.WriteMessages()
}