package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/phoenix/platform/pkg/http/response"
	"github.com/phoenix/platform/projects/phoenix-api/internal/config"
	"github.com/phoenix/platform/projects/phoenix-api/internal/controller"
	"github.com/phoenix/platform/projects/phoenix-api/internal/services"
	"github.com/phoenix/platform/projects/phoenix-api/internal/store"
	"github.com/phoenix/platform/projects/phoenix-api/internal/tasks"
	phoenixws "github.com/phoenix/platform/projects/phoenix-api/internal/websocket"
	"github.com/rs/zerolog/log"
)

type Server struct {
	store            store.Store
	hub              *phoenixws.Hub
	config           *config.Config
	taskQueue        *tasks.Queue
	expController    *controller.ExperimentController
	metricsCollector *services.MetricsCollector
	analysisService  *services.AnalysisService
	templateRenderer *services.PipelineTemplateRenderer
	costService      *services.CostService
	jwtService       *services.JWTService
	wsUpgrader       websocket.Upgrader
}

func NewServer(store store.Store, hub *phoenixws.Hub, config *config.Config) (*Server, error) {
	taskQueue := tasks.NewQueue(store)
	expController := controller.NewExperimentController(store, taskQueue)

	// Initialize metrics collector
	metricsCollector, err := services.NewMetricsCollector(store, config.PrometheusURL)
	if err != nil {
		return nil, err
	}

	// TODO: Wire metrics collector to state machine for auto-start
	// For now, metrics collection can be started manually via API

	// Initialize analysis service
	analysisService, err := services.NewAnalysisService(store, config.PrometheusURL)
	if err != nil {
		return nil, err
	}

	// Initialize template renderer
	templateRenderer := services.NewPipelineTemplateRenderer()

	// Load built-in templates
	for name, tmpl := range templateRenderer.GetBuiltinTemplates() {
		if err := templateRenderer.LoadTemplate(name, tmpl); err != nil {
			log.Error().Err(err).Str("template", name).Msg("Failed to load built-in template")
		}
	}

	// Initialize cost service
	costService := services.NewCostService(store, config.CostRates)

	// Initialize JWT service
	jwtSecret := []byte(config.JWTSecret)
	jwtService := services.NewJWTService(jwtSecret, "phoenix-platform", store)

	// Initialize WebSocket upgrader
	wsUpgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins for development (should be restricted in production)
			return true
		},
	}

	return &Server{
		store:            store,
		hub:              hub,
		config:           config,
		taskQueue:        taskQueue,
		expController:    expController,
		metricsCollector: metricsCollector,
		analysisService:  analysisService,
		templateRenderer: templateRenderer,
		costService:      costService,
		jwtService:       jwtService,
		wsUpgrader:       wsUpgrader,
	}, nil
}

// GetTaskQueue returns the task queue instance
func (s *Server) GetTaskQueue() *tasks.Queue {
	return s.taskQueue
}

func (s *Server) SetupRoutes(r chi.Router) {
	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Authentication endpoints (no auth middleware)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", s.handleLogin)
			r.Post("/refresh", s.handleRefreshToken)
			r.Post("/logout", s.handleLogout)
			r.Post("/register", s.handleRegister) // Optional, for development
		})

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
			r.Get("/{id}/metrics", s.handleGetExperimentMetrics)
			r.Post("/{id}/analyze", s.handleAnalyzeExperiment)
			r.Get("/{id}/cost-analysis", s.handleGetCostAnalysis)
			// UI-focused experiment endpoints
			r.Post("/wizard", s.handleCreateExperimentWizard)
			r.Post("/{id}/rollback", s.handleInstantRollback)
		})

		// Pipeline endpoints (existing from platform-api)
		r.Route("/pipelines", func(r chi.Router) {
			r.Get("/", s.handleListPipelines)
			r.Get("/{id}", s.handleGetPipeline)
			r.Get("/status", s.handleGetPipelineStatus)
			r.Post("/validate", s.handleValidatePipeline)
			r.Post("/render", s.handleRenderPipeline)
			// UI-focused pipeline endpoints
			r.Get("/templates", s.handleGetPipelineTemplates)
			r.Post("/preview", s.handlePreviewPipelineImpact)
			r.Post("/quick-deploy", s.handleQuickDeploy)

			// Pipeline deployment endpoints (nested under /pipelines)
			r.Route("/deployments", func(r chi.Router) {
				r.Post("/", s.handleCreateDeployment)
				r.Get("/", s.handleListDeployments)
				r.Get("/{id}", s.handleGetDeployment)
				r.Put("/{id}", s.handleUpdateDeployment)
				r.Delete("/{id}", s.handleDeleteDeployment)
				r.Post("/{id}/rollback", s.handleRollbackDeployment)
				r.Get("/{id}/status", s.handleGetDeploymentStatus)
				r.Get("/{id}/config", s.handleGetPipelineConfig)
				r.Get("/{id}/versions", s.handleListDeploymentVersions)
			})
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

		r.Route("/tasks", func(r chi.Router) {
			r.Get("/active", s.handleGetActiveTasks)
			r.Get("/queue", s.handleGetTaskQueue)
		})

		r.Get("/cost-analytics", s.handleGetCostAnalytics)
		r.Get("/cost-flow", s.handleGetMetricCostFlow) // Add top-level cost-flow route

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

		// WebSocket endpoint
		r.Get("/ws", s.handleWebSocket)
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

// Compatibility wrappers for existing code
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	response.JSON(w, status, data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	response.Error(w, status, message)
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow connections from any origin for now
			// TODO: Implement proper CORS checking in production
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}

	// Create new client and register with hub
	client := phoenixws.NewClient(conn, s.hub)

	// Register client with hub
	s.hub.Register <- client

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()

	log.Info().Str("remote_addr", r.RemoteAddr).Msg("WebSocket client connected")
}
