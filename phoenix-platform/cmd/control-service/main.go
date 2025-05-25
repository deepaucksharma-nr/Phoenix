package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	grpcserver "github.com/phoenix/platform/cmd/control-service/internal/grpc"
	controllerv1 "github.com/phoenix/platform/api/proto/v1/controller"
)

const (
	defaultGRPCPort   = ":50053"
	defaultHTTPPort   = ":8083"
	shutdownTimeout   = 30 * time.Second
	serviceName       = "control-service"
	serviceVersion    = "0.1.0"
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

	logger.Info("starting control service",
		zap.String("service", serviceName),
		zap.String("version", serviceVersion),
	)

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		logger.Fatal("failed to load configuration", zap.Error(err))
	}

	// Initialize metrics server
	metricsServer := initMetricsServer(cfg.MetricsPort, logger)
	go func() {
		logger.Info("starting metrics server", zap.String("port", cfg.MetricsPort))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server failed", zap.Error(err))
		}
	}()

	// Initialize gRPC server
	grpcServer := grpc.NewServer()
	
	// Register controller service
	controllerService := grpcserver.NewControllerServer(logger)
	controllerv1.RegisterControllerServiceServer(grpcServer, controllerService)
	
	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	
	// Enable reflection for grpcurl debugging
	reflection.Register(grpcServer)

	// Start gRPC server
	listener, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		logger.Fatal("failed to listen", zap.String("port", cfg.GRPCPort), zap.Error(err))
	}

	go func() {
		logger.Info("starting gRPC server", zap.String("port", cfg.GRPCPort))
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatal("failed to serve gRPC", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down control service")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown gRPC server
	grpcServer.GracefulStop()
	
	// Shutdown metrics server
	if err := metricsServer.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown metrics server", zap.Error(err))
	}

	logger.Info("control service stopped")
}

func initLogger() (*zap.Logger, error) {
	env := os.Getenv("ENVIRONMENT")
	if env == "development" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

type Config struct {
	GRPCPort           string
	MetricsPort        string
	Environment        string
}

func loadConfig() (*Config, error) {
	cfg := &Config{
		GRPCPort:          getEnvOrDefault("GRPC_PORT", defaultGRPCPort),
		MetricsPort:       getEnvOrDefault("METRICS_PORT", defaultHTTPPort),
		Environment:       getEnvOrDefault("ENVIRONMENT", "production"),
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initMetricsServer(port string, logger *zap.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	return &http.Server{
		Addr:    port,
		Handler: mux,
	}
}