package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/phoenix/platform/pkg/models"
)

// PipelineDeploymentService handles pipeline deployment operations
type PipelineDeploymentService struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPipelineDeploymentService creates a new pipeline deployment service
func NewPipelineDeploymentService(db *sql.DB, logger *zap.Logger) *PipelineDeploymentService {
	return &PipelineDeploymentService{
		db:     db,
		logger: logger,
	}
}

// CreateDeployment creates a new pipeline deployment
func (s *PipelineDeploymentService) CreateDeployment(ctx context.Context, req *models.CreateDeploymentRequest) (*models.PipelineDeployment, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	// Check for existing deployment with same name in namespace
	exists, err := s.deploymentExists(ctx, req.DeploymentName, req.Namespace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check existing deployment: %v", err)
	}
	if exists {
		return nil, status.Errorf(codes.AlreadyExists, "deployment %s already exists in namespace %s", req.DeploymentName, req.Namespace)
	}

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

	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert deployment
	query := `
		INSERT INTO pipeline_deployments (
			id, deployment_name, pipeline_name, namespace,
			target_nodes, parameters, resources, status, phase,
			created_at, updated_at, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)`

	targetNodesJSON, _ := json.Marshal(deployment.TargetNodes)
	parametersJSON, _ := json.Marshal(deployment.Parameters)
	resourcesJSON, _ := json.Marshal(deployment.Resources)

	_, err = tx.ExecContext(ctx, query,
		deployment.ID, deployment.DeploymentName, deployment.PipelineName,
		deployment.Namespace, targetNodesJSON, parametersJSON, resourcesJSON,
		deployment.Status, deployment.Phase, deployment.CreatedAt,
		deployment.UpdatedAt, deployment.CreatedBy,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create deployment: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	s.logger.Info("pipeline deployment created",
		zap.String("deployment_id", deployment.ID),
		zap.String("deployment_name", deployment.DeploymentName),
		zap.String("pipeline", deployment.PipelineName),
	)

	return deployment, nil
}

// ListDeployments lists pipeline deployments with optional filters
func (s *PipelineDeploymentService) ListDeployments(ctx context.Context, req *models.ListDeploymentsRequest) (*models.ListDeploymentsResponse, error) {
	query := `
		SELECT id, deployment_name, pipeline_name, namespace,
		       target_nodes, parameters, resources, status, phase,
		       instances, metrics, created_at, updated_at, created_by
		FROM pipeline_deployments
		WHERE deleted_at IS NULL`

	args := []interface{}{}
	argPos := 1

	// Add filters
	if req.Namespace != "" {
		query += fmt.Sprintf(" AND namespace = $%d", argPos)
		args = append(args, req.Namespace)
		argPos++
	}

	if req.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, req.Status)
		argPos++
	}

	if req.PipelineName != "" {
		query += fmt.Sprintf(" AND pipeline_name = $%d", argPos)
		args = append(args, req.PipelineName)
		argPos++
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

	// Add pagination
	if req.PageSize > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, req.PageSize)
		argPos++
	}

	if req.PageToken != "" {
		// Decode page token (assuming it's a timestamp)
		query += fmt.Sprintf(" AND created_at < $%d", argPos)
		args = append(args, req.PageToken)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query deployments: %v", err)
	}
	defer rows.Close()

	deployments := []*models.PipelineDeployment{}
	for rows.Next() {
		deployment := &models.PipelineDeployment{}
		var targetNodesJSON, parametersJSON, resourcesJSON, instancesJSON, metricsJSON []byte
		var createdBy sql.NullString

		err := rows.Scan(
			&deployment.ID, &deployment.DeploymentName, &deployment.PipelineName,
			&deployment.Namespace, &targetNodesJSON, &parametersJSON, &resourcesJSON,
			&deployment.Status, &deployment.Phase, &instancesJSON, &metricsJSON,
			&deployment.CreatedAt, &deployment.UpdatedAt, &createdBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan deployment: %v", err)
		}

		// Unmarshal JSON fields
		json.Unmarshal(targetNodesJSON, &deployment.TargetNodes)
		json.Unmarshal(parametersJSON, &deployment.Parameters)
		json.Unmarshal(resourcesJSON, &deployment.Resources)
		json.Unmarshal(instancesJSON, &deployment.Instances)
		json.Unmarshal(metricsJSON, &deployment.Metrics)
		deployment.CreatedBy = createdBy.String

		deployments = append(deployments, deployment)
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM pipeline_deployments WHERE deleted_at IS NULL`
	if req.Namespace != "" {
		countQuery += " AND namespace = '" + req.Namespace + "'"
	}
	if req.Status != "" {
		countQuery += " AND status = '" + req.Status + "'"
	}

	var totalCount int
	err = s.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		s.logger.Warn("failed to get total count", zap.Error(err))
	}

	response := &models.ListDeploymentsResponse{
		Deployments: deployments,
		Total:       totalCount,
		Page:        req.Page,
		PerPage:     req.PageSize,
	}

	// Set next page token if there are more results
	if len(deployments) == req.PageSize && len(deployments) > 0 {
		lastDeployment := deployments[len(deployments)-1]
		response.NextPageToken = lastDeployment.CreatedAt.Format(time.RFC3339Nano)
	}

	return response, nil
}

// GetDeployment retrieves a single deployment by ID
func (s *PipelineDeploymentService) GetDeployment(ctx context.Context, deploymentID string) (*models.PipelineDeployment, error) {
	query := `
		SELECT id, deployment_name, pipeline_name, namespace,
		       target_nodes, parameters, resources, status, phase,
		       instances, metrics, created_at, updated_at, created_by
		FROM pipeline_deployments
		WHERE id = $1 AND deleted_at IS NULL`

	deployment := &models.PipelineDeployment{}
	var targetNodesJSON, parametersJSON, resourcesJSON, instancesJSON, metricsJSON []byte
	var createdBy sql.NullString

	err := s.db.QueryRowContext(ctx, query, deploymentID).Scan(
		&deployment.ID, &deployment.DeploymentName, &deployment.PipelineName,
		&deployment.Namespace, &targetNodesJSON, &parametersJSON, &resourcesJSON,
		&deployment.Status, &deployment.Phase, &instancesJSON, &metricsJSON,
		&deployment.CreatedAt, &deployment.UpdatedAt, &createdBy,
	)
	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "deployment not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get deployment: %v", err)
	}

	// Unmarshal JSON fields
	json.Unmarshal(targetNodesJSON, &deployment.TargetNodes)
	json.Unmarshal(parametersJSON, &deployment.Parameters)
	json.Unmarshal(resourcesJSON, &deployment.Resources)
	json.Unmarshal(instancesJSON, &deployment.Instances)
	json.Unmarshal(metricsJSON, &deployment.Metrics)
	deployment.CreatedBy = createdBy.String

	return deployment, nil
}

// UpdateDeployment updates a deployment
func (s *PipelineDeploymentService) UpdateDeployment(ctx context.Context, deploymentID string, updates *models.UpdateDeploymentRequest) error {
	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Build update query dynamically
	query := "UPDATE pipeline_deployments SET updated_at = NOW()"
	args := []interface{}{}
	argPos := 1

	if updates.Parameters != nil {
		parametersJSON, _ := json.Marshal(updates.Parameters)
		query += fmt.Sprintf(", parameters = $%d", argPos)
		args = append(args, parametersJSON)
		argPos++
	}

	if updates.Resources != nil {
		resourcesJSON, _ := json.Marshal(updates.Resources)
		query += fmt.Sprintf(", resources = $%d", argPos)
		args = append(args, resourcesJSON)
		argPos++
	}

	if updates.Status != "" {
		query += fmt.Sprintf(", status = $%d", argPos)
		args = append(args, updates.Status)
		argPos++
	}

	if updates.Phase != "" {
		query += fmt.Sprintf(", phase = $%d", argPos)
		args = append(args, updates.Phase)
		argPos++
	}

	// Add WHERE clause
	query += fmt.Sprintf(" WHERE id = $%d AND deleted_at IS NULL", argPos)
	args = append(args, deploymentID)

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to update deployment: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return status.Errorf(codes.NotFound, "deployment not found")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	s.logger.Info("pipeline deployment updated",
		zap.String("deployment_id", deploymentID),
		zap.Any("updates", updates),
	)

	return nil
}

// DeleteDeployment soft deletes a deployment
func (s *PipelineDeploymentService) DeleteDeployment(ctx context.Context, deploymentID string) error {
	query := `
		UPDATE pipeline_deployments 
		SET deleted_at = NOW(), status = 'deleting', phase = 'terminating'
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := s.db.ExecContext(ctx, query, deploymentID)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to delete deployment: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return status.Errorf(codes.NotFound, "deployment not found")
	}

	s.logger.Info("pipeline deployment deleted",
		zap.String("deployment_id", deploymentID),
	)

	return nil
}

