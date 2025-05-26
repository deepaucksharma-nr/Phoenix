package client

import (
	"time"
)

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Experiment represents an experiment
type Experiment struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	BaselinePipeline  string                 `json:"baseline_pipeline"`
	CandidatePipeline string                 `json:"candidate_pipeline"`
	Status            string                 `json:"status"`
	TargetNodes       map[string]string      `json:"target_nodes"`
	Parameters        map[string]interface{} `json:"parameters"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	StartedAt         *time.Time             `json:"started_at,omitempty"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	Results           *ExperimentResults     `json:"results,omitempty"`
}

// CreateExperimentRequest represents a request to create an experiment
type CreateExperimentRequest struct {
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	BaselinePipeline  string                 `json:"baseline_pipeline"`
	CandidatePipeline string                 `json:"candidate_pipeline"`
	TargetNodes       map[string]string      `json:"target_nodes"`
	Duration          time.Duration          `json:"duration"`
	Parameters        map[string]interface{} `json:"parameters"`
}

// ListExperimentsRequest represents a request to list experiments
type ListExperimentsRequest struct {
	Status   string `json:"status,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
	Page     int    `json:"page,omitempty"`
}

// ExperimentResults represents the results of an experiment
type ExperimentResults struct {
	BaselineMetrics      MetricsSummary `json:"baseline_metrics"`
	CandidateMetrics     MetricsSummary `json:"candidate_metrics"`
	CostReduction        float64        `json:"cost_reduction"`
	CardinalityReduction float64        `json:"cardinality_reduction"`
	Summary              string         `json:"summary"`
	Recommendation       string         `json:"recommendation"`
}

// MetricsSummary represents a summary of metrics
type MetricsSummary struct {
	Cardinality    int64   `json:"cardinality"`
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    float64 `json:"memory_usage"`
	NetworkTraffic float64 `json:"network_traffic"`
	ErrorRate      float64 `json:"error_rate"`
}

// ExperimentMetrics represents detailed metrics for an experiment
type ExperimentMetrics struct {
	Baseline  TimeSeriesData `json:"baseline"`
	Candidate TimeSeriesData `json:"candidate"`
}

// TimeSeriesData represents time series metric data
type TimeSeriesData struct {
	Cardinality    []MetricPoint `json:"cardinality"`
	CPUUsage       []MetricPoint `json:"cpu_usage"`
	MemoryUsage    []MetricPoint `json:"memory_usage"`
	NetworkTraffic []MetricPoint `json:"network_traffic"`
}

// MetricPoint represents a single metric data point
type MetricPoint struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
}

// OverlapResult represents the result of an overlap check
type OverlapResult struct {
	HasOverlap        bool     `json:"has_overlap"`
	ConflictingExpIDs []string `json:"conflicting_exp_ids"`
	AffectedNodes     []string `json:"affected_nodes"`
	OverlapType       string   `json:"overlap_type"`
	Severity          string   `json:"severity"`
	Message           string   `json:"message"`
	Suggestions       []string `json:"suggestions"`
}

// Pipeline represents a pipeline configuration
type Pipeline struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Template    string                 `json:"template"`
	Parameters  map[string]interface{} `json:"parameters"`
	Config      string                 `json:"config"`
}

// PipelineDeployment represents a direct pipeline deployment
type PipelineDeployment struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	DeploymentName string                 `json:"deployment_name"`
	Pipeline       string                 `json:"pipeline"`
	PipelineName   string                 `json:"pipeline_name"`
	Namespace      string                 `json:"namespace"`
	TargetNodes    map[string]string      `json:"target_nodes"`
	Parameters     map[string]interface{} `json:"parameters"`
	Status         string                 `json:"status"`
	Phase          string                 `json:"phase"`
	Instances      *DeploymentInstances   `json:"instances,omitempty"`
	Metrics        *DeploymentMetrics     `json:"metrics,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// CreatePipelineDeploymentRequest represents a request to create a pipeline deployment
type CreatePipelineDeploymentRequest struct {
	DeploymentName string                 `json:"deployment_name"`
	PipelineName   string                 `json:"pipeline_name"`
	Namespace      string                 `json:"namespace"`
	TargetNodes    map[string]string      `json:"target_nodes"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	Resources      *ResourceRequirements  `json:"resources,omitempty"`
}

// ListPipelineDeploymentsRequest represents a request to list pipeline deployments
type ListPipelineDeploymentsRequest struct {
	Namespace string `json:"namespace,omitempty"`
	Status    string `json:"status,omitempty"`
}

// DeploymentInstances tracks deployment instance counts
type DeploymentInstances struct {
	Desired int `json:"desired"`
	Ready   int `json:"ready"`
	Updated int `json:"updated"`
}

// DeploymentMetrics contains the latest metrics for a deployment
type DeploymentMetrics struct {
	Cardinality int64     `json:"cardinality"`
	Throughput  string    `json:"throughput"`
	ErrorRate   float64   `json:"error_rate"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	LastUpdated time.Time `json:"last_updated"`
}

// ResourceRequirements defines resource requirements and limits
type ResourceRequirements struct {
	Requests ResourceList `json:"requests,omitempty"`
	Limits   ResourceList `json:"limits,omitempty"`
}

// ResourceList defines CPU and memory resources
type ResourceList struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// RollbackPipelineRequest represents a request to rollback a pipeline
type RollbackPipelineRequest struct {
	Version string `json:"version,omitempty"`
}

// DeploymentStatusResponse represents aggregated status for a deployment
type DeploymentStatusResponse struct {
	DeploymentID   string               `json:"deployment_id"`
	DeploymentName string               `json:"deployment_name"`
	PipelineName   string               `json:"pipeline_name"`
	Namespace      string               `json:"namespace"`
	Status         string               `json:"status"`
	Phase          string               `json:"phase"`
	Instances      *DeploymentInstances `json:"instances,omitempty"`
	Metrics        *DeploymentMetrics   `json:"metrics,omitempty"`
	HealthStatus   string               `json:"health_status,omitempty"`
	LastUpdated    time.Time            `json:"last_updated"`
}
