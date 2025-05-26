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

	"github.com/phoenix/platform/cmd/controller/internal/clients"
	"github.com/phoenix/platform/cmd/controller/internal/controller"
	grpcserver "github.com/phoenix/platform/cmd/controller/internal/grpc"
	"github.com/phoenix/platform/cmd/controller/internal/store"
	pb "github.com/phoenix/platform/pkg/api/v1"
)

const (
	defaultGRPCPort   = ":50051"
	defaultHTTPPort   = ":8081"
	shutdownTimeout   = 30 * time.Second
	serviceName       = "experiment-controller"
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

	logger.Info("starting experiment controller",
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

	// Initialize database
	postgresStore, err := initDatabase(cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
	}
	defer func() {
		if err := postgresStore.Close(); err != nil {
			logger.Error("failed to close database", zap.Error(err))
		}
	}()

	// Initialize clients
	generatorClient, kubernetesClient, err := initClients(cfg, logger)
	if err != nil {
		logger.Fatal("failed to initialize clients", zap.Error(err))
	}
	defer func() {
		if err := generatorClient.Close(); err != nil {
			logger.Error("failed to close generator client", zap.Error(err))
		}
	}()

	// Initialize experiment controller
	expController := initExperimentController(logger, postgresStore)

	// Initialize state machine
	stateMachine := initStateMachine(logger, expController, generatorClient, kubernetesClient)

	// Initialize scheduler
	scheduler := initScheduler(logger, expController, stateMachine)
	scheduler.Start(context.Background())
	defer scheduler.Stop()

	// Initialize gRPC server
	grpcServer := grpc.NewServer()
	
	// Register experiment service using simple implementation
	expService := grpcserver.NewSimpleExperimentServer(logger, expController)
	pb.RegisterExperimentServiceServer(grpcServer, expService)
	
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

	logger.Info("shutting down experiment controller")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown gRPC server
	grpcServer.GracefulStop()
	
	// Shutdown metrics server
	if err := metricsServer.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown metrics server", zap.Error(err))
	}

	logger.Info("experiment controller stopped")
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
	DatabaseURL        string
	Environment        string
	GeneratorEndpoint  string
}

func loadConfig() (*Config, error) {
	cfg := &Config{
		GRPCPort:          getEnvOrDefault("GRPC_PORT", defaultGRPCPort),
		MetricsPort:       getEnvOrDefault("METRICS_PORT", defaultHTTPPort),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		Environment:       getEnvOrDefault("ENVIRONMENT", "production"),
		GeneratorEndpoint: getEnvOrDefault("GENERATOR_ENDPOINT", "localhost:50052"),
	}

	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = "postgres://phoenix:phoenix@localhost:5432/phoenix?sslmode=disable"
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
		// TODO: Add readiness checks (DB connection, etc.)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	return &http.Server{
		Addr:    port,
		Handler: mux,
	}
}

func initDatabase(connectionString string, logger *zap.Logger) (*store.PostgresStore, error) {
	logger.Info("initializing database connection")
	return store.NewPostgresStore(connectionString, logger)
}

func initExperimentController(logger *zap.Logger, store controller.ExperimentStore) *controller.ExperimentController {
	logger.Info("initializing experiment controller")
	return controller.NewExperimentController(logger, store)
}

func initClients(cfg *Config, logger *zap.Logger) (*clients.GeneratorClient, *clients.KubernetesClient, error) {
	logger.Info("initializing clients")
	
	// Initialize generator client
	generatorClient := clients.NewGeneratorClient(logger, cfg.GeneratorEndpoint)
	
	// Connect to generator service (with timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := generatorClient.Connect(ctx); err != nil {
		logger.Warn("failed to connect to generator service, will retry later", zap.Error(err))
		// Don't fail startup if generator is not available
	}
	
	// Initialize Kubernetes client
	kubernetesClient, err := clients.NewKubernetesClient(logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize kubernetes client: %w", err)
	}
	
	return generatorClient, kubernetesClient, nil
}

func initStateMachine(logger *zap.Logger, expController *controller.ExperimentController, generatorClient *clients.GeneratorClient, kubernetesClient *clients.KubernetesClient) *controller.StateMachine {
	logger.Info("initializing state machine")
	return controller.NewStateMachine(logger, expController, generatorClient, kubernetesClient)
}

func initScheduler(logger *zap.Logger, expController *controller.ExperimentController, stateMachine *controller.StateMachine) *controller.Scheduler {
	logger.Info("initializing scheduler")
	// Run reconciliation every 30 seconds
	return controller.NewScheduler(logger, expController, stateMachine, 30*time.Second)
}
