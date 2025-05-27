package store

import (
	"context"
	"time"

	"github.com/phoenix/platform/pkg/common/models"
	internalModels "github.com/phoenix/platform/projects/phoenix-api/internal/models"
)

// Store defines the complete interface for all storage operations
type Store interface {
	// Experiment operations
	CreateExperiment(ctx context.Context, experiment *internalModels.Experiment) error
	GetExperiment(ctx context.Context, experimentID string) (*internalModels.Experiment, error)
	ListExperiments(ctx context.Context) ([]*internalModels.Experiment, error)
	UpdateExperiment(ctx context.Context, experiment *internalModels.Experiment) error
	UpdateExperimentPhase(ctx context.Context, experimentID string, phase string) error
	DeleteExperiment(ctx context.Context, experimentID string) error

	// Pipeline deployment operations
	CreateDeployment(ctx context.Context, deployment *models.PipelineDeployment) error
	GetDeployment(ctx context.Context, deploymentID string) (*models.PipelineDeployment, error)
	ListDeployments(ctx context.Context, req *models.ListDeploymentsRequest) ([]*models.PipelineDeployment, int, error)
	UpdateDeployment(ctx context.Context, deploymentID string, update *models.UpdateDeploymentRequest) error
	DeleteDeployment(ctx context.Context, deploymentID string) error
	UpdateDeploymentMetrics(ctx context.Context, deploymentID string, metrics *models.DeploymentMetrics) error

	// Task operations
	CreateTask(ctx context.Context, task *internalModels.Task) error
	GetTask(ctx context.Context, taskID string) (*internalModels.Task, error)
	ListTasks(ctx context.Context, filters map[string]interface{}) ([]*internalModels.Task, error)
	UpdateTask(ctx context.Context, task *internalModels.Task) error
	GetPendingTasksForHost(ctx context.Context, hostID string) ([]*internalModels.Task, error)
	GetTasksByExperiment(ctx context.Context, experimentID string) ([]*internalModels.Task, error)
	GetTaskStats(ctx context.Context) (map[string]interface{}, error)
	GetStaleTasks(ctx context.Context, threshold time.Duration) ([]*internalModels.Task, error)
	DeleteOldTasks(ctx context.Context, before time.Time) error

	// Agent operations
	UpsertAgent(ctx context.Context, agent *internalModels.AgentStatus) error
	GetAgent(ctx context.Context, hostID string) (*internalModels.AgentStatus, error)
	ListAgents(ctx context.Context) ([]*internalModels.AgentStatus, error)
	UpdateAgentHeartbeat(ctx context.Context, heartbeat *internalModels.AgentHeartbeat) error
	CacheMetric(ctx context.Context, hostID string, metric map[string]interface{}) error

	// Event operations
	CreateExperimentEvent(ctx context.Context, event *internalModels.ExperimentEvent) error
	ListExperimentEvents(ctx context.Context, experimentID string) ([]*internalModels.ExperimentEvent, error)
	
	// Metrics operations
	GetExperimentMetrics(ctx context.Context, experimentID string) (map[string]interface{}, error)
	
	// UI-specific operations
	GetMetricCostFlow(ctx context.Context) (map[string]interface{}, error)
	GetCardinalityBreakdown(ctx context.Context, namespace, service string) (map[string]interface{}, error)
	GetAllAgents(ctx context.Context) ([]*internalModels.AgentStatus, error)
	GetAgentsWithLocation(ctx context.Context) (map[string]interface{}, error)
	GetActiveTasks(ctx context.Context, status, hostID string, limit int) ([]*internalModels.Task, error)
	GetCostAnalytics(ctx context.Context, period string) (map[string]interface{}, error)
	GetTaskQueueStatus(ctx context.Context) (map[string]interface{}, error)
	GetPipelineTemplates(ctx context.Context) ([]*PipelineTemplate, error)
	
	// User operations
	CreateUser(ctx context.Context, user *internalModels.User) error
	GetUser(ctx context.Context, userID string) (*internalModels.User, error)
	GetUserByUsername(ctx context.Context, username string) (*internalModels.User, error)
	UpdateUserLastLogin(ctx context.Context, userID string) error
	
	// Deployment versioning operations
	RecordDeploymentVersion(ctx context.Context, deploymentID, pipelineConfig string, parameters map[string]interface{}, deployedBy string, notes string) (int, error)
	GetDeploymentVersion(ctx context.Context, deploymentID string, version int) (*DeploymentVersion, error)
	ListDeploymentVersions(ctx context.Context, deploymentID string) ([]*DeploymentVersion, error)
	GetDeploymentHistory(ctx context.Context, deploymentID string, version int) (*models.PipelineDeployment, error)
	
	// Token blacklist operations
	BlacklistToken(ctx context.Context, jti, userID string, expiresAt time.Time, reason string) error
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)
	CleanupExpiredTokens(ctx context.Context) error
}

// PipelineDeploymentStore defines the interface for pipeline deployment storage
type PipelineDeploymentStore interface {
	CreateDeployment(ctx context.Context, deployment *models.PipelineDeployment) error
	GetDeployment(ctx context.Context, deploymentID string) (*models.PipelineDeployment, error)
	ListDeployments(ctx context.Context, req *models.ListDeploymentsRequest) ([]*models.PipelineDeployment, int, error)
	UpdateDeployment(ctx context.Context, deploymentID string, update *models.UpdateDeploymentRequest) error
	DeleteDeployment(ctx context.Context, deploymentID string) error
	UpdateDeploymentMetrics(ctx context.Context, deploymentID string, metrics *models.DeploymentMetrics) error
	GetDeploymentHistory(ctx context.Context, deploymentID string, version int) (*models.PipelineDeployment, error)
}
