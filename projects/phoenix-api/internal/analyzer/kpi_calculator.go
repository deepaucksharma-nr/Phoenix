package analyzer

import (
	"context"
	"fmt"
	"time"

	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/rs/zerolog/log"
)

type KPICalculator struct {
	promClient v1.API
}

func NewKPICalculator(promURL string) (*KPICalculator, error) {
	client, err := api.NewClient(api.Config{
		Address: promURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	return &KPICalculator{
		promClient: v1.NewAPI(client),
	}, nil
}

// CalculateExperimentKPIs calculates all KPIs for an experiment
func (k *KPICalculator) CalculateExperimentKPIs(ctx context.Context, expID string, duration time.Duration) (*models.KPIResult, error) {
	endTime := time.Now()
	startTime := endTime.Add(-duration)

	result := &models.KPIResult{
		ExperimentID: expID,
		CalculatedAt: time.Now(),
		Errors:       []string{},
	}

	// Calculate cardinality reduction
	cardinalityReduction, err := k.calculateCardinalityReduction(ctx, expID, endTime)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("cardinality calculation failed: %v", err))
	} else {
		result.CardinalityReduction = cardinalityReduction
	}

	// Calculate CPU usage
	cpuBaseline, cpuCandidate, err := k.calculateResourceUsage(ctx, expID, "cpu", startTime, endTime)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("CPU usage calculation failed: %v", err))
	} else {
		result.CPUUsage.Baseline = cpuBaseline
		result.CPUUsage.Candidate = cpuCandidate
		if cpuBaseline > 0 {
			result.CPUUsage.Reduction = ((cpuBaseline - cpuCandidate) / cpuBaseline) * 100
		}
	}

	// Calculate memory usage
	memBaseline, memCandidate, err := k.calculateResourceUsage(ctx, expID, "memory", startTime, endTime)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("memory usage calculation failed: %v", err))
	} else {
		result.MemoryUsage.Baseline = memBaseline
		result.MemoryUsage.Candidate = memCandidate
		if memBaseline > 0 {
			result.MemoryUsage.Reduction = ((memBaseline - memCandidate) / memBaseline) * 100
		}
	}

	// Calculate ingest rate
	ingestBaseline, ingestCandidate, err := k.calculateIngestRate(ctx, expID, startTime, endTime)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("ingest rate calculation failed: %v", err))
	} else {
		result.IngestRate.Baseline = ingestBaseline
		result.IngestRate.Candidate = ingestCandidate
		if ingestBaseline > 0 {
			result.IngestRate.Reduction = ((ingestBaseline - ingestCandidate) / ingestBaseline) * 100
		}
	}

	// Calculate cost reduction based on actual metrics ingestion rates
	if result.IngestRate.Baseline > 0 && result.IngestRate.Candidate > 0 {
		// Calculate monthly costs for baseline and candidate
		baselineMonthlyCost := k.CalculateEstimatedCost(ctx, result.IngestRate.Baseline)
		candidateMonthlyCost := k.CalculateEstimatedCost(ctx, result.IngestRate.Candidate)

		// Calculate cost reduction percentage
		result.CostReduction = k.CalculateROI(baselineMonthlyCost, candidateMonthlyCost)

		log.Info().
			Str("experiment_id", expID).
			Float64("baseline_cost", baselineMonthlyCost).
			Float64("candidate_cost", candidateMonthlyCost).
			Float64("cost_reduction", result.CostReduction).
			Msg("Calculated cost metrics")
	} else {
		// Fallback to weighted model if ingestion rates unavailable
		result.CostReduction = (result.CardinalityReduction * 0.7) +
			(result.CPUUsage.Reduction * 0.2) +
			(result.MemoryUsage.Reduction * 0.1)
	}

	// Calculate data accuracy (simplified - check if key metrics are still present)
	accuracy, err := k.calculateDataAccuracy(ctx, expID, endTime)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("accuracy calculation failed: %v", err))
	} else {
		result.DataAccuracy = accuracy
	}

	return result, nil
}

