package interfaces

import (
	"context"
	"time"
)

// EventBus defines the interface for event-driven communication between services
// This enables asynchronous, decoupled communication in the Phoenix platform
type EventBus interface {
	// Publish sends an event to the event bus
	Publish(ctx context.Context, event Event) error
	
	// Subscribe creates a subscription to events matching the filter
	Subscribe(ctx context.Context, filter EventFilter) (<-chan Event, error)
	
	// Unsubscribe removes a subscription
	Unsubscribe(ctx context.Context, subscriptionID string) error
	
	// PublishBatch sends multiple events atomically
	PublishBatch(ctx context.Context, events []Event) error
}

// Event represents a domain event in the system
type Event interface {
	// GetID returns the unique event ID
	GetID() string
	
	// GetType returns the event type
	GetType() string
	
	// GetSource returns the source service that generated the event
	GetSource() string
	
	// GetTimestamp returns when the event occurred
	GetTimestamp() time.Time
	
	// GetData returns the event payload
	GetData() interface{}
	
	// GetMetadata returns event metadata
	GetMetadata() map[string]string
}

// BaseEvent provides a standard implementation of the Event interface
type BaseEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      interface{}            `json:"data"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
}

func (e *BaseEvent) GetID() string                    { return e.ID }
func (e *BaseEvent) GetType() string                  { return e.Type }
func (e *BaseEvent) GetSource() string                { return e.Source }
func (e *BaseEvent) GetTimestamp() time.Time          { return e.Timestamp }
func (e *BaseEvent) GetData() interface{}             { return e.Data }
func (e *BaseEvent) GetMetadata() map[string]string   { return e.Metadata }

// EventFilter defines criteria for filtering events
type EventFilter struct {
	Types      []string          `json:"types,omitempty"`
	Sources    []string          `json:"sources,omitempty"`
	StartTime  *time.Time        `json:"start_time,omitempty"`
	EndTime    *time.Time        `json:"end_time,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Common event types in the Phoenix platform
const (
	// Experiment events
	EventTypeExperimentCreated     = "experiment.created"
	EventTypeExperimentStarted     = "experiment.started"
	EventTypeExperimentCompleted   = "experiment.completed"
	EventTypeExperimentFailed      = "experiment.failed"
	EventTypeExperimentCancelled   = "experiment.cancelled"
	EventTypeExperimentStateChange = "experiment.state_changed"
	
	// Pipeline events
	EventTypePipelineCreated       = "pipeline.created"
	EventTypePipelineDeployed      = "pipeline.deployed"
	EventTypePipelineDeleted       = "pipeline.deleted"
	EventTypePipelineStatusChanged = "pipeline.status_changed"
	
	// Configuration events
	EventTypeConfigGenerated       = "config.generated"
	EventTypeConfigValidated       = "config.validated"
	EventTypeConfigDeployed        = "config.deployed"
	
	// Metrics events
	EventTypeMetricsCollected      = "metrics.collected"
	EventTypeAnomalyDetected       = "metrics.anomaly_detected"
	EventTypeThresholdExceeded     = "metrics.threshold_exceeded"
	
	// System events
	EventTypeServiceStarted        = "service.started"
	EventTypeServiceStopped        = "service.stopped"
	EventTypeHealthCheckFailed     = "service.health_check_failed"
)

// Event payload types
type ExperimentCreatedEvent struct {
	ExperimentID string            `json:"experiment_id"`
	Name         string            `json:"name"`
	CreatedBy    string            `json:"created_by"`
	Config       *ExperimentConfig `json:"config"`
}

type ExperimentStateChangedEvent struct {
	ExperimentID string          `json:"experiment_id"`
	FromState    ExperimentState `json:"from_state"`
	ToState      ExperimentState `json:"to_state"`
	Reason       string          `json:"reason,omitempty"`
}

