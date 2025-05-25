package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("starting loadsim operator",
		zap.String("version", "0.1.0"),
		zap.String("status", "not-implemented"),
	)

	// TODO: Implement loadsim operator
	// 1. Set up controller-runtime manager
	// 2. Register LoadSimulationJob controller
	// 3. Create/manage simulator Jobs
	// 4. Clean up completed jobs

	logger.Warn("loadsim operator not yet implemented")
	os.Exit(0)
}