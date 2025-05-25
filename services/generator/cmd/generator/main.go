package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

type PipelineConfig struct {
    APIVersion string            `json:"apiVersion"`
    Kind       string            `json:"kind"`
    Metadata   map[string]string `json:"metadata"`
    Spec       map[string]interface{} `json:"spec"`
}

func main() {
    log.Println("Starting Phoenix Config Generator...")

    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    
    // Health check
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"healthy","service":"generator"}`))
    })
    
    // Generator endpoints
    r.Post("/generate", generateConfig)
    r.Get("/templates", listTemplates)
    
    srv := &http.Server{
        Addr:    ":8083",
        Handler: r,
    }
    
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Generator server failed: %v", err)
        }
    }()
    
    log.Println("Generator started on :8083")
    
    // Wait for interrupt
    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
    <-done
    
    log.Println("Shutting down generator...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Generator forced to shutdown: %v", err)
    }
    
    log.Println("Generator stopped")
}

func generateConfig(w http.ResponseWriter, r *http.Request) {
    // Mock config generation
    config := PipelineConfig{
        APIVersion: "phoenix.io/v1alpha1",
        Kind:       "PhoenixProcessPipeline",
        Metadata: map[string]string{
            "name":      "generated-pipeline",
            "namespace": "phoenix-system",
        },
        Spec: map[string]interface{}{
            "pipeline": map[string]interface{}{
                "receivers": []string{"otlp"},
                "processors": []string{"batch", "memory_limiter"},
                "exporters": []string{"prometheus"},
            },
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(config)
}

func listTemplates(w http.ResponseWriter, r *http.Request) {
    templates := map[string][]string{
        "templates": {
            "baseline-v1",
            "optimized-v1",
            "experimental-v1",
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(templates)
}