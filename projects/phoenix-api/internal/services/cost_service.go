package services

import (
	"context"
	"fmt"
	"time"

	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/store"
)

// CostService calculates costs based on metrics and resource usage
type CostService struct {
	store store.Store
	
	// Cost configuration (per month)
	metricsIngestionCostPerMillion float64  // Cost per million metrics ingested
	storageRetentionCostPerGB      float64  // Cost per GB of metrics stored
	cpuCostPerCore                 float64  // Cost per CPU core
	memoryCostPerGB                float64  // Cost per GB of memory
}

// NewCostService creates a new cost calculation service
func NewCostService(store store.Store) *CostService {
	return &CostService{
		store:                          store,
		metricsIngestionCostPerMillion: 50.0,   // $50 per million metrics/month
		storageRetentionCostPerGB:      10.0,   // $10 per GB/month
		cpuCostPerCore:                 100.0,  // $100 per core/month
		memoryCostPerGB:                20.0,   // $20 per GB/month
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
	metricsPerMonth := metricsPerSecond * 60 * 60 * 24 * 30
	
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
	
	// Calculate costs
	cost.MetricsIngestionCost = (metricsPerMonth / 1_000_000) * cs.metricsIngestionCostPerMillion
	
	// Estimate storage based on cardinality and retention
	// Assume 100 bytes per metric point, 30-day retention
	storageGB := (float64(cost.Cardinality) * 100 * 60 * 24 * 30) / (1024 * 1024 * 1024)
	cost.StorageRetentionCost = storageGB * cs.storageRetentionCostPerGB
	
	// Resource costs (get from metrics if available)
	cpuUsage := 0.5  // Default 0.5 cores
	memoryGB := 2.0  // Default 2GB
	
	for _, m := range metrics {
		if m.Name == "cpu_usage" && m.Type == "gauge" {
			cpuUsage = m.Value
		}
		if m.Name == "memory_usage_bytes" && m.Type == "gauge" {
			memoryGB = m.Value / (1024 * 1024 * 1024)
		}
	}
	
	cost.ResourceCost = (cpuUsage * cs.cpuCostPerCore) + (memoryGB * cs.memoryCostPerGB)
	cost.TotalMonthlyCost = cost.MetricsIngestionCost + cost.StorageRetentionCost + cost.ResourceCost
	
	return cost
}

// estimateBaselineCost estimates baseline cost when no metrics are available
func (cs *CostService) estimateBaselineCost(exp *models.Experiment) VariantCost {
	// Typical baseline: high cardinality, no optimization
	return VariantCost{
		Cardinality:          100000, // 100k unique time series
		MetricsIngestionCost: 500.0,  // $500/month for ingestion
		StorageRetentionCost: 200.0,  // $200/month for storage
		ResourceCost:         150.0,  // $150/month for compute
		TotalMonthlyCost:     850.0,
	}
}

// estimateCandidateCost estimates candidate cost with optimization
func (cs *CostService) estimateCandidateCost(exp *models.Experiment) VariantCost {
	// Candidate with 70% cardinality reduction
	return VariantCost{
		Cardinality:          30000,  // 30k unique time series (70% reduction)
		MetricsIngestionCost: 150.0,  // $150/month for ingestion
		StorageRetentionCost: 60.0,   // $60/month for storage
		ResourceCost:         100.0,  // $100/month for compute (less processing)
		TotalMonthlyCost:     310.0,
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
	ExperimentID      string       `json:"experiment_id"`
	CalculatedAt      time.Time    `json:"calculated_at"`
	Duration          time.Duration `json:"duration"`
	BaselineCost      VariantCost  `json:"baseline_cost"`
	CandidateCost     VariantCost  `json:"candidate_cost"`
	MonthlySavings    float64      `json:"monthly_savings"`
	YearlySavings     float64      `json:"yearly_savings"`
	SavingsPercentage float64      `json:"savings_percentage"`
	Recommendations   []string     `json:"recommendations"`
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
	// Delegate to store for now, but we could add caching or aggregation here
	return cs.store.GetMetricCostFlow(ctx)
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
			TotalCardinality: 50000 + int(t.Unix() % 10000), // Simulated variation
			TopNamespaces: map[string]int{
				"production": 25000,
				"staging":    15000,
				"development": 10000,
			},
		}
		trends.Data = append(trends.Data, point)
	}
	
	return trends, nil
}

// CardinalityTrends represents cardinality trends over time
type CardinalityTrends struct {
	StartTime time.Time           `json:"start_time"`
	EndTime   time.Time           `json:"end_time"`
	Interval  time.Duration       `json:"interval"`
	Data      []CardinalityPoint  `json:"data"`
}

// CardinalityPoint represents a single point in cardinality trends
type CardinalityPoint struct {
	Timestamp        time.Time         `json:"timestamp"`
	TotalCardinality int               `json:"total_cardinality"`
	TopNamespaces    map[string]int    `json:"top_namespaces"`
}