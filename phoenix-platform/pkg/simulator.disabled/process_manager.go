package simulator

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// createProcess creates a new simulated process based on pattern
func (s *ProcessSimulator) createProcess(pattern ProcessPattern, index int) *SimulatedProcess {
	name := fmt.Sprintf(pattern.NameTemplate, index)
	
	// Handle templates with multiple placeholders
	if strings.Contains(name, "%!") {
		name = fmt.Sprintf(pattern.NameTemplate, randomString(6), index)
	}

	lifetime := pattern.Lifetime
	if lifetime == 0 {
		lifetime = s.config.Duration // Default to full simulation duration
	}

	// Set initial resource usage based on pattern
	cpu, mem := s.getInitialResources(pattern.CPUPattern, pattern.MemPattern)

	return &SimulatedProcess{
		Name:        name,
		CPUPattern:  pattern.CPUPattern,
		MemPattern:  pattern.MemPattern,
		StartTime:   time.Now(),
		Lifetime:    lifetime,
		Priority:    pattern.Priority,
		CPUUsage:    cpu,
		MemoryUsage: mem,
	}
}

// startProcess starts a simulated process
func (s *ProcessSimulator) startProcess(proc *SimulatedProcess) error {
	// Try to use stress-ng for realistic resource usage
	args := []string{
		"--cpu", "1",
		"--cpu-load", fmt.Sprintf("%.0f", proc.CPUUsage),
		"--vm", "1",
		"--vm-bytes", fmt.Sprintf("%.0fM", proc.MemoryUsage),
		"--timeout", "0", // Run indefinitely
		"--quiet",
	}

	cmd := exec.Command("stress-ng", args...)
	
	// Set process name in environment
	cmd.Env = append(os.Environ(), 
		fmt.Sprintf("PROCESS_NAME=%s", proc.Name),
		fmt.Sprintf("PROCESS_PRIORITY=%s", proc.Priority),
	)
	
	// Set process group so we can kill all children
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		// If stress-ng is not available, create a simple busy process
		cmd = exec.Command("sh", "-c", fmt.Sprintf(
			`while true; do 
				# Simulate CPU usage
				for i in $(seq 1 %d); do
					echo "Process %s (priority: %s) running" > /dev/null
				done
				sleep 0.1
			done`, int(proc.CPUUsage), proc.Name, proc.Priority))
		
		cmd.Env = append(os.Environ(), 
			fmt.Sprintf("PROCESS_NAME=%s", proc.Name),
			fmt.Sprintf("PROCESS_PRIORITY=%s", proc.Priority),
		)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		
		if err := cmd.Start(); err != nil {
			return err
		}
	}

	proc.cmd = cmd
	proc.PID = cmd.Process.Pid

	s.mu.Lock()
	s.processes[proc.Name] = proc
	s.mu.Unlock()

	s.logger.Debug("started process",
		zap.String("name", proc.Name),
		zap.Int("pid", proc.PID),
		zap.String("priority", proc.Priority),
		zap.Float64("cpu", proc.CPUUsage),
		zap.Float64("memory", proc.MemoryUsage),
	)

	return nil
}

// stopProcess stops a simulated process
func (s *ProcessSimulator) stopProcess(proc *SimulatedProcess) {
	if proc.cmd != nil && proc.cmd.Process != nil {
		// Record lifetime
		lifetime := time.Since(proc.StartTime).Seconds()
		s.processLifetime.Observe(lifetime)

		// Kill the process group
		syscall.Kill(-proc.cmd.Process.Pid, syscall.SIGTERM)
		
		// Wait briefly for graceful shutdown
		done := make(chan error, 1)
		go func() {
			done <- proc.cmd.Wait()
		}()
		
		select {
		case <-done:
			// Process exited
		case <-time.After(2 * time.Second):
			// Force kill if still running
			syscall.Kill(-proc.cmd.Process.Pid, syscall.SIGKILL)
		}

		s.logger.Debug("stopped process",
			zap.String("name", proc.Name),
			zap.Float64("lifetime_seconds", lifetime),
		)
	}
}

// updateProcesses updates process resource patterns
func (s *ProcessSimulator) updateProcesses() {
	s.mu.Lock()
	defer s.mu.Unlock()

	elapsed := time.Since(s.startTime)

	for _, proc := range s.processes {
		// Update CPU usage based on pattern
		proc.CPUUsage = s.getCPUUsage(proc.CPUPattern, elapsed)
		
		// Update memory usage based on pattern
		proc.MemoryUsage = s.getMemoryUsage(proc.MemPattern, elapsed)

		// Simulate chaos if enabled
		if s.config.Parameters != nil {
			if chaos, ok := s.config.Parameters["enable_chaos"].(bool); ok && chaos {
				s.applyChaos(proc)
			}
		}
	}
}

// checkLifetimes checks and handles process lifetimes
func (s *ProcessSimulator) checkLifetimes(profile *Profile) {
	s.mu.Lock()
	defer s.mu.Unlock()

	toRestart := []ProcessPattern{}

	for name, proc := range s.processes {
		if proc.Lifetime > 0 && time.Since(proc.StartTime) > proc.Lifetime {
			s.logger.Debug("process lifetime expired",
				zap.String("name", name),
				zap.Duration("lifetime", proc.Lifetime),
			)
			
			s.stopProcess(proc)
			delete(s.processes, name)
			
			// Find the pattern to restart
			for _, pattern := range profile.Patterns {
				if matchesPattern(name, pattern.NameTemplate) {
					toRestart = append(toRestart, pattern)
					break
				}
			}
		}
	}

	// Start replacements outside the lock
	s.mu.Unlock()
	for _, pattern := range toRestart {
		newProc := s.createProcess(pattern, rand.Intn(10000))
		s.startProcess(newProc)
		s.processChurn.Inc()
	}
	s.mu.Lock()
}

