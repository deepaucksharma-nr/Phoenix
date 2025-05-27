package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/phoenix/platform/projects/phoenix-api/internal/analyzer"
	internalModels "github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/store"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/rs/zerolog/log"
)

type MetricsCollector struct {
	store        store.Store
	promClient   v1.API
	kpiCalc      *analyzer.KPICalculator
	collectors   map[string]*experimentCollector
	mu           sync.RWMutex
	pollInterval time.Duration
}

type experimentCollector struct {
	experimentID string
	cancel       context.CancelFunc
	metrics      chan *internalModels.Metric
}

type CollectedMetrics struct {
	ExperimentID string
	Timestamp    time.Time
	Baseline     MetricSet
	Candidate    MetricSet
}

type MetricSet struct {
	Cardinality       int64
	IngestRate        float64
	ResourceUsage     ResourceMetrics
	SampleMetrics     map[string]float64
	ProcessingLatency float64
}

type ResourceMetrics struct {
	CPUUsage    float64
	MemoryUsage float64
	NetworkIO   float64
}

func NewMetricsCollector(store store.Store, promURL string) (*MetricsCollector, error) {
	client, err := api.NewClient(api.Config{
		Address: promURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	kpiCalc, err := analyzer.NewKPICalculator(promURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create KPI calculator: %w", err)
	}

	return &MetricsCollector{
		store:        store,
		promClient:   v1.NewAPI(client),
		kpiCalc:      kpiCalc,
		collectors:   make(map[string]*experimentCollector),
		pollInterval: 15 * time.Second,
	}, nil
}

// StartCollection starts metrics collection for an experiment
func (mc *MetricsCollector) StartCollection(ctx context.Context, experimentID string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.collectors[experimentID]; exists {
		return fmt.Errorf("collection already started for experiment %s", experimentID)
	}

	ctx, cancel := context.WithCancel(ctx)
	collector := &experimentCollector{
		experimentID: experimentID,
		cancel:       cancel,
		metrics:      make(chan *internalModels.Metric, 100),
	}

	mc.collectors[experimentID] = collector

	// Start collection goroutine
	go mc.collectMetrics(ctx, collector)

	log.Info().
		Str("experiment_id", experimentID).
		Msg("Started metrics collection")

	return nil
}

// StopCollection stops metrics collection for an experiment
func (mc *MetricsCollector) StopCollection(experimentID string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	collector, exists := mc.collectors[experimentID]
	if !exists {
		return fmt.Errorf("no collection found for experiment %s", experimentID)
	}

	collector.cancel()
	close(collector.metrics)
	delete(mc.collectors, experimentID)

	log.Info().
		Str("experiment_id", experimentID).
		Msg("Stopped metrics collection")

	return nil
}

// collectMetrics runs the metrics collection loop for an experiment
func (mc *MetricsCollector) collectMetrics(ctx context.Context, collector *experimentCollector) {
	ticker := time.NewTicker(mc.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics, err := mc.fetchExperimentMetrics(ctx, collector.experimentID)
			if err != nil {
				log.Error().
					Err(err).
					Str("experiment_id", collector.experimentID).
					Msg("Failed to fetch metrics")
				continue
			}

			// Store metrics in database
			if err := mc.storeMetrics(ctx, metrics); err != nil {
				log.Error().
					Err(err).
					Str("experiment_id", collector.experimentID).
					Msg("Failed to store metrics")
			}

			// Calculate and store KPIs periodically (every 5 minutes)
			if time.Now().Unix()%(5*60) < int64(mc.pollInterval.Seconds()) {
				go mc.calculateAndStoreKPIs(context.Background(), collector.experimentID)
			}
		}
	}
}

// fetchExperimentMetrics queries Prometheus for experiment metrics
func (mc *MetricsCollector) fetchExperimentMetrics(ctx context.Context, experimentID string) (*CollectedMetrics, error) {
	now := time.Now()
	metrics := &CollectedMetrics{
		ExperimentID: experimentID,
		Timestamp:    now,
	}

	// Fetch baseline metrics
	baseline, err := mc.fetchVariantMetrics(ctx, experimentID, "baseline", now)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch baseline metrics: %w", err)
	}
	metrics.Baseline = baseline

	// Fetch candidate metrics
	candidate, err := mc.fetchVariantMetrics(ctx, experimentID, "candidate", now)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch candidate metrics: %w", err)
	}
	metrics.Candidate = candidate

	return metrics, nil
}

