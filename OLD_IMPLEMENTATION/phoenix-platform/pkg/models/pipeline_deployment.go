package models

import (
	"time"
)

// Pipeline deployment status constants
const (
	DeploymentStatusPending  = "pending"
	DeploymentStatusActive   = "active"
	DeploymentStatusUpdating = "updating"
	DeploymentStatusDeleting = "deleting"
	DeploymentStatusFailed   = "failed"
)

// Pipeline deployment phase constants
const (
	DeploymentPhasePending     = "pending"
	DeploymentPhaseDeploying   = "deploying"
	DeploymentPhaseRunning     = "running"
	DeploymentPhaseUpdating    = "updating"
	DeploymentPhaseTerminating = "terminating"
	DeploymentPhaseFailed      = "failed"
)

// PipelineDeployment represents a direct pipeline deployment
type PipelineDeployment struct {
	ID             string                 `json:"id"`
	DeploymentName string                 `json:"deployment_name"`
	PipelineName   string                 `json:"pipeline_name"`
	Namespace      string                 `json:"namespace"`
	TargetNodes    map[string]string      `json:"target_nodes"`
	Parameters     map[string]interface{} `json:"parameters"`
	Resources      *ResourceRequirements  `json:"resources,omitempty"`
	Status         string                 `json:"status"`
	Phase          string                 `json:"phase"`
	Instances      *DeploymentInstances   `json:"instances,omitempty"`
	Metrics        *DeploymentMetrics     `json:"metrics,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	DeletedAt      *time.Time             `json:"deleted_at,omitempty"`
	CreatedBy      string                 `json:"created_by,omitempty"`
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

// DeploymentInstances tracks deployment instance counts
type DeploymentInstances struct {
	Desired int `json:"desired"`
	Ready   int `json:"ready"`
	Updated int `json:"updated"`
}

// DeploymentMetrics contains the latest metrics for a deployment
type DeploymentMetrics struct {
	Cardinality    int64   `json:"cardinality"`
	Throughput     string  `json:"throughput"`
	ErrorRate      float64 `json:"error_rate"`
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    float64 `json:"memory_usage"`
	LastUpdated    time.Time `json:"last_updated"`
}

// CreateDeploymentRequest represents a request to create a pipeline deployment
type CreateDeploymentRequest struct {
	DeploymentName string                 `json:"deployment_name"`
	PipelineName   string                 `json:"pipeline_name"`
	Namespace      string                 `json:"namespace"`
	TargetNodes    map[string]string      `json:"target_nodes"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	Resources      *ResourceRequirements  `json:"resources,omitempty"`
	Replicas       int                    `json:"replicas,omitempty"`
	CreatedBy      string                 `json:"created_by,omitempty"`
}

// UpdateDeploymentRequest represents a request to update a deployment
type UpdateDeploymentRequest struct {
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Resources  *ResourceRequirements  `json:"resources,omitempty"`
	Status     string                 `json:"status,omitempty"`
	Phase      string                 `json:"phase,omitempty"`
}

// ListDeploymentsRequest represents a request to list deployments
type ListDeploymentsRequest struct {
	Namespace    string `json:"namespace,omitempty"`
	Status       string `json:"status,omitempty"`
	PipelineName string `json:"pipeline_name,omitempty"`
	PageSize     int    `json:"page_size,omitempty"`
	Page         int    `json:"page,omitempty"`
	PageToken    string `json:"page_token,omitempty"`
}

// ListDeploymentsResponse represents a response containing deployments
type ListDeploymentsResponse struct {
	Deployments   []*PipelineDeployment `json:"deployments"`
	Total         int                   `json:"total"`
	Page          int                   `json:"page"`
	PerPage       int                   `json:"per_page"`
	NextPageToken string                `json:"next_page_token,omitempty"`
}