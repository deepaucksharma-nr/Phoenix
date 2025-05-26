package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/phoenix/platform/packages/go-common/models"
	commonstore "github.com/phoenix/platform/packages/go-common/store"
	"github.com/phoenix/platform/projects/platform-api/internal/services"
	"github.com/phoenix/platform/projects/platform-api/internal/store"
	ws "github.com/phoenix/platform/projects/platform-api/internal/websocket"
)

const (
	defaultHTTPPort = 8080
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	// Initialize store
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://phoenix:phoenix@localhost/phoenix?sslmode=disable"
	}

	postgresStore, err := commonstore.NewPostgresStore(dbURL)
	if err != nil {
		logger.Fatal("failed to initialize store", zap.Error(err))
	}
	defer postgresStore.Close()

	// Create experiment service
	experimentService := services.NewExperimentService(postgresStore, nil, logger)
	
	// Create pipeline deployment store and service
	pipelineStore := store.NewPostgresPipelineDeploymentStore(postgresStore)
	pipelineService := services.NewPipelineDeploymentService(pipelineStore, logger)

	// Create WebSocket hub
	wsHub := ws.NewHub(logger)
	go wsHub.Run()

	// Create HTTP router
	r := chi.NewRouter()
	
	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Experiments
		r.Route("/experiments", func(r chi.Router) {
			r.Get("/", listExperiments(experimentService))
			r.Post("/", createExperiment(experimentService))
			r.Get("/{id}", getExperiment(experimentService))
			r.Delete("/{id}", deleteExperiment(experimentService))
			r.Put("/{id}/status", updateExperimentStatus(experimentService))
		})
		
		// Pipeline Deployments
		r.Route("/pipelines/deployments", func(r chi.Router) {
			r.Get("/", listDeployments(pipelineService))
			r.Post("/", createDeployment(pipelineService))
			r.Get("/{id}", getDeployment(pipelineService))
			r.Put("/{id}", updateDeployment(pipelineService))
			r.Delete("/{id}", deleteDeployment(pipelineService))
		})
	})

	// WebSocket endpoint
	r.HandleFunc("/ws", handleWebSocket(wsHub, logger))

	// Start HTTP server
	httpAddr := fmt.Sprintf(":%d", defaultHTTPPort)
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		logger.Info("starting HTTP server", zap.String("addr", httpAddr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	logger.Info("shutting down servers...")
	
	// Gracefully shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}
	
	logger.Info("servers shut down successfully")
}

// HTTP handlers
func listExperiments(svc *services.ExperimentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		experiments, err := svc.ListExperiments(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(experiments)
	}
}

func createExperiment(svc *services.ExperimentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name              string            `json:"name"`
			Description       string            `json:"description"`
			BaselinePipeline  string            `json:"baseline_pipeline"`
			CandidatePipeline string            `json:"candidate_pipeline"`
			TargetNodes       map[string]string `json:"target_nodes"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		experiment, err := svc.CreateExperiment(r.Context(), req.Name, req.Description, 
			req.BaselinePipeline, req.CandidatePipeline, req.TargetNodes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(experiment)
	}
}

func getExperiment(svc *services.ExperimentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		experiment, err := svc.GetExperiment(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(experiment)
	}
}

func deleteExperiment(svc *services.ExperimentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := svc.DeleteExperiment(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func updateExperimentStatus(svc *services.ExperimentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var req struct {
			Status string `json:"status"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := svc.UpdateExperimentStatus(r.Context(), id, req.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusOK)
	}
}

// Pipeline deployment handlers
func listDeployments(svc *services.PipelineDeploymentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &models.ListDeploymentsRequest{
			Namespace:    r.URL.Query().Get("namespace"),
			Status:       r.URL.Query().Get("status"),
			PipelineName: r.URL.Query().Get("pipeline_name"),
		}
		
		resp, err := svc.ListDeployments(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func createDeployment(svc *services.PipelineDeploymentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateDeploymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		deployment, err := svc.CreateDeployment(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(deployment)
	}
}

func getDeployment(svc *services.PipelineDeploymentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		deployment, err := svc.GetDeployment(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(deployment)
	}
}

func updateDeployment(svc *services.PipelineDeploymentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var req models.UpdateDeploymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		if err := svc.UpdateDeployment(r.Context(), id, &req); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusOK)
	}
}

func deleteDeployment(svc *services.PipelineDeploymentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := svc.DeleteDeployment(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, implement proper origin checking
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// handleWebSocket handles WebSocket connections
func handleWebSocket(hub *ws.Hub, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error("WebSocket upgrade failed", zap.Error(err))
			return
		}

		client := ws.NewClient(conn, hub)
		hub.Register <- client

		// Allow collection of memory referenced by the caller by doing all work in
		// new goroutines.
		go client.WritePump()
		go client.ReadPump()
	}
}