// fetchVariantMetrics fetches metrics for a specific variant
func (mc *MetricsCollector) fetchVariantMetrics(ctx context.Context, experimentID, variant string, timestamp time.Time) (MetricSet, error) {
	set := MetricSet{
		SampleMetrics: make(map[string]float64),
	}

	// Get cardinality
	cardQuery := fmt.Sprintf(`count(count by (__name__)({experiment_id="%s",variant="%s"}))`, experimentID, variant)
	cardinality, err := mc.queryScalar(ctx, cardQuery, timestamp)
	if err == nil {
		set.Cardinality = int64(cardinality)
	}

	// Get ingest rate
	ingestQuery := fmt.Sprintf(`sum(rate(prometheus_tsdb_samples_appended_total{experiment_id="%s",variant="%s"}[1m]))`, experimentID, variant)
	ingestRate, err := mc.queryScalar(ctx, ingestQuery, timestamp)
	if err == nil {
		set.IngestRate = ingestRate
	}

	// Get resource usage
	cpuQuery := fmt.Sprintf(`avg(rate(container_cpu_usage_seconds_total{container_name=~"otel-collector.*",experiment_id="%s",variant="%s"}[1m]))`, experimentID, variant)
	cpu, err := mc.queryScalar(ctx, cpuQuery, timestamp)
	if err == nil {
		set.ResourceUsage.CPUUsage = cpu
	}

	memQuery := fmt.Sprintf(`avg(container_memory_usage_bytes{container_name=~"otel-collector.*",experiment_id="%s",variant="%s"})`, experimentID, variant)
	mem, err := mc.queryScalar(ctx, memQuery, timestamp)
	if err == nil {
		set.ResourceUsage.MemoryUsage = mem
	}

	// Get sample metrics
	sampleMetrics := []string{
		"http_server_duration_seconds",
		"http_server_request_count_total",
		"process_cpu_seconds_total",
		"process_resident_memory_bytes",
	}

	for _, metric := range sampleMetrics {
		query := fmt.Sprintf(`avg(%s{experiment_id="%s",variant="%s"})`, metric, experimentID, variant)
		val, err := mc.queryScalar(ctx, query, timestamp)
		if err == nil {
			set.SampleMetrics[metric] = val
		}
	}

	// Get processing latency
	latencyQuery := fmt.Sprintf(`histogram_quantile(0.95, rate(otelcol_processor_process_duration_seconds_bucket{experiment_id="%s",variant="%s"}[1m]))`, experimentID, variant)
	latency, err := mc.queryScalar(ctx, latencyQuery, timestamp)
	if err == nil {
		set.ProcessingLatency = latency
	}

	return set, nil
}

// storeMetrics stores collected metrics in the database
func (mc *MetricsCollector) storeMetrics(ctx context.Context, metrics *CollectedMetrics) error {
	// Store baseline metrics
	baselineData := map[string]interface{}{
		"experiment_id": metrics.ExperimentID,
		"variant":       "baseline",
		"timestamp":     metrics.Timestamp,
		"cardinality":   metrics.Baseline.Cardinality,
		"ingest_rate":   metrics.Baseline.IngestRate,
		"cpu_usage":     metrics.Baseline.ResourceUsage.CPUUsage,
		"memory_usage":  metrics.Baseline.ResourceUsage.MemoryUsage,
		"metadata":      metrics.Baseline.SampleMetrics,
	}

	if err := mc.store.CacheMetric(ctx, metrics.ExperimentID, baselineData); err != nil {
		return fmt.Errorf("failed to store baseline metrics: %w", err)
	}

	// Store candidate metrics
	candidateData := map[string]interface{}{
		"experiment_id": metrics.ExperimentID,
		"variant":       "candidate",
		"timestamp":     metrics.Timestamp,
		"cardinality":   metrics.Candidate.Cardinality,
		"ingest_rate":   metrics.Candidate.IngestRate,
		"cpu_usage":     metrics.Candidate.ResourceUsage.CPUUsage,
		"memory_usage":  metrics.Candidate.ResourceUsage.MemoryUsage,
		"metadata":      metrics.Candidate.SampleMetrics,
	}

	if err := mc.store.CacheMetric(ctx, metrics.ExperimentID, candidateData); err != nil {
		return fmt.Errorf("failed to store candidate metrics: %w", err)
	}

	return nil
}

// calculateAndStoreKPIs calculates KPIs for an experiment and stores them
func (mc *MetricsCollector) calculateAndStoreKPIs(ctx context.Context, experimentID string) {
	kpis, err := mc.kpiCalc.CalculateExperimentKPIs(ctx, experimentID, 5*time.Minute)
	if err != nil {
		log.Error().
			Err(err).
			Str("experiment_id", experimentID).
			Msg("Failed to calculate KPIs")
		return
	}

	// Store KPIs in the database (you might want to add a KPI table)
	log.Info().
		Str("experiment_id", experimentID).
		Float64("cardinality_reduction", kpis.CardinalityReduction).
		Float64("cost_reduction", kpis.CostReduction).
		Float64("cpu_reduction", kpis.CPUUsage.Reduction).
		Float64("memory_reduction", kpis.MemoryUsage.Reduction).
		Float64("data_accuracy", kpis.DataAccuracy).
		Msg("Calculated KPIs")
}

// GetLatestMetrics returns the latest metrics for an experiment
func (mc *MetricsCollector) GetLatestMetrics(ctx context.Context, experimentID string, limit int) ([]*internalModels.Metric, error) {
	// This would need to be implemented based on your actual metric storage
	return nil, fmt.Errorf("not implemented")
}

// GetMetricsInRange returns metrics within a time range
func (mc *MetricsCollector) GetMetricsInRange(ctx context.Context, experimentID string, start, end time.Time) ([]*internalModels.Metric, error) {
	// This would need to be implemented based on your actual metric storage
	return nil, fmt.Errorf("not implemented")
}

// queryScalar executes a Prometheus query and returns a scalar value
func (mc *MetricsCollector) queryScalar(ctx context.Context, query string, timestamp time.Time) (float64, error) {
	result, warnings, err := mc.promClient.Query(ctx, query, timestamp)
	if err != nil {
		return 0, err
	}

	if len(warnings) > 0 {
		log.Warn().Strs("warnings", warnings).Str("query", query).Msg("Prometheus query warnings")
	}

	switch v := result.(type) {
	case model.Vector:
		if len(v) > 0 {
			return float64(v[0].Value), nil
		}
	case *model.Scalar:
		return float64(v.Value), nil
	}

	return 0, nil
}
