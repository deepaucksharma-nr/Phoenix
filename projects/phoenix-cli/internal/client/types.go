package client

import (
	"fmt"
	"time"
)

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// APIError represents an error returned by the API
type APIError struct {
	StatusCode int
	Message    string
	Details    string
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s - %s", e.StatusCode, e.Message, e.Details)
}

// LoginRequest represents a request to login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a response to a login request
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// SuccessCriteria defines the criteria for a successful experiment
type SuccessCriteria struct {
	MaxErrorRate            float64 `json:"max_error_rate,omitempty"`
	MinThroughput           float64 `json:"min_throughput,omitempty"`
	MaxLatency              float64 `json:"max_latency,omitempty"`
	MinCostReduction        float64 `json:"min_cost_reduction,omitempty"`
	MaxDataLoss             float64 `json:"max_data_loss,omitempty"`
	RequireStatSignificance bool    `json:"require_stat_significance,omitempty"`
}

// Experiment represents an experiment
type Experiment struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	BaselinePipeline  string                 `json:"baseline_pipeline"`
	CandidatePipeline string                 `json:"candidate_pipeline"`
	Phase             string                 `json:"phase"`
	Status            string                 `json:"status,omitempty"` // Deprecated: use Phase
	TargetNodes       map[string]string      `json:"target_nodes"`
	Parameters        map[string]interface{} `json:"parameters"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	StartedAt         *time.Time             `json:"started_at,omitempty"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	Results           *ExperimentResults     `json:"results,omitempty"`
	Namespace         string                 `json:"namespace,omitempty"`
}

// CreateExperimentRequest represents a request to create an experiment
type CreateExperimentRequest struct {
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	BaselinePipeline  string                 `json:"baseline_pipeline"`
	CandidatePipeline string                 `json:"candidate_pipeline"`
	TargetNodes       map[string]string      `json:"target_nodes"`
	Duration          interface{}            `json:"duration"` // Can be time.Duration or string
	Parameters        map[string]interface{} `json:"parameters"`
	Namespace         string                 `json:"namespace,omitempty"`
	PipelineA         string                 `json:"pipeline_a,omitempty"`
	PipelineB         string                 `json:"pipeline_b,omitempty"`
	TrafficSplit      interface{}            `json:"traffic_split,omitempty"` // Can be float64 or string
	Selector          string                 `json:"selector,omitempty"`
	SuccessCriteria   *SuccessCriteria       `json:"success_criteria,omitempty"`
	Metadata          map[string]string      `json:"metadata,omitempty"`
}

// ListExperimentsRequest represents a request to list experiments
type ListExperimentsRequest struct {
	Status   string `json:"status,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
	Page     int    `json:"page,omitempty"`
}

// ListExperimentsResponse represents a response containing experiments
type ListExperimentsResponse struct {
	Experiments []Experiment `json:"experiments"`
	Total       int          `json:"total"`
	Page        int          `json:"page"`
	PageSize    int          `json:"page_size"`
}

// PipelineMetrics represents metrics for a pipeline
type PipelineMetrics struct {
	Cardinality         int64     `json:"cardinality"`
	Throughput          float64   `json:"throughput"`
	ErrorRate           float64   `json:"error_rate"`
	Latency             float64   `json:"latency"`
	CostPerHour         float64   `json:"cost_per_hour"`
	DataLossPercent     float64   `json:"data_loss_percent"`
	DataPointsPerSecond float64   `json:"data_points_per_second"`
	BytesPerSecond      float64   `json:"bytes_per_second"`
	P50Latency          float64   `json:"p50_latency"`
	P95Latency          float64   `json:"p95_latency"`
	P99Latency          float64   `json:"p99_latency"`
	Timestamp           time.Time `json:"timestamp"`
}

// MetricsSummary provides a summary of experiment metrics
type MetricsSummary struct {
	Cardinality             int64   `json:"cardinality"`
	CPUUsage                float64 `json:"cpu_usage"`
	MemoryUsage             float64 `json:"memory_usage"`
	NetworkTraffic          float64 `json:"network_traffic"`
	ErrorRate               float64 `json:"error_rate"`
	CostReductionPercent    float64 `json:"cost_reduction_percent"`
	DataLossPercent         float64 `json:"data_loss_percent"`
	ProgressPercent         float64 `json:"progress_percent"`
	EstimatedMonthlySavings float64 `json:"estimated_monthly_savings"`
	DataProcessedGB         float64 `json:"data_processed_gb"`
	ActiveCollectors        int     `json:"active_collectors"`
}

// ExperimentMetrics represents metrics for an experiment
type ExperimentMetrics struct {
	ExperimentID string          `json:"experiment_id"`
	Summary      MetricsSummary  `json:"summary"`
	PipelineA    PipelineMetrics `json:"pipeline_a"`
	PipelineB    PipelineMetrics `json:"pipeline_b"`
	Baseline     TimeSeriesData  `json:"baseline"`
	Candidate    TimeSeriesData  `json:"candidate"`
	Timestamp    time.Time       `json:"timestamp"`
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