// UpdateDeploymentStatus updates the status and phase of a deployment
func (s *PipelineDeploymentService) UpdateDeploymentStatus(ctx context.Context, deploymentID string, status string, phase string, instances *models.DeploymentInstances) error {
	query := `
		UPDATE pipeline_deployments 
		SET status = $1, phase = $2, instances = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL`

	instancesJSON, _ := json.Marshal(instances)

	_, err := s.db.ExecContext(ctx, query, status, phase, instancesJSON, deploymentID)
	if err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	return nil
}

// UpdateDeploymentMetrics updates the metrics for a deployment
func (s *PipelineDeploymentService) UpdateDeploymentMetrics(ctx context.Context, deploymentID string, metrics *models.DeploymentMetrics) error {
	query := `
		UPDATE pipeline_deployments 
		SET metrics = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL`

	metricsJSON, _ := json.Marshal(metrics)

	_, err := s.db.ExecContext(ctx, query, metricsJSON, deploymentID)
	if err != nil {
		return fmt.Errorf("failed to update deployment metrics: %w", err)
	}

	return nil
}

// Helper functions

func (s *PipelineDeploymentService) validateCreateRequest(req *models.CreateDeploymentRequest) error {
	if req.DeploymentName == "" {
		return fmt.Errorf("deployment name is required")
	}
	if req.PipelineName == "" {
		return fmt.Errorf("pipeline name is required")
	}
	if req.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if len(req.TargetNodes) == 0 {
		return fmt.Errorf("at least one target node selector is required")
	}
	return nil
}

func (s *PipelineDeploymentService) deploymentExists(ctx context.Context, name, namespace string) (bool, error) {
	query := `SELECT COUNT(*) FROM pipeline_deployments WHERE deployment_name = $1 AND namespace = $2 AND deleted_at IS NULL`
	
	var count int
	err := s.db.QueryRowContext(ctx, query, name, namespace).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}