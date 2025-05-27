package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/phoenix/platform/pkg/database"

	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/rs/zerolog/log"
)

// Task operations
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
	var resultJSON database.NullString
	var assignedAt, startedAt, completedAt database.NullTime
	var errorMessage database.NullString

	err := row.Scan(
		&task.ID, &task.HostID, &task.ExperimentID, &task.Type, &task.Action,
		&configJSON, &task.Priority, &task.Status,
		&assignedAt, &startedAt, &completedAt,
		&resultJSON, &errorMessage, &task.RetryCount,
		&task.CreatedAt, &task.UpdatedAt,
	)

	if err == database.ErrNoRows {
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
		var resultJSON database.NullString
		var assignedAt, startedAt, completedAt database.NullTime
		var errorMessage database.NullString

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

func (s *CompositeStore) GetPendingTasksForHost(ctx context.Context, hostID string) ([]*models.Task, error) {
	// Get up to 10 pending tasks for the host, ordered by priority
	filters := map[string]interface{}{
		"host_id": hostID,
		"status":  "pending",
		"limit":   10,
	}

	return s.ListTasks(ctx, filters)
}

func (s *CompositeStore) GetTasksByExperiment(ctx context.Context, experimentID string) ([]*models.Task, error) {
	filters := map[string]interface{}{
		"experiment_id": experimentID,
	}

	return s.ListTasks(ctx, filters)
}

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
		var resultJSON database.NullString
		var assignedAt, startedAt, completedAt database.NullTime
		var errorMessage database.NullString

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

// Agent operations
func (s *CompositeStore) UpsertAgent(ctx context.Context, agent *models.AgentStatus) error {
	capabilitiesJSON, err := json.Marshal(agent.Capabilities)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	activeTasksJSON, err := json.Marshal(agent.ActiveTasks)
	if err != nil {
		return fmt.Errorf("failed to marshal active_tasks: %w", err)
	}

	resourceUsageJSON, err := json.Marshal(agent.ResourceUsage)
	if err != nil {
		return fmt.Errorf("failed to marshal resource_usage: %w", err)
	}

	metadataJSON, err := json.Marshal(agent.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO agents (
			host_id, hostname, ip_address, agent_version, started_at,
			last_heartbeat, status, capabilities, active_tasks,
			resource_usage, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (host_id) DO UPDATE SET
			hostname = EXCLUDED.hostname,
			ip_address = EXCLUDED.ip_address,
			agent_version = EXCLUDED.agent_version,
			last_heartbeat = EXCLUDED.last_heartbeat,
			status = EXCLUDED.status,
			capabilities = EXCLUDED.capabilities,
			active_tasks = EXCLUDED.active_tasks,
			resource_usage = EXCLUDED.resource_usage,
			metadata = EXCLUDED.metadata,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err = s.pipelineStore.db.DB().ExecContext(ctx, query,
		agent.HostID, agent.Hostname, agent.IPAddress, agent.AgentVersion,
		agent.StartedAt, agent.LastHeartbeat, agent.Status,
		string(capabilitiesJSON), string(activeTasksJSON),
		string(resourceUsageJSON), string(metadataJSON),
	)

	if err != nil {
		return fmt.Errorf("failed to upsert agent: %w", err)
	}

	return nil
}

func (s *CompositeStore) GetAgent(ctx context.Context, hostID string) (*models.AgentStatus, error) {
	query := `
		SELECT host_id, hostname, ip_address, agent_version, started_at,
		       last_heartbeat, status, capabilities, active_tasks,
		       resource_usage, metadata, created_at, updated_at
		FROM agents WHERE host_id = $1
	`

	row := s.pipelineStore.db.DB().QueryRowContext(ctx, query, hostID)

	var agent models.AgentStatus
	var capabilitiesJSON, activeTasksJSON, resourceUsageJSON, metadataJSON string
	var startedAt database.NullTime

	err := row.Scan(
		&agent.HostID, &agent.Hostname, &agent.IPAddress, &agent.AgentVersion,
		&startedAt, &agent.LastHeartbeat, &agent.Status,
		&capabilitiesJSON, &activeTasksJSON, &resourceUsageJSON, &metadataJSON,
		&agent.CreatedAt, &agent.UpdatedAt,
	)

	if err == database.ErrNoRows {
		return nil, fmt.Errorf("agent not found: %s", hostID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	if startedAt.Valid {
		agent.StartedAt = &startedAt.Time
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(capabilitiesJSON), &agent.Capabilities); err != nil {
		agent.Capabilities = make(map[string]interface{})
	}
	if err := json.Unmarshal([]byte(activeTasksJSON), &agent.ActiveTasks); err != nil {
		agent.ActiveTasks = []string{}
	}
	if err := json.Unmarshal([]byte(resourceUsageJSON), &agent.ResourceUsage); err != nil {
		agent.ResourceUsage = models.ResourceUsage{}
	}
	if err := json.Unmarshal([]byte(metadataJSON), &agent.Metadata); err != nil {
		agent.Metadata = make(map[string]interface{})
	}

	return &agent, nil
}

func (s *CompositeStore) ListAgents(ctx context.Context) ([]*models.AgentStatus, error) {
	query := `
		SELECT host_id, hostname, ip_address, agent_version, started_at,
		       last_heartbeat, status, capabilities, active_tasks,
		       resource_usage, metadata, created_at, updated_at
		FROM agents ORDER BY hostname
	`

	rows, err := s.pipelineStore.db.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	var agents []*models.AgentStatus

	for rows.Next() {
		var agent models.AgentStatus
		var capabilitiesJSON, activeTasksJSON, resourceUsageJSON, metadataJSON string
		var startedAt database.NullTime

		err := rows.Scan(
			&agent.HostID, &agent.Hostname, &agent.IPAddress, &agent.AgentVersion,
			&startedAt, &agent.LastHeartbeat, &agent.Status,
			&capabilitiesJSON, &activeTasksJSON, &resourceUsageJSON, &metadataJSON,
			&agent.CreatedAt, &agent.UpdatedAt,
		)

		if err != nil {
			log.Error().Err(err).Msg("Failed to scan agent row")
			continue
		}

		if startedAt.Valid {
			agent.StartedAt = &startedAt.Time
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(capabilitiesJSON), &agent.Capabilities); err != nil {
			agent.Capabilities = make(map[string]interface{})
		}
		if err := json.Unmarshal([]byte(activeTasksJSON), &agent.ActiveTasks); err != nil {
			agent.ActiveTasks = []string{}
		}
		if err := json.Unmarshal([]byte(resourceUsageJSON), &agent.ResourceUsage); err != nil {
			agent.ResourceUsage = models.ResourceUsage{}
		}
		if err := json.Unmarshal([]byte(metadataJSON), &agent.Metadata); err != nil {
			agent.Metadata = make(map[string]interface{})
		}

		agents = append(agents, &agent)
	}

	return agents, nil
}

func (s *CompositeStore) UpdateAgentHeartbeat(ctx context.Context, heartbeat *models.AgentHeartbeat) error {
	// First check if agent exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM agents WHERE host_id = $1)`
	err := s.pipelineStore.db.DB().QueryRowContext(ctx, checkQuery, heartbeat.HostID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check agent existence: %w", err)
	}

	if !exists {
		// Create new agent entry
		agent := &models.AgentStatus{
			HostID:        heartbeat.HostID,
			Hostname:      heartbeat.HostID, // Will be updated on next full update
			AgentVersion:  heartbeat.AgentVersion,
			Status:        heartbeat.Status,
			LastHeartbeat: heartbeat.LastHeartbeat,
			ActiveTasks:   heartbeat.ActiveTasks,
			ResourceUsage: heartbeat.ResourceUsage,
			Capabilities:  make(map[string]interface{}),
			Metadata:      make(map[string]interface{}),
		}
		return s.UpsertAgent(ctx, agent)
	}

	// Update existing agent
	activeTasksJSON, err := json.Marshal(heartbeat.ActiveTasks)
	if err != nil {
		return fmt.Errorf("failed to marshal active_tasks: %w", err)
	}

	resourceUsageJSON, err := json.Marshal(heartbeat.ResourceUsage)
	if err != nil {
		return fmt.Errorf("failed to marshal resource_usage: %w", err)
	}

	query := `
		UPDATE agents SET
			agent_version = $2,
			last_heartbeat = $3,
			status = $4,
			active_tasks = $5,
			resource_usage = $6,
			updated_at = CURRENT_TIMESTAMP
		WHERE host_id = $1
	`

	_, err = s.pipelineStore.db.DB().ExecContext(ctx, query,
		heartbeat.HostID, heartbeat.AgentVersion, heartbeat.LastHeartbeat,
		heartbeat.Status, string(activeTasksJSON), string(resourceUsageJSON),
	)

	if err != nil {
		return fmt.Errorf("failed to update agent heartbeat: %w", err)
	}

	return nil
}

func (s *CompositeStore) GetAllAgents(ctx context.Context) ([]*models.AgentStatus, error) {
	return s.ListAgents(ctx)
}

func (s *CompositeStore) GetAgentsWithLocation(ctx context.Context) (map[string]interface{}, error) {
	agents, err := s.ListAgents(ctx)
	if err != nil {
		return nil, err
	}

	// Filter agents that have location metadata and format for UI
	var agentsWithLocation []map[string]interface{}
	for _, agent := range agents {
		if _, hasLat := agent.Metadata["latitude"]; hasLat {
			if _, hasLon := agent.Metadata["longitude"]; hasLon {
				agentMap := map[string]interface{}{
					"host_id":        agent.HostID,
					"hostname":       agent.Hostname,
					"status":         agent.Status,
					"latitude":       agent.Metadata["latitude"],
					"longitude":      agent.Metadata["longitude"],
					"region":         agent.Metadata["region"],
					"zone":           agent.Metadata["zone"],
					"last_heartbeat": agent.LastHeartbeat,
				}
				agentsWithLocation = append(agentsWithLocation, agentMap)
			}
		}
	}

	return map[string]interface{}{
		"agents": agentsWithLocation,
		"total":  len(agentsWithLocation),
	}, nil
}

func (s *CompositeStore) CacheMetric(ctx context.Context, hostID string, metric map[string]interface{}) error {
	// Extract required fields
	experimentID, _ := metric["experiment_id"].(string)
	metricName, _ := metric["metric_name"].(string)
	variant, _ := metric["variant"].(string)
	value, _ := metric["value"].(float64)
	timestamp, ok := metric["timestamp"].(time.Time)
	if !ok {
		timestamp = time.Now()
	}

	labelsJSON, err := json.Marshal(metric["labels"])
	if err != nil {
		labelsJSON = []byte("{}")
	}

	query := `
		INSERT INTO metric_cache (
			experiment_id, timestamp, metric_name, variant,
			host_id, value, labels
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = s.pipelineStore.db.DB().ExecContext(ctx, query,
		experimentID, timestamp, metricName, variant,
		hostID, value, string(labelsJSON),
	)

	if err != nil {
		return fmt.Errorf("failed to cache metric: %w", err)
	}

	return nil
}

// GetStaleTasks returns tasks that have been assigned but not updated within the threshold
func (s *CompositeStore) GetStaleTasks(ctx context.Context, threshold time.Duration) ([]*models.Task, error) {
	query := `
		SELECT id, host_id, experiment_id, task_type, action, config,
		       priority, status, assigned_at, started_at, completed_at,
		       result, error_message, retry_count, created_at, updated_at
		FROM tasks 
		WHERE status IN ('assigned', 'running')
		AND assigned_at IS NOT NULL
		AND assigned_at < $1
		ORDER BY assigned_at ASC
	`

	cutoffTime := time.Now().Add(-threshold)

	rows, err := s.pipelineStore.db.DB().QueryContext(ctx, query, cutoffTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get stale tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task

	for rows.Next() {
		var task models.Task
		var configJSON string
		var resultJSON database.NullString
		var assignedAt, startedAt, completedAt database.NullTime
		var errorMessage database.NullString

		err := rows.Scan(
			&task.ID, &task.HostID, &task.ExperimentID, &task.Type, &task.Action,
			&configJSON, &task.Priority, &task.Status,
			&assignedAt, &startedAt, &completedAt,
			&resultJSON, &errorMessage, &task.RetryCount,
			&task.CreatedAt, &task.UpdatedAt,
		)

		if err != nil {
			log.Error().Err(err).Msg("Failed to scan stale task row")
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

// DeleteOldTasks deletes tasks created before the specified time
func (s *CompositeStore) DeleteOldTasks(ctx context.Context, before time.Time) error {
	query := `
		DELETE FROM tasks 
		WHERE created_at < $1
		AND status IN ('completed', 'failed')
	`

	result, err := s.pipelineStore.db.DB().ExecContext(ctx, query, before)
	if err != nil {
		return fmt.Errorf("failed to delete old tasks: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	log.Info().
		Int64("deleted_count", rowsAffected).
		Time("before", before).
		Msg("Deleted old tasks")

	return nil
}