func (k *KPICalculator) calculateCardinalityReduction(ctx context.Context, expID string, timestamp time.Time) (float64, error) {
	// Use Phoenix-specific cardinality metrics
	baselineQuery := fmt.Sprintf(`
		phoenix_observer_kpi_store_phoenix_pipeline_output_cardinality_estimate{
			experiment_id="%s",
			pipeline="metrics/full_fidelity"
		}
	`, expID)

	baselineCardinality, err := k.queryScalar(ctx, baselineQuery, timestamp)
	if err != nil {
		// Fallback to counting unique series
		baselineQuery = fmt.Sprintf(`
			count(count by (__name__)({experiment_id="%s",variant="baseline"}))
		`, expID)
		baselineCardinality, err = k.queryScalar(ctx, baselineQuery, timestamp)
		if err != nil {
			return 0, fmt.Errorf("baseline cardinality query failed: %w", err)
		}
	}

	// Count unique series for candidate
	candidateQuery := fmt.Sprintf(`
		phoenix_observer_kpi_store_phoenix_pipeline_output_cardinality_estimate{
			experiment_id="%s",
			pipeline="metrics/optimised"
		}
	`, expID)

	candidateCardinality, err := k.queryScalar(ctx, candidateQuery, timestamp)
	if err != nil {
		// Fallback to counting unique series
		candidateQuery = fmt.Sprintf(`
			count(count by (__name__)({experiment_id="%s",variant="candidate"}))
		`, expID)
		candidateCardinality, err = k.queryScalar(ctx, candidateQuery, timestamp)
		if err != nil {
			return 0, fmt.Errorf("candidate cardinality query failed: %w", err)
		}
	}

	if baselineCardinality == 0 {
		return 0, nil
	}

	reduction := ((baselineCardinality - candidateCardinality) / baselineCardinality) * 100

	log.Info().
		Str("experiment_id", expID).
		Float64("baseline", baselineCardinality).
		Float64("candidate", candidateCardinality).
		Float64("reduction", reduction).
		Msg("Calculated cardinality reduction")

	return reduction, nil
}

func (k *KPICalculator) calculateResourceUsage(ctx context.Context, expID string, resource string, start, end time.Time) (baseline, candidate float64, err error) {
	var query string

	switch resource {
	case "cpu":
		// Use Phoenix-specific CPU metrics
		query = `avg(rate(process_cpu_seconds_total{
			job=~"phoenix-collector.*",
			experiment_id="%s",
			variant="%s"
		}[5m]))`
	case "memory":
		// Use Phoenix-specific memory metrics
		query = `avg(process_resident_memory_bytes{
			job=~"phoenix-collector.*",
			experiment_id="%s",
			variant="%s"
		})`
	default:
		return 0, 0, fmt.Errorf("unknown resource type: %s", resource)
	}

	// Query baseline
	baselineQuery := fmt.Sprintf(query, expID, "baseline")
	baseline, err = k.queryRangeAvg(ctx, baselineQuery, start, end)
	if err != nil {
		// Fallback to agent metrics
		if resource == "cpu" {
			baselineQuery = fmt.Sprintf(`avg(agent.cpu.percent{experiment_id="%s",variant="baseline"})`, expID)
		} else {
			baselineQuery = fmt.Sprintf(`avg(agent.memory.used_bytes{experiment_id="%s",variant="baseline"})`, expID)
		}
		baseline, err = k.queryRangeAvg(ctx, baselineQuery, start, end)
		if err != nil {
			return 0, 0, fmt.Errorf("baseline %s query failed: %w", resource, err)
		}
	}

	// Query candidate
	candidateQuery := fmt.Sprintf(query, expID, "candidate")
	candidate, err = k.queryRangeAvg(ctx, candidateQuery, start, end)
	if err != nil {
		// Fallback to agent metrics
		if resource == "cpu" {
			candidateQuery = fmt.Sprintf(`avg(agent.cpu.percent{experiment_id="%s",variant="candidate"})`, expID)
		} else {
			candidateQuery = fmt.Sprintf(`avg(agent.memory.used_bytes{experiment_id="%s",variant="candidate"})`, expID)
		}
		candidate, err = k.queryRangeAvg(ctx, candidateQuery, start, end)
		if err != nil {
			return 0, 0, fmt.Errorf("candidate %s query failed: %w", resource, err)
		}
	}

	return baseline, candidate, nil
}

