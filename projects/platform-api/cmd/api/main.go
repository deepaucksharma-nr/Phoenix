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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

github.com/phoenix-vnext/platform/packages/go-common/store
	"github.com/phoenix-vnext/platform/projects/platform-api/internal/services"
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

	postgresStore, err := store.NewPostgresStore(dbURL)
	if err != nil {
		logger.Fatal("failed to initialize store", zap.Error(err))
	}
	defer postgresStore.Close()

	// Create experiment service
	experimentService := services.NewExperimentService(postgresStore, nil, logger)

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
	})

	// Start HTTP server
	httpAddr := fmt.Sprintf(":%d", defaultHTTPPort)
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: r,
	}

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("shutting down servers...")
		cancel()
	}()

	logger.Info("starting HTTP server", zap.String("addr", httpAddr))
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("HTTP server error", zap.Error(err))
	}
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