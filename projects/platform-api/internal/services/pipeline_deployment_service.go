package services

import (
	"context"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	"github.com/phoenix/platform/pkg/common/models"
	"github.com/phoenix/platform/projects/platform-api/internal/store"
	"go.uber.org/zap"
)

type PipelineDeploymentService struct {
	store  store.PipelineDeploymentStore
	logger *zap.Logger
}

func NewPipelineDeploymentService(store store.PipelineDeploymentStore, logger *zap.Logger) *PipelineDeploymentService {
	return &PipelineDeploymentService{
		store:  store,
		logger: logger,
	}
}

func (s *PipelineDeploymentService) CreateDeployment(ctx context.Context, req *models.CreateDeploymentRequest) (*models.PipelineDeployment, error) {
	s.logger.Info("creating pipeline deployment",
		zap.String("deployment_name", req.DeploymentName),
		zap.String("pipeline_name", req.PipelineName),
		zap.String("namespace", req.Namespace))

	// Create deployment model
	deployment := &models.PipelineDeployment{
		ID:             uuid.New().String(),
		DeploymentName: req.DeploymentName,
		PipelineName:   req.PipelineName,
		Namespace:      req.Namespace,
		TargetNodes:    req.TargetNodes,
		Parameters:     req.Parameters,
		Resources:      req.Resources,
		Status:         models.DeploymentStatusPending,
		Phase:          models.DeploymentPhasePending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      req.CreatedBy,
	}

	// Initialize instances if replicas specified
	if req.Replicas > 0 {
		deployment.Instances = &models.DeploymentInstances{
			Desired: req.Replicas,
			Ready:   0,
			Updated: 0,
		}
	}

	// Save to store
	if err := s.store.CreateDeployment(ctx, deployment); err != nil {
		s.logger.Error("failed to create deployment", zap.Error(err))
		return nil, err
	}

	s.logger.Info("deployment created successfully", zap.String("id", deployment.ID))
	return deployment, nil
}

func (s *PipelineDeploymentService) ListDeployments(ctx context.Context, req *models.ListDeploymentsRequest) (*models.ListDeploymentsResponse, error) {
	s.logger.Info("listing pipeline deployments",
		zap.String("namespace", req.Namespace),
		zap.String("status", req.Status),
		zap.String("pipeline_name", req.PipelineName))

	// Set default page size if not specified
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	// Get deployments from store
	deployments, total, err := s.store.ListDeployments(ctx, req)
	if err != nil {
		s.logger.Error("failed to list deployments", zap.Error(err))
		return nil, err
	}

	// Build response
	response := &models.ListDeploymentsResponse{
		Deployments: deployments,
		Total:       total,
		Page:        req.Page,
		PerPage:     req.PageSize,
	}

	// Calculate next page token if there are more results
	if req.Page*req.PageSize < total {
		response.NextPageToken = fmt.Sprintf("%d", req.Page+1)
	}

	s.logger.Info("deployments listed successfully",
		zap.Int("count", len(deployments)),
		zap.Int("total", total))

	return response, nil
}

func (s *PipelineDeploymentService) GetDeployment(ctx context.Context, deploymentID string) (*models.PipelineDeployment, error) {
	s.logger.Info("getting pipeline deployment", zap.String("deployment_id", deploymentID))

	deployment, err := s.store.GetDeployment(ctx, deploymentID)
	if err != nil {
		s.logger.Error("failed to get deployment", zap.Error(err))
		return nil, err
	}

	return deployment, nil
}

func (s *PipelineDeploymentService) UpdateDeployment(ctx context.Context, deploymentID string, req *models.UpdateDeploymentRequest) error {
	s.logger.Info("updating pipeline deployment",
		zap.String("deployment_id", deploymentID),
		zap.String("status", req.Status),
		zap.String("phase", req.Phase))

	if err := s.store.UpdateDeployment(ctx, deploymentID, req); err != nil {
		s.logger.Error("failed to update deployment", zap.Error(err))
		return err
	}

	s.logger.Info("deployment updated successfully", zap.String("deployment_id", deploymentID))
	return nil
}

