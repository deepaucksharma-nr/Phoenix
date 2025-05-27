package interfaces

import (
	"context"
	"time"
)

// PipelineService defines the interface for pipeline management operations
// This interface is consumed by the API and implemented by the Pipeline Manager
type PipelineService interface {
	// CreatePipeline creates a new pipeline configuration
	CreatePipeline(ctx context.Context, req *CreatePipelineRequest) (*Pipeline, error)
	
	// GetPipeline retrieves a pipeline by ID
	GetPipeline(ctx context.Context, id string) (*Pipeline, error)
	
	// ListPipelines returns available pipelines
	ListPipelines(ctx context.Context, filter *PipelineFilter) (*PipelineList, error)
	
	// ValidatePipeline validates a pipeline configuration
	ValidatePipeline(ctx context.Context, config *PipelineConfig) (*ValidationResult, error)
	
	// DeployPipeline deploys a pipeline to target nodes
	DeployPipeline(ctx context.Context, id string, targets []string) (*DeploymentStatus, error)
	
	// GetPipelineStatus retrieves the deployment status of a pipeline
	GetPipelineStatus(ctx context.Context, id string) (*DeploymentStatus, error)
	
	// DeletePipeline removes a pipeline and its deployments
	DeletePipeline(ctx context.Context, id string) error
}

// ConfigGenerator defines the interface for generating OTel configurations
// This interface is implemented by the Config Generator service
type ConfigGenerator interface {
	// GenerateConfig creates an OTel configuration from a pipeline definition
	GenerateConfig(ctx context.Context, pipeline *Pipeline) (*GeneratedConfig, error)
	
	// OptimizeConfig applies optimization strategies to a configuration
	OptimizeConfig(ctx context.Context, config *PipelineConfig, strategy OptimizationStrategy) (*PipelineConfig, error)
	
	// ValidateConfig checks if a configuration is valid
	ValidateConfig(ctx context.Context, config string) (*ValidationResult, error)
	
	// GetTemplates returns available pipeline templates
	GetTemplates(ctx context.Context) ([]*PipelineTemplate, error)
	
	// ApplyTemplate applies a template with variables
	ApplyTemplate(ctx context.Context, templateID string, variables map[string]interface{}) (*PipelineConfig, error)
}

// PipelineOperator defines the interface for pipeline deployment operations
// This interface represents the contract for deploying pipelines to agents
type PipelineOperator interface {
	// DeployPipeline deploys a pipeline configuration to target agents
	DeployPipeline(ctx context.Context, deployment *PipelineDeployment) error
	
	// UpdatePipeline updates an existing pipeline deployment
	UpdatePipeline(ctx context.Context, deployment *PipelineDeployment) error
	
	// DeletePipeline removes a pipeline deployment
	DeletePipeline(ctx context.Context, deploymentID string) error
	
	// GetPipelineDeployment retrieves a pipeline deployment
	GetPipelineDeployment(ctx context.Context, deploymentID string) (*PipelineDeployment, error)
	
	// ListPipelineDeployments lists pipeline deployments
	ListPipelineDeployments(ctx context.Context, filter map[string]string) ([]*PipelineDeployment, error)
	
	// GetDeploymentStatus retrieves the status of a pipeline deployment
	GetDeploymentStatus(ctx context.Context, deploymentID string) (*PipelineStatus, error)
}

// Pipeline represents a complete pipeline configuration
type Pipeline struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        PipelineType           `json:"type"`
	Config      *PipelineConfig        `json:"config"`
	Template    string                 `json:"template,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	Version     string                 `json:"version"`
}

// PipelineType defines the type of pipeline
type PipelineType string

const (
	PipelineTypeProcess PipelineType = "process"
	PipelineTypeTrace   PipelineType = "trace"
	PipelineTypeLog     PipelineType = "log"
	PipelineTypeCustom  PipelineType = "custom"
)

// PipelineConfig contains the OpenTelemetry collector configuration
type PipelineConfig struct {
	Receivers  map[string]interface{} `json:"receivers"`
	Processors map[string]interface{} `json:"processors"`
	Exporters  map[string]interface{} `json:"exporters"`
	Service    *ServiceConfig         `json:"service"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// ServiceConfig defines the service pipelines configuration
type ServiceConfig struct {
	Extensions []string                        `json:"extensions,omitempty"`
	Pipelines  map[string]*ServicePipeline     `json:"pipelines"`
	Telemetry  map[string]interface{}          `json:"telemetry,omitempty"`
}

// ServicePipeline defines a single pipeline within the service
type ServicePipeline struct {
	Receivers  []string `json:"receivers"`
	Processors []string `json:"processors,omitempty"`
	Exporters  []string `json:"exporters"`
}

