package loadgen

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// BaseLoadPattern provides common functionality for load patterns
type BaseLoadPattern struct {
	spawner    ProcessSpawner
	config     ProfileConfig
	processes  sync.Map // map[int]*Process
	metrics    LoadMetrics
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	startTime  time.Time
}

// NewBaseLoadPattern creates a new base load pattern
func NewBaseLoadPattern(spawner ProcessSpawner, config ProfileConfig) *BaseLoadPattern {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseLoadPattern{
		spawner:   spawner,
		config:    config,
		ctx:       ctx,
		cancel:    cancel,
		startTime: time.Now(),
		metrics: LoadMetrics{
			ProcessesByTag: make(map[string]int),
			StartTime:      time.Now(),
		},
	}
}

// Stop stops the load pattern
func (p *BaseLoadPattern) Stop() error {
	p.cancel()
	
	// Kill all processes
	p.processes.Range(func(key, value interface{}) bool {
		pid := key.(int)
		p.spawner.KillProcess(pid)
		return true
	})
	
	return nil
}

// GetMetrics returns current metrics
func (p *BaseLoadPattern) GetMetrics() *LoadMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	metrics := p.metrics
	metrics.Duration = time.Since(p.startTime)
	
	return &metrics
}

// updateMetrics updates the metrics based on current processes
func (p *BaseLoadPattern) updateMetrics() {
	processes, _ := p.spawner.ListProcesses()
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.metrics.ProcessCount = len(processes)
	p.metrics.TotalCPU = 0
	p.metrics.TotalMemoryMB = 0
	p.metrics.ProcessesByTag = make(map[string]int)
	
	for _, proc := range processes {
		p.metrics.TotalCPU += proc.CPUPercent
		p.metrics.TotalMemoryMB += proc.MemoryMB
		
		for tag, value := range proc.Tags {
			key := fmt.Sprintf("%s:%s", tag, value)
			p.metrics.ProcessesByTag[key]++
		}
	}
}

// generateValue generates a value based on distribution
func generateValue(dist Distribution) float64 {
	switch dist.Type {
	case "uniform":
		return dist.Min + rand.Float64()*(dist.Max-dist.Min)
	case "normal":
		// Box-Muller transform for normal distribution
		u1 := rand.Float64()
		u2 := rand.Float64()
		z0 := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
		value := dist.Mean + z0*dist.StdDev
		// Clamp to min/max
		if value < dist.Min {
			value = dist.Min
		} else if value > dist.Max {
			value = dist.Max
		}
		return value
	case "exponential":
		lambda := 1.0 / dist.Mean
		value := -math.Log(rand.Float64()) / lambda
		// Clamp to min/max
		if value < dist.Min {
			value = dist.Min
		} else if value > dist.Max {
			value = dist.Max
		}
		return value
	default:
		// Default to uniform
		return dist.Min + rand.Float64()*(dist.Max-dist.Min)
	}
}

// RealisticLoadPattern simulates realistic process behavior
type RealisticLoadPattern struct {
	*BaseLoadPattern
}

// NewRealisticLoadPattern creates a realistic load pattern
func NewRealisticLoadPattern(spawner ProcessSpawner, config ProfileConfig) *RealisticLoadPattern {
	return &RealisticLoadPattern{
		BaseLoadPattern: NewBaseLoadPattern(spawner, config),
	}
}

// Generate creates load based on the realistic pattern
func (p *RealisticLoadPattern) Generate(ctx context.Context) error {
	// Start with some long-running processes
	longRunningCount := int(float64(p.config.ProcessCount) * 0.7)
	shortLivedCount := p.config.ProcessCount - longRunningCount
	
	// Spawn long-running processes
	for i := 0; i < longRunningCount; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			processName := fmt.Sprintf("%s-long-%d", p.config.Name, i)
			cpu := generateValue(p.config.CPUDistribution)
			memory := uint64(generateValue(p.config.MemoryDistribution))
			
			config := ProcessConfig{
				Name:      processName,
				CPUTarget: cpu,
				MemoryMB:  memory,
				Duration:  0, // Long-running
				Tags:      p.config.Tags,
			}
			
			proc, err := p.spawner.SpawnProcess(config)
			if err != nil {
				return fmt.Errorf("failed to spawn process: %w", err)
			}
			
			p.processes.Store(proc.PID, proc)
		}
		
		// Small delay between spawns
		time.Sleep(10 * time.Millisecond)
	}
	
	// Continuously spawn short-lived processes
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Spawn short-lived processes based on churn rate
			numToSpawn := int(float64(shortLivedCount) * p.config.ProcessChurnRate)
			
			for i := 0; i < numToSpawn; i++ {
				processName := fmt.Sprintf("%s-short-%d", p.config.Name, rand.Intn(10000))
				cpu := generateValue(p.config.CPUDistribution)
				memory := uint64(generateValue(p.config.MemoryDistribution))
				duration := p.config.DurationRange.Min + 
					time.Duration(rand.Int63n(int64(p.config.DurationRange.Max-p.config.DurationRange.Min)))
				
				config := ProcessConfig{
					Name:      processName,
					CPUTarget: cpu,
					MemoryMB:  memory,
					Duration:  duration,
					Tags:      p.config.Tags,
				}
				
				proc, err := p.spawner.SpawnProcess(config)
				if err != nil {
					continue // Log but continue
				}
				
				p.processes.Store(proc.PID, proc)
				
				// Clean up after duration
				go func(pid int) {
					time.Sleep(duration)
					p.processes.Delete(pid)
				}(proc.PID)
			}
			
			// Update metrics
			p.updateMetrics()
		}
	}
}

