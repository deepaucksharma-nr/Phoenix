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

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/phoenix/platform/pkg/generator"
)

const (
	defaultHTTPPort = ":8082"
	shutdownTimeout = 30 * time.Second
	serviceName     = "config-generator"
	serviceVersion  = "0.1.0"
)

func main() {
	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			// Ignore sync errors
		}
	}()

	logger.Info("starting config generator",
		zap.String("service", serviceName),
		zap.String("version", serviceVersion),
	)

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		logger.Fatal("failed to load configuration", zap.Error(err))
	}

	// Initialize generator service
	generatorService := generator.NewService(logger, cfg.GitRepoURL, cfg.GitToken)

	// Initialize HTTP server with API routes
	mux := http.NewServeMux()
	
	// Health endpoints
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	
	// API endpoints
	mux.HandleFunc("/api/v1/generate", handleGenerateConfig(logger, generatorService))
	mux.HandleFunc("/api/v1/templates", handleListTemplates(logger, generatorService))

	// Start HTTP server
	httpServer := &http.Server{
		Addr:    cfg.HTTPPort,
		Handler: mux,
	}

	go func() {
		logger.Info("starting HTTP server", zap.String("port", cfg.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to serve HTTP", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down config generator")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown HTTP server", zap.Error(err))
	}

	logger.Info("config generator stopped")
}

func initLogger() (*zap.Logger, error) {
	env := os.Getenv("ENVIRONMENT")
	if env == "development" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

type Config struct {
	HTTPPort    string
	GitRepoURL  string
	GitToken    string
	Environment string
}

func loadConfig() (*Config, error) {
	cfg := &Config{
		HTTPPort:    getEnvOrDefault("HTTP_PORT", defaultHTTPPort),
		GitRepoURL:  getEnvOrDefault("GIT_REPO_URL", "https://github.com/phoenix/configs"),
		GitToken:    os.Getenv("GIT_TOKEN"),
		Environment: getEnvOrDefault("ENVIRONMENT", "production"),
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// handleGenerateConfig handles HTTP requests for config generation
func handleGenerateConfig(logger *zap.Logger, generatorService *generator.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req generator.GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("failed to decode request", zap.Error(err))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		logger.Info("generating config",
			zap.String("experiment_id", req.ExperimentID),
			zap.String("baseline", req.BaselinePipeline),
			zap.String("candidate", req.CandidatePipeline),
		)

		// Generate configurations
		resp, err := generatorService.GenerateExperimentConfig(r.Context(), &req)
		if err != nil {
			logger.Error("failed to generate config", zap.Error(err))
			http.Error(w, fmt.Sprintf("Failed to generate config: %v", err), http.StatusInternalServerError)
			return
		}

		// Create git pull request
		if err := generatorService.CreateGitPR(r.Context(), resp.GitBranch, resp.KubernetesManifests); err != nil {
			logger.Error("failed to create git PR", zap.Error(err))
			// Don't fail the request, just log the error
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("failed to encode response", zap.Error(err))
		}
	}
}

// handleListTemplates handles HTTP requests for listing available templates
func handleListTemplates(logger *zap.Logger, generatorService *generator.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		templates := generatorService.ListTemplates()
		
		// Convert to response format
		type templateResponse struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		
		response := struct {
			Templates []templateResponse `json:"templates"`
		}{
			Templates: make([]templateResponse, len(templates)),
		}
		
		for i, tmpl := range templates {
			response.Templates[i] = templateResponse{
				Name:        tmpl.Name,
				Description: tmpl.Description,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("failed to encode response", zap.Error(err))
		}
	}
}