package simulator

import (
	"context"
	"fmt"
	"math/rand"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/phoenix/platform/pkg/interfaces"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// ProcessSimulator implements the interfaces.LoadSimulator interface
type ProcessSimulator struct {
	mu           sync.RWMutex
	logger       *zap.Logger
	eventBus     interfaces.EventBus
	processes    map[string]*SimulatedProcess
	config       *interfaces.SimulationConfig
	startTime    time.Time
	status       interfaces.SimulationStatus
	results      *interfaces.SimulationResults
	
	// Metrics
	processCount     prometheus.Gauge
	cpuUsage         prometheus.Gauge
	memoryUsage      prometheus.Gauge
	processChurn     prometheus.Counter
	processLifetime  prometheus.Histogram
}

// SimulatedProcess represents a single simulated process
type SimulatedProcess struct {
	Name       string
	PID        int
	CPUPattern string
	MemPattern string
	StartTime  time.Time
	Lifetime   time.Duration
	Priority   string // critical, high, medium, low
	cmd        *exec.Cmd
	
	// Current metrics
	CPUUsage    float64
	MemoryUsage float64
}

// NewProcessSimulator creates a new process simulator
func NewProcessSimulator(logger *zap.Logger, eventBus interfaces.EventBus) interfaces.LoadSimulator {
	return &ProcessSimulator{
		logger:    logger,
		eventBus:  eventBus,
		processes: make(map[string]*SimulatedProcess),
		status:    interfaces.SimulationStatusPending,
		
		// Initialize Prometheus metrics
		processCount: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "phoenix_simulator_process_count",
			Help: "Current number of simulated processes",
		}),
		cpuUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "phoenix_simulator_cpu_usage_percent",
			Help: "Total CPU usage of simulated processes",
		}),
		memoryUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "phoenix_simulator_memory_usage_mb",
			Help: "Total memory usage of simulated processes in MB",
		}),
		processChurn: promauto.NewCounter(prometheus.CounterOpts{
			Name: "phoenix_simulator_process_churn_total",
			Help: "Total number of process restarts",
		}),
		processLifetime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "phoenix_simulator_process_lifetime_seconds",
			Help:    "Lifetime of simulated processes in seconds",
			Buckets: []float64{10, 30, 60, 300, 600, 1800, 3600},
		}),
	}
}

// CreateSimulation creates a new simulation
func (s *ProcessSimulator) CreateSimulation(ctx context.Context, config *interfaces.SimulationConfig) (*interfaces.Simulation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("creating simulation",
		zap.String("name", config.Name),
		zap.String("type", string(config.Type)),
	)

	// Validate configuration
	if err := s.validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	s.config = config
	s.status = interfaces.SimulationStatusPending
	s.startTime = time.Now()

	simulation := &interfaces.Simulation{
		ID:        fmt.Sprintf("sim-%d", time.Now().UnixNano()),
		Name:      config.Name,
		Status:    s.status,
		Config:    config,
		CreatedAt: s.startTime,
		UpdatedAt: s.startTime,
	}

	// Publish event
	s.publishEvent(interfaces.EventTypeSimulationCreated, map[string]interface{}{
		"simulation_id": simulation.ID,
		"name":          simulation.Name,
		"type":          config.Type,
	})

	return simulation, nil
}

// StartSimulation starts the simulation
func (s *ProcessSimulator) StartSimulation(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status != interfaces.SimulationStatusPending {
		return fmt.Errorf("simulation is not in pending state")
	}

	s.logger.Info("starting simulation", zap.String("id", id))
	s.status = interfaces.SimulationStatusRunning
	s.startTime = time.Now()

	// Start simulation in background
	go s.runSimulation(ctx)

	// Publish event
	s.publishEvent(interfaces.EventTypeSimulationStarted, map[string]interface{}{
		"simulation_id": id,
		"start_time":    s.startTime,
	})

	return nil
}