func (k *KPICalculator) calculateIngestRate(ctx context.Context, expID string, start, end time.Time) (baseline, candidate float64, err error) {
	// Use Phoenix-specific OTEL collector metrics
	query := `sum(rate(otelcol_processor_accepted_metric_points{
		experiment_id="%s",
		variant="%s"
	}[5m]))`

	// Query baseline
	baselineQuery := fmt.Sprintf(query, expID, "baseline")
	baseline, err = k.queryRangeAvg(ctx, baselineQuery, start, end)
	if err != nil {
		// Fallback to receiver metrics
		baselineQuery = fmt.Sprintf(`sum(rate(otelcol_receiver_accepted_metric_points{experiment_id="%s",variant="baseline"}[5m]))`, expID)
		baseline, err = k.queryRangeAvg(ctx, baselineQuery, start, end)
		if err != nil {
			// Final fallback to generic metrics
			baselineQuery = fmt.Sprintf(`sum(rate(up{experiment_id="%s",variant="baseline"}[5m])) * 1000`, expID)
			baseline, err = k.queryRangeAvg(ctx, baselineQuery, start, end)
			if err != nil {
				return 0, 0, fmt.Errorf("baseline ingest query failed: %w", err)
			}
		}
	}

	// Query candidate
	candidateQuery := fmt.Sprintf(query, expID, "candidate")
	candidate, err = k.queryRangeAvg(ctx, candidateQuery, start, end)
	if err != nil {
		// Fallback to receiver metrics
		candidateQuery = fmt.Sprintf(`sum(rate(otelcol_receiver_accepted_metric_points{experiment_id="%s",variant="candidate"}[5m]))`, expID)
		candidate, err = k.queryRangeAvg(ctx, candidateQuery, start, end)
		if err != nil {
			// Final fallback to generic metrics
			candidateQuery = fmt.Sprintf(`sum(rate(up{experiment_id="%s",variant="candidate"}[5m])) * 1000`, expID)
			candidate, err = k.queryRangeAvg(ctx, candidateQuery, start, end)
			if err != nil {
				return 0, 0, fmt.Errorf("candidate ingest query failed: %w", err)
			}
		}
	}

	return baseline, candidate, nil
}

func (k *KPICalculator) calculateDataAccuracy(ctx context.Context, expID string, timestamp time.Time) (float64, error) {
	// Check if key metrics are present in both baseline and candidate
	// These are critical business metrics that should be preserved
	keyMetrics := []string{
		// HTTP metrics
		"http_server_duration_seconds",
		"http_server_request_count_total",
		"http_server_active_requests",
		// System metrics
		"process_cpu_seconds_total",
		"process_resident_memory_bytes",
		"go_goroutines",
		// Business metrics
		"phoenix_api_request_duration_seconds",
		"phoenix_experiment_active",
		"phoenix_pipeline_deployments_total",
		// Agent metrics
		"agent_cpu_percent",
		"agent_memory_used_bytes",
		"agent_uptime_seconds",
	}

	baselinePresent := 0
	candidatePresent := 0

	for _, metric := range keyMetrics {
		// Check baseline
		baselineQuery := fmt.Sprintf(`count(%s{experiment_id="%s",variant="baseline"})`, metric, expID)
		if val, err := k.queryScalar(ctx, baselineQuery, timestamp); err == nil && val > 0 {
			baselinePresent++
		}

		// Check candidate
		candidateQuery := fmt.Sprintf(`count(%s{experiment_id="%s",variant="candidate"})`, metric, expID)
		if val, err := k.queryScalar(ctx, candidateQuery, timestamp); err == nil && val > 0 {
			candidatePresent++
		}
	}

	if baselinePresent == 0 {
		return 100, nil // If no baseline metrics, assume 100% accuracy
	}

	accuracy := (float64(candidatePresent) / float64(baselinePresent)) * 100
	return accuracy, nil
}

