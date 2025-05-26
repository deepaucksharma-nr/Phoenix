package interfaces

import (
	"context"
	"time"
)

// ExperimentService defines the core interface for experiment management
// This interface is implemented by the Experiment Controller and consumed by the API service
type ExperimentService interface {
	// CreateExperiment creates a new experiment with the given configuration
	CreateExperiment(ctx context.Context, req *CreateExperimentRequest) (*Experiment, error)
	
	// GetExperiment retrieves an experiment by ID
	GetExperiment(ctx context.Context, id string) (*Experiment, error)
	
	// UpdateExperiment updates an existing experiment
	UpdateExperiment(ctx context.Context, id string, req *UpdateExperimentRequest) (*Experiment, error)
	
	// DeleteExperiment deletes an experiment
	DeleteExperiment(ctx context.Context, id string) error
	
	// ListExperiments returns a paginated list of experiments
	ListExperiments(ctx context.Context, filter *ExperimentFilter) (*ExperimentList, error)
	
	// StartExperiment transitions an experiment to running state
	StartExperiment(ctx context.Context, id string) error
	
	// StopExperiment stops a running experiment
	StopExperiment(ctx context.Context, id string) error
	
	// GetExperimentResults retrieves the results/metrics for an experiment
	GetExperimentResults(ctx context.Context, id string) (*ExperimentResults, error)
	
	// PromoteExperiment promotes the candidate pipeline to production
	PromoteExperiment(ctx context.Context, id string) error
}

// ExperimentStore defines the persistence interface for experiments
// This interface is implemented by the store package and consumed by the controller
type ExperimentStore interface {
	// CreateExperiment persists a new experiment
	CreateExperiment(ctx context.Context, exp *Experiment) error
	
	// GetExperiment retrieves an experiment by ID
	GetExperiment(ctx context.Context, id string) (*Experiment, error)
	
	// UpdateExperiment updates an existing experiment
	UpdateExperiment(ctx context.Context, exp *Experiment) error
	
	// DeleteExperiment removes an experiment
	DeleteExperiment(ctx context.Context, id string) error
	
	// ListExperiments returns experiments matching the filter
	ListExperiments(ctx context.Context, filter *ExperimentFilter) ([]*Experiment, error)
	
	// UpdateExperimentState updates only the state of an experiment
	UpdateExperimentState(ctx context.Context, id string, state ExperimentState) error
	
	// GetExperimentsByState retrieves all experiments in a given state
	GetExperimentsByState(ctx context.Context, state ExperimentState) ([]*Experiment, error)
}

// Experiment represents a complete experiment configuration and state
type Experiment struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	State             ExperimentState        `json:"state"`
	BaselinePipeline  string                 `json:"baseline_pipeline"`
	CandidatePipeline string                 `json:"candidate_pipeline"`
	TargetNodes       []string               `json:"target_nodes"`
	Config            *ExperimentConfig      `json:"config"`
	Results           *ExperimentResults     `json:"results,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	StartedAt         *time.Time             `json:"started_at,omitempty"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	CreatedBy         string                 `json:"created_by"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ExperimentState represents the lifecycle state of an experiment
type ExperimentState string

const (
	ExperimentStatePending   ExperimentState = "pending"
	ExperimentStateRunning   ExperimentState = "running"
	ExperimentStateCompleted ExperimentState = "completed"
	ExperimentStateFailed    ExperimentState = "failed"
	ExperimentStateCancelled ExperimentState = "cancelled"
)

// ExperimentConfig contains the configuration for an experiment
type ExperimentConfig struct {
	Duration          time.Duration          `json:"duration"`
	TrafficSplit      *TrafficSplit          `json:"traffic_split"`
	SuccessCriteria   *SuccessCriteria       `json:"success_criteria"`
	LoadProfile       string                 `json:"load_profile"`
	PipelineVariables map[string]interface{} `json:"pipeline_variables,omitempty"`
}

// TrafficSplit defines how traffic is distributed between baseline and candidate
type TrafficSplit struct {
	BaselinePercentage  int `json:"baseline_percentage"`
	CandidatePercentage int `json:"candidate_percentage"`
}

// SuccessCriteria defines what constitutes a successful experiment
type SuccessCriteria struct {
	MinCardinalityReduction float64 `json:"min_cardinality_reduction"`
	MaxLatencyIncrease      float64 `json:"max_latency_increase"`
	MaxErrorRate            float64 `json:"max_error_rate"`
	CriticalProcessCoverage float64 `json:"critical_process_coverage"`
}

// ExperimentResults contains the metrics and analysis from an experiment
type ExperimentResults struct {
	BaselineMetrics   *PipelineMetrics       `json:"baseline_metrics"`
	CandidateMetrics  *PipelineMetrics       `json:"candidate_metrics"`
	Comparison        *MetricsComparison     `json:"comparison"`
	Recommendation    string                 `json:"recommendation"`
	AnalysisTimestamp time.Time              `json:"analysis_timestamp"`
	RawData           map[string]interface{} `json:"raw_data,omitempty"`
}

// PipelineMetrics contains metrics for a single pipeline
type PipelineMetrics struct {
	TimeSeriesCount    int64   `json:"time_series_count"`
	AvgLatency         float64 `json:"avg_latency_ms"`
	P99Latency         float64 `json:"p99_latency_ms"`
	ErrorRate          float64 `json:"error_rate"`
	ProcessesCovered   int64   `json:"processes_covered"`
	CriticalProcesses  int64   `json:"critical_processes"`
	DataPointsPerMin   int64   `json:"data_points_per_min"`
	EstimatedCostPerHr float64 `json:"estimated_cost_per_hr"`
}

// MetricsComparison contains the comparison between baseline and candidate
type MetricsComparison struct {
	CardinalityReduction   float64 `json:"cardinality_reduction_percent"`
	LatencyIncrease        float64 `json:"latency_increase_percent"`
	ErrorRateDiff          float64 `json:"error_rate_diff"`
	CriticalProcessCoverage float64 `json:"critical_process_coverage_percent"`
	CostSavings            float64 `json:"cost_savings_percent"`
	MeetsSuccessCriteria   bool    `json:"meets_success_criteria"`
}

// Request/Response types for API operations
type CreateExperimentRequest struct {
	Name              string            `json:"name" validate:"required,min=3,max=100"`
	Description       string            `json:"description" validate:"max=500"`
	BaselinePipeline  string            `json:"baseline_pipeline" validate:"required"`
	CandidatePipeline string            `json:"candidate_pipeline" validate:"required"`
	TargetNodes       []string          `json:"target_nodes" validate:"required,min=1"`
	Config            *ExperimentConfig `json:"config" validate:"required"`
}

type UpdateExperimentRequest struct {
	Description *string           `json:"description,omitempty"`
	Config      *ExperimentConfig `json:"config,omitempty"`
}

type ExperimentFilter struct {
	States    []ExperimentState `json:"states,omitempty"`
	CreatedBy string            `json:"created_by,omitempty"`
	PageSize  int               `json:"page_size,omitempty"`
	PageToken string            `json:"page_token,omitempty"`
}

type ExperimentList struct {
	Experiments   []*Experiment `json:"experiments"`
	NextPageToken string        `json:"next_page_token,omitempty"`
	TotalCount    int           `json:"total_count"`
}