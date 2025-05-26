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
	
	// Calculate cost reduction (based on cardinality and resource usage)
	if result.CardinalityReduction > 0 {
		// Simple cost model: 70% weight on cardinality, 20% on CPU, 10% on memory
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
	// Count unique series for baseline
	baselineQuery := fmt.Sprintf(`
		count(count by (__name__)({experiment_id="%s",variant="baseline"}))
	`, expID)
	
	baselineCardinality, err := k.queryScalar(ctx, baselineQuery, timestamp)
	if err != nil {
		return 0, fmt.Errorf("baseline cardinality query failed: %w", err)
	}
	
	// Count unique series for candidate
	candidateQuery := fmt.Sprintf(`
		count(count by (__name__)({experiment_id="%s",variant="candidate"}))
	`, expID)
	
	candidateCardinality, err := k.queryScalar(ctx, candidateQuery, timestamp)
	if err != nil {
		return 0, fmt.Errorf("candidate cardinality query failed: %w", err)
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
		query = `avg(rate(container_cpu_usage_seconds_total{
			container_name=~"otel-collector.*",
			experiment_id="%s",
			variant="%s"
		}[5m]))`
	case "memory":
		query = `avg(container_memory_usage_bytes{
			container_name=~"otel-collector.*",
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
		return 0, 0, fmt.Errorf("baseline %s query failed: %w", resource, err)
	}
	
	// Query candidate
	candidateQuery := fmt.Sprintf(query, expID, "candidate")
	candidate, err = k.queryRangeAvg(ctx, candidateQuery, start, end)
	if err != nil {
		return 0, 0, fmt.Errorf("candidate %s query failed: %w", resource, err)
	}
	
	return baseline, candidate, nil
}

func (k *KPICalculator) calculateIngestRate(ctx context.Context, expID string, start, end time.Time) (baseline, candidate float64, err error) {
	// Calculate metrics ingestion rate
	query := `sum(rate(prometheus_tsdb_ingestion_rate{
		job="pushgateway",
		experiment_id="%s",
		variant="%s"
	}[5m]))`
	
	// Query baseline
	baselineQuery := fmt.Sprintf(query, expID, "baseline")
	baseline, err = k.queryRangeAvg(ctx, baselineQuery, start, end)
	if err != nil {
		// Fallback to counting samples
		baselineQuery = fmt.Sprintf(`sum(rate(up{experiment_id="%s",variant="baseline"}[5m]))`, expID)
		baseline, err = k.queryRangeAvg(ctx, baselineQuery, start, end)
		if err != nil {
			return 0, 0, fmt.Errorf("baseline ingest query failed: %w", err)
		}
	}
	
	// Query candidate
	candidateQuery := fmt.Sprintf(query, expID, "candidate")
	candidate, err = k.queryRangeAvg(ctx, candidateQuery, start, end)
	if err != nil {
		// Fallback to counting samples
		candidateQuery = fmt.Sprintf(`sum(rate(up{experiment_id="%s",variant="candidate"}[5m]))`, expID)
		candidate, err = k.queryRangeAvg(ctx, candidateQuery, start, end)
		if err != nil {
			return 0, 0, fmt.Errorf("candidate ingest query failed: %w", err)
		}
	}
	
	return baseline, candidate, nil
}

func (k *KPICalculator) calculateDataAccuracy(ctx context.Context, expID string, timestamp time.Time) (float64, error) {
	// Check if key metrics are present in both baseline and candidate
	keyMetrics := []string{
		"http_server_duration_seconds",
		"http_server_request_count_total",
		"process_cpu_seconds_total",
		"process_resident_memory_bytes",
	}
	
	baselinePresent := 0
	candidatePresent := 0
	
	for _, metric := range keyMetrics {
		// Check baseline
		baselineQuery := fmt.Sprintf(`%s{experiment_id="%s",variant="baseline"}`, metric, expID)
		if val, err := k.queryScalar(ctx, baselineQuery, timestamp); err == nil && val > 0 {
			baselinePresent++
		}
		
		// Check candidate
		candidateQuery := fmt.Sprintf(`%s{experiment_id="%s",variant="candidate"}`, metric, expID)
		if val, err := k.queryScalar(ctx, candidateQuery, timestamp); err == nil && val > 0 {
			candidatePresent++
		}
	}
	
	if baselinePresent == 0 {
		return 0, nil
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