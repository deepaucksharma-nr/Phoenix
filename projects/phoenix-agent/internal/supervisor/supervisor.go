package supervisor

import (
	"context"
	"fmt"
	"sync"

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
	case "loadsim":
		return s.executeLoadSimTask(ctx, task)
	case "command":
		return s.executeCommandTask(ctx, task)
	default:
		return nil, fmt.Errorf("unknown task type: %s", task.Type)
	}
}

func (s *Supervisor) executeCollectorTask(ctx context.Context, task *poller.Task) (map[string]interface{}, error) {
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

func (s *Supervisor) executeCommandTask(ctx context.Context, task *poller.Task) (map[string]interface{}, error) {
	// Execute arbitrary commands (for future extensibility)
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