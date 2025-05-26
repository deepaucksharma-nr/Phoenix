package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/phoenix-vnext/platform/packages/go-common/models"
	"github.com/phoenix-vnext/platform/packages/go-common/store"
	pb "github.com/phoenix-vnext/platform/packages/contracts/proto/v1"
	"github.com/phoenix-vnext/platform/services/api/internal/services"
)

const (
	defaultGRPCPort = 5050
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

	// Get database connection for pipeline deployment service
	db := postgresStore.DB()

	// Create gRPC server (simplified, without auth for now)
	grpcServer := grpc.NewServer()

	// Register services
	experimentService := services.NewExperimentService(postgresStore, nil, logger)
	pb.RegisterExperimentServiceServer(grpcServer, experimentService)

	// Enable reflection
	reflection.Register(grpcServer)

	// Start gRPC server
	grpcPort := getEnvInt("GRPC_PORT", defaultGRPCPort)
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	go func() {
		logger.Info("starting gRPC server", zap.Int("port", grpcPort))
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Fatal("failed to serve gRPC", zap.Error(err))
		}
	}()

	// Create HTTP server
	httpPort := getEnvInt("HTTP_PORT", defaultHTTPPort)
	httpServer := createHTTPServer(httpPort, grpcPort, logger, db)

	// Start HTTP server
	go func() {
		logger.Info("starting HTTP server", zap.Int("port", httpPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to serve HTTP", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down servers...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown HTTP server", zap.Error(err))
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	logger.Info("servers stopped")
}

func createHTTPServer(httpPort, grpcPort int, logger *zap.Logger, db *sql.DB) *http.Server {
	// Create router
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))
	router.Use(middleware.Timeout(60 * time.Second))

	// CORS
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "3600")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Health check
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Metrics
	router.Handle("/metrics", promhttp.Handler())

	// gRPC-Gateway
	ctx := context.Background()
	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	endpoint := fmt.Sprintf("localhost:%d", grpcPort)

	err := pb.RegisterExperimentServiceHandlerFromEndpoint(ctx, gwmux, endpoint, opts)
	if err != nil {
		logger.Fatal("failed to register gateway", zap.Error(err))
	}

	// Mount API routes
	router.Mount("/api/v1", gwmux)

	// Create pipeline deployment service
	pipelineService := services.NewPipelineDeploymentService(db, logger)

	// Pipeline routes
	router.Route("/api/v1/pipelines", func(r chi.Router) {
		// Pipeline templates (static for now)
		r.Get("/", listPipelines)
		r.Get("/{name}", getPipeline)
		r.Post("/validate", validatePipeline)
		
		// Pipeline deployments
		r.Route("/deployments", func(r chi.Router) {
			r.Post("/", createPipelineDeploymentHandler(pipelineService, logger))
			r.Get("/", listPipelineDeploymentsHandler(pipelineService, logger))
			r.Get("/{id}", getPipelineDeploymentHandler(pipelineService, logger))
			r.Patch("/{id}", updatePipelineDeploymentHandler(pipelineService, logger))
			r.Delete("/{id}", deletePipelineDeploymentHandler(pipelineService, logger))
		})
	})

	// WebSocket handler - temporarily commented out
	// wsHandler := api.NewWebSocketHandler(logger)
	// router.HandleFunc("/ws", wsHandler.ServeHTTP)

	// Static files (dashboard)
	if os.Getenv("SERVE_STATIC") == "true" {
		fileServer := http.FileServer(http.Dir("./dist"))
		router.Handle("/*", fileServer)
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", httpPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Pipeline template handlers (static data for now)
func listPipelines(w http.ResponseWriter, r *http.Request) {
	pipelines := []map[string]interface{}{
		{
			"name":        "process-baseline-v1",
			"description": "Baseline configuration with no optimization",
			"type":        "baseline",
		},
		{
			"name":        "process-topk-v1",
			"description": "Keep only top K resource-consuming processes",
			"type":        "optimization",
			"parameters": map[string]interface{}{
				"top_k": map[string]interface{}{
					"type":        "integer",
					"default":     10,
					"description": "Number of top processes to keep",
				},
			},
		},
		{
			"name":        "process-priority-filter-v1",
			"description": "Keep only critical processes",
			"type":        "optimization",
			"parameters": map[string]interface{}{
				"critical_processes": map[string]interface{}{
					"type":        "array",
					"required":    true,
					"description": "List of critical process names to keep",
				},
			},
		},
		{
			"name":        "process-aggregated-v1",
			"description": "Aggregate metrics by process name",
			"type":        "optimization",
		},
		{
			"name":        "process-adaptive-v1",
			"description": "Adaptive filtering based on resource usage",
			"type":        "optimization",
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pipelines": pipelines,
	})
}

func getPipeline(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	
	pipeline := map[string]interface{}{
		"name":        name,
		"description": "Pipeline template",
		"config":      "# OpenTelemetry Collector configuration\n# This would contain the actual YAML config",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pipeline)
}

func validatePipeline(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Config string `json:"config"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// TODO: Implement actual validation
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":   true,
		"message": "Pipeline configuration is valid",
	})
}

// Pipeline deployment handlers
func createPipelineDeploymentHandler(service *services.PipelineDeploymentService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateDeploymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		deployment, err := service.CreateDeployment(r.Context(), &req)
		if err != nil {
			logger.Error("failed to create deployment", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(deployment)
	}
}

func listPipelineDeploymentsHandler(service *services.PipelineDeploymentService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := models.ListDeploymentsRequest{
			Namespace:    r.URL.Query().Get("namespace"),
			Status:       r.URL.Query().Get("status"),
			PipelineName: r.URL.Query().Get("pipeline"),
		}
		
		// Parse page size
		if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
			var pageSize int
			fmt.Sscanf(pageSizeStr, "%d", &pageSize)
			req.PageSize = pageSize
		}
		if req.PageSize <= 0 || req.PageSize > 100 {
			req.PageSize = 20
		}
		
		resp, err := service.ListDeployments(r.Context(), &req)
		if err != nil {
			logger.Error("failed to list deployments", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func getPipelineDeploymentHandler(service *services.PipelineDeploymentService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deploymentID := chi.URLParam(r, "id")
		
		deployment, err := service.GetDeployment(r.Context(), deploymentID)
		if err != nil {
			if err.Error() == "deployment not found" {
				http.Error(w, "Deployment not found", http.StatusNotFound)
				return
			}
			logger.Error("failed to get deployment", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(deployment)
	}
}

func updatePipelineDeploymentHandler(service *services.PipelineDeploymentService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deploymentID := chi.URLParam(r, "id")
		
		var req models.UpdateDeploymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		err := service.UpdateDeployment(r.Context(), deploymentID, &req)
		if err != nil {
			if err.Error() == "deployment not found" {
				http.Error(w, "Deployment not found", http.StatusNotFound)
				return
			}
			logger.Error("failed to update deployment", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"deployment_id": deploymentID,
			"status":        "updating",
			"message":       "Deployment update initiated",
		})
	}
}

func deletePipelineDeploymentHandler(service *services.PipelineDeploymentService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deploymentID := chi.URLParam(r, "id")
		
		err := service.DeleteDeployment(r.Context(), deploymentID)
		if err != nil {
			if err.Error() == "deployment not found" {
				http.Error(w, "Deployment not found", http.StatusNotFound)
				return
			}
			logger.Error("failed to delete deployment", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"deployment_id": deploymentID,
			"status":        "deleting",
			"message":       "Deployment removal initiated",
		})
	}
}