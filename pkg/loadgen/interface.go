package loadgen

import (
	"context"
	"time"
)

// Process represents a simulated process
type Process struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	PID        int               `json:"pid"`
	StartTime  time.Time         `json:"start_time"`
	CPUPercent float64           `json:"cpu_percent"`
	MemoryMB   uint64            `json:"memory_mb"`
	State      ProcessState      `json:"state"`
	Tags       map[string]string `json:"tags"`
}

// ProcessState represents the state of a process
type ProcessState string

const (
	ProcessStateRunning  ProcessState = "running"
	ProcessStateSleeping ProcessState = "sleeping"
	ProcessStateZombie   ProcessState = "zombie"
	ProcessStateStopped  ProcessState = "stopped"
)

// LoadPattern defines a pattern for generating load
type LoadPattern interface {
	// Generate creates load based on the pattern
	Generate(ctx context.Context) error
	// Stop stops generating load
	Stop() error
	// GetMetrics returns current metrics
	GetMetrics() *LoadMetrics
}

// LoadMetrics contains metrics about generated load
type LoadMetrics struct {
	ProcessCount   int                    `json:"process_count"`
	TotalCPU       float64                `json:"total_cpu"`
	TotalMemoryMB  uint64                 `json:"total_memory_mb"`
	ProcessesByTag map[string]int         `json:"processes_by_tag"`
	StartTime      time.Time              `json:"start_time"`
	Duration       time.Duration          `json:"duration"`
}

// ProcessSpawner manages process lifecycle
type ProcessSpawner interface {
	// SpawnProcess creates a new simulated process
	SpawnProcess(config ProcessConfig) (*Process, error)
	// KillProcess terminates a simulated process
	KillProcess(pid int) error
	// ListProcesses returns all active processes
	ListProcesses() ([]*Process, error)
	// UpdateProcess updates process metrics
	UpdateProcess(pid int, cpu float64, memory uint64) error
}

// ProcessConfig defines configuration for a process
type ProcessConfig struct {
	Name      string            `json:"name"`
	CPUTarget float64           `json:"cpu_target"`
	MemoryMB  uint64            `json:"memory_mb"`
	Duration  time.Duration     `json:"duration"`
	Tags      map[string]string `json:"tags"`
}

// ProfileConfig defines configuration for a load profile
type ProfileConfig struct {
	Name               string                 `json:"name"`
	ProcessCount       int                    `json:"process_count"`
	ProcessChurnRate   float64                `json:"process_churn_rate"`
	CPUDistribution    Distribution           `json:"cpu_distribution"`
	MemoryDistribution Distribution           `json:"memory_distribution"`
	DurationRange      DurationRange          `json:"duration_range"`
	Tags               map[string]string      `json:"tags"`
	CustomPatterns     []CustomPattern        `json:"custom_patterns,omitempty"`
}

// Distribution defines a statistical distribution
type Distribution struct {
	Type   string  `json:"type"` // "uniform", "normal", "exponential"
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Mean   float64 `json:"mean,omitempty"`
	StdDev float64 `json:"std_dev,omitempty"`
}

// DurationRange defines min and max duration
type DurationRange struct {
	Min time.Duration `json:"min"`
	Max time.Duration `json:"max"`
}

// CustomPattern allows for custom load patterns
type CustomPattern struct {
	Name       string                 `json:"name"`
	Expression string                 `json:"expression"`
	Parameters map[string]interface{} `json:"parameters"`
}