// PipelineTemplate represents a pre-defined pipeline template
type PipelineTemplate struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	Type              PipelineType           `json:"type"`
	Category          string                 `json:"category"`
	OptimizationLevel string                 `json:"optimization_level"`
	Variables         []*TemplateVariable    `json:"variables"`
	Config            *PipelineConfig        `json:"config"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// TemplateVariable defines a variable in a pipeline template
type TemplateVariable struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Type         string      `json:"type"`
	Default      interface{} `json:"default,omitempty"`
	Required     bool        `json:"required"`
	ValidValues  []string    `json:"valid_values,omitempty"`
}

// PipelineDeployment represents a pipeline deployment configuration
type PipelineDeployment struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	TargetEnv         string                 `json:"target_env"`
	TargetNodes       map[string]string      `json:"target_nodes,omitempty"`
	PipelineRef       string                 `json:"pipeline_ref"`
	ConfigOverrides   map[string]interface{} `json:"config_overrides,omitempty"`
	ConfigVariables   map[string]string      `json:"config_variables,omitempty"`
	Resources         *ResourceSpec          `json:"resources,omitempty"`
}

// PipelineStatus represents the deployment status of a pipeline
type PipelineStatus struct {
	Phase              string                 `json:"phase"`
	ObservedGeneration int64                  `json:"observedGeneration"`
	Conditions         []PipelineCondition    `json:"conditions"`
	CollectorStatus    *CollectorStatus       `json:"collectorStatus,omitempty"`
	LastUpdated        time.Time              `json:"lastUpdated"`
}

// PipelineCondition represents a condition of the pipeline
type PipelineCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
}

// CollectorStatus contains status information about deployed collectors
type CollectorStatus struct {
	DesiredInstances   int32                  `json:"desired_instances"`
	RunningInstances   int32                  `json:"running_instances"`
	HealthyInstances   int32                  `json:"healthy_instances"`
	FailedInstances    int32                  `json:"failed_instances"`
	AgentStatus        map[string]string      `json:"agent_status,omitempty"`
}

// DeploymentStatus represents the overall deployment status
type DeploymentStatus struct {
	PipelineID   string                    `json:"pipeline_id"`
	Phase        string                    `json:"phase"`
	Message      string                    `json:"message,omitempty"`
	NodeStatuses map[string]*NodeStatus    `json:"node_statuses"`
	StartedAt    time.Time                 `json:"started_at"`
	CompletedAt  *time.Time                `json:"completed_at,omitempty"`
}

// NodeStatus represents the deployment status on a specific node
type NodeStatus struct {
	NodeName    string    `json:"node_name"`
	AgentID     string    `json:"agent_id"`
	Phase       string    `json:"phase"`
	Message     string    `json:"message,omitempty"`
	CollectorID string    `json:"collector_id,omitempty"`
	LastUpdated time.Time `json:"last_updated"`
}

// ResourceSpec defines resource requirements for collectors
type ResourceSpec struct {
	CPURequest    string `json:"cpu_request,omitempty"`
	CPULimit      string `json:"cpu_limit,omitempty"`
	MemoryRequest string `json:"memory_request,omitempty"`
	MemoryLimit   string `json:"memory_limit,omitempty"`
}

// GeneratedConfig represents the output of config generation
type GeneratedConfig struct {
	YAMLConfig     string                 `json:"yaml_config"`
	ConfigChecksum string                 `json:"config_checksum"`
	Optimizations  []string               `json:"optimizations_applied,omitempty"`
	Warnings       []string               `json:"warnings,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationResult represents the result of configuration validation
type ValidationResult struct {
	Valid   bool               `json:"valid"`
	Errors  []ValidationError  `json:"errors,omitempty"`
	Warnings []ValidationWarning `json:"warnings,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Path    string `json:"path"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// OptimizationStrategy defines the optimization approach
type OptimizationStrategy string

const (
	OptimizationStrategyBalanced   OptimizationStrategy = "balanced"
	OptimizationStrategyAggressive OptimizationStrategy = "aggressive"
	OptimizationStrategyMinimal    OptimizationStrategy = "minimal"
	OptimizationStrategyCustom     OptimizationStrategy = "custom"
)

// Request/Response types
type CreatePipelineRequest struct {
	Name        string                 `json:"name" validate:"required,min=3,max=100"`
	Description string                 `json:"description" validate:"max=500"`
	Type        PipelineType           `json:"type" validate:"required"`
	Template    string                 `json:"template,omitempty"`
	Config      *PipelineConfig        `json:"config,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
}

type PipelineFilter struct {
	Type      PipelineType `json:"type,omitempty"`
	Template  string       `json:"template,omitempty"`
	CreatedBy string       `json:"created_by,omitempty"`
	PageSize  int          `json:"page_size,omitempty"`
	PageToken string       `json:"page_token,omitempty"`
}

type PipelineList struct {
	Pipelines     []*Pipeline `json:"pipelines"`
	NextPageToken string      `json:"next_page_token,omitempty"`
	TotalCount    int         `json:"total_count"`
}