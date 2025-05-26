package websocket

import (
	"encoding/json"
	"time"
)

// EventType represents the type of WebSocket event
type EventType string

const (
	// Event types for real-time updates
	EventAgentStatus      EventType = "agent_status"
	EventExperimentUpdate EventType = "experiment_update"
	EventMetricFlow       EventType = "metric_flow"
	EventTaskProgress     EventType = "task_progress"
	EventCostUpdate       EventType = "cost_update"
	EventAlert            EventType = "alert"
	EventPipelineStatus   EventType = "pipeline_status"
)

// Event represents a WebSocket event
type Event struct {
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// AgentStatusUpdate represents real-time agent state
type AgentStatusUpdate struct {
	HostID          string       `json:"host_id"`
	Status          string       `json:"status"` // healthy, updating, offline
	ActiveTasks     []TaskInfo   `json:"active_tasks"`
	Metrics         AgentMetrics `json:"metrics"`
	CostSavings     float64      `json:"cost_savings"`
	LastHeartbeat   time.Time    `json:"last_heartbeat"`
	Location        *Location    `json:"location,omitempty"`
}

// TaskInfo represents a task being executed by an agent
type TaskInfo struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Progress   int    `json:"progress"`
	StartedAt  time.Time `json:"started_at"`
}

// AgentMetrics represents real-time agent performance metrics
type AgentMetrics struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryMB      int64   `json:"memory_mb"`
	MetricsPerSec int64   `json:"metrics_per_sec"`
	DroppedCount  int64   `json:"dropped_count"`
}

// Location represents agent geographical location for map view
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Region    string  `json:"region"`
	Zone      string  `json:"zone"`
}

// MetricFlowUpdate represents real-time metric costs
type MetricFlowUpdate struct {
	Timestamp     time.Time              `json:"timestamp"`
	TotalCostRate float64                `json:"total_cost_rate"` // cost per minute
	TopMetrics    []MetricCostBreakdown  `json:"top_metrics"`
	ByService     map[string]float64     `json:"by_service"`
	ByNamespace   map[string]float64     `json:"by_namespace"`
}

// MetricCostBreakdown represents cost breakdown for a single metric
type MetricCostBreakdown struct {
	MetricName    string  `json:"metric_name"`
	CostPerMinute float64 `json:"cost_per_minute"`
	Cardinality   int64   `json:"cardinality"`
	Percentage    float64 `json:"percentage"`
	Labels        map[string]string `json:"labels"`
}

// ExperimentUpdateEvent represents experiment state changes
type ExperimentUpdateEvent struct {
	ExperimentID string            `json:"experiment_id"`
	Name         string            `json:"name"`
	Status       string            `json:"status"`
	Progress     int               `json:"progress"`
	Metrics      ExperimentMetrics `json:"metrics"`
	Actions      []string          `json:"available_actions"`
}

// ExperimentMetrics represents real-time experiment KPIs
type ExperimentMetrics struct {
	BaselineCost    float64 `json:"baseline_cost"`
	CandidateCost   float64 `json:"candidate_cost"`
	SavingsPercent  float64 `json:"savings_percent"`
	CoveragePercent float64 `json:"coverage_percent"`
	LatencyImpactMs float64 `json:"latency_impact_ms"`
	CPUImpact       float64 `json:"cpu_impact"`
}

// TaskProgressUpdate represents task execution progress
type TaskProgressUpdate struct {
	TaskID         string        `json:"task_id"`
	Type           string        `json:"type"`
	Description    string        `json:"description"`
	Progress       int           `json:"progress"`
	TotalHosts     int           `json:"total_hosts"`
	CompletedHosts int           `json:"completed_hosts"`
	FailedHosts    int           `json:"failed_hosts"`
	ETA            time.Duration `json:"eta"`
	Status         string        `json:"status"`
}

// AlertEvent represents system alerts
type AlertEvent struct {
	ID          string    `json:"id"`
	Severity    string    `json:"severity"` // info, warning, error, critical
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	Source      string    `json:"source"`
	Timestamp   time.Time `json:"timestamp"`
	ActionItems []string  `json:"action_items,omitempty"`
}

// Message represents a WebSocket message from client
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Subscription represents a client subscription request
type Subscription struct {
	Events    []EventType `json:"events"`
	Filters   Filters     `json:"filters"`
}

// Filters for event subscriptions
type Filters struct {
	Experiments []string `json:"experiments,omitempty"`
	Hosts       []string `json:"hosts,omitempty"`
	Services    []string `json:"services,omitempty"`
	Namespaces  []string `json:"namespaces,omitempty"`
}