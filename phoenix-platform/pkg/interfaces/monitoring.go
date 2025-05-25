package interfaces

import (
	"context"
	"time"
)

// MonitoringService defines the interface for metrics and monitoring operations
// This interface is used for collecting and analyzing experiment metrics
type MonitoringService interface {
	// QueryMetrics retrieves metrics for a given time range and filters
	QueryMetrics(ctx context.Context, query *MetricsQuery) (*MetricsResult, error)
	
	// GetRealtimeMetrics streams real-time metrics for an experiment
	GetRealtimeMetrics(ctx context.Context, experimentID string) (<-chan *MetricUpdate, error)
	
	// CompareMetrics compares metrics between baseline and candidate
	CompareMetrics(ctx context.Context, req *CompareMetricsRequest) (*MetricsComparison, error)
	
	// GenerateReport creates a detailed analysis report
	GenerateReport(ctx context.Context, experimentID string) (*AnalysisReport, error)
	
	// SetAlert creates or updates an alert rule
	SetAlert(ctx context.Context, alert *AlertRule) error
	
	// GetAlerts retrieves active alerts for an experiment
	GetAlerts(ctx context.Context, experimentID string) ([]*Alert, error)
}

// MetricsCollector defines the interface for collecting metrics from collectors
// This interface is implemented by components that gather metrics from OTel collectors
type MetricsCollector interface {
	// CollectNodeMetrics gathers metrics from a specific node
	CollectNodeMetrics(ctx context.Context, nodeID string) (*NodeMetrics, error)
	
	// CollectPipelineMetrics gathers metrics for a pipeline
	CollectPipelineMetrics(ctx context.Context, pipelineID string) (*PipelineMetrics, error)
	
	// StartCollection begins continuous metrics collection
	StartCollection(ctx context.Context, config *CollectionConfig) error
	
	// StopCollection stops metrics collection
	StopCollection(ctx context.Context, collectionID string) error
}

// CostAnalyzer defines the interface for cost analysis operations
type CostAnalyzer interface {
	// EstimateCost calculates the estimated cost for a pipeline configuration
	EstimateCost(ctx context.Context, metrics *PipelineMetrics) (*CostEstimate, error)
	
	// CompareCosts compares costs between baseline and candidate
	CompareCosts(ctx context.Context, baseline, candidate *PipelineMetrics) (*CostComparison, error)
	
	// ProjectSavings projects cost savings over time
	ProjectSavings(ctx context.Context, comparison *CostComparison, duration time.Duration) (*SavingsProjection, error)
}

// MetricsQuery represents a query for retrieving metrics
type MetricsQuery struct {
	ExperimentID string            `json:"experiment_id"`
	PipelineType string            `json:"pipeline_type"` // baseline or candidate
	MetricNames  []string          `json:"metric_names"`
	StartTime    time.Time         `json:"start_time"`
	EndTime      time.Time         `json:"end_time"`
	Aggregation  AggregationType   `json:"aggregation,omitempty"`
	GroupBy      []string          `json:"group_by,omitempty"`
	Filters      map[string]string `json:"filters,omitempty"`
}

// MetricsResult contains the result of a metrics query
type MetricsResult struct {
	Series    []*MetricSeries        `json:"series"`
	Summary   *MetricsSummary        `json:"summary,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MetricSeries represents a time series of metric values
type MetricSeries struct {
	Name       string            `json:"name"`
	Labels     map[string]string `json:"labels"`
	DataPoints []*DataPoint      `json:"data_points"`
}

// DataPoint represents a single metric measurement
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// MetricsSummary provides aggregate statistics
type MetricsSummary struct {
	TotalDataPoints int64              `json:"total_data_points"`
	TimeRange       *TimeRange         `json:"time_range"`
	Aggregates      map[string]float64 `json:"aggregates"`
}

// TimeRange represents a time interval
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// MetricUpdate represents a real-time metric update
type MetricUpdate struct {
	ExperimentID string                 `json:"experiment_id"`
	PipelineType string                 `json:"pipeline_type"`
	Timestamp    time.Time              `json:"timestamp"`
	Metrics      map[string]float64     `json:"metrics"`
	Labels       map[string]string      `json:"labels,omitempty"`
}

// CompareMetricsRequest defines parameters for metric comparison
type CompareMetricsRequest struct {
	ExperimentID      string    `json:"experiment_id"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
	ComparisonMetrics []string  `json:"comparison_metrics,omitempty"`
}

// AnalysisReport represents a comprehensive analysis of an experiment
type AnalysisReport struct {
	ExperimentID      string                 `json:"experiment_id"`
	GeneratedAt       time.Time              `json:"generated_at"`
	Summary           *ReportSummary         `json:"summary"`
	MetricsAnalysis   *MetricsAnalysis       `json:"metrics_analysis"`
	CostAnalysis      *CostAnalysis          `json:"cost_analysis"`
	Recommendations   []*Recommendation      `json:"recommendations"`
	Visualizations    []*Visualization       `json:"visualizations,omitempty"`
}

