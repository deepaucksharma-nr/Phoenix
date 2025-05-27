package services

import (
	"context"
	"fmt"
	"time"

	"github.com/phoenix/platform/projects/phoenix-api/internal/metrics"
	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/store"
	"github.com/rs/zerolog/log"
)

// AnalysisService handles experiment analysis and KPI calculation
type AnalysisService struct {
	store     store.Store
	collector *metrics.Collector
	costModel *CostModel
}

// NewAnalysisService creates a new analysis service
func NewAnalysisService(store store.Store, promURL string) (*AnalysisService, error) {
	collector, err := metrics.NewCollector(promURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics collector: %w", err)
	}

	return &AnalysisService{
		store:     store,
		collector: collector,
		costModel: NewCostModel(),
	}, nil
}

// AnalyzeExperiment performs full analysis of an experiment
func (s *AnalysisService) AnalyzeExperiment(ctx context.Context, experimentID string) (*models.KPIResult, error) {
	// Get experiment details
	exp, err := s.store.GetExperiment(ctx, experimentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get experiment: %w", err)
	}

	// Determine time range
	timeRange := 30 * time.Minute // Default
	if exp.Config.Duration > 0 {
		timeRange = exp.Config.Duration
	}

	// Collect metrics
	expMetrics, err := s.collector.CollectExperimentMetrics(ctx, experimentID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to collect metrics: %w", err)
	}

	// Calculate KPIs
	result := &models.KPIResult{
		ExperimentID: experimentID,
		CalculatedAt: time.Now(),
		Errors:       []string{},
	}

	// Cardinality reduction
	if expMetrics.Baseline.Cardinality > 0 {
		result.CardinalityReduction = float64(expMetrics.Baseline.Cardinality-expMetrics.Candidate.Cardinality) /
			float64(expMetrics.Baseline.Cardinality) * 100
	}

	// CPU usage
	result.CPUUsage.Baseline = expMetrics.Baseline.CPUUsage
	result.CPUUsage.Candidate = expMetrics.Candidate.CPUUsage
	if expMetrics.Baseline.CPUUsage > 0 {
		result.CPUUsage.Reduction = (expMetrics.Baseline.CPUUsage - expMetrics.Candidate.CPUUsage) /
			expMetrics.Baseline.CPUUsage * 100
	}

	// Memory usage
	result.MemoryUsage.Baseline = expMetrics.Baseline.MemoryUsageMB
	result.MemoryUsage.Candidate = expMetrics.Candidate.MemoryUsageMB
	if expMetrics.Baseline.MemoryUsageMB > 0 {
		result.MemoryUsage.Reduction = (expMetrics.Baseline.MemoryUsageMB - expMetrics.Candidate.MemoryUsageMB) /
			expMetrics.Baseline.MemoryUsageMB * 100
	}

	// Ingest rate
	result.IngestRate.Baseline = expMetrics.Baseline.IngestRate
	result.IngestRate.Candidate = expMetrics.Candidate.IngestRate
	if expMetrics.Baseline.IngestRate > 0 {
		result.IngestRate.Reduction = (expMetrics.Baseline.IngestRate - expMetrics.Candidate.IngestRate) /
			expMetrics.Baseline.IngestRate * 100
	}

	// Calculate cost reduction
	result.CostReduction = s.costModel.CalculateCostReduction(
		expMetrics.Baseline.Cardinality,
		expMetrics.Candidate.Cardinality,
		result.CPUUsage.Reduction,
		result.MemoryUsage.Reduction,
	)

	// Check critical metrics if specified
	if len(exp.Config.CriticalProcesses) > 0 {
		criticalCheck, err := s.collector.CheckCriticalMetrics(
			ctx, experimentID, "candidate",
			exp.Config.CriticalProcesses, time.Now(),
		)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("critical metrics check failed: %v", err))
		} else {
			allPresent := true
			for process, present := range criticalCheck {
				if !present {
					allPresent = false
					result.Errors = append(result.Errors,
						fmt.Sprintf("critical process '%s' metrics missing in candidate", process))
				}
			}
			if allPresent {
				result.DataAccuracy = 100.0
			} else {
				result.DataAccuracy = 90.0 // Penalize for missing critical metrics
			}
		}
	} else {
		// Default accuracy based on error rate
		if expMetrics.Candidate.ErrorRate < 1.0 {
			result.DataAccuracy = 99.0
		} else {
			result.DataAccuracy = 100.0 - expMetrics.Candidate.ErrorRate
		}
	}

	// Store results
	if err := s.storeResults(ctx, exp, result); err != nil {
		log.Error().Err(err).Msg("Failed to store analysis results")
	}

	// Create experiment event
	event := &models.ExperimentEvent{
		ExperimentID: experimentID,
		EventType:    "analysis_completed",
		Phase:        "analyzing",
		Message: fmt.Sprintf("Analysis completed: %.1f%% cardinality reduction, %.1f%% cost savings",
			result.CardinalityReduction, result.CostReduction),
		Metadata: map[string]interface{}{
			"cardinality_reduction": result.CardinalityReduction,
			"cost_reduction":        result.CostReduction,
			"data_accuracy":         result.DataAccuracy,
		},
	}

	if err := s.store.CreateExperimentEvent(ctx, event); err != nil {
		log.Error().Err(err).Msg("Failed to create analysis event")
	}

	return result, nil
}

