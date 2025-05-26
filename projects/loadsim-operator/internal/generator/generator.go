package generator

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// LoadGenerator manages process generation for load simulation
type LoadGenerator struct {
	logger       *zap.Logger
	profile      string
	processCount int
	duration     time.Duration
	churnRate    float64
	processes    map[string]*Process
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// Process represents a simulated process
type Process struct {
	Name      string
	PID       int
	Cmd       *exec.Cmd
	StartTime time.Time
	CPULoad   float64
	MemLoad   float64
}

// NewLoadGenerator creates a new load generator
func NewLoadGenerator(logger *zap.Logger, profile string, processCount int, duration time.Duration) *LoadGenerator {
	ctx, cancel := context.WithCancel(context.Background())
	return &LoadGenerator{
		logger:       logger,
		profile:      profile,
		processCount: processCount,
		duration:     duration,
		processes:    make(map[string]*Process),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start begins the load generation
func (lg *LoadGenerator) Start() error {
	lg.logger.Info("starting load generator",
		zap.String("profile", lg.profile),
		zap.Int("processCount", lg.processCount),
		zap.Duration("duration", lg.duration))

	switch lg.profile {
	case "realistic":
		return lg.runRealisticProfile()
	case "high-cardinality":
		return lg.runHighCardinalityProfile()
	case "process-churn":
		return lg.runProcessChurnProfile()
	case "custom":
		return lg.runCustomProfile()
	default:
		return fmt.Errorf("unknown profile: %s", lg.profile)
	}
}

// Stop gracefully stops all processes
func (lg *LoadGenerator) Stop() error {
	lg.logger.Info("stopping load generator")
	lg.cancel()
	lg.wg.Wait()

	lg.mu.Lock()
	defer lg.mu.Unlock()

	for name, proc := range lg.processes {
		if proc.Cmd != nil && proc.Cmd.Process != nil {
			lg.logger.Debug("killing process", zap.String("name", name))
			proc.Cmd.Process.Kill()
		}
	}

	return nil
}

// runRealisticProfile simulates a mix of long-running and short-lived processes
func (lg *LoadGenerator) runRealisticProfile() error {
	// 70% long-running processes, 30% short-lived
	longRunning := int(float64(lg.processCount) * 0.7)
	shortLived := lg.processCount - longRunning

	// Start long-running processes
	for i := 0; i < longRunning; i++ {
		name := fmt.Sprintf("webapp-%d", i)
		lg.wg.Add(1)
		go lg.createProcess(name, lg.duration, "steady", "steady")
	}

	// Start short-lived process spawner
	lg.wg.Add(1)
	go func() {
		defer lg.wg.Done()
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		counter := 0
		for {
			select {
			case <-lg.ctx.Done():
				return
			case <-ticker.C:
				for j := 0; j < shortLived/6; j++ {
					name := fmt.Sprintf("job-%d-%d", counter, j)
					lg.wg.Add(1)
					go lg.createProcess(name, 30*time.Second, "spiky", "growing")
					counter++
				}
			}
		}
	}()

	// Wait for duration
	time.Sleep(lg.duration)
	lg.Stop()
	return nil
}

// runHighCardinalityProfile creates many unique process names
func (lg *LoadGenerator) runHighCardinalityProfile() error {
	services := []string{"api", "worker", "cache", "db", "queue", "stream", "batch", "cron"}
	environments := []string{"prod", "staging", "dev", "test"}
	regions := []string{"us-east", "us-west", "eu-west", "ap-south"}
	
	processesPerCombination := lg.processCount / (len(services) * len(environments) * len(regions))
	if processesPerCombination < 1 {
		processesPerCombination = 1
	}

	for _, service := range services {
		for _, env := range environments {
			for _, region := range regions {
				for i := 0; i < processesPerCombination; i++ {
					name := fmt.Sprintf("%s-%s-%s-%d", service, env, region, i)
					lg.wg.Add(1)
					go lg.createProcess(name, lg.duration, "random", "random")
				}
			}
		}
	}

	// Wait for duration
	time.Sleep(lg.duration)
	lg.Stop()
	return nil
}

// runProcessChurnProfile rapidly creates and destroys processes
func (lg *LoadGenerator) runProcessChurnProfile() error {
	lg.churnRate = 0.8 // 80% churn rate

	// Create initial batch
	for i := 0; i < lg.processCount; i++ {
		name := fmt.Sprintf("churn-proc-%d", i)
		lg.wg.Add(1)
		go lg.createProcess(name, 5*time.Second, "spiky", "spiky")
	}

	// Churn loop
	lg.wg.Add(1)
	go func() {
		defer lg.wg.Done()
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		counter := lg.processCount
		for {
			select {
			case <-lg.ctx.Done():
				return
			case <-ticker.C:
				// Kill some processes
				lg.mu.RLock()
				toKill := int(float64(len(lg.processes)) * lg.churnRate)
				killed := 0
				for name, proc := range lg.processes {
					if killed >= toKill {
						break
					}
					if proc.Cmd != nil && proc.Cmd.Process != nil {
						proc.Cmd.Process.Kill()
						delete(lg.processes, name)
						killed++
					}
				}
				lg.mu.RUnlock()

				// Create new processes
				for i := 0; i < killed; i++ {
					name := fmt.Sprintf("churn-proc-%d", counter)
					counter++
					lg.wg.Add(1)
					go lg.createProcess(name, 5*time.Second, "spiky", "spiky")
				}
			}
		}
	}()

	// Wait for duration
	time.Sleep(lg.duration)
	lg.Stop()
	return nil
}

// runCustomProfile runs user-defined patterns
func (lg *LoadGenerator) runCustomProfile() error {
	// For now, use realistic profile as fallback
	// TODO: Implement custom profile parsing from environment
	return lg.runRealisticProfile()
}

// createProcess creates a single simulated process
func (lg *LoadGenerator) createProcess(name string, lifetime time.Duration, cpuPattern, memPattern string) {
	defer lg.wg.Done()

	// Create a simple stress command that uses CPU and memory
	cmd := exec.CommandContext(lg.ctx, "sh", "-c", fmt.Sprintf(`
		# Process: %s
		while true; do
			# CPU load simulation
			dd if=/dev/zero of=/dev/null bs=1M count=100 2>/dev/null
			# Memory allocation simulation
			head -c %dM /dev/zero | tail
			sleep 1
		done
	`, name, rand.Intn(50)+10))

	proc := &Process{
		Name:      name,
		Cmd:       cmd,
		StartTime: time.Now(),
		CPULoad:   lg.getCPULoad(cpuPattern),
		MemLoad:   lg.getMemLoad(memPattern),
	}

	// Set process group to make cleanup easier
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	lg.mu.Lock()
	lg.processes[name] = proc
	lg.mu.Unlock()

	// Start the process
	if err := cmd.Start(); err != nil {
		lg.logger.Error("failed to start process", zap.String("name", name), zap.Error(err))
		return
	}

	proc.PID = cmd.Process.Pid
	lg.logger.Debug("started process", zap.String("name", name), zap.Int("pid", proc.PID))

	// Wait for lifetime or context cancellation
	timer := time.NewTimer(lifetime)
	defer timer.Stop()

	select {
	case <-timer.C:
		// Lifetime expired
	case <-lg.ctx.Done():
		// Context cancelled
	}

	// Clean up
	if cmd.Process != nil {
		cmd.Process.Kill()
	}

	lg.mu.Lock()
	delete(lg.processes, name)
	lg.mu.Unlock()

	lg.logger.Debug("stopped process", zap.String("name", name))
}

// getCPULoad returns CPU load based on pattern
func (lg *LoadGenerator) getCPULoad(pattern string) float64 {
	switch pattern {
	case "steady":
		return 0.3 + rand.Float64()*0.1 // 30-40%
	case "spiky":
		if rand.Float64() < 0.2 {
			return 0.8 + rand.Float64()*0.2 // 80-100% spike
		}
		return 0.1 + rand.Float64()*0.2 // 10-30% normal
	case "growing":
		return 0.1 + rand.Float64()*0.6 // 10-70% growing over time
	case "random":
		return rand.Float64() // 0-100%
	default:
		return 0.3
	}
}

// getMemLoad returns memory load based on pattern
func (lg *LoadGenerator) getMemLoad(pattern string) float64 {
	switch pattern {
	case "steady":
		return 0.2 + rand.Float64()*0.1 // 20-30%
	case "spiky":
		if rand.Float64() < 0.1 {
			return 0.7 + rand.Float64()*0.3 // 70-100% spike
		}
		return 0.1 + rand.Float64()*0.1 // 10-20% normal
	case "growing":
		return 0.1 + rand.Float64()*0.7 // 10-80% growing over time
	case "random":
		return rand.Float64() // 0-100%
	default:
		return 0.2
	}
}

// GetActiveProcessCount returns the current number of active processes
func (lg *LoadGenerator) GetActiveProcessCount() int {
	lg.mu.RLock()
	defer lg.mu.RUnlock()
	return len(lg.processes)
}

// GetProcessList returns a list of active process names
func (lg *LoadGenerator) GetProcessList() []string {
	lg.mu.RLock()
	defer lg.mu.RUnlock()

	names := make([]string, 0, len(lg.processes))
	for name := range lg.processes {
		names = append(names, name)
	}
	return names
}