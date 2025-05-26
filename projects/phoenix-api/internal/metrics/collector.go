package metrics

import (
	"context"
	"fmt"
	"time"
	
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/rs/zerolog/log"
)

// Collector interfaces with Prometheus to collect metrics
type Collector struct {
	promAPI v1.API
}

// NewCollector creates a new metrics collector
func NewCollector(promURL string) (*Collector, error) {
	client, err := api.NewClient(api.Config{
		Address: promURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}
	
	return &Collector{
		promAPI: v1.NewAPI(client),
	}, nil
}

// CollectExperimentMetrics collects all metrics for an experiment
func (c *Collector) CollectExperimentMetrics(ctx context.Context, experimentID string, timeRange time.Duration) (*ExperimentMetrics, error) {
	endTime := time.Now()
	startTime := endTime.Add(-timeRange)
	
	metrics := &ExperimentMetrics{
		ExperimentID: experimentID,
		StartTime:    startTime,
		EndTime:      endTime,
		Baseline:     &PipelineMetrics{},
		Candidate:    &PipelineMetrics{},
	}
	
	// Collect baseline metrics
	if err := c.collectPipelineMetrics(ctx, experimentID, "baseline", startTime, endTime, metrics.Baseline); err != nil {
		return nil, fmt.Errorf("failed to collect baseline metrics: %w", err)
	}
	
	// Collect candidate metrics
	if err := c.collectPipelineMetrics(ctx, experimentID, "candidate", startTime, endTime, metrics.Candidate); err != nil {
		return nil, fmt.Errorf("failed to collect candidate metrics: %w", err)
	}
	
	return metrics, nil
}

// collectPipelineMetrics collects metrics for a specific pipeline variant
func (c *Collector) collectPipelineMetrics(ctx context.Context, experimentID, variant string, start, end time.Time, pm *PipelineMetrics) error {
	pm.Variant = variant
	
	// Get cardinality (unique time series count)
	cardinality, err := c.getCardinality(ctx, experimentID, variant, end)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get cardinality")
	} else {
		pm.Cardinality = int64(cardinality)
	}
	
	// Get CPU usage
	cpuUsage, err := c.getResourceUsage(ctx, experimentID, variant, "cpu", start, end)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get CPU usage")
	} else {
		pm.CPUUsage = cpuUsage
	}
	
	// Get memory usage
	memUsage, err := c.getResourceUsage(ctx, experimentID, variant, "memory", start, end)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get memory usage")
	} else {
		pm.MemoryUsageMB = memUsage
	}
	
	// Get ingestion rate
	ingestRate, err := c.getIngestRate(ctx, experimentID, variant, start, end)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get ingest rate")
	} else {
		pm.IngestRate = ingestRate
	}
	
	// Get error rate
	errorRate, err := c.getErrorRate(ctx, experimentID, variant, start, end)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get error rate")
	} else {
		pm.ErrorRate = errorRate
	}
	
	// Get top metrics by cardinality
	topMetrics, err := c.getTopMetrics(ctx, experimentID, variant, end, 10)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get top metrics")
	} else {
		pm.TopMetrics = topMetrics
	}
	
	return nil
}

// getCardinality returns the number of unique time series
func (c *Collector) getCardinality(ctx context.Context, experimentID, variant string, timestamp time.Time) (float64, error) {
	query := fmt.Sprintf(`
		count(
			group by (__name__, job, instance) (
				{experiment_id="%s", variant="%s"}
			)
		)
	`, experimentID, variant)
	
	return c.queryScalar(ctx, query, timestamp)
}

// getResourceUsage returns CPU or memory usage
func (c *Collector) getResourceUsage(ctx context.Context, experimentID, variant, resource string, start, end time.Time) (float64, error) {
	var query string
	
	switch resource {
	case "cpu":
		query = fmt.Sprintf(`
			avg_over_time(
				otelcol_process_cpu_seconds{
					experiment_id="%s",
					variant="%s"
				}[%s]
			)
		`, experimentID, variant, end.Sub(start).String())
		
	case "memory":
		query = fmt.Sprintf(`
			avg_over_time(
				otelcol_process_memory_rss{
					experiment_id="%s",
					variant="%s"
				}[%s]
			) / 1024 / 1024
		`, experimentID, variant, end.Sub(start).String())
		
	default:
		return 0, fmt.Errorf("unknown resource type: %s", resource)
	}
	
	return c.queryScalar(ctx, query, end)
}