// ReportSummary provides high-level experiment results
type ReportSummary struct {
	Duration               time.Duration `json:"duration"`
	Status                 string        `json:"status"`
	MeetsSuccessCriteria   bool          `json:"meets_success_criteria"`
	CardinalityReduction   float64       `json:"cardinality_reduction_percent"`
	EstimatedMonthlySavings float64      `json:"estimated_monthly_savings_usd"`
	KeyFindings            []string      `json:"key_findings"`
}

// MetricsAnalysis contains detailed metrics analysis
type MetricsAnalysis struct {
	BaselineAnalysis   *PipelineAnalysis      `json:"baseline_analysis"`
	CandidateAnalysis  *PipelineAnalysis      `json:"candidate_analysis"`
	Comparison         *DetailedComparison    `json:"comparison"`
	AnomaliesDetected  []*Anomaly             `json:"anomalies,omitempty"`
}

// PipelineAnalysis represents analysis of a single pipeline
type PipelineAnalysis struct {
	AvgTimeSeriesCount float64                `json:"avg_time_series_count"`
	PeakTimeSeriesCount int64                 `json:"peak_time_series_count"`
	AvgLatency         float64                `json:"avg_latency_ms"`
	P50Latency         float64                `json:"p50_latency_ms"`
	P95Latency         float64                `json:"p95_latency_ms"`
	P99Latency         float64                `json:"p99_latency_ms"`
	ErrorRate          float64                `json:"error_rate"`
	ProcessCoverage    *ProcessCoverageStats  `json:"process_coverage"`
	ResourceUsage      *ResourceUsageStats    `json:"resource_usage"`
}

// ProcessCoverageStats contains process coverage statistics
type ProcessCoverageStats struct {
	TotalProcesses      int64    `json:"total_processes"`
	CoveredProcesses    int64    `json:"covered_processes"`
	CriticalProcesses   int64    `json:"critical_processes"`
	CriticalCoverage    float64  `json:"critical_coverage_percent"`
	TopMissedProcesses  []string `json:"top_missed_processes,omitempty"`
}

// ResourceUsageStats contains resource utilization statistics
type ResourceUsageStats struct {
	AvgCPUPercent    float64 `json:"avg_cpu_percent"`
	PeakCPUPercent   float64 `json:"peak_cpu_percent"`
	AvgMemoryMB      float64 `json:"avg_memory_mb"`
	PeakMemoryMB     float64 `json:"peak_memory_mb"`
	NetworkBandwidth float64 `json:"network_bandwidth_mbps"`
}

// DetailedComparison provides detailed comparison between pipelines
type DetailedComparison struct {
	CardinalityReduction   float64                       `json:"cardinality_reduction_percent"`
	LatencyImpact          *LatencyImpact                `json:"latency_impact"`
	ProcessCoverageImpact  *ProcessCoverageImpact        `json:"process_coverage_impact"`
	ResourceImpact         *ResourceImpact               `json:"resource_impact"`
	DataQualityAssessment  *DataQualityAssessment        `json:"data_quality_assessment"`
}

// LatencyImpact represents the impact on latency
type LatencyImpact struct {
	AvgLatencyIncrease float64 `json:"avg_latency_increase_percent"`
	P95LatencyIncrease float64 `json:"p95_latency_increase_percent"`
	P99LatencyIncrease float64 `json:"p99_latency_increase_percent"`
	Acceptable         bool    `json:"acceptable"`
}

// ProcessCoverageImpact represents the impact on process coverage
type ProcessCoverageImpact struct {
	ProcessesLost          int64    `json:"processes_lost"`
	CriticalProcessesLost  int64    `json:"critical_processes_lost"`
	CoverageReduction      float64  `json:"coverage_reduction_percent"`
	LostProcessNames       []string `json:"lost_process_names,omitempty"`
}

// ResourceImpact represents the impact on resource usage
type ResourceImpact struct {
	CPUReduction    float64 `json:"cpu_reduction_percent"`
	MemoryReduction float64 `json:"memory_reduction_percent"`
	NetworkReduction float64 `json:"network_reduction_percent"`
}

// DataQualityAssessment evaluates the quality of data after optimization
type DataQualityAssessment struct {
	Score              float64             `json:"score"` // 0-100
	MetricCompleteness float64             `json:"metric_completeness_percent"`
	DataFreshness      float64             `json:"data_freshness_seconds"`
	QualityIssues      []string            `json:"quality_issues,omitempty"`
}

// Anomaly represents an detected anomaly during the experiment
type Anomaly struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Impact      string    `json:"impact,omitempty"`
}

// Recommendation represents a recommendation based on analysis
type Recommendation struct {
	Priority    string `json:"priority"` // high, medium, low
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact,omitempty"`
	Action      string `json:"action,omitempty"`
}

