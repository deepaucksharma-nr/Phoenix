package supervisor

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/phoenix/platform/projects/phoenix-agent/internal/config"
	"github.com/phoenix/platform/projects/phoenix-agent/internal/poller"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type Supervisor struct {
	config           *config.Config
	collectorManager *CollectorManager
	loadSimManager   *LoadSimManager
	activeTasks      sync.Map
	mu               sync.RWMutex
}

func NewSupervisor(cfg *config.Config) *Supervisor {
	return &Supervisor{
		config:           cfg,
		collectorManager: NewCollectorManager(cfg),
		loadSimManager:   NewLoadSimManager(),
	}
}

// ExecuteTask executes a task based on its type
func (s *Supervisor) ExecuteTask(ctx context.Context, task *poller.Task) (map[string]interface{}, error) {
	// Track active task
	s.activeTasks.Store(task.ID, task)
	defer s.activeTasks.Delete(task.ID)

	switch task.Type {
	case "collector":
		return s.executeCollectorTask(ctx, task)
	case "deployment":
		return s.executePipelineDeploymentTask(ctx, task)
	case "loadsim":
		return s.executeLoadSimTask(ctx, task)
	case "command":
		return s.executeCommandTask(ctx, task)
	default:
		return nil, fmt.Errorf("unknown task type: %s", task.Type)
	}
}

func (s *Supervisor) executeCollectorTask(ctx context.Context, task *poller.Task) (map[string]interface{}, error) {
	// Pass context to the collector operations in the future
	// For now, just acknowledge the parameter to suppress warning
	_ = ctx

	config := task.Config

	id, ok := config["id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing collector id in config")
	}

	variant, ok := config["variant"].(string)
	if !ok {
		return nil, fmt.Errorf("missing variant in config")
	}

	switch task.Action {
	case "start":
		configURL, ok := config["configUrl"].(string)
		if !ok {
			return nil, fmt.Errorf("missing configUrl in config")
		}

		vars, _ := config["vars"].(map[string]string)

		if err := s.collectorManager.Start(id, variant, configURL, vars); err != nil {
			return nil, fmt.Errorf("failed to start collector: %w", err)
		}

		return map[string]interface{}{
			"status": "started",
			"pid":    s.collectorManager.GetProcessInfo(id),
		}, nil

	case "stop":
		if err := s.collectorManager.Stop(id); err != nil {
			return nil, fmt.Errorf("failed to stop collector: %w", err)
		}

		return map[string]interface{}{
			"status": "stopped",
		}, nil

	case "update":
		// Stop and restart with new config
		s.collectorManager.Stop(id)

		configURL, ok := config["configUrl"].(string)
		if !ok {
			return nil, fmt.Errorf("missing configUrl in config")
		}

		vars, _ := config["vars"].(map[string]string)

		if err := s.collectorManager.Start(id, variant, configURL, vars); err != nil {
			return nil, fmt.Errorf("failed to update collector: %w", err)
		}

		return map[string]interface{}{
			"status": "updated",
			"pid":    s.collectorManager.GetProcessInfo(id),
		}, nil

	default:
		return nil, fmt.Errorf("unknown collector action: %s", task.Action)
	}
}

func (s *Supervisor) executeLoadSimTask(ctx context.Context, task *poller.Task) (map[string]interface{}, error) {
	// Pass context to load sim operations in the future
	// For now, just acknowledge the parameter to suppress warning
	_ = ctx

	config := task.Config

	switch task.Action {
	case "start":
		profile, ok := config["profile"].(string)
		if !ok {
			return nil, fmt.Errorf("missing profile in config")
		}

		durationStr, ok := config["duration"].(string)
		if !ok {
			durationStr = "60s"
		}

		if err := s.loadSimManager.Start(profile, durationStr); err != nil {
			return nil, fmt.Errorf("failed to start load simulation: %w", err)
		}

		return map[string]interface{}{
			"status":  "started",
			"profile": profile,
		}, nil

	case "stop":
		if err := s.loadSimManager.Stop(); err != nil {
			return nil, fmt.Errorf("failed to stop load simulation: %w", err)
		}

		return map[string]interface{}{
			"status": "stopped",
		}, nil

	default:
		return nil, fmt.Errorf("unknown loadsim action: %s", task.Action)
	}
}