func (k *KPICalculator) queryScalar(ctx context.Context, query string, timestamp time.Time) (float64, error) {
	result, warnings, err := k.promClient.Query(ctx, query, timestamp)
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

func (k *KPICalculator) queryRangeAvg(ctx context.Context, query string, start, end time.Time) (float64, error) {
	r := v1.Range{
		Start: start,
		End:   end,
		Step:  30 * time.Second,
	}

	result, warnings, err := k.promClient.QueryRange(ctx, query, r)
	if err != nil {
		return 0, err
	}

	if len(warnings) > 0 {
		log.Warn().Strs("warnings", warnings).Str("query", query).Msg("Prometheus query warnings")
	}

	switch v := result.(type) {
	case model.Matrix:
		if len(v) > 0 && len(v[0].Values) > 0 {
			sum := 0.0
			for _, sample := range v[0].Values {
				sum += float64(sample.Value)
			}
			return sum / float64(len(v[0].Values)), nil
		}
	}

	return 0, nil
}

// CalculateEstimatedCost calculates the estimated cost based on metrics volume
func (k *KPICalculator) CalculateEstimatedCost(ctx context.Context, metricsPerSecond float64) float64 {
	// Cost model based on industry standards:
	// - $0.10 per million datapoints ingested
	// - Additional storage costs: $0.05 per million datapoints stored
	// - Processing overhead: 20% of base cost

	const (
		costPerMillionDatapoints = 0.10           // USD
		storageMultiplier        = 0.05           // USD per million stored
		processingOverhead       = 0.20           // 20% overhead
		secondsPerMonth          = 30 * 24 * 3600 // ~2.6M seconds
	)

	// Calculate monthly datapoints
	monthlyDatapoints := metricsPerSecond * secondsPerMonth
	millionDatapoints := monthlyDatapoints / 1_000_000

	// Base ingestion cost
	baseCost := millionDatapoints * costPerMillionDatapoints

	// Storage cost (assuming 30-day retention)
	storageCost := millionDatapoints * storageMultiplier

	// Processing overhead
	overhead := (baseCost + storageCost) * processingOverhead

	// Total monthly cost
	totalCost := baseCost + storageCost + overhead

	return totalCost
}

// CalculateROI calculates the return on investment for the optimization
func (k *KPICalculator) CalculateROI(baselineCost, optimizedCost float64) float64 {
	if baselineCost == 0 {
		return 0
	}

	savings := baselineCost - optimizedCost
	roi := (savings / baselineCost) * 100

	return roi
}

// GetAdditionalMetrics fetches additional performance metrics
func (k *KPICalculator) GetAdditionalMetrics(ctx context.Context, expID string, duration time.Duration) map[string]float64 {
	metrics := make(map[string]float64)
	endTime := time.Now()
	startTime := endTime.Add(-duration)

	// P99 latency for pipeline processing
	p99Query := fmt.Sprintf(`histogram_quantile(0.99, 
		sum(rate(otelcol_processor_batch_batch_send_size_bucket{experiment_id="%s"}[5m])) by (le, variant)
	)`, expID)

	if val, err := k.queryRangeAvg(ctx, p99Query, startTime, endTime); err == nil {
		metrics["p99_latency_ms"] = val * 1000 // Convert to milliseconds
	}

	// Error rate
	errorQuery := fmt.Sprintf(`sum(rate(otelcol_processor_refused_metric_points{experiment_id="%s"}[5m])) / 
		(sum(rate(otelcol_receiver_accepted_metric_points{experiment_id="%s"}[5m])) + 0.1)`, expID, expID)

	if val, err := k.queryRangeAvg(ctx, errorQuery, startTime, endTime); err == nil {
		metrics["error_rate"] = val * 100 // Convert to percentage
	}

	// Pipeline efficiency
	efficiencyQuery := fmt.Sprintf(`
		(sum(rate(otelcol_processor_accepted_metric_points{experiment_id="%s"}[5m])) /
		 (sum(rate(otelcol_processor_accepted_metric_points{experiment_id="%s"}[5m])) + 
		  sum(rate(otelcol_processor_refused_metric_points{experiment_id="%s"}[5m])) + 0.1)
		) * 100`, expID, expID, expID)

	if val, err := k.queryRangeAvg(ctx, efficiencyQuery, startTime, endTime); err == nil {
		metrics["pipeline_efficiency"] = val
	}

	return metrics
}
