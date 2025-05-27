package services

import (
	"context"
	"fmt"
	"time"

	"github.com/phoenix/platform/projects/phoenix-api/internal/config"
	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/store"
)

// CostService calculates costs based on metrics and resource usage
type CostService struct {
	store store.Store

	// Cost configuration (per month)
	metricsIngestionCostPerMillion float64 // Cost per million metrics ingested
	storageRetentionCostPerGB      float64 // Cost per GB of metrics stored
	cpuCostPerCore                 float64 // Cost per CPU core
	memoryCostPerGB                float64 // Cost per GB of memory
}

// NewCostService creates a new cost calculation service
func NewCostService(store store.Store, costRates config.CostRates) *CostService {
	// Use industry-standard defaults if not configured
	ingestionCost := costRates.MetricsIngestionPerMillion
	if ingestionCost == 0 {
		ingestionCost = 0.10 // $0.10 per million datapoints
	}

	storageCost := costRates.StorageRetentionPerGB
	if storageCost == 0 {
		storageCost = 0.05 // $0.05 per GB stored
	}

	cpuCost := costRates.CPUCostPerCore
	if cpuCost == 0 {
		cpuCost = 50.0 // $50 per core per month
	}

	memoryCost := costRates.MemoryCostPerGB
	if memoryCost == 0 {
		memoryCost = 10.0 // $10 per GB per month
	}

	return &CostService{
		store:                          store,
		metricsIngestionCostPerMillion: ingestionCost,
		storageRetentionCostPerGB:      storageCost,
		cpuCostPerCore:                 cpuCost,
		memoryCostPerGB:                memoryCost,
	}
}

// CalculateExperimentCostSavings calculates the cost savings for an experiment
func (cs *CostService) CalculateExperimentCostSavings(ctx context.Context, experimentID string) (*CostAnalysis, error) {
	// Get experiment
	exp, err := cs.store.GetExperiment(ctx, experimentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get experiment: %w", err)
	}

	// Get metrics for baseline and candidate
	// TODO: Implement GetExperimentMetrics in store
	var metrics []*models.Metric
	// For now, we'll use estimates

	// Calculate costs
	analysis := &CostAnalysis{
		ExperimentID: experimentID,
		CalculatedAt: time.Now(),
		Duration:     exp.Config.Duration,
	}

	// If we have real metrics, use them
	if metrics != nil && len(metrics) > 0 {
		baselineMetrics := filterMetricsByVariant(metrics, "baseline")
		candidateMetrics := filterMetricsByVariant(metrics, "candidate")

		analysis.BaselineCost = cs.calculateVariantCost(baselineMetrics)
		analysis.CandidateCost = cs.calculateVariantCost(candidateMetrics)
	} else {
		// Use estimates based on typical scenarios
		analysis.BaselineCost = cs.estimateBaselineCost(exp)
		analysis.CandidateCost = cs.estimateCandidateCost(exp)
	}

	// Calculate savings
	analysis.MonthlySavings = analysis.BaselineCost.TotalMonthlyCost - analysis.CandidateCost.TotalMonthlyCost
	analysis.YearlySavings = analysis.MonthlySavings * 12
	analysis.SavingsPercentage = (analysis.MonthlySavings / analysis.BaselineCost.TotalMonthlyCost) * 100

	// Add recommendations
	analysis.Recommendations = cs.generateRecommendations(analysis)

	return analysis, nil
}

