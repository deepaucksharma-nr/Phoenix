package interfaces

import (
	"context"
	"time"
)

// SimulationService defines the interface for load simulation operations
// This interface is used by the Process Simulator component
type SimulationService interface {
	// CreateSimulation creates a new simulation job
	CreateSimulation(ctx context.Context, req *CreateSimulationRequest) (*Simulation, error)
	
	// GetSimulation retrieves a simulation by ID
	GetSimulation(ctx context.Context, id string) (*Simulation, error)
	
	// ListSimulations lists all simulations
	ListSimulations(ctx context.Context, filter *SimulationFilter) (*SimulationList, error)
	
	// StartSimulation starts a simulation
	StartSimulation(ctx context.Context, id string) error
	
	// StopSimulation stops a running simulation
	StopSimulation(ctx context.Context, id string) error
	
	// GetSimulationMetrics retrieves metrics from a simulation
	GetSimulationMetrics(ctx context.Context, id string) (*SimulationMetrics, error)
}

// LoadGenerator defines the interface for generating load patterns
type LoadGenerator interface {
	// GenerateLoad starts generating load based on the profile
	GenerateLoad(ctx context.Context, profile *LoadProfile) error
	
	// StopLoad stops load generation
	StopLoad(ctx context.Context) error
	
	// GetStatus returns the current load generation status
	GetStatus(ctx context.Context) (*LoadStatus, error)
	
	// AdjustLoad dynamically adjusts the load
	AdjustLoad(ctx context.Context, adjustment *LoadAdjustment) error
}

// ProcessSimulator simulates process behavior
type ProcessSimulator interface {
	// SimulateProcesses creates and manages simulated processes
	SimulateProcesses(ctx context.Context, config *ProcessSimulationConfig) error
	
	// GetProcessList returns the list of simulated processes
	GetProcessList(ctx context.Context) ([]*SimulatedProcess, error)
	
	// KillProcess terminates a simulated process
	KillProcess(ctx context.Context, pid int) error
	
	// SpawnProcess creates a new simulated process
	SpawnProcess(ctx context.Context, spec *ProcessSpec) (*SimulatedProcess, error)
}

