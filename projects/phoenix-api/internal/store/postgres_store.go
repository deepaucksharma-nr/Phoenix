package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/phoenix/platform/pkg/common/models"
	commonstore "github.com/phoenix/platform/pkg/common/store"
)

// PostgresPipelineDeploymentStore implements PipelineDeploymentStore using PostgreSQL
type PostgresPipelineDeploymentStore struct {
	db *commonstore.PostgresStore
}

// NewPostgresPipelineDeploymentStore creates a new PostgreSQL-backed pipeline deployment store
func NewPostgresPipelineDeploymentStore(db *commonstore.PostgresStore) *PostgresPipelineDeploymentStore {
	store := &PostgresPipelineDeploymentStore{db: db}
	
	// Create tables if they don't exist
	ctx := context.Background()
	if err := store.createTables(ctx); err != nil {
		// Log error but don't fail - tables might already exist
		fmt.Printf("Warning: failed to create tables: %v\n", err)
	}
	
	return store
}

// CreateDeployment creates a new pipeline deployment
func (s *PostgresPipelineDeploymentStore) CreateDeployment(ctx context.Context, deployment *models.PipelineDeployment) error {
	targetNodesJSON, err := json.Marshal(deployment.TargetNodes)
	if err != nil {
		return fmt.Errorf("failed to marshal target_nodes: %w", err)
	}
	
	parametersJSON, err := json.Marshal(deployment.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}
	
	resourcesJSON, err := json.Marshal(deployment.Resources)
	if err != nil {
		return fmt.Errorf("failed to marshal resources: %w", err)
	}
	
	query := `
		INSERT INTO pipeline_deployments (
			id, deployment_name, pipeline_name, namespace, target_nodes,
			parameters, resources, status, phase, created_at, updated_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	
	_, err = s.db.DB().ExecContext(ctx, query,
		deployment.ID, deployment.DeploymentName, deployment.PipelineName, 
		deployment.Namespace, string(targetNodesJSON), string(parametersJSON),
		string(resourcesJSON), deployment.Status, deployment.Phase,
		deployment.CreatedAt, deployment.UpdatedAt, deployment.CreatedBy,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}
	
	return nil
}

// GetDeployment retrieves a deployment by ID
func (s *PostgresPipelineDeploymentStore) GetDeployment(ctx context.Context, deploymentID string) (*models.PipelineDeployment, error) {
	query := `
		SELECT id, deployment_name, pipeline_name, namespace, target_nodes,
		       parameters, resources, status, phase, created_at, updated_at,
		       deleted_at, created_by
		FROM pipeline_deployments WHERE id = $1 AND deleted_at IS NULL
	`
	
	row := s.db.DB().QueryRowContext(ctx, query, deploymentID)
	
	var deployment models.PipelineDeployment
	var targetNodesJSON, parametersJSON, resourcesJSON string
	var deletedAt *time.Time
	
	err := row.Scan(
		&deployment.ID, &deployment.DeploymentName, &deployment.PipelineName,
		&deployment.Namespace, &targetNodesJSON, &parametersJSON, &resourcesJSON,
		&deployment.Status, &deployment.Phase, &deployment.CreatedAt,
		&deployment.UpdatedAt, &deletedAt, &deployment.CreatedBy,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}
	
	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(targetNodesJSON), &deployment.TargetNodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal target_nodes: %w", err)
	}
	
	if err := json.Unmarshal([]byte(parametersJSON), &deployment.Parameters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}
	
	if err := json.Unmarshal([]byte(resourcesJSON), &deployment.Resources); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resources: %w", err)
	}
	
	deployment.DeletedAt = deletedAt
	
	return &deployment, nil
}

// ListDeployments lists deployments with pagination and filtering
func (s *PostgresPipelineDeploymentStore) ListDeployments(ctx context.Context, req *models.ListDeploymentsRequest) ([]*models.PipelineDeployment, int, error) {
	// Build query with filters
	query := `
		SELECT id, deployment_name, pipeline_name, namespace, target_nodes,
		       parameters, resources, status, phase, created_at, updated_at,
		       deleted_at, created_by
		FROM pipeline_deployments
		WHERE deleted_at IS NULL
	`
	
	args := []interface{}{}
	argCount := 0
	
	// Add filters
	if req.Namespace != "" {
		argCount++
		query += fmt.Sprintf(" AND namespace = $%d", argCount)
		args = append(args, req.Namespace)
	}
	
	if req.Status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, req.Status)
	}
	
	if req.PipelineName != "" {
		argCount++
		query += fmt.Sprintf(" AND pipeline_name = $%d", argCount)
		args = append(args, req.PipelineName)
	}
	
	// Get total count
	countQuery := "SELECT COUNT(*) FROM (" + query + ") as subquery"
	var total int
	err := s.db.DB().QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deployments: %w", err)
	}
	
	// Add pagination
	query += " ORDER BY created_at DESC"
	
	limit := req.PageSize
	if limit <= 0 {
		limit = 20
	}
	offset := (req.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}
	
	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit)
	
	argCount++
	query += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, offset)
	
	rows, err := s.db.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list deployments: %w", err)
	}
	defer rows.Close()
	
	var deployments []*models.PipelineDeployment
	for rows.Next() {
		var deployment models.PipelineDeployment
		var targetNodesJSON, parametersJSON, resourcesJSON string
		var deletedAt *time.Time
		
		err := rows.Scan(
			&deployment.ID, &deployment.DeploymentName, &deployment.PipelineName,
			&deployment.Namespace, &targetNodesJSON, &parametersJSON, &resourcesJSON,
			&deployment.Status, &deployment.Phase, &deployment.CreatedAt,
			&deployment.UpdatedAt, &deletedAt, &deployment.CreatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan deployment row: %w", err)
		}
		
		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(targetNodesJSON), &deployment.TargetNodes); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal target_nodes: %w", err)
		}
		
		if err := json.Unmarshal([]byte(parametersJSON), &deployment.Parameters); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal parameters: %w", err)
		}
		
		if err := json.Unmarshal([]byte(resourcesJSON), &deployment.Resources); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal resources: %w", err)
		}
		
		deployment.DeletedAt = deletedAt
		deployments = append(deployments, &deployment)
	}
	
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate deployment rows: %w", err)
	}
	
	return deployments, total, nil
}

// UpdateDeployment updates a deployment
func (s *PostgresPipelineDeploymentStore) UpdateDeployment(ctx context.Context, deploymentID string, update *models.UpdateDeploymentRequest) error {
	// Build dynamic update query
	setClause := "updated_at = $2"
	args := []interface{}{deploymentID, time.Now()}
	argCount := 2
	
	if update.Parameters != nil {
		argCount++
		parametersJSON, err := json.Marshal(update.Parameters)
		if err != nil {
			return fmt.Errorf("failed to marshal parameters: %w", err)
		}
		setClause += fmt.Sprintf(", parameters = $%d", argCount)
		args = append(args, string(parametersJSON))
	}
	
	if update.Resources != nil {
		argCount++
		resourcesJSON, err := json.Marshal(update.Resources)
		if err != nil {
			return fmt.Errorf("failed to marshal resources: %w", err)
		}
		setClause += fmt.Sprintf(", resources = $%d", argCount)
		args = append(args, string(resourcesJSON))
	}
	
	if update.Status != "" {
		argCount++
		setClause += fmt.Sprintf(", status = $%d", argCount)
		args = append(args, update.Status)
	}
	
	if update.Phase != "" {
		argCount++
		setClause += fmt.Sprintf(", phase = $%d", argCount)
		args = append(args, update.Phase)
	}
	
	query := fmt.Sprintf(`
		UPDATE pipeline_deployments 
		SET %s
		WHERE id = $1 AND deleted_at IS NULL
	`, setClause)
	
	result, err := s.db.DB().ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("deployment not found: %s", deploymentID)
	}
	
	return nil
}

// DeleteDeployment soft deletes a deployment
func (s *PostgresPipelineDeploymentStore) DeleteDeployment(ctx context.Context, deploymentID string) error {
	query := `
		UPDATE pipeline_deployments 
		SET deleted_at = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	now := time.Now()
	result, err := s.db.DB().ExecContext(ctx, query, deploymentID, now, now)
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("deployment not found: %s", deploymentID)
	}
	
	return nil
}

