package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/phoenix-vnext/platform/projects/loadsim-operator/internal/generator"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Get configuration from environment
	experimentID := os.Getenv("EXPERIMENT_ID")
	profile := os.Getenv("PROFILE")
	durationStr := os.Getenv("DURATION")
	processCountStr := os.Getenv("PROCESS_COUNT")

	if experimentID == "" || profile == "" || durationStr == "" {
		logger.Fatal("missing required environment variables",
			zap.String("EXPERIMENT_ID", experimentID),
			zap.String("PROFILE", profile),
			zap.String("DURATION", durationStr))
	}

	// Parse duration
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		logger.Fatal("invalid duration", zap.String("duration", durationStr), zap.Error(err))
	}

	// Parse process count
	processCount := 100 // default
	if processCountStr != "" {
		count, err := strconv.Atoi(processCountStr)
		if err != nil {
			logger.Warn("invalid process count, using default", 
				zap.String("processCount", processCountStr), 
				zap.Error(err))
		} else {
			processCount = count
		}
	}

	logger.Info("starting load simulator",
		zap.String("experimentID", experimentID),
		zap.String("profile", profile),
		zap.Duration("duration", duration),
		zap.Int("processCount", processCount))

	// Create load generator
	lg := generator.NewLoadGenerator(logger, profile, processCount, duration)

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start load generation in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- lg.Start()
	}()

	// Wait for completion or signal
	select {
	case err := <-errChan:
		if err != nil {
			logger.Error("load generator failed", zap.Error(err))
			os.Exit(1)
		}
		logger.Info("load generation completed successfully")
	case sig := <-sigChan:
		logger.Info("received signal, shutting down", zap.String("signal", sig.String()))
		if err := lg.Stop(); err != nil {
			logger.Error("failed to stop load generator", zap.Error(err))
			os.Exit(1)
		}
	}

	logger.Info("load simulator exited")
}