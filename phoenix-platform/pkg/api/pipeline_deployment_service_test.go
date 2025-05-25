package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPipelineDeploymentService(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := NewPipelineDeploymentService(db)
	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
}

func TestPipelineDeploymentService_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := NewPipelineDeploymentService(db)
	ctx := context.Background()

	deployment := &PipelineDeployment{
		Name:        "test-deployment",
		Namespace:   "default",
		Template:    "process-intelligent-v1",
		Config:      json.RawMessage(`{"sampling_rate": 0.1}`),
		Description: "Test deployment",
		CreatedBy:   "user@example.com",
	}

	expectedID := "dep-123"
	mock.ExpectQuery(`INSERT INTO pipeline_deployments`).
		WithArgs(
			deployment.Name,
			deployment.Namespace,
			deployment.Template,
			deployment.Config,
			deployment.Description,
			"active",
			deployment.CreatedBy,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(expectedID, time.Now(), time.Now()))

	err = service.Create(ctx, deployment)
	assert.NoError(t, err)
	assert.Equal(t, expectedID, deployment.ID)
	assert.NotZero(t, deployment.CreatedAt)
	assert.NotZero(t, deployment.UpdatedAt)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPipelineDeploymentService_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := NewPipelineDeploymentService(db)
	ctx := context.Background()

	deploymentID := "dep-123"
	expectedDeployment := &PipelineDeployment{
		ID:          deploymentID,
		Name:        "test-deployment",
		Namespace:   "default",
		Template:    "process-intelligent-v1",
		Config:      json.RawMessage(`{"sampling_rate": 0.1}`),
		Description: "Test deployment",
		Status:      "active",
		CreatedBy:   "user@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "namespace", "template", "config",
		"description", "status", "created_by", "created_at", "updated_at",
	}).AddRow(
		expectedDeployment.ID,
		expectedDeployment.Name,
		expectedDeployment.Namespace,
		expectedDeployment.Template,
		expectedDeployment.Config,
		expectedDeployment.Description,
		expectedDeployment.Status,
		expectedDeployment.CreatedBy,
		expectedDeployment.CreatedAt,
		expectedDeployment.UpdatedAt,
	)

	mock.ExpectQuery(`SELECT (.+) FROM pipeline_deployments WHERE id = \$1`).
		WithArgs(deploymentID).
		WillReturnRows(rows)

	deployment, err := service.Get(ctx, deploymentID)
	assert.NoError(t, err)
	assert.Equal(t, expectedDeployment.ID, deployment.ID)
	assert.Equal(t, expectedDeployment.Name, deployment.Name)
	assert.Equal(t, expectedDeployment.Status, deployment.Status)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPipelineDeploymentService_Get_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := NewPipelineDeploymentService(db)
	ctx := context.Background()

	deploymentID := "non-existent"
	mock.ExpectQuery(`SELECT (.+) FROM pipeline_deployments WHERE id = \$1`).
		WithArgs(deploymentID).
		WillReturnError(sql.ErrNoRows)

	deployment, err := service.Get(ctx, deploymentID)
	assert.Error(t, err)
	assert.Nil(t, deployment)
	assert.Contains(t, err.Error(), "not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPipelineDeploymentService_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := NewPipelineDeploymentService(db)
	ctx := context.Background()

	namespace := "default"
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "name", "namespace", "template", "config",
		"description", "status", "created_by", "created_at", "updated_at",
	}).
		AddRow("dep-1", "deployment-1", namespace, "template-1", `{}`,
			"desc-1", "active", "user1", now, now).
		AddRow("dep-2", "deployment-2", namespace, "template-2", `{}`,
			"desc-2", "active", "user2", now, now)

	mock.ExpectQuery(`SELECT (.+) FROM pipeline_deployments WHERE namespace = \$1`).
		WithArgs(namespace).
		WillReturnRows(rows)

	deployments, err := service.List(ctx, namespace)
	assert.NoError(t, err)
	assert.Len(t, deployments, 2)
	assert.Equal(t, "dep-1", deployments[0].ID)
	assert.Equal(t, "dep-2", deployments[1].ID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPipelineDeploymentService_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := NewPipelineDeploymentService(db)
	ctx := context.Background()

	deploymentID := "dep-123"
	update := &PipelineDeploymentUpdate{
		Config: json.RawMessage(`{"sampling_rate": 0.05}`),
		Reason: "Reducing sampling rate",
	}
	updatedBy := "admin@example.com"

	// Mock transaction
	mock.ExpectBegin()

	// Mock deployment history insert
	mock.ExpectExec(`INSERT INTO pipeline_deployment_history`).
		WithArgs(sqlmock.AnyArg(), deploymentID, sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), "update", update.Reason, updatedBy).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock deployment update
	mock.ExpectExec(`UPDATE pipeline_deployments SET config = \$1, updated_at = CURRENT_TIMESTAMP`).
		WithArgs(update.Config, deploymentID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err = service.Update(ctx, deploymentID, update, updatedBy)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPipelineDeploymentService_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := NewPipelineDeploymentService(db)
	ctx := context.Background()

	deploymentID := "dep-123"
	newStatus := "rollback"
	reason := "Performance regression detected"
	updatedBy := "system"

	// Mock transaction
	mock.ExpectBegin()

	// Mock current deployment query
	currentRows := sqlmock.NewRows([]string{"config", "status"}).
		AddRow(`{"sampling_rate": 0.1}`, "active")
	mock.ExpectQuery(`SELECT config, status FROM pipeline_deployments`).
		WithArgs(deploymentID).
		WillReturnRows(currentRows)

	// Mock deployment history insert
	mock.ExpectExec(`INSERT INTO pipeline_deployment_history`).
		WithArgs(sqlmock.AnyArg(), deploymentID, sqlmock.AnyArg(), "active",
			sqlmock.AnyArg(), newStatus, "status_change", reason, updatedBy).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock deployment status update
	mock.ExpectExec(`UPDATE pipeline_deployments SET status = \$1`).
		WithArgs(newStatus, deploymentID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err = service.UpdateStatus(ctx, deploymentID, newStatus, reason, updatedBy)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPipelineDeploymentService_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := NewPipelineDeploymentService(db)
	ctx := context.Background()

	deploymentID := "dep-123"

	mock.ExpectExec(`DELETE FROM pipeline_deployments WHERE id = \$1`).
		WithArgs(deploymentID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = service.Delete(ctx, deploymentID)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPipelineDeploymentService_GetHistory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := NewPipelineDeploymentService(db)
	ctx := context.Background()

	deploymentID := "dep-123"
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "deployment_id", "config", "old_status", "new_status",
		"action", "reason", "created_by", "created_at",
	}).
		AddRow("hist-1", deploymentID, `{"sampling_rate": 0.1}`, "", "active",
			"create", "Initial deployment", "user1", now.Add(-2*time.Hour)).
		AddRow("hist-2", deploymentID, `{"sampling_rate": 0.05}`, "active", "active",
			"update", "Reduced sampling", "user2", now.Add(-1*time.Hour))

	mock.ExpectQuery(`SELECT (.+) FROM pipeline_deployment_history WHERE deployment_id = \$1`).
		WithArgs(deploymentID).
		WillReturnRows(rows)

	history, err := service.GetHistory(ctx, deploymentID)
	assert.NoError(t, err)
	assert.Len(t, history, 2)
	assert.Equal(t, "hist-1", history[0].ID)
	assert.Equal(t, "create", history[0].Action)
	assert.Equal(t, "hist-2", history[1].ID)
	assert.Equal(t, "update", history[1].Action)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPipelineDeploymentService_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := NewPipelineDeploymentService(db)
	ctx := context.Background()

	deploymentID := "dep-123"
	historyID := "hist-1"
	reason := "Reverting to previous configuration"
	rolledBackBy := "admin@example.com"

	// Mock transaction
	mock.ExpectBegin()

	// Mock history entry query
	historyRows := sqlmock.NewRows([]string{"config"}).
		AddRow(`{"sampling_rate": 0.1}`)
	mock.ExpectQuery(`SELECT config FROM pipeline_deployment_history WHERE id = \$1`).
		WithArgs(historyID, deploymentID).
		WillReturnRows(historyRows)

	// Mock current deployment query
	currentRows := sqlmock.NewRows([]string{"config", "status"}).
		AddRow(`{"sampling_rate": 0.05}`, "active")
	mock.ExpectQuery(`SELECT config, status FROM pipeline_deployments WHERE id = \$1`).
		WithArgs(deploymentID).
		WillReturnRows(currentRows)

	// Mock deployment history insert for rollback
	mock.ExpectExec(`INSERT INTO pipeline_deployment_history`).
		WithArgs(sqlmock.AnyArg(), deploymentID, sqlmock.AnyArg(), "active",
			sqlmock.AnyArg(), "active", "rollback", reason, rolledBackBy).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock deployment update
	mock.ExpectExec(`UPDATE pipeline_deployments SET config = \$1`).
		WithArgs(`{"sampling_rate": 0.1}`, deploymentID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err = service.Rollback(ctx, deploymentID, historyID, reason, rolledBackBy)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPipelineDeploymentService_Export(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := NewPipelineDeploymentService(db)
	ctx := context.Background()

	deploymentID := "dep-123"
	now := time.Now()

	// Mock deployment query
	deploymentRows := sqlmock.NewRows([]string{
		"id", "name", "namespace", "template", "config",
		"description", "status", "created_by", "created_at", "updated_at",
	}).AddRow(
		deploymentID, "test-deployment", "default", "process-intelligent-v1",
		`{"sampling_rate": 0.1}`, "Test deployment", "active",
		"user@example.com", now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM pipeline_deployments WHERE id = \$1`).
		WithArgs(deploymentID).
		WillReturnRows(deploymentRows)

	// Mock history query
	historyRows := sqlmock.NewRows([]string{
		"id", "deployment_id", "config", "old_status", "new_status",
		"action", "reason", "created_by", "created_at",
	}).
		AddRow("hist-1", deploymentID, `{"sampling_rate": 0.1}`, "", "active",
			"create", "Initial deployment", "user@example.com", now)

	mock.ExpectQuery(`SELECT (.+) FROM pipeline_deployment_history WHERE deployment_id = \$1`).
		WithArgs(deploymentID).
		WillReturnRows(historyRows)

	export, err := service.Export(ctx, deploymentID)
	assert.NoError(t, err)
	assert.NotNil(t, export)
	assert.Equal(t, deploymentID, export.Deployment.ID)
	assert.Equal(t, "test-deployment", export.Deployment.Name)
	assert.Len(t, export.History, 1)
	assert.NotZero(t, export.ExportedAt)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}