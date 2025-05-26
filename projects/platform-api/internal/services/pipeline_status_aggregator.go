package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/phoenix/platform/packages/go-common/models"
	"github.com/phoenix/platform/projects/platform-api/internal/store"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PipelineStatusAggregator aggregates status from multiple sources
type PipelineStatusAggregator struct {
	store            store.PipelineDeploymentStore
	metricsCollector MetricsCollector
	k8sClient        kubernetes.Interface
	logger           *zap.Logger
}

// MetricsCollector interface for collecting metrics
type MetricsCollector interface {
	// GetPipelineMetrics retrieves metrics for a pipeline deployment
	GetPipelineMetrics(ctx context.Context, deploymentID string) (*models.DeploymentMetrics, error)
	// GetCollectorHealth retrieves health status of collectors (placeholder for now)
	GetCollectorHealth(ctx context.Context, namespace string, selector map[string]string) (string, error)
}

// NewPipelineStatusAggregator creates a new status aggregator
func NewPipelineStatusAggregator(
	store store.PipelineDeploymentStore,
	metricsCollector MetricsCollector,
	k8sClient kubernetes.Interface,
	logger *zap.Logger,
) *PipelineStatusAggregator {
	return &PipelineStatusAggregator{
		store:            store,
		metricsCollector: metricsCollector,
		k8sClient:        k8sClient,
		logger:           logger,
	}
}

// AggregatedStatus contains comprehensive pipeline status
type AggregatedStatus struct {
	DeploymentID      string                        `json:"deployment_id"`
	DeploymentName    string                        `json:"deployment_name"`
	PipelineName      string                        `json:"pipeline_name"`
	Namespace         string                        `json:"namespace"`
	Status            string                        `json:"status"`
	Phase             string                        `json:"phase"`
	Instances         *models.DeploymentInstances   `json:"instances,omitempty"`
	Metrics           *models.DeploymentMetrics     `json:"metrics,omitempty"`
	HealthStatus      string                        `json:"health_status,omitempty"`
	CollectorStatuses []CollectorStatus             `json:"collector_statuses,omitempty"`
	LastUpdated       time.Time                     `json:"last_updated"`
	Summary           StatusSummary                 `json:"summary"`
}

// CollectorStatus represents individual collector status
type CollectorStatus struct {
	PodName     string                 `json:"pod_name"`
	NodeName    string                 `json:"node_name"`
	Status      string                 `json:"status"`
	Ready       bool                   `json:"ready"`
	RestartCount int32                 `json:"restart_count"`
	StartTime   *time.Time             `json:"start_time,omitempty"`
	Conditions  []v1.PodCondition      `json:"conditions,omitempty"`
	Metrics     *CollectorMetrics      `json:"metrics,omitempty"`
}

// CollectorMetrics contains collector-specific metrics
type CollectorMetrics struct {
	ProcessedMetrics   int64   `json:"processed_metrics"`
	DroppedMetrics     int64   `json:"dropped_metrics"`
	ExportErrors       int64   `json:"export_errors"`
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryUsageMB      float64 `json:"memory_usage_mb"`
}

// StatusSummary provides a high-level summary
type StatusSummary struct {
	HealthyCollectors   int     `json:"healthy_collectors"`
	UnhealthyCollectors int     `json:"unhealthy_collectors"`
	TotalMetricsRate    float64 `json:"total_metrics_rate"`
	ErrorRate           float64 `json:"error_rate"`
	CardinalityReduction float64 `json:"cardinality_reduction,omitempty"`
	IsHealthy           bool    `json:"is_healthy"`
	Issues              []string `json:"issues,omitempty"`
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

	// Collect metrics
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

	// Collect health status
	wg.Add(1)
	go func() {
		defer wg.Done()
		selector := map[string]string{
			"deployment": deployment.DeploymentName,
			"pipeline":   deployment.PipelineName,
		}
		health, err := a.metricsCollector.GetCollectorHealth(ctx, deployment.Namespace, selector)
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

	// Collect Kubernetes pod status
	wg.Add(1)
	go func() {
		defer wg.Done()
		collectorStatuses, err := a.getCollectorStatuses(ctx, deployment)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("pod status collection failed: %w", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		status.CollectorStatuses = collectorStatuses
		mu.Unlock()
	}()

	wg.Wait()

	// Log any errors that occurred
	for _, err := range errors {
		a.logger.Warn("error during status aggregation", zap.Error(err))
	}

	// Generate summary
	status.Summary = a.generateSummary(status)

	return status, nil
}

// getCollectorStatuses retrieves status of collector pods
func (a *PipelineStatusAggregator) getCollectorStatuses(ctx context.Context, deployment *models.PipelineDeployment) ([]CollectorStatus, error) {
	// List pods with deployment labels
	labelSelector := fmt.Sprintf("deployment=%s,pipeline=%s", deployment.DeploymentName, deployment.PipelineName)
	pods, err := a.k8sClient.CoreV1().Pods(deployment.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	statuses := make([]CollectorStatus, 0, len(pods.Items))
	for _, pod := range pods.Items {
		status := CollectorStatus{
			PodName:    pod.Name,
			NodeName:   pod.Spec.NodeName,
			Status:     string(pod.Status.Phase),
			Ready:      isPodReady(&pod),
			Conditions: pod.Status.Conditions,
		}

		// Get container restart count
		for _, containerStatus := range pod.Status.ContainerStatuses {
			status.RestartCount += containerStatus.RestartCount
		}

		// Get start time
		if pod.Status.StartTime != nil {
			startTime := pod.Status.StartTime.Time
			status.StartTime = &startTime
		}

		// TODO: Get collector-specific metrics from Prometheus or internal metrics endpoint
		// This would require additional integration with metrics backend

		statuses = append(statuses, status)
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
				fmt.Sprintf("Collector %s has high restart count: %d", collector.PodName, collector.RestartCount))
		}
	}

	// Calculate metrics if available
	if status.Metrics != nil {
		// TODO: Add MetricsPerSecond field to DeploymentMetrics model
	// summary.TotalMetricsRate = status.Metrics.MetricsPerSecond
		summary.ErrorRate = status.Metrics.ErrorRate
		// TODO: Add CardinalityReduction field to DeploymentMetrics model
		// summary.CardinalityReduction = status.Metrics.CardinalityReduction

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

// isPodReady checks if a pod is ready
func isPodReady(pod *v1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodReady {
			return condition.Status == v1.ConditionTrue
		}
	}
	return false
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
		// TODO: Add UpdatedBy field to UpdateDeploymentRequest model
	// UpdatedBy: "system-aggregator",
	}

	// Update status based on summary
	if !aggregatedStatus.Summary.IsHealthy {
		updateReq.Status = models.DeploymentStatusFailed // Use existing status instead of DeploymentStatusDegraded
		// TODO: Add StatusMessage field to UpdateDeploymentRequest model
	} else if aggregatedStatus.Summary.HealthyCollectors == 0 {
		updateReq.Status = models.DeploymentStatusFailed
		// TODO: Add StatusMessage field to UpdateDeploymentRequest model
	} else {
		updateReq.Status = models.DeploymentStatusActive // Use existing status
		// TODO: Add StatusMessage field to UpdateDeploymentRequest model
	}

	// Update metrics if available
	if aggregatedStatus.Metrics != nil {
		// TODO: Implement UpdateDeploymentMetrics method in store
		a.logger.Debug("would update deployment metrics here", 
			zap.String("deployment_id", deploymentID),
			zap.Float64("cardinality", float64(aggregatedStatus.Metrics.Cardinality)))
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