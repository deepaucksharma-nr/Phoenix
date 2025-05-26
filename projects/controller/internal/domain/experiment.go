package domain

import (
	"context"
	"time"
)

// Experiment represents an A/B testing experiment aligned with proto definitions
type Experiment struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Description         string            `json:"description"`
	BaselinePipelineID  string            `json:"baseline_pipeline_id"`
	CandidatePipelineID string            `json:"candidate_pipeline_id"`
	TrafficPercentage   int               `json:"traffic_percentage"`
	TargetServices      []string          `json:"target_services"`
	State               string            `json:"state"`
	StateMessage        string            `json:"state_message"`
	CreatedAt           time.Time         `json:"created_at"`
	StartedAt           time.Time         `json:"started_at"`
	EndedAt             time.Time         `json:"ended_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
	Results             *ExperimentResults `json:"results,omitempty"`
	Labels              map[string]string `json:"labels"`
	Annotations         map[string]string `json:"annotations"`
}

// ExperimentResults represents the results of an experiment
type ExperimentResults struct {
	MetricsComparison   *MetricsComparison `json:"metrics_comparison,omitempty"`
	BaselineCost        *CostBreakdown     `json:"baseline_cost,omitempty"`
	CandidateCost       *CostBreakdown     `json:"candidate_cost,omitempty"`
	CostReductionPercent float64           `json:"cost_reduction_percent"`
	Performance         *PerformanceAnalysis `json:"performance,omitempty"`
	Recommendation      *Recommendation    `json:"recommendation,omitempty"`
	ConfidenceLevel     float64           `json:"confidence_level"`
	PValue              float64           `json:"p_value"`
}

// MetricsComparison compares baseline and candidate metrics
type MetricsComparison struct {
	BaselineTotalMetrics   int64              `json:"baseline_total_metrics"`
	CandidateTotalMetrics  int64              `json:"candidate_total_metrics"`
	ReductionPercentage    float64            `json:"reduction_percentage"`
	ByMetric               []MetricComparison `json:"by_metric"`
}

// MetricComparison compares a specific metric
type MetricComparison struct {
	MetricName           string  `json:"metric_name"`
	BaselineCount        int64   `json:"baseline_count"`
	CandidateCount       int64   `json:"candidate_count"`
	ReductionPercentage  float64 `json:"reduction_percentage"`
	IsDropped            bool    `json:"is_dropped"`
}

// CostBreakdown represents cost analysis
type CostBreakdown struct {
	TotalCost   float64            `json:"total_cost"`
	ComputeCost float64            `json:"compute_cost"`
	StorageCost float64            `json:"storage_cost"`
	NetworkCost float64            `json:"network_cost"`
	CustomCosts map[string]float64 `json:"custom_costs"`
}

// PerformanceAnalysis represents performance comparison
type PerformanceAnalysis struct {
	BaselineLatencyP50 float64             `json:"baseline_latency_p50"`
	BaselineLatencyP95 float64             `json:"baseline_latency_p95"`
	BaselineLatencyP99 float64             `json:"baseline_latency_p99"`
	CandidateLatencyP50 float64            `json:"candidate_latency_p50"`
	CandidateLatencyP95 float64            `json:"candidate_latency_p95"`
	CandidateLatencyP99 float64            `json:"candidate_latency_p99"`
	BaselineResources   *ResourceUtilization `json:"baseline_resources,omitempty"`
	CandidateResources  *ResourceUtilization `json:"candidate_resources,omitempty"`
}

// ResourceUtilization represents resource usage
type ResourceUtilization struct {
	CPUPercentage         float64 `json:"cpu_percentage"`
	MemoryPercentage      float64 `json:"memory_percentage"`
	DiskPercentage        float64 `json:"disk_percentage"`
	NetworkBandwidthMbps  float64 `json:"network_bandwidth_mbps"`
}

// Recommendation represents an experiment recommendation
type Recommendation struct {
	Type              string            `json:"type"`
	Reason            string            `json:"reason"`
	Warnings          []string          `json:"warnings"`
	SuggestedChanges  map[string]string `json:"suggested_changes"`
}

// ExperimentRepository defines the interface for experiment persistence
type ExperimentRepository interface {
	Create(ctx context.Context, exp *Experiment) (*Experiment, error)
	GetByID(ctx context.Context, id string) (*Experiment, error)
	Update(ctx context.Context, exp *Experiment) (*Experiment, error)
	List(ctx context.Context, filters map[string]interface{}, pageSize int32, pageToken string) ([]*Experiment, string, error)
	UpdateState(ctx context.Context, id string, state string, message string) (*Experiment, error)
}

// ExperimentService defines the business logic for experiments
type ExperimentService interface {
	CreateExperiment(ctx context.Context, exp *Experiment) (*Experiment, error)
	GetExperiment(ctx context.Context, id string) (*Experiment, error)
	ListExperiments(ctx context.Context, states []string, labels map[string]string, pageSize int32, pageToken string) ([]*Experiment, string, error)
	UpdateExperimentState(ctx context.Context, id string, state string, reason string) (*Experiment, error)
}

// States
const (
	StatePending      = "pending"
	StateInitializing = "initializing"
	StateRunning      = "running"
	StatePausing      = "pausing"
	StatePaused       = "paused"
	StateResuming     = "resuming"
	StateCompleting   = "completing"
	StateCompleted    = "completed"
	StateFailed       = "failed"
	StateCancelled    = "cancelled"
)

// Recommendation types
const (
	RecommendationAdoptCandidate = "adopt_candidate"
	RecommendationKeepBaseline   = "keep_baseline"
	RecommendationNeedsTuning    = "needs_tuning"
	RecommendationInconclusive   = "inconclusive"
)