func (s *PipelineDeploymentService) DeleteDeployment(ctx context.Context, deploymentID string) error {
	s.logger.Info("deleting pipeline deployment", zap.String("deployment_id", deploymentID))

	if err := s.store.DeleteDeployment(ctx, deploymentID); err != nil {
		s.logger.Error("failed to delete deployment", zap.Error(err))
		return err
	}

	s.logger.Info("deployment deleted successfully", zap.String("deployment_id", deploymentID))
	return nil
}

// GetDeploymentStatus retrieves the current status of a deployment
func (s *PipelineDeploymentService) GetDeploymentStatus(ctx context.Context, deploymentID string) (*models.DeploymentStatus, error) {
	s.logger.Info("getting deployment status", zap.String("deployment_id", deploymentID))

	deployment, err := s.store.GetDeployment(ctx, deploymentID)
	if err != nil {
		s.logger.Error("failed to get deployment", zap.Error(err))
		return nil, err
	}

	// Build status response
	status := &models.DeploymentStatus{
		DeploymentID: deployment.ID,
		Status:       deployment.Status,
		Phase:        deployment.Phase,
		LastUpdated:  deployment.UpdatedAt,
	}

	// Add metrics if available
	if deployment.Metrics != nil {
		status.Metrics = deployment.Metrics
	}

	// Add instance information if available
	if deployment.Instances != nil {
		status.Instances = deployment.Instances
	}

	// Add instances information if available
	if deployment.Instances != nil {
		status.Instances = deployment.Instances
	}

	return status, nil
}

// RollbackDeployment rolls back a deployment to a previous version
func (s *PipelineDeploymentService) RollbackDeployment(ctx context.Context, deploymentID string, version int) error {
	s.logger.Info("rolling back deployment",
		zap.String("deployment_id", deploymentID),
		zap.Int("target_version", version))

	// Get current deployment to validate it exists
	_, err := s.store.GetDeployment(ctx, deploymentID)
	if err != nil {
		s.logger.Error("failed to get deployment", zap.Error(err))
		return err
	}

	// TODO: Implement version validation once Version field is added to PipelineDeployment model
	// For now, just validate that version is positive
	if version <= 0 {
		return fmt.Errorf("invalid version %d: must be positive", version)
	}

	// TODO: Implement deployment history retrieval once GetDeploymentHistory method is added to store
	// For now, we'll create a basic rollback update

	// Create rollback update request
	updateReq := &models.UpdateDeploymentRequest{
		Status: models.DeploymentStatusUpdating,
		Phase:  models.DeploymentPhaseUpdating,
		
		// TODO: Restore configuration from history once history functionality is implemented
		// For now, we'll just update the status to indicate rollback in progress
	}

	// Update deployment with rollback configuration
	if err := s.store.UpdateDeployment(ctx, deploymentID, updateReq); err != nil {
		s.logger.Error("failed to update deployment for rollback", zap.Error(err))
		return err
	}

	// TODO: Record rollback event once DeploymentEvent model and store methods are implemented
	s.logger.Debug("rollback event would be recorded here", 
		zap.String("deployment_id", deploymentID),
		zap.Int("target_version", version))

	s.logger.Info("deployment rollback initiated successfully",
		zap.String("deployment_id", deploymentID),
		zap.Int("target_version", version))

	return nil
}

// UpdateDeploymentMetrics updates the metrics for a deployment
func (s *PipelineDeploymentService) UpdateDeploymentMetrics(ctx context.Context, deploymentID string, metrics *models.DeploymentMetrics) error {
	s.logger.Info("updating deployment metrics", zap.String("deployment_id", deploymentID))

	// TODO: Update metrics in store once UpdateDeploymentMetrics method is implemented
	s.logger.Debug("would update deployment metrics here", 
		zap.String("deployment_id", deploymentID),
		zap.Float64("cardinality", float64(metrics.Cardinality)))

	return nil
}

// TODO: UpdateDeploymentHealth will be implemented once DeploymentHealth model is defined