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
	"google.golang.org/grpc/reflection"

	generatorv1 "github.com/phoenix/platform/pkg/grpc/proto/v1"
	"github.com/phoenix/platform/projects/generator/internal/config"
	generatorgrpc "github.com/phoenix/platform/projects/generator/internal/grpc"
)

// simpleConfigManager implements the ConfigManager interface with basic functionality
type simpleConfigManager struct {
	templates map[string]*config.Template
}

func newSimpleConfigManager() *simpleConfigManager {
	return &simpleConfigManager{
		templates: make(map[string]*config.Template),
	}
}

func (m *simpleConfigManager) GenerateConfig(ctx context.Context, req config.GenerateRequest) (*config.GeneratedConfig, error) {
	// Simple implementation - just return a basic config
	return &config.GeneratedConfig{
		ID:           fmt.Sprintf("config-%s-%d", req.ExperimentID, time.Now().Unix()),
		ExperimentID: req.ExperimentID,
		Content:      fmt.Sprintf("# Generated config for experiment %s\n# Template: %s\n", req.ExperimentID, req.Template),
		Version:      "v1",
	}, nil
}

func (m *simpleConfigManager) ValidateConfig(ctx context.Context, cfg string) error {
	if cfg == "" {
		return fmt.Errorf("configuration cannot be empty")
	}
	return nil
}

func (m *simpleConfigManager) GetTemplate(ctx context.Context, name string) (*config.Template, error) {
	if tmpl, ok := m.templates[name]; ok {
		return tmpl, nil
	}
	return nil, fmt.Errorf("template %s not found", name)
}

func (m *simpleConfigManager) ListTemplates(ctx context.Context) ([]*config.Template, error) {
	var templates []*config.Template
	for _, tmpl := range m.templates {
		templates = append(templates, tmpl)
	}
	return templates, nil
}

func (m *simpleConfigManager) CreateTemplate(ctx context.Context, tmpl *config.Template) error {
	m.templates[tmpl.Name] = tmpl
	return nil
}

func (m *simpleConfigManager) UpdateTemplate(ctx context.Context, name string, tmpl *config.Template) error {
	if _, ok := m.templates[name]; !ok {
		return fmt.Errorf("template %s not found", name)
	}
	m.templates[name] = tmpl
	return nil
}

func (m *simpleConfigManager) DeleteTemplate(ctx context.Context, name string) error {
	delete(m.templates, name)
	return nil
}

func main() {
	// Initialize zap logger
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

	logger.Info("Starting Phoenix Configuration Generator...")

	// Initialize config manager
	configManager := newSimpleConfigManager()

	// Create default templates
	configManager.CreateTemplate(context.Background(), &config.Template{
		Name:        "otel-collector",
		Description: "OpenTelemetry Collector configuration template",
		Content:     "# OTEL Collector Configuration\n",
		Version:     "v1",
	})

	// Create gRPC server
	grpcServer := grpc.NewServer()
	generatorServer := generatorgrpc.NewGeneratorServer(configManager)
	generatorv1.RegisterGeneratorServiceServer(grpcServer, generatorServer)
	
	// Register reflection service for gRPC debugging
	reflection.Register(grpcServer)
	
	// Start gRPC server
	grpcPort := getEnvInt("GRPC_PORT", 50052)
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
		w.Write([]byte(`{"status":"healthy","service":"generator"}`))
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready","service":"generator"}`))
	})
	
	// Start HTTP server for health checks
	httpPort := getEnvInt("HTTP_PORT", 8081)
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
	
	logger.Info("Shutting down generator...")
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Stop gRPC server
	grpcServer.GracefulStop()
	
	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("HTTP server forced to shutdown", zap.Error(err))
	}
	
	logger.Info("Generator stopped")
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