// simulateChurn simulates process churn
func (s *ProcessSimulator) simulateChurn(profile *Profile) {
	s.mu.Lock()
	
	processCount := len(s.processes)
	churns := int(float64(processCount) * profile.ChurnRate / 60) // Per minute
	
	if churns == 0 {
		s.mu.Unlock()
		return
	}

	s.logger.Info("simulating process churn",
		zap.Int("processes", churns),
		zap.Float64("rate", profile.ChurnRate),
	)

	// Select random processes to restart
	names := make([]string, 0, processCount)
	for name := range s.processes {
		names = append(names, name)
	}

	toRestart := []struct {
		pattern ProcessPattern
		proc    *SimulatedProcess
	}{}

	for i := 0; i < churns && i < len(names); i++ {
		idx := rand.Intn(len(names))
		name := names[idx]
		proc := s.processes[name]
		
		if proc != nil && proc.Priority != "critical" { // Don't churn critical processes
			s.stopProcess(proc)
			delete(s.processes, name)
			
			// Find the pattern to restart
			for _, pattern := range profile.Patterns {
				if matchesPattern(name, pattern.NameTemplate) {
					toRestart = append(toRestart, struct {
						pattern ProcessPattern
						proc    *SimulatedProcess
					}{pattern, proc})
					break
				}
			}
		}
	}
	
	s.mu.Unlock()

	// Start replacements
	for _, item := range toRestart {
		newProc := s.createProcess(item.pattern, rand.Intn(10000))
		s.startProcess(newProc)
		s.processChurn.Inc()
	}
}

// cleanup stops all processes
func (s *ProcessSimulator) cleanup() error {
	s.logger.Info("cleaning up processes")
	
	s.mu.Lock()
	defer s.mu.Unlock()

	for name, proc := range s.processes {
		s.logger.Debug("stopping process", zap.String("name", name))
		s.stopProcess(proc)
	}

	s.processes = make(map[string]*SimulatedProcess)
	s.processCount.Set(0)
	
	return nil
}

// Resource calculation helpers

func (s *ProcessSimulator) getInitialResources(cpuPattern, memPattern string) (cpu, mem float64) {
	switch cpuPattern {
	case "steady":
		cpu = 20.0
	case "spiky":
		cpu = 30.0
	case "growing":
		cpu = 10.0
	case "random":
		cpu = float64(10 + rand.Intn(40))
	default:
		cpu = 20.0
	}

	switch memPattern {
	case "steady":
		mem = 50.0
	case "spiky":
		mem = 75.0
	case "growing":
		mem = 30.0
	case "random":
		mem = float64(20 + rand.Intn(100))
	default:
		mem = 50.0
	}

	return cpu, mem
}

func (s *ProcessSimulator) getCPUUsage(pattern string, elapsed time.Duration) float64 {
	switch pattern {
	case "steady":
		return 20.0 + rand.Float64()*5 // 20-25%
	case "spiky":
		// Spike every 30 seconds
		if int(elapsed.Seconds())%30 < 5 {
			return 70.0 + rand.Float64()*20 // 70-90%
		}
		return 10.0 + rand.Float64()*10 // 10-20%
	case "growing":
		// Increases over time up to 80%
		growth := elapsed.Minutes() * 2
		return min(80.0, 10.0+growth+rand.Float64()*5)
	case "random":
		return rand.Float64() * 100 // 0-100%
	default:
		return 20.0
	}
}

func (s *ProcessSimulator) getMemoryUsage(pattern string, elapsed time.Duration) float64 {
	switch pattern {
	case "steady":
		return 50.0 + rand.Float64()*10 // 50-60MB
	case "spiky":
		// Memory spike every minute
		if int(elapsed.Seconds())%60 < 10 {
			return 150.0 + rand.Float64()*50 // 150-200MB
		}
		return 40.0 + rand.Float64()*20 // 40-60MB
	case "growing":
		// Memory leak simulation
		growth := elapsed.Minutes() * 10
		return min(500.0, 50.0+growth+rand.Float64()*10)
	case "random":
		return 10.0 + rand.Float64()*200 // 10-210MB
	default:
		return 50.0
	}
}

// applyChaos applies chaos engineering effects
func (s *ProcessSimulator) applyChaos(proc *SimulatedProcess) {
	// Random CPU spike
	if rand.Float64() < 0.05 { // 5% chance
		proc.CPUUsage = 90.0 + rand.Float64()*10
		s.logger.Debug("chaos: CPU spike",
			zap.String("process", proc.Name),
			zap.Float64("cpu", proc.CPUUsage),
		)
	}

	// Random memory spike
	if rand.Float64() < 0.03 { // 3% chance
		proc.MemoryUsage *= 2.5
		s.logger.Debug("chaos: memory spike",
			zap.String("process", proc.Name),
			zap.Float64("memory", proc.MemoryUsage),
		)
	}
}

// Helper functions

func matchesPattern(name, pattern string) bool {
	// Simple pattern matching - could be improved with regex
	basePattern := strings.Split(pattern, "-%")[0]
	return strings.HasPrefix(name, basePattern)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}