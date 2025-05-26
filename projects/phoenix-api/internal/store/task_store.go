package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	
	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/rs/zerolog/log"
)

// CreateTask creates a new task
func (s *CompositeStore) CreateTask(ctx context.Context, task *models.Task) error {
	configJSON, err := json.Marshal(task.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	query := `
		INSERT INTO tasks (
			host_id, experiment_id, task_type, action, config,
			priority, status, retry_count
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	
	err = s.pipelineStore.db.DB().QueryRowContext(ctx, query,
		task.HostID, task.ExperimentID, task.Type, task.Action,
		string(configJSON), task.Priority, task.Status, task.RetryCount,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	
	return nil
}

// GetTask retrieves a task by ID
func (s *CompositeStore) GetTask(ctx context.Context, taskID string) (*models.Task, error) {
	query := `
		SELECT id, host_id, experiment_id, task_type, action, config,
		       priority, status, assigned_at, started_at, completed_at,
		       result, error_message, retry_count, created_at, updated_at
		FROM tasks WHERE id = $1
	`
	
	row := s.pipelineStore.db.DB().QueryRowContext(ctx, query, taskID)
	
	var task models.Task
	var configJSON string
	var resultJSON sql.NullString
	var assignedAt, startedAt, completedAt sql.NullTime
	var errorMessage sql.NullString
	
	err := row.Scan(
		&task.ID, &task.HostID, &task.ExperimentID, &task.Type, &task.Action,
		&configJSON, &task.Priority, &task.Status,
		&assignedAt, &startedAt, &completedAt,
		&resultJSON, &errorMessage, &task.RetryCount,
		&task.CreatedAt, &task.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	
	// Handle nullable fields
	if assignedAt.Valid {
		task.AssignedAt = &assignedAt.Time
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	if errorMessage.Valid {
		task.ErrorMessage = errorMessage.String
	}
	
	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(configJSON), &task.Config); err != nil {
		task.Config = make(map[string]interface{})
	}
	if resultJSON.Valid {
		if err := json.Unmarshal([]byte(resultJSON.String), &task.Result); err != nil {
			task.Result = make(map[string]interface{})
		}
	}
	
	return &task, nil
}

// ListTasks retrieves tasks based on filters
func (s *CompositeStore) ListTasks(ctx context.Context, filters map[string]interface{}) ([]*models.Task, error) {
	query := `
		SELECT id, host_id, experiment_id, task_type, action, config,
		       priority, status, assigned_at, started_at, completed_at,
		       result, error_message, retry_count, created_at, updated_at
		FROM tasks WHERE 1=1
	`
	
	var args []interface{}
	argCount := 0
	
	// Build dynamic query based on filters
	if hostID, ok := filters["host_id"].(string); ok && hostID != "" {
		argCount++
		query += fmt.Sprintf(" AND host_id = $%d", argCount)
		args = append(args, hostID)
	}
	
	if experimentID, ok := filters["experiment_id"].(string); ok && experimentID != "" {
		argCount++
		query += fmt.Sprintf(" AND experiment_id = $%d", argCount)
		args = append(args, experimentID)
	}
	
	if status, ok := filters["status"].(string); ok && status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}
	
	if taskType, ok := filters["type"].(string); ok && taskType != "" {
		argCount++
		query += fmt.Sprintf(" AND task_type = $%d", argCount)
		args = append(args, taskType)
	}
	
	// Add ordering
	query += " ORDER BY priority DESC, created_at ASC"
	
	// Add limit if specified
	if limit, ok := filters["limit"].(int); ok && limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}
	
	rows, err := s.pipelineStore.db.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()
	
	var tasks []*models.Task
	
	for rows.Next() {
		var task models.Task
		var configJSON string
		var resultJSON sql.NullString
		var assignedAt, startedAt, completedAt sql.NullTime
		var errorMessage sql.NullString
		
		err := rows.Scan(
			&task.ID, &task.HostID, &task.ExperimentID, &task.Type, &task.Action,
			&configJSON, &task.Priority, &task.Status,
			&assignedAt, &startedAt, &completedAt,
			&resultJSON, &errorMessage, &task.RetryCount,
			&task.CreatedAt, &task.UpdatedAt,
		)
		
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan task row")
			continue
		}
		
		// Handle nullable fields
		if assignedAt.Valid {
			task.AssignedAt = &assignedAt.Time
		}
		if startedAt.Valid {
			task.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}
		if errorMessage.Valid {
			task.ErrorMessage = errorMessage.String
		}
		
		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(configJSON), &task.Config); err != nil {
			task.Config = make(map[string]interface{})
		}
		if resultJSON.Valid {
			if err := json.Unmarshal([]byte(resultJSON.String), &task.Result); err != nil {
				task.Result = make(map[string]interface{})
			}
		}
		
		tasks = append(tasks, &task)
	}
	
	return tasks, nil
}

// UpdateTask updates a task
func (s *CompositeStore) UpdateTask(ctx context.Context, task *models.Task) error {
	configJSON, err := json.Marshal(task.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	resultJSON, err := json.Marshal(task.Result)
	if err != nil {
		resultJSON = []byte("null")
	}
	
	query := `
		UPDATE tasks SET
			status = $2,
			assigned_at = $3,
			started_at = $4,
			completed_at = $5,
			result = $6,
			error_message = $7,
			retry_count = $8,
			config = $9,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	
	_, err = s.pipelineStore.db.DB().ExecContext(ctx, query,
		task.ID, task.Status, task.AssignedAt, task.StartedAt,
		task.CompletedAt, string(resultJSON), task.ErrorMessage,
		task.RetryCount, string(configJSON),
	)
	
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	
	return nil
}

// GetPendingTasksForHost retrieves pending tasks for a specific host
func (s *CompositeStore) GetPendingTasksForHost(ctx context.Context, hostID string) ([]*models.Task, error) {
	// Get up to 10 pending tasks for the host, ordered by priority
	filters := map[string]interface{}{
		"host_id": hostID,
		"status":  "pending",
		"limit":   10,
	}
	
	return s.ListTasks(ctx, filters)
}

// GetTasksByExperiment retrieves all tasks for an experiment
func (s *CompositeStore) GetTasksByExperiment(ctx context.Context, experimentID string) ([]*models.Task, error) {
	filters := map[string]interface{}{
		"experiment_id": experimentID,
	}
	
	return s.ListTasks(ctx, filters)
}

// GetTaskStats retrieves task statistics
func (s *CompositeStore) GetTaskStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
			COUNT(CASE WHEN status = 'assigned' THEN 1 END) as assigned,
			COUNT(CASE WHEN status = 'running' THEN 1 END) as running,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed,
			COUNT(DISTINCT host_id) as unique_hosts,
			COUNT(DISTINCT experiment_id) as unique_experiments
		FROM tasks
		WHERE created_at > NOW() - INTERVAL '24 hours'
	`
	
	var stats struct {
		Total             int
		Pending           int
		Assigned          int
		Running           int
		Completed         int
		Failed            int
		UniqueHosts       int
		UniqueExperiments int
	}
	
	err := s.pipelineStore.db.DB().QueryRowContext(ctx, query).Scan(
		&stats.Total, &stats.Pending, &stats.Assigned, &stats.Running,
		&stats.Completed, &stats.Failed, &stats.UniqueHosts, &stats.UniqueExperiments,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get task stats: %w", err)
	}
	
	return map[string]interface{}{
		"total":              stats.Total,
		"pending":            stats.Pending,
		"assigned":           stats.Assigned,
		"running":            stats.Running,
		"completed":          stats.Completed,
		"failed":             stats.Failed,
		"unique_hosts":       stats.UniqueHosts,
		"unique_experiments": stats.UniqueExperiments,
	}, nil
}

// GetActiveTasks retrieves active tasks with filters
func (s *CompositeStore) GetActiveTasks(ctx context.Context, status, hostID string, limit int) ([]*models.Task, error) {
	filters := make(map[string]interface{})
	
	if status != "" {
		// Handle "active" as a special case meaning non-completed tasks
		if status == "active" {
			return s.getActiveTasksSpecial(ctx, hostID, limit)
		}
		filters["status"] = status
	}
	
	if hostID != "" {
		filters["host_id"] = hostID
	}
	
	if limit > 0 {
		filters["limit"] = limit
	}
	
	return s.ListTasks(ctx, filters)
}

// getActiveTasksSpecial handles the special "active" status query
func (s *CompositeStore) getActiveTasksSpecial(ctx context.Context, hostID string, limit int) ([]*models.Task, error) {
	query := `
		SELECT id, host_id, experiment_id, task_type, action, config,
		       priority, status, assigned_at, started_at, completed_at,
		       result, error_message, retry_count, created_at, updated_at
		FROM tasks 
		WHERE status IN ('pending', 'assigned', 'running')
	`
	
	var args []interface{}
	argCount := 0
	
	if hostID != "" {
		argCount++
		query += fmt.Sprintf(" AND host_id = $%d", argCount)
		args = append(args, hostID)
	}
	
	query += " ORDER BY priority DESC, created_at ASC"
	
	if limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}
	
	rows, err := s.pipelineStore.db.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get active tasks: %w", err)
	}
	defer rows.Close()
	
	var tasks []*models.Task
	
	for rows.Next() {
		var task models.Task
		var configJSON string
		var resultJSON sql.NullString
		var assignedAt, startedAt, completedAt sql.NullTime
		var errorMessage sql.NullString
		
		err := rows.Scan(
			&task.ID, &task.HostID, &task.ExperimentID, &task.Type, &task.Action,
			&configJSON, &task.Priority, &task.Status,
			&assignedAt, &startedAt, &completedAt,
			&resultJSON, &errorMessage, &task.RetryCount,
			&task.CreatedAt, &task.UpdatedAt,
		)
		
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan task row")
			continue
		}
		
		// Handle nullable fields
		if assignedAt.Valid {
			task.AssignedAt = &assignedAt.Time
		}
		if startedAt.Valid {
			task.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}
		if errorMessage.Valid {
			task.ErrorMessage = errorMessage.String
		}
		
		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(configJSON), &task.Config); err != nil {
			task.Config = make(map[string]interface{})
		}
		if resultJSON.Valid {
			if err := json.Unmarshal([]byte(resultJSON.String), &task.Result); err != nil {
				task.Result = make(map[string]interface{})
			}
		}
		
		tasks = append(tasks, &task)
	}
	
	return tasks, nil
}