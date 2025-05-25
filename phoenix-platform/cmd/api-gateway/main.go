package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/phoenix/platform/cmd/api-gateway/internal/handlers"
	"github.com/phoenix/platform/cmd/api-gateway/internal/middleware"
	"github.com/phoenix/platform/pkg/auth"
	"github.com/phoenix/platform/pkg/clients"
)

const (
	defaultHTTPPort   = ":8080"
	shutdownTimeout   = 30 * time.Second
	serviceName       = "api-gateway"
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

	logger.Info("starting API gateway",
		zap.String("service", serviceName),
		zap.String("version", serviceVersion),
	)

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		logger.Fatal("failed to load configuration", zap.Error(err))
	}

	// Initialize JWT Manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, 24*time.Hour)

	// Initialize Phoenix clients
	phoenixClient, err := initPhoenixClient(cfg, logger)
	if err != nil {
		logger.Fatal("failed to initialize Phoenix client", zap.Error(err))
	}
	defer phoenixClient.Close()

	// Initialize Gin router
	router := setupRouter(cfg, logger, jwtManager, phoenixClient)

	// Start HTTP server
	srv := &http.Server{
		Addr:         cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting HTTP server", zap.String("port", cfg.HTTPPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to serve HTTP", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down API gateway")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown HTTP server", zap.Error(err))
	}

	logger.Info("API gateway stopped")
}

func setupRouter(cfg *Config, logger *zap.Logger, jwtManager *auth.JWTManager, phoenixClient *clients.PhoenixClient) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogging(logger))
	router.Use(middleware.CORS())

	// Health check endpoints (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// Metrics endpoint (no auth required)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(jwtManager, logger)
	experimentHandler := handlers.NewExperimentHandler(logger, phoenixClient.Experiment)
	generatorHandler := handlers.NewGeneratorHandler(logger, phoenixClient.Generator)
	controlHandler := handlers.NewControlHandler(logger, phoenixClient.Controller)

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Auth routes (no auth required)
	authRoutes := v1.Group("/auth")
	{
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/refresh", authHandler.Refresh)
	}

	// Protected routes
	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware(middleware.AuthConfig{
		JWTManager: jwtManager,
		SkipPaths:  []string{"/api/v1/auth/"},
		RequiredRoles: map[string][]string{
			// Admin endpoints
			"/api/v1/experiments": {"admin", "user"},
			"/api/v1/templates":   {"admin"},
			"/api/v1/control":     {"admin"},
		},
		Logger: logger,
	}))

	// Auth endpoints (require authentication)
	protected.GET("/auth/me", authHandler.Me)
	protected.POST("/auth/logout", authHandler.Logout)

	// Experiment endpoints
	experiments := protected.Group("/experiments")
	{
		experiments.GET("", experimentHandler.ListExperiments)
		experiments.POST("", experimentHandler.CreateExperiment)
		experiments.GET("/:id", experimentHandler.GetExperiment)
		experiments.PUT("/:id", experimentHandler.UpdateExperiment)
		experiments.DELETE("/:id", experimentHandler.DeleteExperiment)
		experiments.GET("/:id/status", experimentHandler.GetExperimentStatus)
		experiments.POST("/:id/start", middleware.RequireRoles("admin"), experimentHandler.StartExperiment)
		experiments.POST("/:id/stop", middleware.RequireRoles("admin"), experimentHandler.StopExperiment)
	}

	// Generator endpoints
	generator := protected.Group("/generator")
	{
		generator.POST("/generate", generatorHandler.GenerateConfig)
		generator.POST("/validate", generatorHandler.ValidateConfig)
	}

	// Template endpoints
	templates := protected.Group("/templates")
	templates.Use(middleware.RequireRoles("admin"))
	{
		templates.GET("", generatorHandler.ListTemplates)
		templates.POST("", generatorHandler.CreateTemplate)
		templates.GET("/:name", generatorHandler.GetTemplate)
		templates.PUT("/:name", generatorHandler.UpdateTemplate)
		templates.DELETE("/:name", generatorHandler.DeleteTemplate)
	}

	// Control endpoints
	control := protected.Group("/control")
	control.Use(middleware.RequireRoles("admin"))
	{
		control.POST("/signals", controlHandler.ApplyControlSignal)
		control.GET("/signals/:id", controlHandler.GetControlSignal)
		control.GET("/experiments/:id/signals", controlHandler.ListControlSignals)
		control.GET("/experiments/:id/drift", controlHandler.GetDriftReport)
	}

	return router
}

func initLogger() (*zap.Logger, error) {
	env := os.Getenv("ENVIRONMENT")
	if env == "development" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

type Config struct {
	HTTPPort       string
	ExperimentAddr string
	GeneratorAddr  string
	ControllerAddr string
	Environment    string
	JWTSecret      string
}

func loadConfig() (*Config, error) {
	cfg := &Config{
		HTTPPort:       getEnvOrDefault("HTTP_PORT", defaultHTTPPort),
		ExperimentAddr: getEnvOrDefault("EXPERIMENT_ADDR", "localhost:50051"),
		GeneratorAddr:  getEnvOrDefault("GENERATOR_ADDR", "localhost:50052"),
		ControllerAddr: getEnvOrDefault("CONTROLLER_ADDR", "localhost:50053"),
		Environment:    getEnvOrDefault("ENVIRONMENT", "production"),
		JWTSecret:      getEnvOrDefault("JWT_SECRET", "phoenix-secret-key-change-in-production"),
	}

	// Validate JWT secret in production
	if cfg.Environment == "production" && cfg.JWTSecret == "phoenix-secret-key-change-in-production" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initPhoenixClient(cfg *Config, logger *zap.Logger) (*clients.PhoenixClient, error) {
	clientCfg := &clients.Config{
		ExperimentAddr: cfg.ExperimentAddr,
		GeneratorAddr:  cfg.GeneratorAddr,
		ControllerAddr: cfg.ControllerAddr,
		TLSEnabled:     false,
	}

	logger.Info("connecting to Phoenix services",
		zap.String("experiment", cfg.ExperimentAddr),
		zap.String("generator", cfg.GeneratorAddr),
		zap.String("controller", cfg.ControllerAddr),
	)

	return clients.NewPhoenixClient(clientCfg)
}