// calculateVariantCost calculates cost for a pipeline variant based on metrics
func (cs *CostService) calculateVariantCost(metrics []*models.Metric) VariantCost {
	var cost VariantCost

	// Calculate metrics ingestion rate (metrics per second)
	metricsPerSecond := float64(len(metrics)) / 300.0 // Assuming 5-minute window
	const secondsPerMonth = 30 * 24 * 3600            // ~2.6M seconds
	metricsPerMonth := metricsPerSecond * float64(secondsPerMonth)

	// Calculate cardinality
	uniqueMetrics := make(map[string]bool)
	for _, m := range metrics {
		key := m.Name
		for k, v := range m.Labels {
			key += fmt.Sprintf("_%s_%s", k, v)
		}
		uniqueMetrics[key] = true
	}
	cost.Cardinality = len(uniqueMetrics)

	// Calculate costs using same model as KPI calculator
	millionDatapoints := metricsPerMonth / 1_000_000

	// Base ingestion cost
	cost.MetricsIngestionCost = millionDatapoints * cs.metricsIngestionCostPerMillion

	// Storage cost (assuming 30-day retention, 8 bytes per datapoint)
	storageGB := (float64(cost.Cardinality) * 8 * 60 * 60 * 24 * 30) / (1024 * 1024 * 1024)
	cost.StorageRetentionCost = storageGB * cs.storageRetentionCostPerGB

	// Resource costs (get from metrics if available)
	cpuUsage := 0.5 // Default 0.5 cores
	memoryGB := 2.0 // Default 2GB

	for _, m := range metrics {
		if m.Name == "process_cpu_seconds_total" && m.Type == "counter" {
			// Convert CPU seconds to cores (assuming 5-minute rate)
			cpuUsage = m.Value / 300.0
		}
		if m.Name == "process_resident_memory_bytes" && m.Type == "gauge" {
			memoryGB = m.Value / (1024 * 1024 * 1024)
		}
		// Also check agent metrics
		if m.Name == "agent.cpu.percent" && m.Type == "gauge" {
			cpuUsage = m.Value / 100.0 // Convert percentage to cores
		}
		if m.Name == "agent.memory.used_bytes" && m.Type == "gauge" {
			memoryGB = m.Value / (1024 * 1024 * 1024)
		}
	}

	cost.ResourceCost = (cpuUsage * cs.cpuCostPerCore) + (memoryGB * cs.memoryCostPerGB)

	// Add 20% processing overhead (same as KPI calculator)
	processingOverhead := (cost.MetricsIngestionCost + cost.StorageRetentionCost) * 0.20

	cost.TotalMonthlyCost = cost.MetricsIngestionCost + cost.StorageRetentionCost + cost.ResourceCost + processingOverhead

	return cost
}

// estimateBaselineCost estimates baseline cost when no metrics are available
func (cs *CostService) estimateBaselineCost(exp *models.Experiment) VariantCost {
	// Typical baseline: high cardinality, no optimization
	// Assume 10k metrics per second for a medium-sized deployment
	const (
		baselineMetricsPerSecond = 10000.0
		baselineCardinality      = 100000
		secondsPerMonth          = 30 * 24 * 3600
	)

	monthlyDatapoints := baselineMetricsPerSecond * secondsPerMonth
	millionDatapoints := monthlyDatapoints / 1_000_000

	// Calculate costs
	ingestionCost := millionDatapoints * cs.metricsIngestionCostPerMillion

	// Storage: 8 bytes per datapoint, 30-day retention
	storageGB := (float64(baselineCardinality) * 8 * 60 * 60 * 24 * 30) / (1024 * 1024 * 1024)
	storageCost := storageGB * cs.storageRetentionCostPerGB

	// Resources: typical baseline usage
	resourceCost := (2.0 * cs.cpuCostPerCore) + (8.0 * cs.memoryCostPerGB) // 2 cores, 8GB RAM

	// Add 20% processing overhead
	processingOverhead := (ingestionCost + storageCost) * 0.20

	return VariantCost{
		Cardinality:          baselineCardinality,
		MetricsIngestionCost: ingestionCost,
		StorageRetentionCost: storageCost,
		ResourceCost:         resourceCost,
		TotalMonthlyCost:     ingestionCost + storageCost + resourceCost + processingOverhead,
	}
}