// Simulation represents a load simulation job
type Simulation struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Profile     *LoadProfile        `json:"profile"`
	Target      *SimulationTarget   `json:"target"`
	State       SimulationState     `json:"state"`
	CreatedAt   time.Time           `json:"created_at"`
	StartedAt   *time.Time          `json:"started_at,omitempty"`
	StoppedAt   *time.Time          `json:"stopped_at,omitempty"`
	CreatedBy   string              `json:"created_by"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SimulationState represents the state of a simulation
type SimulationState string

const (
	SimulationStatePending  SimulationState = "pending"
	SimulationStateRunning  SimulationState = "running"
	SimulationStateStopped  SimulationState = "stopped"
	SimulationStateFailed   SimulationState = "failed"
)

// LoadProfile defines the characteristics of the load to generate
type LoadProfile struct {
	Name            string                 `json:"name"`
	Type            LoadProfileType        `json:"type"`
	ProcessCount    *ProcessCountConfig    `json:"process_count"`
	ResourceUsage   *ResourceUsageConfig   `json:"resource_usage"`
	ProcessChurn    *ProcessChurnConfig    `json:"process_churn"`
	Duration        time.Duration          `json:"duration,omitempty"`
	RampUpTime      time.Duration          `json:"ramp_up_time,omitempty"`
	CustomPatterns  []*CustomPattern       `json:"custom_patterns,omitempty"`
}

// LoadProfileType defines types of load profiles
type LoadProfileType string

const (
	LoadProfileTypeRealistic       LoadProfileType = "realistic"
	LoadProfileTypeHighCardinality LoadProfileType = "high_cardinality"
	LoadProfileTypeHighChurn       LoadProfileType = "high_churn"
	LoadProfileTypeStress          LoadProfileType = "stress"
	LoadProfileTypeCustom          LoadProfileType = "custom"
)

// ProcessCountConfig defines process count parameters
type ProcessCountConfig struct {
	Initial      int           `json:"initial"`
	Min          int           `json:"min"`
	Max          int           `json:"max"`
	Distribution Distribution  `json:"distribution"`
}

// ResourceUsageConfig defines resource usage parameters
type ResourceUsageConfig struct {
	CPUUsage    *ResourcePattern `json:"cpu_usage"`
	MemoryUsage *ResourcePattern `json:"memory_usage"`
	DiskIO      *ResourcePattern `json:"disk_io,omitempty"`
	NetworkIO   *ResourcePattern `json:"network_io,omitempty"`
}

// ResourcePattern defines a resource usage pattern
type ResourcePattern struct {
	BaseValue    float64      `json:"base_value"`
	Variance     float64      `json:"variance"`
	Pattern      PatternType  `json:"pattern"`
	Period       time.Duration `json:"period,omitempty"`
}

// PatternType defines pattern types
type PatternType string

const (
	PatternTypeConstant    PatternType = "constant"
	PatternTypeSinusoidal  PatternType = "sinusoidal"
	PatternTypeSquareWave  PatternType = "square_wave"
	PatternTypeRandom      PatternType = "random"
	PatternTypeSpike       PatternType = "spike"
)

// ProcessChurnConfig defines process lifecycle parameters
type ProcessChurnConfig struct {
	SpawnRate       float64        `json:"spawn_rate"`       // processes per second
	KillRate        float64        `json:"kill_rate"`        // processes per second
	LifetimeMin     time.Duration  `json:"lifetime_min"`
	LifetimeMax     time.Duration  `json:"lifetime_max"`
	LifetimeDist    Distribution   `json:"lifetime_dist"`
}

// Distribution defines statistical distributions
type Distribution string

const (
	DistributionUniform     Distribution = "uniform"
	DistributionNormal      Distribution = "normal"
	DistributionExponential Distribution = "exponential"
	DistributionPoisson     Distribution = "poisson"
)

// CustomPattern allows defining custom load patterns
type CustomPattern struct {
	Name       string                 `json:"name"`
	StartTime  time.Duration          `json:"start_time"`
	Duration   time.Duration          `json:"duration"`
	Parameters map[string]interface{} `json:"parameters"`
}

// SimulationTarget defines where the simulation runs
type SimulationTarget struct {
	Type      TargetType `json:"type"`
	Namespace string     `json:"namespace,omitempty"`
	Nodes     []string   `json:"nodes,omitempty"`
	Selector  map[string]string `json:"selector,omitempty"`
}

// TargetType defines simulation target types
type TargetType string

const (
	TargetTypeLocal      TargetType = "local"
	TargetTypeKubernetes TargetType = "kubernetes"
	TargetTypeDocker     TargetType = "docker"
)

// ProcessSimulationConfig configures process simulation
type ProcessSimulationConfig struct {
	ProcessTypes    []*ProcessTypeConfig   `json:"process_types"`
	SystemProcesses []*SystemProcessConfig `json:"system_processes,omitempty"`
	ChaosConfig     *ChaosConfig           `json:"chaos_config,omitempty"`
}

// ProcessTypeConfig defines a type of process to simulate
type ProcessTypeConfig struct {
	Name            string           `json:"name"`
	Command         string           `json:"command"`
	Args            []string         `json:"args,omitempty"`
	Count           int              `json:"count"`
	ResourceProfile *ResourceProfile `json:"resource_profile"`
	Labels          map[string]string `json:"labels,omitempty"`
}

// SystemProcessConfig simulates system processes
type SystemProcessConfig struct {
	Name     string `json:"name"`
	PID      int    `json:"pid"`
	PPID     int    `json:"ppid"`
	User     string `json:"user"`
	Critical bool   `json:"critical"`
}

// ResourceProfile defines resource consumption profile
type ResourceProfile struct {
	CPUCores      float64 `json:"cpu_cores"`
	MemoryMB      int     `json:"memory_mb"`
	DiskIOMBps    float64 `json:"disk_io_mbps,omitempty"`
	NetworkMBps   float64 `json:"network_mbps,omitempty"`
}

// ChaosConfig defines chaos engineering parameters
type ChaosConfig struct {
	Enabled          bool              `json:"enabled"`
	ProcessCrashes   *CrashConfig      `json:"process_crashes,omitempty"`
	ResourceSpikes   *SpikeConfig      `json:"resource_spikes,omitempty"`
	NetworkIssues    *NetworkConfig    `json:"network_issues,omitempty"`
}

// CrashConfig defines process crash behavior
type CrashConfig struct {
	Probability float64       `json:"probability"` // 0-1
	Pattern     CrashPattern  `json:"pattern"`
	MTBF        time.Duration `json:"mtbf,omitempty"` // mean time between failures
}

// CrashPattern defines crash patterns
type CrashPattern string

const (
	CrashPatternRandom      CrashPattern = "random"
	CrashPatternPeriodic    CrashPattern = "periodic"
	CrashPatternCascading   CrashPattern = "cascading"
)

// SpikeConfig defines resource spike behavior
type SpikeConfig struct {
	Probability  float64       `json:"probability"`
	Magnitude    float64       `json:"magnitude"`    // multiplier
	Duration     time.Duration `json:"duration"`
	ResourceType string        `json:"resource_type"` // cpu, memory, all
}

// NetworkConfig defines network chaos parameters
type NetworkConfig struct {
	PacketLoss   float64       `json:"packet_loss"`   // percentage
	Latency      time.Duration `json:"latency"`
	Jitter       time.Duration `json:"jitter"`
	Bandwidth    float64       `json:"bandwidth_mbps"`
}

// SimulatedProcess represents a simulated process
type SimulatedProcess struct {
	PID             int               `json:"pid"`
	PPID            int               `json:"ppid"`
	Name            string            `json:"name"`
	Command         string            `json:"command"`
	User            string            `json:"user"`
	State           ProcessState      `json:"state"`
	StartTime       time.Time         `json:"start_time"`
	CPUUsage        float64           `json:"cpu_usage"`
	MemoryUsage     int64             `json:"memory_usage_bytes"`
	ThreadCount     int               `json:"thread_count"`
	OpenFiles       int               `json:"open_files"`
	Labels          map[string]string `json:"labels,omitempty"`
}

// ProcessState represents process state
type ProcessState string

const (
	ProcessStateRunning  ProcessState = "running"
	ProcessStateSleeping ProcessState = "sleeping"
	ProcessStateZombie   ProcessState = "zombie"
	ProcessStateStopped  ProcessState = "stopped"
)

// ProcessSpec defines specifications for a new process
type ProcessSpec struct {
	Name            string            `json:"name"`
	Command         string            `json:"command"`
	Args            []string          `json:"args,omitempty"`
	User            string            `json:"user,omitempty"`
	ResourceProfile *ResourceProfile  `json:"resource_profile"`
	Labels          map[string]string `json:"labels,omitempty"`
}

// LoadStatus represents the current load generation status
type LoadStatus struct {
	Active          bool                   `json:"active"`
	CurrentLoad     *LoadMetrics           `json:"current_load"`
	StartTime       time.Time              `json:"start_time"`
	Duration        time.Duration          `json:"duration"`
	ProcessCount    int                    `json:"process_count"`
	Errors          []string               `json:"errors,omitempty"`
}

// LoadMetrics contains current load metrics
type LoadMetrics struct {
	ProcessCount    int     `json:"process_count"`
	CPUUsage        float64 `json:"cpu_usage_percent"`
	MemoryUsage     int64   `json:"memory_usage_bytes"`
	SpawnRate       float64 `json:"spawn_rate"`
	KillRate        float64 `json:"kill_rate"`
	AvgProcessLife  float64 `json:"avg_process_life_seconds"`
}

// LoadAdjustment defines load adjustment parameters
type LoadAdjustment struct {
	ProcessCountDelta int                    `json:"process_count_delta,omitempty"`
	CPUUsageTarget    *float64               `json:"cpu_usage_target,omitempty"`
	MemoryUsageTarget *int64                 `json:"memory_usage_target,omitempty"`
	Duration          time.Duration          `json:"duration,omitempty"`
}

// SimulationMetrics contains metrics from a simulation
type SimulationMetrics struct {
	Duration              time.Duration          `json:"duration"`
	TotalProcessesCreated int64                  `json:"total_processes_created"`
	TotalProcessesKilled  int64                  `json:"total_processes_killed"`
	AvgProcessCount       float64                `json:"avg_process_count"`
	PeakProcessCount      int                    `json:"peak_process_count"`
	AvgCPUUsage           float64                `json:"avg_cpu_usage"`
	PeakCPUUsage          float64                `json:"peak_cpu_usage"`
	AvgMemoryUsage        int64                  `json:"avg_memory_usage"`
	PeakMemoryUsage       int64                  `json:"peak_memory_usage"`
	ErrorCount            int                    `json:"error_count"`
	CustomMetrics         map[string]interface{} `json:"custom_metrics,omitempty"`
}

// Request/Response types
type CreateSimulationRequest struct {
	Name        string            `json:"name" validate:"required,min=3,max=100"`
	Description string            `json:"description" validate:"max=500"`
	Profile     *LoadProfile      `json:"profile" validate:"required"`
	Target      *SimulationTarget `json:"target" validate:"required"`
	AutoStart   bool              `json:"auto_start"`
}

type SimulationFilter struct {
	States    []SimulationState `json:"states,omitempty"`
	CreatedBy string            `json:"created_by,omitempty"`
	PageSize  int               `json:"page_size,omitempty"`
	PageToken string            `json:"page_token,omitempty"`
}

type SimulationList struct {
	Simulations   []*Simulation `json:"simulations"`
	NextPageToken string        `json:"next_page_token,omitempty"`
	TotalCount    int           `json:"total_count"`
}