package loadgen

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// MemoryProcessSpawner implements ProcessSpawner in memory
type MemoryProcessSpawner struct {
	mu         sync.RWMutex
	processes  map[int]*Process
	nextPID    atomic.Int32
	maxPID     int32
}

// NewMemoryProcessSpawner creates a new in-memory process spawner
func NewMemoryProcessSpawner() *MemoryProcessSpawner {
	spawner := &MemoryProcessSpawner{
		processes: make(map[int]*Process),
		maxPID:    99999,
	}
	spawner.nextPID.Store(1000) // Start PIDs at 1000
	return spawner
}

// SpawnProcess creates a new simulated process
func (s *MemoryProcessSpawner) SpawnProcess(config ProcessConfig) (*Process, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate new PID
	pid := int(s.nextPID.Add(1))
	if int32(pid) > s.maxPID {
		return nil, fmt.Errorf("PID limit reached")
	}

	// Create process
	process := &Process{
		ID:         fmt.Sprintf("proc-%d", pid),
		Name:       config.Name,
		PID:        pid,
		StartTime:  time.Now(),
		CPUPercent: config.CPUTarget,
		MemoryMB:   config.MemoryMB,
		State:      ProcessStateRunning,
		Tags:       config.Tags,
	}

	s.processes[pid] = process

	// If duration is specified, schedule termination
	if config.Duration > 0 {
		go func() {
			time.Sleep(config.Duration)
			s.KillProcess(pid)
		}()
	}

	return process, nil
}

// KillProcess terminates a simulated process
func (s *MemoryProcessSpawner) KillProcess(pid int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	process, exists := s.processes[pid]
	if !exists {
		return fmt.Errorf("process with PID %d not found", pid)
	}

	// Mark as zombie briefly before removing
	process.State = ProcessStateZombie
	
	// Remove after a short delay to simulate cleanup
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.mu.Lock()
		delete(s.processes, pid)
		s.mu.Unlock()
	}()

	return nil
}

// ListProcesses returns all active processes
func (s *MemoryProcessSpawner) ListProcesses() ([]*Process, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	processes := make([]*Process, 0, len(s.processes))
	for _, p := range s.processes {
		// Create a copy to avoid data races
		processCopy := *p
		processes = append(processes, &processCopy)
	}

	return processes, nil
}

// UpdateProcess updates process metrics
func (s *MemoryProcessSpawner) UpdateProcess(pid int, cpu float64, memory uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	process, exists := s.processes[pid]
	if !exists {
		return fmt.Errorf("process with PID %d not found", pid)
	}

	process.CPUPercent = cpu
	process.MemoryMB = memory

	return nil
}

// SimulateProcessActivity simulates realistic process behavior
func (s *MemoryProcessSpawner) SimulateProcessActivity(ctx context.Context, updateInterval time.Duration) {
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			for pid, process := range s.processes {
				if process.State == ProcessStateRunning {
					// Add some randomness to CPU and memory usage
					cpuVariation := (rand.Float64() - 0.5) * 10 // ±5% variation
					memVariation := (rand.Float64() - 0.5) * 20 // ±10MB variation
					
					newCPU := process.CPUPercent + cpuVariation
					if newCPU < 0 {
						newCPU = 0
					} else if newCPU > 100 {
						newCPU = 100
					}
					
					newMem := float64(process.MemoryMB) + memVariation
					if newMem < 1 {
						newMem = 1
					}
					
					process.CPUPercent = newCPU
					process.MemoryMB = uint64(newMem)
				}
			}
			s.mu.Unlock()
		}
	}
}