// estimateCandidateCost estimates candidate cost with optimization
func (cs *CostService) estimateCandidateCost(exp *models.Experiment) VariantCost {
	// Candidate with 70% cardinality reduction
	const (
		candidateMetricsPerSecond = 3000.0 // 70% reduction
		candidateCardinality      = 30000  // 70% reduction
		secondsPerMonth           = 30 * 24 * 3600
	)

	monthlyDatapoints := candidateMetricsPerSecond * secondsPerMonth
	millionDatapoints := monthlyDatapoints / 1_000_000

	// Calculate costs
	ingestionCost := millionDatapoints * cs.metricsIngestionCostPerMillion

	// Storage: 8 bytes per datapoint, 30-day retention
	storageGB := (float64(candidateCardinality) * 8 * 60 * 60 * 24 * 30) / (1024 * 1024 * 1024)
	storageCost := storageGB * cs.storageRetentionCostPerGB

	// Resources: reduced due to less processing
	resourceCost := (1.0 * cs.cpuCostPerCore) + (4.0 * cs.memoryCostPerGB) // 1 core, 4GB RAM

	// Add 20% processing overhead
	processingOverhead := (ingestionCost + storageCost) * 0.20

	return VariantCost{
		Cardinality:          candidateCardinality,
		MetricsIngestionCost: ingestionCost,
		StorageRetentionCost: storageCost,
		ResourceCost:         resourceCost,
		TotalMonthlyCost:     ingestionCost + storageCost + resourceCost + processingOverhead,
	}
}

// generateRecommendations generates cost optimization recommendations
func (cs *CostService) generateRecommendations(analysis *CostAnalysis) []string {
	recommendations := []string{}

	if analysis.SavingsPercentage > 50 {
		recommendations = append(recommendations,
			"Excellent cost reduction! Consider promoting this configuration to production.")
	}

	if analysis.CandidateCost.Cardinality > 50000 {
		recommendations = append(recommendations,
			"Cardinality is still high. Consider additional label aggregation rules.")
	}

	if analysis.CandidateCost.ResourceCost > analysis.BaselineCost.ResourceCost {
		recommendations = append(recommendations,
			"Resource usage increased. Review processor configurations for efficiency.")
	}

	if analysis.SavingsPercentage < 20 {
		recommendations = append(recommendations,
			"Low savings detected. Consider more aggressive filtering or sampling strategies.")
	}

	return recommendations
}

