package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// StatusAggregator aggregates pipeline status from multiple sources
type StatusAggregator struct {
	k8sClient    kubernetes.Interface
	promClient   v1.API
	logger       *zap.Logger
	namespace    string
	cacheTTL     time.Duration
	cache        map[string]*AggregatedStatus
	cacheMutex   sync.RWMutex
	lastCacheTime time.Time
}

// AggregatedStatus represents the aggregated status of a pipeline deployment
type AggregatedStatus struct {
	DeploymentID   string                 `json:"deployment_id"`
	PipelineName   string                 `json:"pipeline_name"`
	Namespace      string                 `json:"namespace"`
	CRStatus       *CRStatus              `json:"cr_status"`
	CollectorHealth *CollectorHealth       `json:"collector_health"`
	Metrics        *PipelineMetrics       `json:"metrics"`
	ProcessRetention *ProcessRetention     `json:"process_retention"`
	LastUpdated    time.Time              `json:"last_updated"`
	OverallHealth  string                 `json:"overall_health"`
	Issues         []string               `json:"issues,omitempty"`
}

// CRStatus represents the status from the PhoenixProcessPipeline CR
type CRStatus struct {
	Phase      string    `json:"phase"`
	Ready      bool      `json:"ready"`
	Message    string    `json:"message,omitempty"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CollectorHealth represents the health of OTel collectors
type CollectorHealth struct {
	TotalInstances   int     `json:"total_instances"`
	HealthyInstances int     `json:"healthy_instances"`
	CPUUsage         float64 `json:"cpu_usage_percent"`
	MemoryUsage      float64 `json:"memory_usage_percent"`
	RestartCount     int     `json:"restart_count"`
	Uptime           string  `json:"uptime"`
}

// PipelineMetrics represents metrics about the pipeline performance
type PipelineMetrics struct {
	InputRate          float64 `json:"input_rate_per_sec"`
	OutputRate         float64 `json:"output_rate_per_sec"`
	DroppedRate        float64 `json:"dropped_rate_per_sec"`
	ProcessingLatency  float64 `json:"processing_latency_ms"`
	CardinalityBefore  int64   `json:"cardinality_before"`
	CardinalityAfter   int64   `json:"cardinality_after"`
	CardinalityReduction float64 `json:"cardinality_reduction_percent"`
}

// ProcessRetention represents which processes are being retained
type ProcessRetention struct {
	TotalProcesses    int      `json:"total_processes"`
	RetainedProcesses int      `json:"retained_processes"`
	TopProcesses      []string `json:"top_processes"`
	CriticalProcesses []string `json:"critical_processes"`
	FilteringStrategy string   `json:"filtering_strategy"`
}

// NewStatusAggregator creates a new status aggregator
func NewStatusAggregator(k8sClient kubernetes.Interface, promEndpoint, namespace string, logger *zap.Logger) (*StatusAggregator, error) {
	// Create Prometheus client
	promClient, err := api.NewClient(api.Config{
		Address: promEndpoint,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	return &StatusAggregator{
		k8sClient:  k8sClient,
		promClient: v1.NewAPI(promClient),
		logger:     logger,
		namespace:  namespace,
		cacheTTL:   30 * time.Second,
		cache:      make(map[string]*AggregatedStatus),
	}, nil
}

// GetStatus retrieves aggregated status for a pipeline deployment
func (a *StatusAggregator) GetStatus(ctx context.Context, deploymentID string) (*AggregatedStatus, error) {
	// Check cache first
	if cached := a.getCached(deploymentID); cached != nil {
		return cached, nil
	}

	// Aggregate status from various sources
	status := &AggregatedStatus{
		DeploymentID: deploymentID,
		Namespace:    a.namespace,
		LastUpdated:  time.Now(),
		Issues:       []string{},
	}

	// Collect data concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	// Get CR status
	wg.Add(1)
	go func() {
		defer wg.Done()
		if crStatus, err := a.getCRStatus(ctx, deploymentID); err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("CR status: %w", err))
			mu.Unlock()
		} else {
			status.CRStatus = crStatus
		}
	}()

	// Get collector health
	wg.Add(1)
	go func() {
		defer wg.Done()
		if health, err := a.getCollectorHealth(ctx, deploymentID); err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("collector health: %w", err))
			mu.Unlock()
		} else {
			status.CollectorHealth = health
		}
	}()

	// Get pipeline metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		if metrics, err := a.getPipelineMetrics(ctx, deploymentID); err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("pipeline metrics: %w", err))
			mu.Unlock()
		} else {
			status.Metrics = metrics
		}
	}()

	// Get process retention info
	wg.Add(1)
	go func() {
		defer wg.Done()
		if retention, err := a.getProcessRetention(ctx, deploymentID); err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("process retention: %w", err))
			mu.Unlock()
		} else {
			status.ProcessRetention = retention
		}
	}()

	wg.Wait()

	// Log any errors
	for _, err := range errors {
		a.logger.Warn("failed to collect status data", zap.Error(err))
		status.Issues = append(status.Issues, err.Error())
	}

	// Determine overall health
	status.OverallHealth = a.calculateOverallHealth(status)

	// Cache the result
	a.updateCache(deploymentID, status)

	return status, nil
}

// getCRStatus retrieves status from the Kubernetes CR
func (a *StatusAggregator) getCRStatus(ctx context.Context, deploymentID string) (*CRStatus, error) {
	// This would normally query the PhoenixProcessPipeline CR
	// For now, return a mock implementation
	return &CRStatus{
		Phase:     "Running",
		Ready:     true,
		UpdatedAt: time.Now(),
	}, nil
}

// getCollectorHealth retrieves health metrics for collectors
func (a *StatusAggregator) getCollectorHealth(ctx context.Context, deploymentID string) (*CollectorHealth, error) {
	// Query Prometheus for collector metrics
	queries := map[string]string{
		"cpu_usage":     fmt.Sprintf(`avg(rate(container_cpu_usage_seconds_total{pod=~"otelcol-%s-.*"}[5m])) * 100`, deploymentID),
		"memory_usage":  fmt.Sprintf(`avg(container_memory_usage_bytes{pod=~"otelcol-%s-.*"}) / avg(container_spec_memory_limit_bytes{pod=~"otelcol-%s-.*"}) * 100`, deploymentID, deploymentID),
		"restart_count": fmt.Sprintf(`sum(kube_pod_container_status_restarts_total{pod=~"otelcol-%s-.*"})`, deploymentID),
		"uptime":        fmt.Sprintf(`min(time() - kube_pod_start_time{pod=~"otelcol-%s-.*"})`, deploymentID),
	}

	health := &CollectorHealth{}

	for metric, query := range queries {
		result, _, err := a.promClient.Query(ctx, query, time.Now())
		if err != nil {
			a.logger.Warn("failed to query metric", zap.String("metric", metric), zap.Error(err))
			continue
		}

		if vector, ok := result.(model.Vector); ok && len(vector) > 0 {
			value := float64(vector[0].Value)
			switch metric {
			case "cpu_usage":
				health.CPUUsage = value
			case "memory_usage":
				health.MemoryUsage = value
			case "restart_count":
				health.RestartCount = int(value)
			case "uptime":
				health.Uptime = time.Duration(value * float64(time.Second)).String()
			}
		}
	}

	// Get pod count
	pods, err := a.k8sClient.CoreV1().Pods(a.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=otelcol,deployment=%s", deploymentID),
	})
	if err == nil {
		health.TotalInstances = len(pods.Items)
		for _, pod := range pods.Items {
			if pod.Status.Phase == "Running" {
				health.HealthyInstances++
			}
		}
	}

	return health, nil
}

// getPipelineMetrics retrieves pipeline performance metrics
func (a *StatusAggregator) getPipelineMetrics(ctx context.Context, deploymentID string) (*PipelineMetrics, error) {
	metrics := &PipelineMetrics{}

	// Query Prometheus for pipeline metrics
	queries := map[string]string{
		"input_rate":    fmt.Sprintf(`rate(otelcol_receiver_accepted_metric_points{deployment="%s"}[5m])`, deploymentID),
		"output_rate":   fmt.Sprintf(`rate(otelcol_exporter_sent_metric_points{deployment="%s"}[5m])`, deploymentID),
		"dropped_rate":  fmt.Sprintf(`rate(otelcol_processor_dropped_metric_points{deployment="%s"}[5m])`, deploymentID),
		"latency":       fmt.Sprintf(`histogram_quantile(0.95, rate(otelcol_processor_batch_batch_processing_duration_bucket{deployment="%s"}[5m]))`, deploymentID),
		"cardinality_before": fmt.Sprintf(`phoenix_pipeline_cardinality_before{deployment="%s"}`, deploymentID),
		"cardinality_after":  fmt.Sprintf(`phoenix_pipeline_cardinality_after{deployment="%s"}`, deploymentID),
	}

	for metric, query := range queries {
		result, _, err := a.promClient.Query(ctx, query, time.Now())
		if err != nil {
			a.logger.Warn("failed to query metric", zap.String("metric", metric), zap.Error(err))
			continue
		}

		if vector, ok := result.(model.Vector); ok && len(vector) > 0 {
			value := float64(vector[0].Value)
			switch metric {
			case "input_rate":
				metrics.InputRate = value
			case "output_rate":
				metrics.OutputRate = value
			case "dropped_rate":
				metrics.DroppedRate = value
			case "latency":
				metrics.ProcessingLatency = value * 1000 // Convert to ms
			case "cardinality_before":
				metrics.CardinalityBefore = int64(value)
			case "cardinality_after":
				metrics.CardinalityAfter = int64(value)
			}
		}
	}

	// Calculate cardinality reduction
	if metrics.CardinalityBefore > 0 {
		metrics.CardinalityReduction = float64(metrics.CardinalityBefore-metrics.CardinalityAfter) / float64(metrics.CardinalityBefore) * 100
	}

	return metrics, nil
}

// getProcessRetention retrieves information about retained processes
func (a *StatusAggregator) getProcessRetention(ctx context.Context, deploymentID string) (*ProcessRetention, error) {
	retention := &ProcessRetention{
		TopProcesses:      []string{},
		CriticalProcesses: []string{},
	}

	// Query for total process count
	totalQuery := fmt.Sprintf(`count(count by (process_name) (process_cpu_seconds_total{deployment="%s"}))`, deploymentID)
	result, _, err := a.promClient.Query(ctx, totalQuery, time.Now())
	if err == nil {
		if vector, ok := result.(model.Vector); ok && len(vector) > 0 {
			retention.TotalProcesses = int(vector[0].Value)
		}
	}

	// Query for retained process count
	retainedQuery := fmt.Sprintf(`count(count by (process_name) (process_cpu_seconds_total{deployment="%s",phoenix_retained="true"}))`, deploymentID)
	result, _, err = a.promClient.Query(ctx, retainedQuery, time.Now())
	if err == nil {
		if vector, ok := result.(model.Vector); ok && len(vector) > 0 {
			retention.RetainedProcesses = int(vector[0].Value)
		}
	}

	// Get top processes by CPU
	topQuery := fmt.Sprintf(`topk(5, avg by (process_name) (rate(process_cpu_seconds_total{deployment="%s",phoenix_retained="true"}[5m])))`, deploymentID)
	result, _, err = a.promClient.Query(ctx, topQuery, time.Now())
	if err == nil {
		if vector, ok := result.(model.Vector); ok {
			for _, sample := range vector {
				if processName, ok := sample.Metric["process_name"]; ok {
					retention.TopProcesses = append(retention.TopProcesses, string(processName))
				}
			}
		}
	}

	// Determine filtering strategy based on deployment config
	// This would normally be retrieved from the deployment configuration
	retention.FilteringStrategy = "top-k"

	return retention, nil
}

// calculateOverallHealth determines the overall health status
func (a *StatusAggregator) calculateOverallHealth(status *AggregatedStatus) string {
	if status.CRStatus != nil && !status.CRStatus.Ready {
		return "Unhealthy"
	}

	if status.CollectorHealth != nil {
		if status.CollectorHealth.HealthyInstances < status.CollectorHealth.TotalInstances {
			return "Degraded"
		}
		if status.CollectorHealth.RestartCount > 5 {
			return "Unstable"
		}
		if status.CollectorHealth.CPUUsage > 80 || status.CollectorHealth.MemoryUsage > 80 {
			return "Warning"
		}
	}

	if status.Metrics != nil {
		if status.Metrics.DroppedRate > status.Metrics.OutputRate*0.1 { // More than 10% dropped
			return "Degraded"
		}
		if status.Metrics.ProcessingLatency > 1000 { // More than 1 second
			return "Warning"
		}
	}

	return "Healthy"
}

// Cache management methods
func (a *StatusAggregator) getCached(deploymentID string) *AggregatedStatus {
	a.cacheMutex.RLock()
	defer a.cacheMutex.RUnlock()

	if time.Since(a.lastCacheTime) > a.cacheTTL {
		return nil
	}

	return a.cache[deploymentID]
}

func (a *StatusAggregator) updateCache(deploymentID string, status *AggregatedStatus) {
	a.cacheMutex.Lock()
	defer a.cacheMutex.Unlock()

	a.cache[deploymentID] = status
	a.lastCacheTime = time.Now()
}