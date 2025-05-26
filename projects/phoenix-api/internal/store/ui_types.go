package store

import (
	"time"
	
	"github.com/phoenix/platform/pkg/common/websocket"
)

// MetricCostFlow represents the current metric cost breakdown
type MetricCostFlow struct {
	TotalCostPerMinute float64                   `json:"total_cost_per_minute"`
	TopMetrics         []MetricCostDetail        `json:"top_metrics"`
	ByService          map[string]float64        `json:"by_service"`
	ByNamespace        map[string]float64        `json:"by_namespace"`
	LastUpdated        time.Time                 `json:"last_updated"`
}

// MetricCostDetail represents cost details for a single metric
type MetricCostDetail struct {
	Name          string            `json:"name"`
	CostPerMinute float64           `json:"cost_per_minute"`
	Cardinality   int64             `json:"cardinality"`
	Percentage    float64           `json:"percentage"`
	Labels        map[string]string `json:"labels"`
}

// CardinalityBreakdown represents cardinality analysis
type CardinalityBreakdown struct {
	TotalCardinality int64                      `json:"total_cardinality"`
	ByMetric         map[string]int64           `json:"by_metric"`
	ByLabel          map[string]int64           `json:"by_label"`
	TopContributors  []CardinalityContributor  `json:"top_contributors"`
	Timestamp        time.Time                  `json:"timestamp"`
}

// CardinalityContributor represents a high-cardinality source
type CardinalityContributor struct {
	MetricName   string            `json:"metric_name"`
	Labels       map[string]string `json:"labels"`
	Cardinality  int64             `json:"cardinality"`
	Percentage   float64           `json:"percentage"`
}

// PipelineTemplate represents a pre-built pipeline configuration
type PipelineTemplate struct {
	ID                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	Category              string                 `json:"category"`
	Config                map[string]interface{} `json:"config"`
	EstimatedSavings      float64                `json:"estimated_savings_percent"`
	EstimatedCPUImpact    float64                `json:"estimated_cpu_impact"`
	EstimatedMemoryImpact int                    `json:"estimated_memory_impact_mb"`
	UIPreview             UIPreview              `json:"ui_preview"`
}

// UIPreview represents visual preview data for a pipeline
type UIPreview struct {
	ProcessorBlocks []ProcessorBlock `json:"processor_blocks"`
	FlowDirection   string           `json:"flow_direction"` // horizontal, vertical
	Color           string           `json:"color"`           // primary color for the pipeline
}

// ProcessorBlock represents a visual block in the pipeline builder
type ProcessorBlock struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Config      map[string]interface{} `json:"config"`
	InputPorts  []string               `json:"input_ports"`
	OutputPorts []string               `json:"output_ports"`
	Position    Position               `json:"position"`
}

// Position represents x,y coordinates for visual elements
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// CostAnalytics represents aggregated cost analytics
type CostAnalytics struct {
	Period                string                   `json:"period"`
	TotalCost             float64                  `json:"total_cost"`
	TotalSavings          float64                  `json:"total_savings"`
	SavingsPercent        float64                  `json:"savings_percent"`
	CostTrend             []CostDataPoint          `json:"cost_trend"`
	SavingsByPipeline     map[string]float64       `json:"savings_by_pipeline"`
	SavingsByService      map[string]float64       `json:"savings_by_service"`
	TopCostDrivers        []CostDriver             `json:"top_cost_drivers"`
	ProjectedMonthlyCost  float64                  `json:"projected_monthly_cost"`
	ProjectedSavings      float64                  `json:"projected_savings"`
}

// CostDataPoint represents a single point in cost trend
type CostDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Cost      float64   `json:"cost"`
	Savings   float64   `json:"savings"`
}

// CostDriver represents a major cost contributor
type CostDriver struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"` // metric, service, namespace
	Cost        float64 `json:"cost"`
	Percentage  float64 `json:"percentage"`
	Trend       string  `json:"trend"` // increasing, decreasing, stable
}

// TaskQueueStatus represents the current state of the task queue
type TaskQueueStatus struct {
	PendingTasks   int                    `json:"pending_tasks"`
	RunningTasks   int                    `json:"running_tasks"`
	CompletedTasks int                    `json:"completed_tasks"`
	FailedTasks    int                    `json:"failed_tasks"`
	TasksByType    map[string]int         `json:"tasks_by_type"`
	TasksByHost    map[string]int         `json:"tasks_by_host"`
	AverageWaitTime time.Duration         `json:"average_wait_time"`
	QueuedTasks    []websocket.TaskInfo   `json:"queued_tasks"`
}