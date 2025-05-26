package models

import (
	"time"
)

// Experiment represents an experiment in the system
type Experiment struct {
	ID          string             `json:"id" db:"id"`
	Name        string             `json:"name" db:"name"`
	Description string             `json:"description" db:"description"`
	Phase       string             `json:"phase" db:"phase"`
	Config      ExperimentConfig   `json:"config" db:"config"`
	Status      ExperimentStatus   `json:"status" db:"status"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" db:"updated_at"`
}

// ExperimentConfig contains the configuration for an experiment
type ExperimentConfig struct {
	TargetHosts        []string          `json:"target_hosts"`
	BaselineTemplate   PipelineTemplate  `json:"baseline_template"`
	CandidateTemplate  PipelineTemplate  `json:"candidate_template"`
	LoadProfile        string            `json:"load_profile,omitempty"`
	Duration           time.Duration     `json:"duration"`
	WarmupDuration     time.Duration     `json:"warmup_duration"`
}

// ExperimentStatus represents the current status of an experiment
type ExperimentStatus struct {
	StartTime      *time.Time             `json:"start_time,omitempty"`
	EndTime        *time.Time             `json:"end_time,omitempty"`
	KPIs           map[string]float64     `json:"kpis,omitempty"`
	Error          string                 `json:"error,omitempty"`
	ActiveHosts    int                    `json:"active_hosts"`
}

// ExperimentEvent represents an event in the experiment lifecycle
type ExperimentEvent struct {
	ID           int                    `json:"id" db:"id"`
	ExperimentID string                 `json:"experiment_id" db:"experiment_id"`
	EventType    string                 `json:"event_type" db:"event_type"`
	Phase        string                 `json:"phase" db:"phase"`
	Message      string                 `json:"message" db:"message"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// PipelineTemplate represents a reusable pipeline configuration
type PipelineTemplate struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	URL         string                 `json:"url" db:"config_url"`
	ConfigURL   string                 `json:"config_url" db:"config_url"`
	Variables   map[string]string      `json:"variables" db:"variables"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// Task represents a unit of work for an agent
type Task struct {
	ID           string                 `json:"id" db:"id"`
	HostID       string                 `json:"host_id" db:"host_id"`
	ExperimentID string                 `json:"experiment_id" db:"experiment_id"`
	Type         string                 `json:"type" db:"task_type"`
	Action       string                 `json:"action" db:"action"`
	Config       map[string]interface{} `json:"config" db:"config"`
	Priority     int                    `json:"priority" db:"priority"`
	Status       string                 `json:"status" db:"status"`
	AssignedAt   *time.Time             `json:"assigned_at,omitempty" db:"assigned_at"`
	StartedAt    *time.Time             `json:"started_at,omitempty" db:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	Result       map[string]interface{} `json:"result,omitempty" db:"result"`
	ErrorMessage string                 `json:"error_message,omitempty" db:"error_message"`
	RetryCount   int                    `json:"retry_count" db:"retry_count"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// AgentStatus represents the current status of an agent
type AgentStatus struct {
	HostID        string                 `json:"host_id" db:"host_id"`
	Hostname      string                 `json:"hostname" db:"hostname"`
	IPAddress     string                 `json:"ip_address" db:"ip_address"`
	AgentVersion  string                 `json:"agent_version" db:"agent_version"`
	StartedAt     *time.Time             `json:"started_at" db:"started_at"`
	LastHeartbeat time.Time              `json:"last_heartbeat" db:"last_heartbeat"`
	Status        string                 `json:"status" db:"status"`
	Capabilities  map[string]interface{} `json:"capabilities" db:"capabilities"`
	ActiveTasks   []string               `json:"active_tasks" db:"active_tasks"`
	ResourceUsage ResourceUsage          `json:"resource_usage" db:"resource_usage"`
	Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// AgentHeartbeat represents a heartbeat from an agent
type AgentHeartbeat struct {
	HostID        string                 `json:"host_id"`
	AgentVersion  string                 `json:"agent_version"`
	Status        string                 `json:"status"`
	ActiveTasks   []string               `json:"active_tasks"`
	ResourceUsage ResourceUsage          `json:"resource_usage"`
	LastHeartbeat time.Time              `json:"-"`
}

// ResourceUsage represents resource usage metrics
type ResourceUsage struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	MemoryBytes   int64   `json:"memory_bytes"`
	DiskPercent   float64 `json:"disk_percent"`
	DiskBytes     int64   `json:"disk_bytes"`
}

// ActivePipeline represents a running pipeline on a host
type ActivePipeline struct {
	ID           string                 `json:"id" db:"id"`
	HostID       string                 `json:"host_id" db:"host_id"`
	ExperimentID string                 `json:"experiment_id" db:"experiment_id"`
	Variant      string                 `json:"variant" db:"variant"`
	ConfigURL    string                 `json:"config_url" db:"config_url"`
	ConfigHash   string                 `json:"config_hash" db:"config_hash"`
	ProcessInfo  map[string]interface{} `json:"process_info" db:"process_info"`
	MetricsInfo  map[string]interface{} `json:"metrics_info" db:"metrics_info"`
	Status       string                 `json:"status" db:"status"`
	StartedAt    time.Time              `json:"started_at" db:"started_at"`
	StoppedAt    *time.Time             `json:"stopped_at,omitempty" db:"stopped_at"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// MetricCache represents cached metrics for faster queries
type MetricCache struct {
	ID           int                    `json:"id" db:"id"`
	ExperimentID string                 `json:"experiment_id" db:"experiment_id"`
	Timestamp    time.Time              `json:"timestamp" db:"timestamp"`
	MetricName   string                 `json:"metric_name" db:"metric_name"`
	Variant      string                 `json:"variant" db:"variant"`
	HostID       string                 `json:"host_id" db:"host_id"`
	Value        float64                `json:"value" db:"value"`
	Labels       map[string]string      `json:"labels" db:"labels"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// KPIResult represents the calculated KPIs for an experiment
type KPIResult struct {
	ExperimentID         string    `json:"experiment_id"`
	CalculatedAt         time.Time `json:"calculated_at"`
	CardinalityReduction float64   `json:"cardinality_reduction"`
	CostReduction        float64   `json:"cost_reduction"`
	CPUUsage             struct {
		Baseline  float64 `json:"baseline"`
		Candidate float64 `json:"candidate"`
		Reduction float64 `json:"reduction"`
	} `json:"cpu_usage"`
	MemoryUsage struct {
		Baseline  float64 `json:"baseline"`
		Candidate float64 `json:"candidate"`
		Reduction float64 `json:"reduction"`
	} `json:"memory_usage"`
	IngestRate struct {
		Baseline  float64 `json:"baseline"`
		Candidate float64 `json:"candidate"`
		Reduction float64 `json:"reduction"`
	} `json:"ingest_rate"`
	DataAccuracy float64              `json:"data_accuracy"`
	Errors       []string             `json:"errors,omitempty"`
}