// HighCardinalityLoadPattern creates many unique process names
type HighCardinalityLoadPattern struct {
	*BaseLoadPattern
	nameCounter atomic.Int64
}

// NewHighCardinalityLoadPattern creates a high cardinality load pattern
func NewHighCardinalityLoadPattern(spawner ProcessSpawner, config ProfileConfig) *HighCardinalityLoadPattern {
	return &HighCardinalityLoadPattern{
		BaseLoadPattern: NewBaseLoadPattern(spawner, config),
	}
}

// Generate creates load with high cardinality
func (p *HighCardinalityLoadPattern) Generate(ctx context.Context) error {
	// Spawn processes with unique names
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	processesPerTick := p.config.ProcessCount / 10
	if processesPerTick < 1 {
		processesPerTick = 1
	}
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Count current processes
			currentCount := 0
			p.processes.Range(func(key, value interface{}) bool {
				currentCount++
				return true
			})
			
			// Spawn new processes if below target
			if currentCount < p.config.ProcessCount {
				for i := 0; i < processesPerTick && currentCount < p.config.ProcessCount; i++ {
					// Generate unique name with timestamp and counter
					counter := p.nameCounter.Add(1)
					processName := fmt.Sprintf("%s-%d-%d-%s", 
						p.config.Name, 
						time.Now().UnixNano(), 
						counter,
						generateRandomString(8))
					
					cpu := generateValue(p.config.CPUDistribution)
					memory := uint64(generateValue(p.config.MemoryDistribution))
					duration := p.config.DurationRange.Min + 
						time.Duration(rand.Int63n(int64(p.config.DurationRange.Max-p.config.DurationRange.Min)))
					
					// Add random tags for more cardinality
					tags := make(map[string]string)
					for k, v := range p.config.Tags {
						tags[k] = v
					}
					tags["instance_id"] = generateRandomString(12)
					tags["version"] = fmt.Sprintf("v%d.%d.%d", rand.Intn(10), rand.Intn(10), rand.Intn(100))
					
					config := ProcessConfig{
						Name:      processName,
						CPUTarget: cpu,
						MemoryMB:  memory,
						Duration:  duration,
						Tags:      tags,
					}
					
					proc, err := p.spawner.SpawnProcess(config)
					if err != nil {
						continue
					}
					
					p.processes.Store(proc.PID, proc)
					currentCount++
					
					// Clean up after duration
					go func(pid int) {
						time.Sleep(duration)
						p.processes.Delete(pid)
					}(proc.PID)
				}
			}
			
			// Update metrics
			p.updateMetrics()
		}
	}
}

// ProcessChurnLoadPattern creates rapid process creation/destruction
type ProcessChurnLoadPattern struct {
	*BaseLoadPattern
}

// NewProcessChurnLoadPattern creates a process churn load pattern
func NewProcessChurnLoadPattern(spawner ProcessSpawner, config ProfileConfig) *ProcessChurnLoadPattern {
	return &ProcessChurnLoadPattern{
		BaseLoadPattern: NewBaseLoadPattern(spawner, config),
	}
}

// Generate creates load with high process churn
func (p *ProcessChurnLoadPattern) Generate(ctx context.Context) error {
	// Very short-lived processes with high spawn rate
	spawnInterval := time.Duration(float64(time.Second) / (p.config.ProcessChurnRate * float64(p.config.ProcessCount)))
	ticker := time.NewTicker(spawnInterval)
	defer ticker.Stop()
	
	processCounter := 0
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			processCounter++
			processName := fmt.Sprintf("%s-churn-%d", p.config.Name, processCounter)
			
			cpu := generateValue(p.config.CPUDistribution)
			memory := uint64(generateValue(p.config.MemoryDistribution))
			
			// Very short duration for high churn
			duration := time.Duration(rand.Int63n(int64(5 * time.Second)))
			
			config := ProcessConfig{
				Name:      processName,
				CPUTarget: cpu,
				MemoryMB:  memory,
				Duration:  duration,
				Tags:      p.config.Tags,
			}
			
			proc, err := p.spawner.SpawnProcess(config)
			if err != nil {
				continue
			}
			
			p.processes.Store(proc.PID, proc)
			
			// Clean up after duration
			go func(pid int) {
				time.Sleep(duration)
				p.processes.Delete(pid)
				p.spawner.KillProcess(pid)
			}(proc.PID)
			
			// Update metrics periodically
			if processCounter%10 == 0 {
				p.updateMetrics()
			}
		}
	}
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}