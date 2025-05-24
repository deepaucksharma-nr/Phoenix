package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("starting experiment controller",
		zap.String("version", "0.1.0"),
		zap.String("status", "not-implemented"),
	)

	// TODO: Implement experiment controller
	// 1. Set up Kubernetes client
	// 2. Watch PhoenixExperiment CRDs
	// 3. Reconcile experiments
	// 4. Update experiment status

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down experiment controller")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30)
	defer cancel()

	// TODO: Implement graceful shutdown
	_ = ctx

	logger.Info("experiment controller stopped")
}