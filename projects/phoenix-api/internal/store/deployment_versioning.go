package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/phoenix/platform/pkg/database" // Only for type reference (database.ErrNoRows)
	"time"

	"github.com/phoenix/platform/pkg/common/models"
)

// DeploymentVersion represents a version of a deployment
type DeploymentVersion struct {
	ID                  int64                  `json:"id" db:"id"`
	DeploymentID        string                 `json:"deployment_id" db:"deployment_id"`
	Version             int                    `json:"version" db:"version"`
	PipelineConfig      string                 `json:"pipeline_config" db:"pipeline_config"`
	Parameters          map[string]interface{} `json:"parameters" db:"parameters"`
	DeployedBy          string                 `json:"deployed_by" db:"deployed_by"`
	DeployedAt          time.Time              `json:"deployed_at" db:"deployed_at"`
	Status              string                 `json:"status" db:"status"`
	RollbackFromVersion *int                   `json:"rollback_from_version,omitempty" db:"rollback_from_version"`
	Notes               string                 `json:"notes,omitempty" db:"notes"`
}

// RecordDeploymentVersion records a new version of a deployment
func (s *PostgresPipelineDeploymentStore) RecordDeploymentVersion(ctx context.Context, deploymentID, pipelineConfig string, parameters map[string]interface{}, deployedBy string, notes string) (int, error) {
	// Convert parameters to JSON
	paramsJSON, err := json.Marshal(parameters)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	query := `SELECT record_deployment_version($1, $2, $3, $4, $5)`

	var version int
	err = s.db.DB().QueryRowContext(ctx, query,
		deploymentID, pipelineConfig, string(paramsJSON), deployedBy, notes,
	).Scan(&version)

	if err != nil {
		return 0, fmt.Errorf("failed to record deployment version: %w", err)
	}

	return version, nil
}

// GetDeploymentVersion retrieves a specific version of a deployment
func (s *PostgresPipelineDeploymentStore) GetDeploymentVersion(ctx context.Context, deploymentID string, version int) (*DeploymentVersion, error) {
	query := `
		SELECT id, deployment_id, version, pipeline_config, parameters,
		       deployed_by, deployed_at, status, rollback_from_version, notes
		FROM deployment_versions
		WHERE deployment_id = $1 AND version = $2
	`

	var dv DeploymentVersion
	var paramsJSON string

	err := s.db.DB().QueryRowContext(ctx, query, deploymentID, version).Scan(
		&dv.ID, &dv.DeploymentID, &dv.Version, &dv.PipelineConfig, &paramsJSON,
		&dv.DeployedBy, &dv.DeployedAt, &dv.Status, &dv.RollbackFromVersion, &dv.Notes,
	)

	if err == database.ErrNoRows {
		return nil, fmt.Errorf("deployment version not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment version: %w", err)
	}

	// Unmarshal parameters
	if err := json.Unmarshal([]byte(paramsJSON), &dv.Parameters); err != nil {
		dv.Parameters = make(map[string]interface{})
	}

	return &dv, nil
}

// ListDeploymentVersions retrieves the version history for a deployment
func (s *PostgresPipelineDeploymentStore) ListDeploymentVersions(ctx context.Context, deploymentID string) ([]*DeploymentVersion, error) {
	limit := 20
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT id, deployment_id, version, pipeline_config, parameters,
		       deployed_by, deployed_at, status, rollback_from_version, notes
		FROM deployment_versions
		WHERE deployment_id = $1
		ORDER BY version DESC
		LIMIT $2
	`

	rows, err := s.db.DB().QueryContext(ctx, query, deploymentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment history: %w", err)
	}
	defer rows.Close()

	var versions []*DeploymentVersion

	for rows.Next() {
		var dv DeploymentVersion
		var paramsJSON string

		err := rows.Scan(
			&dv.ID, &dv.DeploymentID, &dv.Version, &dv.PipelineConfig, &paramsJSON,
			&dv.DeployedBy, &dv.DeployedAt, &dv.Status, &dv.RollbackFromVersion, &dv.Notes,
		)
		if err != nil {
			continue
		}

		// Unmarshal parameters
		if err := json.Unmarshal([]byte(paramsJSON), &dv.Parameters); err != nil {
			dv.Parameters = make(map[string]interface{})
		}

		versions = append(versions, &dv)
	}

	return versions, nil
}

// RollbackDeployment rolls back a deployment to a specific version
func (s *PostgresPipelineDeploymentStore) RollbackDeployment(ctx context.Context, deploymentID string, targetVersion int, rolledBackBy string, notes string) error {
	query := `SELECT rollback_deployment_version($1, $2, $3, $4)`

	var success bool
	err := s.db.DB().QueryRowContext(ctx, query,
		deploymentID, targetVersion, rolledBackBy, notes,
	).Scan(&success)

	if err != nil {
		return fmt.Errorf("failed to rollback deployment: %w", err)
	}

	if !success {
		return fmt.Errorf("rollback failed - target version not found")
	}

	return nil
}

// GetCurrentVersion gets the current version number for a deployment
func (s *PostgresPipelineDeploymentStore) GetCurrentVersion(ctx context.Context, deploymentID string) (int, error) {
	query := `SELECT current_version FROM pipeline_deployments WHERE id = $1`

	var version int
	err := s.db.DB().QueryRowContext(ctx, query, deploymentID).Scan(&version)

	if err == database.ErrNoRows {
		return 0, fmt.Errorf("deployment not found")
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get current version: %w", err)
	}

	return version, nil
}

// Implement GetDeploymentHistory for composite store to satisfy interface
func (s *CompositeStore) GetDeploymentHistory(ctx context.Context, deploymentID string, version int) (*models.PipelineDeployment, error) {
	// Get the deployment version
	dv, err := s.pipelineStore.GetDeploymentVersion(ctx, deploymentID, version)
	if err != nil {
		return nil, err
	}

	// Get the base deployment
	deployment, err := s.pipelineStore.GetDeployment(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	// Override with version-specific data
	deployment.Parameters = dv.Parameters
	deployment.UpdatedAt = dv.DeployedAt

	return deployment, nil
}
