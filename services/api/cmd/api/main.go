package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    log.Println("Starting Phoenix API Service...")

    // Create router
    r := chi.NewRouter()
    
    // Middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.RequestID)
    
    // Health check endpoint
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"healthy","service":"api"}`))
    })
    
    // API endpoints
    r.Route("/api/v1", func(r chi.Router) {
        // Experiments endpoints
        r.Route("/experiments", func(r chi.Router) {
            r.Post("/", createExperiment)
            r.Get("/{id}", getExperiment)
            r.Get("/", listExperiments)
            r.Put("/{id}", updateExperiment)
            r.Delete("/{id}", deleteExperiment)
            r.Post("/{id}/start", startExperiment)
            r.Post("/{id}/stop", stopExperiment)
        })
        
        // Pipelines endpoints
        r.Route("/pipelines", func(r chi.Router) {
            r.Get("/", listPipelines)
            r.Get("/{id}", getPipeline)
        })
    })
    
    // Metrics endpoint
    r.Handle("/metrics", promhttp.Handler())
    
    // Start server
    srv := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }
    
    // Graceful shutdown
    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed to start: %v", err)
        }
    }()
    log.Println("API Server started on :8080")
    
    <-done
    log.Println("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }
    
    log.Println("Server stopped")
}

// Handler functions
func createExperiment(w http.ResponseWriter, r *http.Request) {
    // Mock implementation
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    response := fmt.Sprintf(`{
        "id": "exp-%d",
        "name": "e2e-test-experiment",
        "status": "created",
        "created_at": "%s"
    }`, time.Now().Unix(), time.Now().Format(time.RFC3339))
    w.Write([]byte(response))
}

func getExperiment(w http.ResponseWriter, r *http.Request) {
    expID := chi.URLParam(r, "id")
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    response := fmt.Sprintf(`{
        "id": "%s",
        "name": "e2e-test-experiment",
        "status": "running",
        "created_at": "%s"
    }`, expID, time.Now().Format(time.RFC3339))
    w.Write([]byte(response))
}

func listExperiments(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"experiments":[],"total":0}`))
}

func updateExperiment(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotImplemented)
}

func deleteExperiment(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNoContent)
}

func startExperiment(w http.ResponseWriter, r *http.Request) {
    expID := chi.URLParam(r, "id")
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    response := fmt.Sprintf(`{"id":"%s","status":"starting"}`, expID)
    w.Write([]byte(response))
}

func stopExperiment(w http.ResponseWriter, r *http.Request) {
    expID := chi.URLParam(r, "id")
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    response := fmt.Sprintf(`{"id":"%s","status":"stopping"}`, expID)
    w.Write([]byte(response))
}

func listPipelines(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{
        "pipelines": [
            {"id": "baseline-v1", "name": "Baseline Pipeline"},
            {"id": "optimized-v1", "name": "Optimized Pipeline"}
        ]
    }`))
}

func getPipeline(w http.ResponseWriter, r *http.Request) {
    pipelineID := chi.URLParam(r, "id")
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    response := fmt.Sprintf(`{"id":"%s","name":"Pipeline %s"}`, pipelineID, pipelineID)
    w.Write([]byte(response))
}