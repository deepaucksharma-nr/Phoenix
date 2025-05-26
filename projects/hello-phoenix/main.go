package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gorilla/mux"
)

// Service info
var (
    serviceName = "hello-phoenix"
    version     = "1.0.0"
    startTime   = time.Now()
)

// Experiment represents a cost optimization experiment
type Experiment struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Status      string    `json:"status"`
    CostSaving  float64   `json:"cost_saving_percent"`
    CreatedAt   time.Time `json:"created_at"`
}

// In-memory storage for demo
var experiments = []Experiment{
    {
        ID:         "exp-001",
        Name:       "Reduce Prometheus Metrics",
        Status:     "running",
        CostSaving: 45.2,
        CreatedAt:  time.Now().Add(-24 * time.Hour),
    },
    {
        ID:         "exp-002",
        Name:       "Optimize Datadog Tags",
        Status:     "completed",
        CostSaving: 72.8,
        CreatedAt:  time.Now().Add(-48 * time.Hour),
    },
}

func main() {
    router := mux.NewRouter()

    // Health endpoint
    router.HandleFunc("/health", healthHandler).Methods("GET")
    
    // Service info
    router.HandleFunc("/info", infoHandler).Methods("GET")
    
    // API endpoints
    api := router.PathPrefix("/api/v1").Subrouter()
    api.HandleFunc("/experiments", listExperiments).Methods("GET")
    api.HandleFunc("/experiments/{id}", getExperiment).Methods("GET")
    api.HandleFunc("/metrics", getMetrics).Methods("GET")

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("🚀 Phoenix Platform - %s v%s starting on port %s", serviceName, version, port)
    log.Printf("📊 Demonstrating observability cost optimization")
    log.Printf("🔗 Try: http://localhost:%s/info", port)
    
    if err := http.ListenAndServe(":"+port, router); err != nil {
        log.Fatal(err)
    }
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "healthy",
        "uptime": time.Since(startTime).String(),
    })
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "service":     serviceName,
        "version":     version,
        "description": "Phoenix Platform - Reduce observability costs by 90%",
        "features": []string{
            "Metric cardinality reduction",
            "A/B testing for telemetry pipelines",
            "Real-time cost analysis",
            "Automated optimization",
        },
        "endpoints": map[string]string{
            "health":      "/health",
            "experiments": "/api/v1/experiments",
            "metrics":     "/api/v1/metrics",
        },
    })
}

func listExperiments(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "experiments": experiments,
        "total":       len(experiments),
    })
}

func getExperiment(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    
    for _, exp := range experiments {
        if exp.ID == id {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(exp)
            return
        }
    }
    
    http.NotFound(w, r)
}

func getMetrics(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "total_experiments":      len(experiments),
        "average_cost_saving":    59.0,
        "metrics_processed":      1234567,
        "cardinality_reduction":  "87%",
        "monthly_savings_usd":    45000,
        "active_optimizations":   3,
    })
}