// getIngestRate returns the rate of metrics ingestion
func (c *Collector) getIngestRate(ctx context.Context, experimentID, variant string, start, end time.Time) (float64, error) {
	query := fmt.Sprintf(`
		rate(
			otelcol_receiver_accepted_metric_points{
				experiment_id="%s",
				variant="%s"
			}[%s]
		)
	`, experimentID, variant, end.Sub(start).String())
	
	return c.queryScalar(ctx, query, end)
}

// getErrorRate returns the error rate for the pipeline
func (c *Collector) getErrorRate(ctx context.Context, experimentID, variant string, start, end time.Time) (float64, error) {
	query := fmt.Sprintf(`
		rate(
			otelcol_processor_dropped_metric_points{
				experiment_id="%s",
				variant="%s"
			}[%s]
		) / 
		rate(
			otelcol_receiver_accepted_metric_points{
				experiment_id="%s",
				variant="%s"
			}[%s]
		) * 100
	`, experimentID, variant, end.Sub(start).String(), experimentID, variant, end.Sub(start).String())
	
	return c.queryScalar(ctx, query, end)
}

// getTopMetrics returns the top N metrics by cardinality
func (c *Collector) getTopMetrics(ctx context.Context, experimentID, variant string, timestamp time.Time, limit int) ([]MetricInfo, error) {
	query := fmt.Sprintf(`
		topk(%d,
			count by (__name__) (
				{experiment_id="%s", variant="%s"}
			)
		)
	`, limit, experimentID, variant)
	
	value, warnings, err := c.promAPI.Query(ctx, query, timestamp)
	if err != nil {
		return nil, err
	}
	
	if len(warnings) > 0 {
		log.Warn().Strs("warnings", warnings).Msg("Prometheus query warnings")
	}
	
	var metrics []MetricInfo
	
	if vector, ok := value.(model.Vector); ok {
		for _, sample := range vector {
			metricName := string(sample.Metric["__name__"])
			if metricName == "" {
				continue
			}
			
			metrics = append(metrics, MetricInfo{
				Name:        metricName,
				Cardinality: int64(sample.Value),
			})
		}
	}
	
	return metrics, nil
}

// queryScalar executes a Prometheus query and returns a scalar result
func (c *Collector) queryScalar(ctx context.Context, query string, timestamp time.Time) (float64, error) {
	value, warnings, err := c.promAPI.Query(ctx, query, timestamp)
	if err != nil {
		return 0, err
	}
	
	if len(warnings) > 0 {
		log.Warn().Strs("warnings", warnings).Msg("Prometheus query warnings")
	}
	
	switch v := value.(type) {
	case model.Vector:
		if len(v) > 0 {
			return float64(v[0].Value), nil
		}
		return 0, nil
		
	case *model.Scalar:
		return float64(v.Value), nil
		
	default:
		return 0, fmt.Errorf("unexpected value type: %T", value)
	}
}

// CheckCriticalMetrics verifies that critical metrics are still present
func (c *Collector) CheckCriticalMetrics(ctx context.Context, experimentID, variant string, criticalProcesses []string, timestamp time.Time) (map[string]bool, error) {
	results := make(map[string]bool)
	
	for _, process := range criticalProcesses {
		query := fmt.Sprintf(`
			count(
				{
					experiment_id="%s",
					variant="%s",
					process=~".*%s.*"
				}
			)
		`, experimentID, variant, process)
		
		count, err := c.queryScalar(ctx, query, timestamp)
		if err != nil {
			log.Error().Err(err).Str("process", process).Msg("Failed to check critical metric")
			results[process] = false
			continue
		}
		
		results[process] = count > 0
	}
	
	return results, nil
}

// ExperimentMetrics contains metrics for both pipeline variants
type ExperimentMetrics struct {
	ExperimentID string           `json:"experiment_id"`
	StartTime    time.Time        `json:"start_time"`
	EndTime      time.Time        `json:"end_time"`
	Baseline     *PipelineMetrics `json:"baseline"`
	Candidate    *PipelineMetrics `json:"candidate"`
}

// PipelineMetrics contains metrics for a single pipeline variant
type PipelineMetrics struct {
	Variant       string       `json:"variant"`
	Cardinality   int64        `json:"cardinality"`
	CPUUsage      float64      `json:"cpu_usage"`
	MemoryUsageMB float64      `json:"memory_usage_mb"`
	IngestRate    float64      `json:"ingest_rate"`
	ErrorRate     float64      `json:"error_rate"`
	TopMetrics    []MetricInfo `json:"top_metrics"`
}

// MetricInfo contains information about a specific metric
type MetricInfo struct {
	Name        string `json:"name"`
	Cardinality int64  `json:"cardinality"`
}