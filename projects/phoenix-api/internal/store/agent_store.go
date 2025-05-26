package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/rs/zerolog/log"
)

// UpsertAgent creates or updates an agent status
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

// GetAgent retrieves an agent by host ID
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
	var startedAt sql.NullTime
	
	err := row.Scan(
		&agent.HostID, &agent.Hostname, &agent.IPAddress, &agent.AgentVersion,
		&startedAt, &agent.LastHeartbeat, &agent.Status,
		&capabilitiesJSON, &activeTasksJSON, &resourceUsageJSON, &metadataJSON,
		&agent.CreatedAt, &agent.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
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

// ListAgents retrieves all agents
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
		var startedAt sql.NullTime
		
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

// UpdateAgentHeartbeat updates agent heartbeat and status
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

// GetAllAgents is an alias for ListAgents
func (s *CompositeStore) GetAllAgents(ctx context.Context) ([]*models.AgentStatus, error) {
	return s.ListAgents(ctx)
}

// GetAgentsWithLocation retrieves agents with location metadata
func (s *CompositeStore) GetAgentsWithLocation(ctx context.Context) ([]*models.AgentStatus, error) {
	agents, err := s.ListAgents(ctx)
	if err != nil {
		return nil, err
	}
	
	// Filter agents that have location metadata
	var agentsWithLocation []*models.AgentStatus
	for _, agent := range agents {
		if _, hasLat := agent.Metadata["latitude"]; hasLat {
			if _, hasLon := agent.Metadata["longitude"]; hasLon {
				agentsWithLocation = append(agentsWithLocation, agent)
			}
		}
	}
	
	return agentsWithLocation, nil
}

// CacheMetric stores a metric in the cache
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