// StopSimulation stops the simulation
func (s *ProcessSimulator) StopSimulation(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status != interfaces.SimulationStatusRunning {
		return fmt.Errorf("simulation is not running")
	}

	s.logger.Info("stopping simulation", zap.String("id", id))
	s.status = interfaces.SimulationStatusStopping

	// Stop all processes
	go s.cleanup()

	// Publish event
	s.publishEvent(interfaces.EventTypeSimulationStopped, map[string]interface{}{
		"simulation_id": id,
		"stop_time":     time.Now(),
	})

	return nil
}

// GetSimulation returns the current simulation
func (s *ProcessSimulator) GetSimulation(ctx context.Context, id string) (*interfaces.Simulation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &interfaces.Simulation{
		ID:        id,
		Name:      s.config.Name,
		Status:    s.status,
		Config:    s.config,
		CreatedAt: s.startTime,
		UpdatedAt: time.Now(),
	}, nil
}

// ListSimulations returns all simulations (only current one for this implementation)
func (s *ProcessSimulator) ListSimulations(ctx context.Context, filter *interfaces.SimulationFilter) ([]*interfaces.Simulation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config == nil {
		return []*interfaces.Simulation{}, nil
	}

	sim := &interfaces.Simulation{
		ID:        fmt.Sprintf("sim-%d", s.startTime.UnixNano()),
		Name:      s.config.Name,
		Status:    s.status,
		Config:    s.config,
		CreatedAt: s.startTime,
		UpdatedAt: time.Now(),
	}

	return []*interfaces.Simulation{sim}, nil
}

// GetSimulationResults returns simulation results
func (s *ProcessSimulator) GetSimulationResults(ctx context.Context, id string) (*interfaces.SimulationResults, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.results == nil {
		// Generate current results
		s.results = s.generateResults()
	}

	return s.results, nil
}

// runSimulation runs the main simulation loop
func (s *ProcessSimulator) runSimulation(ctx context.Context) {
	s.logger.Info("simulation started",
		zap.String("type", string(s.config.Type)),
		zap.Duration("duration", s.config.Duration),
	)

	// Get profile based on simulation type
	profile := s.getProfile()

	// Start initial processes
	if err := s.startInitialProcesses(profile); err != nil {
		s.logger.Error("failed to start initial processes", zap.Error(err))
		s.status = interfaces.SimulationStatusFailed
		return
	}

	// Run simulation loop
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	metricsTicker := time.NewTicker(10 * time.Second)
	defer metricsTicker.Stop()

	churnTicker := time.NewTicker(1 * time.Minute)
	defer churnTicker.Stop()

	timeout := time.After(s.config.Duration)

	for {
		select {
		case <-ticker.C:
			s.updateProcesses()
			s.checkLifetimes(profile)

		case <-metricsTicker.C:
			s.emitMetrics()

		case <-churnTicker.C:
			s.simulateChurn(profile)

		case <-timeout:
			s.logger.Info("simulation duration reached")
			s.completeSimulation()
			return

		case <-ctx.Done():
			s.logger.Info("context cancelled")
			s.cleanup()
			return
		}
	}
}

// getProfile returns the simulation profile based on type
func (s *ProcessSimulator) getProfile() *Profile {
	switch s.config.Type {
	case interfaces.SimulationTypeRealistic:
		return profiles["realistic"]
	case interfaces.SimulationTypeHighCardinality:
		return profiles["high-cardinality"]
	case interfaces.SimulationTypeHighChurn:
		return profiles["process-churn"]
	case interfaces.SimulationTypeChaos:
		// Chaos profile with random failures
		return &Profile{
			Name: "chaos",
			Patterns: []ProcessPattern{
				{NameTemplate: "chaos-process-%d", CPUPattern: "spiky", MemPattern: "spiky", Count: 50, Priority: "low"},
				{NameTemplate: "critical-service-%d", CPUPattern: "steady", MemPattern: "steady", Count: 5, Priority: "critical"},
			},
			ChurnRate:   0.9, // Very high churn
			ChaosConfig: &ChaosConfig{FailureRate: 0.1, CPUSpikeProbability: 0.2},
		}
	default:
		return profiles["realistic"]
	}
}

