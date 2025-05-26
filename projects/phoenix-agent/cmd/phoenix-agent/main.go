package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/phoenix/platform/projects/phoenix-agent/internal/config"
	"github.com/phoenix/platform/projects/phoenix-agent/internal/metrics"
	"github.com/phoenix/platform/projects/phoenix-agent/internal/poller"
	"github.com/phoenix/platform/projects/phoenix-agent/internal/supervisor"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Command line flags
	var (
		apiURL       = flag.String("api-url", getEnv("PHOENIX_API_URL", "http://phoenix-api:8080"), "Phoenix API URL")
		hostID       = flag.String("host-id", getHostID(), "Unique host identifier")
		pollInterval = flag.Duration("poll-interval", getDurationEnv("POLL_INTERVAL", 15*time.Second), "Task poll interval")
		configDir    = flag.String("config-dir", getEnv("CONFIG_DIR", "/etc/phoenix-agent"), "Directory for agent configs")
		logLevel     = flag.String("log-level", getEnv("LOG_LEVEL", "info"), "Log level (debug, info, warn, error)")
	)
	flag.Parse()

	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	level, err := zerolog.ParseLevel(*logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	
	if getEnv("LOG_FORMAT", "json") == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	log.Info().
		Str("api_url", *apiURL).
		Str("host_id", *hostID).
		Dur("poll_interval", *pollInterval).
		Msg("Starting Phoenix Agent")

	// Initialize configuration
	cfg := &config.Config{
		APIURL:       *apiURL,
		HostID:       *hostID,
		PollInterval: *pollInterval,
		ConfigDir:    *configDir,
	}

	// Initialize components
	apiClient := poller.NewClient(cfg)
	taskSupervisor := supervisor.NewSupervisor(cfg)
	metricsReporter := metrics.NewReporter(cfg, apiClient)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start metrics reporting
	go metricsReporter.Start(ctx)

	// Main polling loop
	go func() {
		ticker := time.NewTicker(cfg.PollInterval)
		defer ticker.Stop()

		// Initial poll immediately
		pollAndExecuteTasks(ctx, apiClient, taskSupervisor)

		for {
			select {
			case <-ticker.C:
				pollAndExecuteTasks(ctx, apiClient, taskSupervisor)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info().Msg("Shutting down agent...")

	// Stop all supervised processes
	taskSupervisor.StopAll()

	// Give processes time to shut down gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	select {
	case <-shutdownCtx.Done():
		log.Warn().Msg("Shutdown timeout exceeded")
	case <-time.After(5 * time.Second):
		log.Info().Msg("Graceful shutdown completed")
	}
}

func pollAndExecuteTasks(ctx context.Context, client *poller.Client, supervisor *supervisor.Supervisor) {
	// Send heartbeat
	if err := client.SendHeartbeat(ctx, supervisor.GetStatus()); err != nil {
		log.Error().Err(err).Msg("Failed to send heartbeat")
	}

	// Get pending tasks
	tasks, err := client.GetTasks(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tasks")
		return
	}

	if len(tasks) == 0 {
		log.Debug().Msg("No pending tasks")
		return
	}

	log.Info().Int("count", len(tasks)).Msg("Received tasks")

	// Execute tasks
	for _, task := range tasks {
		log.Info().
			Str("task_id", task.ID).
			Str("type", task.Type).
			Str("action", task.Action).
			Msg("Executing task")

		// Update task status to running
		if err := client.UpdateTaskStatus(ctx, task.ID, "running", nil, ""); err != nil {
			log.Error().Err(err).Str("task_id", task.ID).Msg("Failed to update task status")
		}

		// Execute task
		result, err := supervisor.ExecuteTask(ctx, task)
		if err != nil {
			log.Error().Err(err).Str("task_id", task.ID).Msg("Task execution failed")
			client.UpdateTaskStatus(ctx, task.ID, "failed", nil, err.Error())
			continue
		}

		// Update task status to completed
		if err := client.UpdateTaskStatus(ctx, task.ID, "completed", result, ""); err != nil {
			log.Error().Err(err).Str("task_id", task.ID).Msg("Failed to update task status")
		}
	}
}

func getHostID() string {
	// Try to get from environment
	if hostID := os.Getenv("PHOENIX_HOST_ID"); hostID != "" {
		return hostID
	}

	// Try to get hostname
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}

	// Fallback to a generated ID
	return fmt.Sprintf("agent-%d", time.Now().Unix())
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}