package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ConfigTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Config      map[string]interface{} `json:"config"`
}

type GenerateRequest struct {
	TemplateID   string                 `json:"template_id"`
	ExperimentID string                 `json:"experiment_id"`
	Parameters   map[string]interface{} `json:"parameters"`
}

type GenerateResponse struct {
	ConfigID string                 `json:"config_id"`
	Config   map[string]interface{} `json:"config"`
	Status   string                 `json:"status"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	// Predefined templates
	templates := map[string]ConfigTemplate{
		"basic-otel": {
			ID:          "basic-otel",
			Name:        "Basic OpenTelemetry",
			Description: "Basic OpenTelemetry collector configuration",
			Type:        "otel-collector",
			Config: map[string]interface{}{
				"receivers": map[string]interface{}{
					"otlp": map[string]interface{}{
						"protocols": map[string]interface{}{
							"grpc": map[string]interface{}{
								"endpoint": "0.0.0.0:4317",
							},
						},
					},
				},
				"processors": map[string]interface{}{
					"batch": map[string]interface{}{},
				},
				"exporters": map[string]interface{}{
					"prometheus": map[string]interface{}{
						"endpoint": "0.0.0.0:8889",
					},
				},
				"service": map[string]interface{}{
					"pipelines": map[string]interface{}{
						"metrics": map[string]interface{}{
							"receivers":  []string{"otlp"},
							"processors": []string{"batch"},
							"exporters":  []string{"prometheus"},
						},
					},
				},
			},
		},
		"cardinality-reduction": {
			ID:          "cardinality-reduction",
			Name:        "Cardinality Reduction",
			Description: "Configuration with cardinality reduction processors",
			Type:        "otel-collector",
			Config: map[string]interface{}{
				"receivers": map[string]interface{}{
					"otlp": map[string]interface{}{
						"protocols": map[string]interface{}{
							"grpc": map[string]interface{}{
								"endpoint": "0.0.0.0:4317",
							},
						},
					},
				},
				"processors": map[string]interface{}{
					"batch": map[string]interface{}{},
					"filter": map[string]interface{}{
						"metrics": map[string]interface{}{
							"exclude": map[string]interface{}{
								"match_type": "regexp",
								"metric_names": []string{
									".*_debug_.*",
									".*_test_.*",
								},
							},
						},
					},
					"attributes": map[string]interface{}{
						"actions": []map[string]interface{}{
							{
								"key":    "environment",
								"action": "upsert",
								"value":  "production",
							},
						},
					},
				},
				"exporters": map[string]interface{}{
					"prometheus": map[string]interface{}{
						"endpoint": "0.0.0.0:8889",
					},
				},
				"service": map[string]interface{}{
					"pipelines": map[string]interface{}{
						"metrics": map[string]interface{}{
							"receivers":  []string{"otlp"},
							"processors": []string{"batch", "filter", "attributes"},
							"exporters":  []string{"prometheus"},
						},
					},
				},
			},
		},
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	r.Get("/templates", func(w http.ResponseWriter, r *http.Request) {
		templateList := make([]ConfigTemplate, 0, len(templates))
		for _, t := range templates {
			templateList = append(templateList, t)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(templateList)
	})

	r.Get("/templates/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		template, ok := templates[id]
		if !ok {
			http.Error(w, "Template not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(template)
	})

	r.Post("/generate", func(w http.ResponseWriter, r *http.Request) {
		var req GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		template, ok := templates[req.TemplateID]
		if !ok {
			http.Error(w, "Template not found", http.StatusNotFound)
			return
		}

		// Generate config based on template
		generatedConfig := make(map[string]interface{})
		for k, v := range template.Config {
			generatedConfig[k] = v
		}

		// Apply any custom parameters
		if req.Parameters != nil {
			// Simple merge - in real implementation, this would be more sophisticated
			for k, v := range req.Parameters {
				generatedConfig[k] = v
			}
		}

		resp := GenerateResponse{
			ConfigID: fmt.Sprintf("config-%s-%d", req.ExperimentID, time.Now().Unix()),
			Config:   generatedConfig,
			Status:   "generated",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Generator service starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Generator service stopped")
}