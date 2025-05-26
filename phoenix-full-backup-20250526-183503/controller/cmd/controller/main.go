package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "google.golang.org/grpc"

    // Phoenix platform imports
    "github.com/phoenix/platform/projects/controller/internal/clients"
    "github.com/phoenix/platform/projects/controller/internal/controller"
    controllergrpc "github.com/phoenix/platform/projects/controller/internal/grpc"
    "github.com/phoenix/platform/projects/controller/internal/store"
)

func main() {
    // Initialize zap logger directly
    zapConfig := zap.NewProductionConfig()
    if getEnv("ENVIRONMENT", "development") == "development" {
        zapConfig = zap.NewDevelopmentConfig()
    }
    zapConfig.Level = zap.NewAtomicLevelAt(getLogLevel(getEnv("LOG_LEVEL", "info")))
    
    logger, err := zapConfig.Build()
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }
    defer logger.Sync()

    logger.Info("Starting Phoenix Experiment Controller...")

    // Initialize database connection
    dbHost := getEnv("DB_HOST", "localhost")
    dbPort := getEnvInt("DB_PORT", 5432)
    dbUser := getEnv("DB_USER", "phoenix")
    dbPassword := getEnv("DB_PASSWORD", "phoenix")
    dbName := getEnv("DB_NAME", "phoenix")
    dbSSLMode := getEnv("DB_SSL_MODE", "disable")

    // Initialize experiment store with connection string
    connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
    
    experimentStore, err := store.NewPostgresStore(connStr, logger)
    if err != nil {
        logger.Fatal("Failed to initialize experiment store", zap.Error(err))
    }
    defer experimentStore.Close()

    // Initialize Kubernetes client
    k8sClient, err := clients.NewKubernetesClient(logger)
    if err != nil {
        logger.Fatal("Failed to initialize Kubernetes client", zap.Error(err))
    }

    // Initialize generator client
    generatorAddr := getEnv("GENERATOR_ADDR", "generator:8081")
    generatorClient := clients.NewGeneratorClient(logger, generatorAddr)

    // Initialize experiment controller
    experimentController := controller.NewExperimentController(
        logger,
        experimentStore,
    )

    // Initialize state machine
    stateMachine := controller.NewStateMachine(logger, experimentController, generatorClient, k8sClient)
    
    // Start the controller's scheduler
    schedulerInterval := time.Duration(getEnvInt("SCHEDULER_INTERVAL_SECONDS", 30)) * time.Second
    scheduler := controller.NewScheduler(logger, experimentController, stateMachine, schedulerInterval)
    go scheduler.Start(context.Background())

    // Create gRPC server
    grpcServer := grpc.NewServer()
    _ = controllergrpc.NewAdapterServer(logger, experimentController)
    
    // TODO: Register the service with the gRPC server once proto files are generated
    // pb.RegisterExperimentServiceServer(grpcServer, adapterServer)
    
    // Start gRPC server
    grpcPort := getEnvInt("GRPC_PORT", 50051)
    grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
    if err != nil {
        logger.Fatal("Failed to listen on gRPC port", zap.Error(err))
    }

    go func() {
        logger.Info("Starting gRPC server", zap.Int("port", grpcPort))
        if err := grpcServer.Serve(grpcListener); err != nil {
            logger.Fatal("Failed to serve gRPC", zap.Error(err))
        }
    }()

    // Create HTTP router for health checks
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"healthy","service":"controller"}`))
    })

    r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
        // Check database connection via experiment store's underlying DB
        if err := experimentStore.DB().PingContext(r.Context()); err != nil {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusServiceUnavailable)
            w.Write([]byte(`{"status":"not ready","service":"controller","error":"database unavailable"}`))
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ready","service":"controller"}`))
    })
    
    // Start HTTP server for health checks
    httpPort := getEnvInt("HTTP_PORT", 8082)
    srv := &http.Server{
        Addr:    fmt.Sprintf(":%d", httpPort),
        Handler: r,
    }
    
    go func() {
        logger.Info("Starting HTTP server", zap.Int("port", httpPort))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("HTTP server failed", zap.Error(err))
        }
    }()
    
    // Wait for interrupt
    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
    <-done
    
    logger.Info("Shutting down controller...")
    
    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Stop gRPC server
    grpcServer.GracefulStop()
    
    // Shutdown HTTP server
    if err := srv.Shutdown(ctx); err != nil {
        logger.Error("HTTP server forced to shutdown", zap.Error(err))
    }
    
    logger.Info("Controller stopped")
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        var intValue int
        fmt.Sscanf(value, "%d", &intValue)
        return intValue
    }
    return defaultValue
}

func getLogLevel(level string) zapcore.Level {
    switch level {
    case "debug":
        return zapcore.DebugLevel
    case "info":
        return zapcore.InfoLevel
    case "warn":
        return zapcore.WarnLevel
    case "error":
        return zapcore.ErrorLevel
    default:
        return zapcore.InfoLevel
    }
}