// createTables creates the pipeline_deployments table if it doesn't exist
func (s *PostgresPipelineDeploymentStore) createTables(ctx context.Context) error {
	schema := `
		CREATE TABLE IF NOT EXISTS pipeline_deployments (
			id VARCHAR(255) PRIMARY KEY,
			deployment_name VARCHAR(255) NOT NULL,
			pipeline_name VARCHAR(255) NOT NULL,
			namespace VARCHAR(255) NOT NULL,
			target_nodes JSONB,
			parameters JSONB,
			resources JSONB,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			phase VARCHAR(50) NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP WITH TIME ZONE,
			created_by VARCHAR(255)
		);
		
		CREATE INDEX IF NOT EXISTS idx_deployments_namespace ON pipeline_deployments(namespace);
		CREATE INDEX IF NOT EXISTS idx_deployments_status ON pipeline_deployments(status);
		CREATE INDEX IF NOT EXISTS idx_deployments_pipeline ON pipeline_deployments(pipeline_name);
		CREATE INDEX IF NOT EXISTS idx_deployments_created_at ON pipeline_deployments(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_deployments_deleted_at ON pipeline_deployments(deleted_at);
	`
	
	_, err := s.db.DB().ExecContext(ctx, schema)
	return err
}

// UpdateDeploymentMetrics updates metrics for a deployment
func (s *PostgresPipelineDeploymentStore) UpdateDeploymentMetrics(ctx context.Context, deploymentID string, metrics *models.DeploymentMetrics) error {
	// TODO: Implement metrics storage
	// For now, just log the metrics
	fmt.Printf("Updating metrics for deployment %s: cardinality=%d, errorRate=%.2f\n", 
		deploymentID, metrics.Cardinality, metrics.ErrorRate)
	return nil
}

// GetDeploymentHistory gets historical deployment configuration
func (s *PostgresPipelineDeploymentStore) GetDeploymentHistory(ctx context.Context, deploymentID string, version int) (*models.PipelineDeployment, error) {
	// TODO: Implement deployment versioning
	// For now, just return the current deployment
	return s.GetDeployment(ctx, deploymentID)
}