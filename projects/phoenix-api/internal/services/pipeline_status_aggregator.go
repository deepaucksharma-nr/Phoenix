package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/phoenix/platform/pkg/common/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/store"
	"go.uber.org/zap"
)

// AgentStatus represents agent status
type AgentStatus struct {
	AgentID   string    `json:"agent_id"`
	HostName  string    `json:"host_name"`
	Status    string    `json:"status"`
	LastSeen  time.Time `json:"last_seen"`
	Pipelines []string  `json:"pipelines"`
}

// AgentCondition represents agent condition
type AgentCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// AgentClient interface for agent operations
type AgentClient interface {
	GetAgentStatus(ctx context.Context, agentID string) (*AgentStatus, error)
	ListAgents(ctx context.Context) ([]*AgentStatus, error)
}

// PipelineStatusAggregator aggregates status from multiple sources
type PipelineStatusAggregator struct {
	store            store.PipelineDeploymentStore
	metricsCollector MetricsCollector
	agentClient      AgentClient
	logger           *zap.Logger
}

// MetricsCollector interface for collecting metrics
type MetricsCollector interface {
	// GetPipelineMetrics retrieves metrics for a pipeline deployment
	GetPipelineMetrics(ctx context.Context, deploymentID string) (*models.DeploymentMetrics, error)
	// GetCollectorHealth retrieves health status of collectors (placeholder for now)
	GetCollectorHealth(ctx context.Context, agentID string) (string, error)
}

// NewPipelineStatusAggregator creates a new status aggregator
func NewPipelineStatusAggregator(
	store store.PipelineDeploymentStore,
	metricsCollector MetricsCollector,
	agentClient AgentClient,
	logger *zap.Logger,
) *PipelineStatusAggregator {
	return &PipelineStatusAggregator{
		store:            store,
		metricsCollector: metricsCollector,
		agentClient:      agentClient,
		logger:           logger,
	}
}

// AggregatedStatus contains comprehensive pipeline status
type AggregatedStatus struct {
	DeploymentID      string                      `json:"deployment_id"`
	DeploymentName    string                      `json:"deployment_name"`
	PipelineName      string                      `json:"pipeline_name"`
	Namespace         string                      `json:"namespace"`
	Status            string                      `json:"status"`
	Phase             string                      `json:"phase"`
	Instances         *models.DeploymentInstances `json:"instances,omitempty"`
	Metrics           *models.DeploymentMetrics   `json:"metrics,omitempty"`
	HealthStatus      string                      `json:"health_status,omitempty"`
	CollectorStatuses []CollectorStatus           `json:"collector_statuses,omitempty"`
	LastUpdated       time.Time                   `json:"last_updated"`
	Summary           StatusSummary               `json:"summary"`
}

// CollectorStatus represents individual collector status
type CollectorStatus struct {
	AgentID      string            `json:"agent_id"`
	HostName     string            `json:"host_name"`
	PodName      string            `json:"pod_name"`
	NodeName     string            `json:"node_name"`
	Status       string            `json:"status"`
	Ready        bool              `json:"ready"`
	RestartCount int32             `json:"restart_count"`
	StartTime    *time.Time        `json:"start_time,omitempty"`
	Conditions   []AgentCondition  `json:"conditions,omitempty"`
	Metrics      *CollectorMetrics `json:"metrics,omitempty"`
}

// CollectorMetrics contains collector-specific metrics
type CollectorMetrics struct {
	ProcessedMetrics int64   `json:"processed_metrics"`
	DroppedMetrics   int64   `json:"dropped_metrics"`
	ExportErrors     int64   `json:"export_errors"`
	CPUUsagePercent  float64 `json:"cpu_usage_percent"`
	MemoryUsageMB    float64 `json:"memory_usage_mb"`
}

// StatusSummary provides a high-level summary
type StatusSummary struct {
	HealthyCollectors    int      `json:"healthy_collectors"`
	UnhealthyCollectors  int      `json:"unhealthy_collectors"`
	TotalMetricsRate     float64  `json:"total_metrics_rate"`
	ErrorRate            float64  `json:"error_rate"`
	CardinalityReduction float64  `json:"cardinality_reduction,omitempty"`
	IsHealthy            bool     `json:"is_healthy"`
	Issues               []string `json:"issues,omitempty"`
}

// GetAggregatedStatus retrieves comprehensive status for a deployment
func (a *PipelineStatusAggregator) GetAggregatedStatus(ctx context.Context, deploymentID string) (*AggregatedStatus, error) {
	a.logger.Info("aggregating pipeline status", zap.String("deployment_id", deploymentID))

	// Get deployment from store
	deployment, err := a.store.GetDeployment(ctx, deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	status := &AggregatedStatus{
		DeploymentID:   deployment.ID,
		DeploymentName: deployment.DeploymentName,
		PipelineName:   deployment.PipelineName,
		Namespace:      deployment.Namespace,
		Status:         deployment.Status,
		Phase:          deployment.Phase,
		Instances:      deployment.Instances,
		LastUpdated:    time.Now(),
	}

	// Collect data from multiple sources concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	// Collect metrics if a collector is configured
	if a.metricsCollector != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metrics, err := a.metricsCollector.GetPipelineMetrics(ctx, deploymentID)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("metrics collection failed: %w", err))
				mu.Unlock()
				return
			}
			mu.Lock()
			status.Metrics = metrics
			mu.Unlock()
		}()
	}

	// Collect health status if a collector is configured
	if a.metricsCollector != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// For now, use deployment ID as agent identifier
			health, err := a.metricsCollector.GetCollectorHealth(ctx, deployment.ID)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("health collection failed: %w", err))
				mu.Unlock()
				return
			}
			mu.Lock()
			status.HealthStatus = health
			mu.Unlock()
		}()
	}

	// Collect agent status if an agent client is configured
	if a.agentClient != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			collectorStatuses, err := a.getCollectorStatuses(ctx, deployment)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("agent status collection failed: %w", err))
				mu.Unlock()
				return
			}
			mu.Lock()
			status.CollectorStatuses = collectorStatuses
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Log any errors that occurred
	for _, err := range errors {
		a.logger.Warn("error during status aggregation", zap.Error(err))
	}

	// Generate summary
	status.Summary = a.generateSummary(status)

	return status, nil
}