func (s *Supervisor) executePipelineDeploymentTask(ctx context.Context, task *poller.Task) (map[string]interface{}, error) {
	// Pass context to pipeline operations in the future
	// For now, just acknowledge the parameter to suppress warning
	_ = ctx

	config := task.Config

	deploymentID, ok := config["deployment_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing deployment_id in config")
	}

	deploymentName, ok := config["deployment_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing deployment_name in config")
	}

	pipelineConfig, ok := config["pipeline_config"].(string)
	if !ok || pipelineConfig == "" {
		return nil, fmt.Errorf("missing or empty pipeline_config in config")
	}

	switch task.Action {
	case "deploy":
		// Create a unique ID for this collector instance
		collectorID := fmt.Sprintf("dep-%s-%s", deploymentID, s.config.HostID)

		// Write pipeline config to a temporary file
		configPath := fmt.Sprintf("/tmp/pipeline-%s.yaml", collectorID)
		if err := os.WriteFile(configPath, []byte(pipelineConfig), 0644); err != nil {
			return nil, fmt.Errorf("failed to write pipeline config: %w", err)
		}

		// Use collector manager to deploy the pipeline
		vars := make(map[string]string)
		if params, ok := config["parameters"].(map[string]interface{}); ok {
			for k, v := range params {
				vars[k] = fmt.Sprintf("%v", v)
			}
		}

		// Add pushgateway URL if provided
		if pushgatewayURL, ok := config["pushgateway_url"].(string); ok {
			vars["METRICS_PUSHGATEWAY_URL"] = pushgatewayURL
		}

		// Start collector with the pipeline config
		if err := s.collectorManager.Start(collectorID, deploymentName, "file://"+configPath, vars); err != nil {
			os.Remove(configPath) // Clean up temp file
			return nil, fmt.Errorf("failed to deploy pipeline: %w", err)
		}

		// Clean up temp file after a delay (give collector time to read it)
		go func() {
			time.Sleep(5 * time.Second)
			os.Remove(configPath)
		}()

		return map[string]interface{}{
			"status":        "deployed",
			"deployment_id": deploymentID,
			"collector_id":  collectorID,
			"pid":           s.collectorManager.GetProcessInfo(collectorID),
		}, nil

	case "undeploy":
		collectorID := fmt.Sprintf("dep-%s-%s", deploymentID, s.config.HostID)

		if err := s.collectorManager.Stop(collectorID); err != nil {
			return nil, fmt.Errorf("failed to undeploy pipeline: %w", err)
		}

		return map[string]interface{}{
			"status":        "undeployed",
			"deployment_id": deploymentID,
		}, nil

	case "update":
		// Stop and redeploy with new config
		collectorID := fmt.Sprintf("dep-%s-%s", deploymentID, s.config.HostID)
		s.collectorManager.Stop(collectorID)

		// Write new pipeline config
		configPath := fmt.Sprintf("/tmp/pipeline-%s.yaml", collectorID)
		if err := os.WriteFile(configPath, []byte(pipelineConfig), 0644); err != nil {
			return nil, fmt.Errorf("failed to write pipeline config: %w", err)
		}

		// Restart with new config
		vars := make(map[string]string)
		if params, ok := config["parameters"].(map[string]interface{}); ok {
			for k, v := range params {
				vars[k] = fmt.Sprintf("%v", v)
			}
		}

		// Add pushgateway URL if provided
		if pushgatewayURL, ok := config["pushgateway_url"].(string); ok {
			vars["METRICS_PUSHGATEWAY_URL"] = pushgatewayURL
		}

		if err := s.collectorManager.Start(collectorID, deploymentName, "file://"+configPath, vars); err != nil {
			os.Remove(configPath)
			return nil, fmt.Errorf("failed to update pipeline: %w", err)
		}

		// Clean up temp file after a delay
		go func() {
			time.Sleep(5 * time.Second)
			os.Remove(configPath)
		}()

		return map[string]interface{}{
			"status":        "updated",
			"deployment_id": deploymentID,
			"collector_id":  collectorID,
			"pid":           s.collectorManager.GetProcessInfo(collectorID),
		}, nil

	default:
		return nil, fmt.Errorf("unknown deployment action: %s", task.Action)
	}
}

func (s *Supervisor) executeCommandTask(ctx context.Context, task *poller.Task) (map[string]interface{}, error) {
	// Execute arbitrary commands (for future extensibility)
	// Using context and task parameters to avoid unused parameter warnings
	_ = ctx
	_ = task

	return map[string]interface{}{
		"status": "not_implemented",
	}, fmt.Errorf("command tasks not yet implemented")
}

// GetStatus returns the current agent status
func (s *Supervisor) GetStatus() *poller.AgentStatus {
	// Get active tasks
	var activeTasks []string
	s.activeTasks.Range(func(key, value interface{}) bool {
		activeTasks = append(activeTasks, key.(string))
		return true
	})

	// Get resource usage
	cpuPercent, _ := cpu.Percent(0, false)
	memInfo, _ := mem.VirtualMemory()

	cpuUsage := 0.0
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	return &poller.AgentStatus{
		Status:      "healthy",
		ActiveTasks: activeTasks,
		ResourceUsage: poller.ResourceUsage{
			CPUPercent:    cpuUsage,
			MemoryPercent: memInfo.UsedPercent,
			MemoryBytes:   int64(memInfo.Used),
		},
	}
}

// StopAll stops all managed processes
func (s *Supervisor) StopAll() {
	log.Info().Msg("Stopping all supervised processes")

	// Stop all collectors
	s.collectorManager.StopAll()

	// Stop load simulation
	s.loadSimManager.Stop()
}

// Shutdown gracefully shuts down the supervisor and all managed processes
func (s *Supervisor) Shutdown(ctx context.Context) error {
	log.Info().Msg("Shutting down supervisor")
	
	// Create error channel to collect shutdown errors
	errChan := make(chan error, 2)
	
	// Shutdown collectors
	go func() {
		s.collectorManager.StopAll()
		errChan <- nil
	}()
	
	// Shutdown load simulation manager
	go func() {
		errChan <- s.loadSimManager.Shutdown(ctx)
	}()
	
	// Wait for both shutdowns or timeout
	var shutdownErr error
	for i := 0; i < 2; i++ {
		select {
		case err := <-errChan:
			if err != nil && shutdownErr == nil {
				shutdownErr = err
			}
		case <-ctx.Done():
			return fmt.Errorf("shutdown timed out: %w", ctx.Err())
		}
	}
	
	if shutdownErr != nil {
		return fmt.Errorf("shutdown error: %w", shutdownErr)
	}
	
	log.Info().Msg("Supervisor shutdown complete")
	return nil
}

// GetMetrics returns metrics from all managed processes
func (s *Supervisor) GetMetrics() []map[string]interface{} {
	var metrics []map[string]interface{}

	// Get collector metrics
	collectorMetrics := s.collectorManager.GetMetrics()
	metrics = append(metrics, collectorMetrics...)

	// Get load sim metrics if running
	if loadSimMetrics := s.loadSimManager.GetMetrics(); loadSimMetrics != nil {
		metrics = append(metrics, loadSimMetrics)
	}

	return metrics
}