// Visualization represents a chart or graph in the report
type Visualization struct {
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	DataURL  string                 `json:"data_url,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

// CostAnalysis contains cost-related analysis
type CostAnalysis struct {
	BaselineCost      *CostBreakdown     `json:"baseline_cost"`
	CandidateCost     *CostBreakdown     `json:"candidate_cost"`
	Comparison        *CostComparison    `json:"comparison"`
	ProjectedSavings  *SavingsProjection `json:"projected_savings"`
}

// CostBreakdown provides detailed cost information
type CostBreakdown struct {
	HourlyCost        float64            `json:"hourly_cost_usd"`
	DailyCost         float64            `json:"daily_cost_usd"`
	MonthlyCost       float64            `json:"monthly_cost_usd"`
	CostPerTimeSeries float64            `json:"cost_per_time_series_usd"`
	CostByComponent   map[string]float64 `json:"cost_by_component,omitempty"`
}

// CostEstimate represents an estimated cost
type CostEstimate struct {
	TimeSeriesCount   int64   `json:"time_series_count"`
	DataPointsPerMin  int64   `json:"data_points_per_min"`
	EstimatedHourlyCost  float64 `json:"estimated_hourly_cost_usd"`
	EstimatedMonthlyCost float64 `json:"estimated_monthly_cost_usd"`
	Confidence        float64 `json:"confidence"` // 0-1
}

// CostComparison compares costs between configurations
type CostComparison struct {
	AbsoluteSavings   float64 `json:"absolute_savings_usd"`
	PercentSavings    float64 `json:"percent_savings"`
	ROIEstimate       float64 `json:"roi_estimate"`
	PaybackPeriodDays int     `json:"payback_period_days"`
}

// SavingsProjection projects savings over time
type SavingsProjection struct {
	OneMonth    float64                `json:"one_month_usd"`
	ThreeMonths float64                `json:"three_months_usd"`
	SixMonths   float64                `json:"six_months_usd"`
	OneYear     float64                `json:"one_year_usd"`
	Assumptions map[string]interface{} `json:"assumptions,omitempty"`
}

// AlertRule defines an alert condition
type AlertRule struct {
	ID           string                 `json:"id"`
	ExperimentID string                 `json:"experiment_id"`
	Name         string                 `json:"name"`
	Condition    string                 `json:"condition"`
	Threshold    float64                `json:"threshold"`
	Duration     time.Duration          `json:"duration"`
	Severity     string                 `json:"severity"`
	Actions      []string               `json:"actions"`
	Enabled      bool                   `json:"enabled"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Alert represents an active alert
type Alert struct {
	ID           string    `json:"id"`
	RuleID       string    `json:"rule_id"`
	ExperimentID string    `json:"experiment_id"`
	Severity     string    `json:"severity"`
	State        string    `json:"state"`
	Message      string    `json:"message"`
	Value        float64   `json:"value"`
	TriggeredAt  time.Time `json:"triggered_at"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}

// NodeMetrics contains metrics for a specific node
type NodeMetrics struct {
	NodeID           string             `json:"node_id"`
	CollectorMetrics *CollectorMetrics  `json:"collector_metrics"`
	SystemMetrics    *SystemMetrics     `json:"system_metrics"`
	Timestamp        time.Time          `json:"timestamp"`
}

// CollectorMetrics contains OpenTelemetry collector metrics
type CollectorMetrics struct {
	ReceivedMetrics   int64   `json:"received_metrics"`
	ProcessedMetrics  int64   `json:"processed_metrics"`
	DroppedMetrics    int64   `json:"dropped_metrics"`
	ExportedMetrics   int64   `json:"exported_metrics"`
	QueueSize         int64   `json:"queue_size"`
	ProcessingLatency float64 `json:"processing_latency_ms"`
}

// SystemMetrics contains system-level metrics
type SystemMetrics struct {
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryUsageMB      float64 `json:"memory_usage_mb"`
	NetworkInMbps      float64 `json:"network_in_mbps"`
	NetworkOutMbps     float64 `json:"network_out_mbps"`
	DiskUsagePercent   float64 `json:"disk_usage_percent"`
}

// CollectionConfig defines configuration for metrics collection
type CollectionConfig struct {
	CollectionID     string            `json:"collection_id"`
	ExperimentID     string            `json:"experiment_id"`
	Interval         time.Duration     `json:"interval"`
	MetricsToCollect []string          `json:"metrics_to_collect"`
	Targets          []string          `json:"targets"`
	RetentionPeriod  time.Duration     `json:"retention_period"`
}

// AggregationType defines how metrics should be aggregated
type AggregationType string

const (
	AggregationTypeAvg   AggregationType = "avg"
	AggregationTypeSum   AggregationType = "sum"
	AggregationTypeMin   AggregationType = "min"
	AggregationTypeMax   AggregationType = "max"
	AggregationTypeCount AggregationType = "count"
	AggregationTypeP50   AggregationType = "p50"
	AggregationTypeP95   AggregationType = "p95"
	AggregationTypeP99   AggregationType = "p99"
)