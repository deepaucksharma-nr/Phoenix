package store

import (
	"context"
	
	"github.com/phoenix/platform/pkg/common/models"
)

// PipelineDeploymentStore defines the interface for pipeline deployment storage
type PipelineDeploymentStore interface {
	CreateDeployment(ctx context.Context, deployment *models.PipelineDeployment) error
	GetDeployment(ctx context.Context, deploymentID string) (*models.PipelineDeployment, error)
	ListDeployments(ctx context.Context, req *models.ListDeploymentsRequest) ([]*models.PipelineDeployment, int, error)
	UpdateDeployment(ctx context.Context, deploymentID string, update *models.UpdateDeploymentRequest) error
	DeleteDeployment(ctx context.Context, deploymentID string) error
}