// storeResults stores the analysis results
func (s *AnalysisService) storeResults(ctx context.Context, exp *models.Experiment, result *models.KPIResult) error {
	// Update experiment with KPI results
	exp.Status.KPIs = map[string]float64{
		"cardinality_reduction": result.CardinalityReduction,
		"cost_reduction":        result.CostReduction,
		"cpu_reduction":         result.CPUUsage.Reduction,
		"memory_reduction":      result.MemoryUsage.Reduction,
		"data_accuracy":         result.DataAccuracy,
	}

	return s.store.UpdateExperiment(ctx, exp)
}

// GetRecommendation provides a recommendation based on analysis results
func (s *AnalysisService) GetRecommendation(result *models.KPIResult) string {
	// Success criteria thresholds
	const (
		minCardinalityReduction = 20.0
		minCostSavings          = 15.0
		minDataAccuracy         = 98.0
		maxCPUIncrease          = 10.0
	)

	if result.DataAccuracy < minDataAccuracy {
		return "DO NOT PROMOTE: Data accuracy below threshold. Critical metrics may be missing."
	}

	if result.CPUUsage.Reduction < -maxCPUIncrease {
		return "CAUTION: Candidate pipeline uses significantly more CPU. Review configuration."
	}

	if result.CardinalityReduction < minCardinalityReduction {
		return "LIMITED BENEFIT: Cardinality reduction below target. Consider more aggressive filtering."
	}

	if result.CostReduction < minCostSavings {
		return "LIMITED BENEFIT: Cost savings below target. May not justify deployment effort."
	}

	if result.CardinalityReduction > 50 && result.DataAccuracy >= 99 {
		return "STRONGLY RECOMMEND: Excellent cardinality reduction with high data accuracy."
	}

	return "RECOMMEND: Candidate pipeline meets success criteria. Ready for promotion."
}

// CostModel calculates cost based on metrics
type CostModel struct {
	// Cost factors (example values - should be configurable)
	CostPerSeries   float64
	CostPerCPUCore  float64
	CostPerGBMemory float64
}

// NewCostModel creates a new cost model with default values
func NewCostModel() *CostModel {
	return &CostModel{
		CostPerSeries:   0.01, // $0.01 per series per month
		CostPerCPUCore:  50.0, // $50 per CPU core per month
		CostPerGBMemory: 5.0,  // $5 per GB memory per month
	}
}

// CalculateCostReduction calculates overall cost reduction
func (m *CostModel) CalculateCostReduction(baselineCardinality, candidateCardinality int64,
	cpuReduction, memoryReduction float64) float64 {

	// Series cost reduction
	seriesCostReduction := float64(baselineCardinality-candidateCardinality) * m.CostPerSeries

	// Resource cost impact (negative reduction means increased cost)
	cpuCostImpact := cpuReduction / 100 * m.CostPerCPUCore
	memoryCostImpact := memoryReduction / 100 * m.CostPerGBMemory

	// Total monthly savings
	totalBaseCost := float64(baselineCardinality)*m.CostPerSeries + m.CostPerCPUCore + m.CostPerGBMemory
	totalSavings := seriesCostReduction + cpuCostImpact + memoryCostImpact

	if totalBaseCost > 0 {
		return (totalSavings / totalBaseCost) * 100
	}

	return 0
}
