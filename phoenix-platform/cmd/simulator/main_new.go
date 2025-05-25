package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/phoenix/platform/pkg/eventbus"
	"github.com/phoenix/platform/pkg/interfaces"
	"github.com/phoenix/platform/pkg/simulator"
	"go.uber.org/zap"
)

const (
	defaultControlPort  = 8090
	defaultMetricsPort  = 8888
	serviceName         = "process-simulator"
	serviceVersion      = "0.1.0"
)

// ServiceContainer holds all service dependencies
type ServiceContainer struct {
	Logger       *zap.Logger
	EventBus     interfaces.EventBus
	Simulator    interfaces.LoadSimulator
	MetricsEmitter *simulator.PrometheusMetricsEmitter
	ControlAPI   *simulator.ControlAPI
}

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

	logger.Info("starting process simulator",
		zap.String("service", serviceName),
		zap.String("version", serviceVersion),
	)

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		logger.Fatal("failed to load configuration", zap.Error(err))
	}

	// Initialize service container
	container := initServiceContainer(cfg, logger)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start metrics emitter
	if err := container.MetricsEmitter.Start(ctx); err != nil {
		logger.Fatal("failed to start metrics emitter", zap.Error(err))
	}

	// Start control API
	if err := container.ControlAPI.Start(ctx); err != nil {
		logger.Fatal("failed to start control API", zap.Error(err))
	}

	// Start event listeners
	go startEventListeners(container, logger)

	// Check if we should auto-start a simulation
	if cfg.AutoStart {
		go autoStartSimulation(container, cfg, logger)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down process simulator")
	cancel()

	// Give services time to shutdown gracefully
	time.Sleep(2 * time.Second)

	logger.Info("process simulator stopped")
}

func initServiceContainer(cfg *Config, logger *zap.Logger) *ServiceContainer {
	// Initialize event bus
	eventBus := eventbus.NewMemoryEventBus()

	// Initialize process simulator
	processSimulator := simulator.NewProcessSimulator(logger, eventBus)

	// Initialize metrics emitter
	metricsEmitter := simulator.NewPrometheusMetricsEmitter(logger, cfg.MetricsPort)

	// Initialize control API
	controlAPI := simulator.NewControlAPI(logger, processSimulator, cfg.ControlPort)

	return &ServiceContainer{
		Logger:         logger,
		EventBus:       eventBus,
		Simulator:      processSimulator,
		MetricsEmitter: metricsEmitter,
		ControlAPI:     controlAPI,
	}
}

func initLogger() (*zap.Logger, error) {
	env := os.Getenv("ENVIRONMENT")
	if env == "development" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

// Config holds simulator configuration
type Config struct {
	ControlPort    int
	MetricsPort    int
	AutoStart      bool
	Profile        string
	Duration       string
	ProcessCount   int
	EnableChaos    bool
	TargetHost     string
}

func loadConfig() (*Config, error) {
	cfg := &Config{
		ControlPort:  getEnvInt("CONTROL_PORT", defaultControlPort),
		MetricsPort:  getEnvInt("METRICS_PORT", defaultMetricsPort),
		AutoStart:    getEnvBool("AUTO_START", false),
		Profile:      getEnvOrDefault("PROFILE", "realistic"),
		Duration:     getEnvOrDefault("DURATION", "1h"),
		ProcessCount: getEnvInt("PROCESS_COUNT", 100),
		EnableChaos:  getEnvBool("ENABLE_CHAOS", false),
		TargetHost:   getEnvOrDefault("TARGET_HOST", "localhost"),
	}

	return cfg, nil
}

// autoStartSimulation automatically starts a simulation based on config
func autoStartSimulation(container *ServiceContainer, cfg *Config, logger *zap.Logger) {
	// Wait a moment for services to initialize
	time.Sleep(5 * time.Second)

	logger.Info("auto-starting simulation",
		zap.String("profile", cfg.Profile),
		zap.String("duration", cfg.Duration),
		zap.Int("process_count", cfg.ProcessCount),
	)

	// Parse duration
	duration, err := time.ParseDuration(cfg.Duration)
	if err != nil {
		logger.Error("invalid duration", zap.Error(err))
		return
	}

	// Create simulation config
	simConfig := &interfaces.SimulationConfig{
		Name:     fmt.Sprintf("auto-sim-%d", time.Now().Unix()),
		Type:     getSimulationType(cfg.Profile),
		Duration: duration,
		Parameters: map[string]interface{}{
			"process_count": float64(cfg.ProcessCount),
			"enable_chaos":  cfg.EnableChaos,
			"target_host":   cfg.TargetHost,
		},
	}

	// Create simulation
	simulation, err := container.Simulator.CreateSimulation(context.Background(), simConfig)
	if err != nil {
		logger.Error("failed to create simulation", zap.Error(err))
		return
	}

	logger.Info("simulation created", zap.String("id", simulation.ID))

	// Start simulation
	if err := container.Simulator.StartSimulation(context.Background(), simulation.ID); err != nil {
		logger.Error("failed to start simulation", zap.Error(err))
		return
	}

	logger.Info("simulation started", zap.String("id", simulation.ID))
}

// startEventListeners starts background event listeners
func startEventListeners(container *ServiceContainer, logger *zap.Logger) {
	ctx := context.Background()

	// Subscribe to simulation events
	events, err := container.EventBus.Subscribe(ctx, interfaces.EventFilter{
		Types: []interfaces.EventType{
			interfaces.EventTypeSimulationStarted,
			interfaces.EventTypeSimulationCompleted,
			interfaces.EventTypeSimulationFailed,
		},
	})
	if err != nil {
		logger.Error("failed to subscribe to events", zap.Error(err))
		return
	}

	// Process events
	for event := range events {
		logger.Info("received event",
			zap.String("type", string(event.Type)),
			zap.String("id", event.ID),
		)

		switch event.Type {
		case interfaces.EventTypeSimulationCompleted:
			// Log results if available
			if results, ok := event.Data["results"].(*interfaces.SimulationResults); ok {
				logger.Info("simulation completed",
					zap.Any("metrics", results.Metrics),
				)
			}
		}
	}
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

func getSimulationType(profile string) interfaces.SimulationType {
	switch profile {
	case "realistic":
		return interfaces.SimulationTypeRealistic
	case "high-cardinality":
		return interfaces.SimulationTypeHighCardinality
	case "process-churn", "high-churn":
		return interfaces.SimulationTypeHighChurn
	case "chaos":
		return interfaces.SimulationTypeChaos
	default:
		return interfaces.SimulationTypeRealistic
	}
}