// getCollectorStatuses retrieves status of agents running collectors
func (a *PipelineStatusAggregator) getCollectorStatuses(ctx context.Context, deployment *models.PipelineDeployment) ([]CollectorStatus, error) {
	if a.agentClient == nil {
		a.logger.Warn("agent client not configured, skipping agent status collection")
		return []CollectorStatus{}, nil
	}

	// List all agents
	agents, err := a.agentClient.ListAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	statuses := make([]CollectorStatus, 0)
	for _, agent := range agents {
		// Check if this agent is running the deployment's pipeline
		for _, pipeline := range agent.Pipelines {
			if pipeline == deployment.PipelineName {
				status := CollectorStatus{
					AgentID:      agent.AgentID,
					HostName:     agent.HostName,
					Status:       agent.Status,
					Ready:        agent.Status == "healthy",
					RestartCount: 0, // TODO: Track agent restarts
					Conditions:   []AgentCondition{},
				}

				// Set start time based on last seen
				startTime := agent.LastSeen
				status.StartTime = &startTime

				// TODO: Get collector-specific metrics from agent metrics endpoint
				// This would require additional integration with metrics backend

				statuses = append(statuses, status)
				break
			}
		}
	}

	return statuses, nil
}

// generateSummary creates a status summary
func (a *PipelineStatusAggregator) generateSummary(status *AggregatedStatus) StatusSummary {
	summary := StatusSummary{
		Issues: []string{},
	}

	// Count healthy/unhealthy collectors
	for _, collector := range status.CollectorStatuses {
		if collector.Ready && collector.Status == "Running" {
			summary.HealthyCollectors++
		} else {
			summary.UnhealthyCollectors++
		}

		// Check for high restart counts
		if collector.RestartCount > 5 {
			summary.Issues = append(summary.Issues,
				fmt.Sprintf("Collector %s has high restart count: %d", collector.AgentID, collector.RestartCount))
		}
	}

	// Calculate metrics if available
	if status.Metrics != nil {
		summary.TotalMetricsRate = status.Metrics.MetricsPerSecond
		summary.ErrorRate = status.Metrics.ErrorRate
		summary.CardinalityReduction = status.Metrics.CardinalityReduction

		// Check for high error rates
		if summary.ErrorRate > 0.05 { // 5% error rate threshold
			summary.Issues = append(summary.Issues,
				fmt.Sprintf("High error rate detected: %.2f%%", summary.ErrorRate*100))
		}
	}

	// Check instance health
	if status.Instances != nil {
		if status.Instances.Ready < status.Instances.Desired {
			summary.Issues = append(summary.Issues,
				fmt.Sprintf("Only %d/%d instances are ready", status.Instances.Ready, status.Instances.Desired))
		}
	}

	// Determine overall health
	summary.IsHealthy = summary.UnhealthyCollectors == 0 &&
		summary.ErrorRate < 0.05 &&
		len(summary.Issues) == 0

	return summary
}


// UpdateDeploymentStatusFromAggregation updates deployment status based on aggregated data
func (a *PipelineStatusAggregator) UpdateDeploymentStatusFromAggregation(ctx context.Context, deploymentID string) error {
	// Get aggregated status
	aggregatedStatus, err := a.GetAggregatedStatus(ctx, deploymentID)
	if err != nil {
		return fmt.Errorf("failed to get aggregated status: %w", err)
	}

	// Prepare update request
	updateReq := &models.UpdateDeploymentRequest{
		UpdatedBy: "system-aggregator",
	}

	// Update status based on summary
	if !aggregatedStatus.Summary.IsHealthy {
		updateReq.Status = models.DeploymentStatusDegraded
		updateReq.StatusMessage = "Deployment is degraded"
	} else if aggregatedStatus.Summary.HealthyCollectors == 0 {
		updateReq.Status = models.DeploymentStatusFailed
		updateReq.StatusMessage = "No healthy collectors"
	} else {
		updateReq.Status = models.DeploymentStatusHealthy
		updateReq.StatusMessage = "Deployment is healthy"
	}

	// Update metrics if available
	if aggregatedStatus.Metrics != nil {
		if err := a.store.UpdateDeploymentMetrics(ctx, deploymentID, aggregatedStatus.Metrics); err != nil {
			a.logger.Error("failed to update deployment metrics", zap.Error(err))
		}
	}

	// TODO: Update health status once store method is implemented
	if aggregatedStatus.HealthStatus != "" {
		a.logger.Debug("health status available", zap.String("status", aggregatedStatus.HealthStatus))
	}

	// Update deployment status
	if err := a.store.UpdateDeployment(ctx, deploymentID, updateReq); err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	a.logger.Info("deployment status updated from aggregation",
		zap.String("deployment_id", deploymentID),
		zap.String("status", updateReq.Status),
		zap.Bool("is_healthy", aggregatedStatus.Summary.IsHealthy))

	return nil
}