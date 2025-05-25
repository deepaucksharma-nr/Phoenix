package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func main() {
    log.Println("Starting Phoenix Experiment Controller...")

    // Create router for health checks
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"healthy","service":"controller"}`))
    })
    
    // Start HTTP server for health checks
    srv := &http.Server{
        Addr:    ":8082",
        Handler: r,
    }
    
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Health server failed: %v", err)
        }
    }()
    
    log.Println("Controller started on :8082")
    
    // Simulate controller work
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                log.Println("Controller: Checking for new experiments...")
                // In real implementation, this would:
                // 1. Query database for pending experiments
                // 2. Create Kubernetes resources
                // 3. Update experiment status
            }
        }
    }()
    
    // Wait for interrupt
    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
    <-done
    
    log.Println("Shutting down controller...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Controller forced to shutdown: %v", err)
    }
    
    log.Println("Controller stopped")
}