// filterMetricsByVariant filters metrics by variant type
func filterMetricsByVariant(metrics []*models.Metric, variant string) []*models.Metric {
	filtered := []*models.Metric{}
	for _, m := range metrics {
		if m.Labels != nil && m.Labels["variant"] == variant {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// CostAnalysis represents the cost analysis for an experiment
type CostAnalysis struct {
	ExperimentID      string        `json:"experiment_id"`
	CalculatedAt      time.Time     `json:"calculated_at"`
	Duration          time.Duration `json:"duration"`
	BaselineCost      VariantCost   `json:"baseline_cost"`
	CandidateCost     VariantCost   `json:"candidate_cost"`
	MonthlySavings    float64       `json:"monthly_savings"`
	YearlySavings     float64       `json:"yearly_savings"`
	SavingsPercentage float64       `json:"savings_percentage"`
	Recommendations   []string      `json:"recommendations"`
}

// VariantCost represents the cost breakdown for a pipeline variant
type VariantCost struct {
	Cardinality          int     `json:"cardinality"`
	MetricsIngestionCost float64 `json:"metrics_ingestion_cost"`
	StorageRetentionCost float64 `json:"storage_retention_cost"`
	ResourceCost         float64 `json:"resource_cost"`
	TotalMonthlyCost     float64 `json:"total_monthly_cost"`
}

// GetRealTimeCostFlow returns real-time cost flow data
func (cs *CostService) GetRealTimeCostFlow(ctx context.Context) (*store.MetricCostFlow, error) {
	// Get the data from store as map and convert it back to struct
	data, err := cs.store.GetMetricCostFlow(ctx)
	if err != nil {
		return nil, err
	}

	// Convert map back to struct for backwards compatibility
	flow := &store.MetricCostFlow{}

	if totalCost, ok := data["total_cost_per_minute"].(float64); ok {
		flow.TotalCostPerMinute = totalCost
	}

	if topMetrics, ok := data["top_metrics"].([]store.MetricCostDetail); ok {
		flow.TopMetrics = topMetrics
	}

	if byService, ok := data["by_service"].(map[string]float64); ok {
		flow.ByService = byService
	}

	if byNamespace, ok := data["by_namespace"].(map[string]float64); ok {
		flow.ByNamespace = byNamespace
	}

	if lastUpdated, ok := data["last_updated"].(time.Time); ok {
		flow.LastUpdated = lastUpdated
	}

	return flow, nil
}

// GetCardinalityTrends returns cardinality trends over time
func (cs *CostService) GetCardinalityTrends(ctx context.Context, duration time.Duration) (*CardinalityTrends, error) {
	// Get metrics from the last duration
	endTime := time.Now()
	startTime := endTime.Add(-duration)

	// This would normally query time-series data
	// For now, return mock trend data
	trends := &CardinalityTrends{
		StartTime: startTime,
		EndTime:   endTime,
		Interval:  5 * time.Minute,
		Data:      []CardinalityPoint{},
	}

	// Generate trend points
	for t := startTime; t.Before(endTime); t = t.Add(5 * time.Minute) {
		point := CardinalityPoint{
			Timestamp:        t,
			TotalCardinality: 50000 + int(t.Unix()%10000), // Simulated variation
			TopNamespaces: map[string]int{
				"production":  25000,
				"staging":     15000,
				"development": 10000,
			},
		}
		trends.Data = append(trends.Data, point)
	}

	return trends, nil
}

// CalculateRealTimeCost calculates the current cost based on real-time metrics rate
func (cs *CostService) CalculateRealTimeCost(metricsPerSecond float64) map[string]float64 {
	const secondsPerMonth = 30 * 24 * 3600

	// Calculate monthly projections
	monthlyDatapoints := metricsPerSecond * float64(secondsPerMonth)
	millionDatapoints := monthlyDatapoints / 1_000_000

	// Calculate cost components
	ingestionCost := millionDatapoints * cs.metricsIngestionCostPerMillion

	// Estimate storage based on typical cardinality ratio (10:1)
	estimatedCardinality := metricsPerSecond * 10
	storageGB := (estimatedCardinality * 8 * 60 * 60 * 24 * 30) / (1024 * 1024 * 1024)
	storageCost := storageGB * cs.storageRetentionCostPerGB

	// Processing overhead
	processingOverhead := (ingestionCost + storageCost) * 0.20

	// Total costs
	totalMonthlyCost := ingestionCost + storageCost + processingOverhead

	return map[string]float64{
		"metrics_per_second":    metricsPerSecond,
		"monthly_datapoints":    monthlyDatapoints,
		"ingestion_cost":        ingestionCost,
		"storage_cost":          storageCost,
		"processing_overhead":   processingOverhead,
		"total_monthly_cost":    totalMonthlyCost,
		"total_yearly_cost":     totalMonthlyCost * 12,
		"cost_per_minute":       totalMonthlyCost / (30 * 24 * 60),
		"cost_per_hour":         totalMonthlyCost / (30 * 24),
		"estimated_cardinality": estimatedCardinality,
	}
}

// CardinalityTrends represents cardinality trends over time
type CardinalityTrends struct {
	StartTime time.Time          `json:"start_time"`
	EndTime   time.Time          `json:"end_time"`
	Interval  time.Duration      `json:"interval"`
	Data      []CardinalityPoint `json:"data"`
}

// CardinalityPoint represents a single point in cardinality trends
type CardinalityPoint struct {
	Timestamp        time.Time      `json:"timestamp"`
	TotalCardinality int            `json:"total_cardinality"`
	TopNamespaces    map[string]int `json:"top_namespaces"`
}