// startInitialProcesses starts the initial set of processes
func (s *ProcessSimulator) startInitialProcesses(profile *Profile) error {
	processIdx := 0
	targetCount := 100 // Default

	if s.config.Parameters != nil {
		if count, ok := s.config.Parameters["process_count"].(float64); ok {
			targetCount = int(count)
		}
	}

	for _, pattern := range profile.Patterns {
		count := pattern.Count
		if targetCount < 100 && pattern.Count > 10 {
			// Scale down for smaller simulations
			count = pattern.Count * targetCount / 100
			if count < 1 {
				count = 1
			}
		}

		for i := 0; i < count && processIdx < targetCount; i++ {
			proc := s.createProcess(pattern, i)
			if err := s.startProcess(proc); err != nil {
				s.logger.Warn("failed to start process",
					zap.String("name", proc.Name),
					zap.Error(err),
				)
				continue
			}
			processIdx++

			// Stagger process creation
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		}
	}

	s.logger.Info("initial processes started", zap.Int("count", processIdx))
	return nil
}

// emitMetrics emits current metrics
func (s *ProcessSimulator) emitMetrics() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalCPU := 0.0
	totalMem := 0.0
	count := float64(len(s.processes))

	for _, proc := range s.processes {
		totalCPU += proc.CPUUsage
		totalMem += proc.MemoryUsage
	}

	// Update Prometheus metrics
	s.processCount.Set(count)
	s.cpuUsage.Set(totalCPU)
	s.memoryUsage.Set(totalMem)

	// Log summary
	if rand.Float64() < 0.1 { // 10% chance to log
		s.logger.Info("simulation metrics",
			zap.Int("processes", len(s.processes)),
			zap.Float64("total_cpu", totalCPU),
			zap.Float64("total_memory_mb", totalMem),
		)
	}
}

// generateResults generates simulation results
func (s *ProcessSimulator) generateResults() *interfaces.SimulationResults {
	totalProcesses := 0
	totalCPU := 0.0
	totalMem := 0.0
	priorities := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	}

	s.mu.RLock()
	for _, proc := range s.processes {
		totalProcesses++
		totalCPU += proc.CPUUsage
		totalMem += proc.MemoryUsage
		priorities[proc.Priority]++
	}
	s.mu.RUnlock()

	return &interfaces.SimulationResults{
		SimulationID: fmt.Sprintf("sim-%d", s.startTime.UnixNano()),
		StartTime:    s.startTime,
		EndTime:      time.Now(),
		Metrics: map[string]interface{}{
			"total_processes":      totalProcesses,
			"processes_created":    int(s.processChurn.Get()),
			"avg_cpu_usage":        totalCPU / float64(totalProcesses),
			"avg_memory_usage_mb":  totalMem / float64(totalProcesses),
			"process_distribution": priorities,
			"simulation_duration":  time.Since(s.startTime).Seconds(),
		},
		Logs: []string{
			fmt.Sprintf("Simulation completed successfully after %v", time.Since(s.startTime)),
			fmt.Sprintf("Total processes created: %d", int(s.processChurn.Get())),
			fmt.Sprintf("Final process count: %d", totalProcesses),
		},
	}
}

// Additional helper methods...

func (s *ProcessSimulator) validateConfig(config *interfaces.SimulationConfig) error {
	if config.Name == "" {
		return fmt.Errorf("simulation name is required")
	}
	if config.Duration <= 0 {
		return fmt.Errorf("duration must be positive")
	}
	return nil
}

func (s *ProcessSimulator) publishEvent(eventType interfaces.EventType, data map[string]interface{}) {
	event := interfaces.Event{
		ID:        fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		Type:      eventType,
		Source:    "simulator",
		Timestamp: time.Now(),
		Data:      data,
	}

	if err := s.eventBus.Publish(context.Background(), event); err != nil {
		s.logger.Error("failed to publish event", zap.Error(err))
	}
}

func (s *ProcessSimulator) completeSimulation() {
	s.mu.Lock()
	s.status = interfaces.SimulationStatusCompleted
	s.results = s.generateResults()
	s.mu.Unlock()

	s.cleanup()

	s.publishEvent(interfaces.EventTypeSimulationCompleted, map[string]interface{}{
		"simulation_id": fmt.Sprintf("sim-%d", s.startTime.UnixNano()),
		"results":       s.results,
	})
}

// Process management methods remain similar to original implementation...