type PipelineDeployedEvent struct {
	PipelineID     string   `json:"pipeline_id"`
	ExperimentID   string   `json:"experiment_id,omitempty"`
	NodeCount      int      `json:"node_count"`
	DeploymentType string   `json:"deployment_type"`
}

type MetricsCollectedEvent struct {
	ExperimentID  string                 `json:"experiment_id"`
	PipelineType  string                 `json:"pipeline_type"`
	MetricsSummary *MetricsSummary       `json:"metrics_summary"`
}

type AnomalyDetectedEvent struct {
	ExperimentID string  `json:"experiment_id"`
	AnomalyType  string  `json:"anomaly_type"`
	Severity     string  `json:"severity"`
	Description  string  `json:"description"`
	Value        float64 `json:"value"`
	Threshold    float64 `json:"threshold"`
}

// EventHandler defines a function that processes events
type EventHandler func(ctx context.Context, event Event) error

// EventProcessor provides event processing capabilities
type EventProcessor interface {
	// RegisterHandler registers a handler for specific event types
	RegisterHandler(eventType string, handler EventHandler) error
	
	// UnregisterHandler removes a handler
	UnregisterHandler(eventType string) error
	
	// ProcessEvent processes an event through registered handlers
	ProcessEvent(ctx context.Context, event Event) error
	
	// Start begins processing events
	Start(ctx context.Context) error
	
	// Stop stops processing events
	Stop(ctx context.Context) error
}

// Workflow orchestrates complex multi-step processes
type Workflow interface {
	// GetID returns the workflow ID
	GetID() string
	
	// GetState returns the current workflow state
	GetState() WorkflowState
	
	// Execute runs the workflow
	Execute(ctx context.Context) error
	
	// GetResult returns the workflow result
	GetResult() (interface{}, error)
	
	// Cancel cancels the workflow
	Cancel(ctx context.Context) error
}

// WorkflowState represents the state of a workflow
type WorkflowState string

const (
	WorkflowStatePending   WorkflowState = "pending"
	WorkflowStateRunning   WorkflowState = "running"
	WorkflowStateCompleted WorkflowState = "completed"
	WorkflowStateFailed    WorkflowState = "failed"
	WorkflowStateCancelled WorkflowState = "cancelled"
)

// WorkflowEngine manages workflow execution
type WorkflowEngine interface {
	// CreateWorkflow creates a new workflow instance
	CreateWorkflow(ctx context.Context, definition *WorkflowDefinition) (Workflow, error)
	
	// GetWorkflow retrieves a workflow by ID
	GetWorkflow(ctx context.Context, id string) (Workflow, error)
	
	// ListWorkflows lists workflows matching the filter
	ListWorkflows(ctx context.Context, filter *WorkflowFilter) ([]*WorkflowInfo, error)
}

// WorkflowDefinition defines a workflow
type WorkflowDefinition struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Steps       []*WorkflowStep        `json:"steps"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
	RetryPolicy *RetryPolicy           `json:"retry_policy,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Config       map[string]interface{} `json:"config"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Timeout      time.Duration          `json:"timeout,omitempty"`
	RetryPolicy  *RetryPolicy           `json:"retry_policy,omitempty"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts int           `json:"max_attempts"`
	Backoff     BackoffPolicy `json:"backoff"`
}

// BackoffPolicy defines backoff strategy
type BackoffPolicy struct {
	Type       string        `json:"type"` // constant, linear, exponential
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay,omitempty"`
	Multiplier   float64       `json:"multiplier,omitempty"`
}

// WorkflowFilter filters workflows
type WorkflowFilter struct {
	States    []WorkflowState `json:"states,omitempty"`
	StartTime *time.Time      `json:"start_time,omitempty"`
	EndTime   *time.Time      `json:"end_time,omitempty"`
	PageSize  int             `json:"page_size,omitempty"`
	PageToken string          `json:"page_token,omitempty"`
}

// WorkflowInfo provides workflow metadata
type WorkflowInfo struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	State      WorkflowState          `json:"state"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}