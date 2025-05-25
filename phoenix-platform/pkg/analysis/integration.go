package analysis

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/phoenix/platform/pkg/models"
)

// ExperimentAnalyzer integrates statistical analysis with the experiment system
type ExperimentAnalyzer struct {
	analyzer StatisticalAnalyzer
}

// NewExperimentAnalyzer creates a new experiment analyzer
func NewExperimentAnalyzer() *ExperimentAnalyzer {
	return &ExperimentAnalyzer{
		analyzer: NewStatisticalAnalyzer(),
	}
}

// AnalyzeExperimentResults analyzes the results of an experiment
func (ea *ExperimentAnalyzer) AnalyzeExperimentResults(ctx context.Context, experiment *models.Experiment, metrics map[string]*MetricData) (*ExperimentAnalysis, error) {
	if experiment == nil {
		return nil, fmt.Errorf("experiment cannot be nil")
	}

	if len(metrics) == 0 {
		return nil, fmt.Errorf("no metrics provided for analysis")
	}

	analysis := &ExperimentAnalysis{
		ExperimentID: experiment.ID,
		Name:         experiment.Name,
		StartTime:    experiment.CreatedAt,
		AnalysisTime: time.Now(),
		Metrics:      make(map[string]*MetricAnalysis),
	}

	// Analyze each metric
	for metricName, data := range metrics {
		metricAnalysis, err := ea.analyzeMetric(metricName, data)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze metric %s: %w", metricName, err)
		}
		analysis.Metrics[metricName] = metricAnalysis
	}

	// Generate overall recommendation
	analysis.Recommendation = ea.generateRecommendation(analysis.Metrics)
	analysis.Confidence = ea.calculateOverallConfidence(analysis.Metrics)
	analysis.SufficientData = ea.hasSufficientData(analysis.Metrics)

	return analysis, nil
}

// analyzeMetric performs statistical analysis on a single metric
func (ea *ExperimentAnalyzer) analyzeMetric(name string, data *MetricData) (*MetricAnalysis, error) {
	if data == nil || len(data.Baseline) == 0 || len(data.Candidate) == 0 {
		return nil, fmt.Errorf("insufficient data for metric %s", name)
	}

	// Perform t-test
	result := ea.analyzer.TTest(data.Baseline, data.Candidate)

	// Calculate percentiles
	baselineP50 := Percentile(data.Baseline, 50)
	baselineP95 := Percentile(data.Baseline, 95)
	baselineP99 := Percentile(data.Baseline, 99)

	candidateP50 := Percentile(data.Candidate, 50)
	candidateP95 := Percentile(data.Candidate, 95)
	candidateP99 := Percentile(data.Candidate, 99)

	return &MetricAnalysis{
		Name:        name,
		Type:        data.Type,
		TestResult:  result,
		BaselineStats: MetricStats{
			Mean:        result.BaselineCI.Mean,
			StdDev:      calculateStdDev(data.Baseline),
			Min:         min(data.Baseline),
			Max:         max(data.Baseline),
			Count:       len(data.Baseline),
			Percentiles: map[string]float64{
				"p50": baselineP50,
				"p95": baselineP95,
				"p99": baselineP99,
			},
		},
		CandidateStats: MetricStats{
			Mean:        result.CandidateCI.Mean,
			StdDev:      calculateStdDev(data.Candidate),
			Min:         min(data.Candidate),
			Max:         max(data.Candidate),
			Count:       len(data.Candidate),
			Percentiles: map[string]float64{
				"p50": candidateP50,
				"p95": candidateP95,
				"p99": candidateP99,
			},
		},
		Improvement: calculateImprovement(data.Type, result),
	}, nil
}

// generateRecommendation creates an overall recommendation based on all metrics
func (ea *ExperimentAnalyzer) generateRecommendation(metrics map[string]*MetricAnalysis) Recommendation {
	criticalRegressions := 0
	significantImprovements := 0
	insufficientData := false

	for _, metric := range metrics {
		if metric.TestResult.BaselineSampleSize < 30 || metric.TestResult.CandidateSampleSize < 30 {
			insufficientData = true
			continue
		}

		if metric.TestResult.Significant {
			if metric.Improvement < -5 { // More than 5% regression
				criticalRegressions++
			} else if metric.Improvement > 5 { // More than 5% improvement
				significantImprovements++
			}
		}
	}

	if insufficientData {
		return RecommendationContinue
	}

	if criticalRegressions > 0 {
		return RecommendationReject
	}

	if significantImprovements > 0 {
		return RecommendationPromote
	}

	return RecommendationNeutral
}

// calculateOverallConfidence calculates the overall confidence in the recommendation
func (ea *ExperimentAnalyzer) calculateOverallConfidence(metrics map[string]*MetricAnalysis) float64 {
	if len(metrics) == 0 {
		return 0
	}

	totalConfidence := 0.0
	validMetrics := 0

	for _, metric := range metrics {
		if metric.TestResult.BaselineSampleSize >= 30 && metric.TestResult.CandidateSampleSize >= 30 {
			// Weight confidence by p-value strength
			if metric.TestResult.PValue < 0.001 {
				totalConfidence += 0.95
			} else if metric.TestResult.PValue < 0.01 {
				totalConfidence += 0.85
			} else if metric.TestResult.PValue < 0.05 {
				totalConfidence += 0.75
			} else {
				totalConfidence += 0.5
			}
			validMetrics++
		}
	}

	if validMetrics == 0 {
		return 0.3 // Low confidence due to insufficient data
	}

	return totalConfidence / float64(validMetrics)
}

// hasSufficientData checks if all metrics have sufficient data
func (ea *ExperimentAnalyzer) hasSufficientData(metrics map[string]*MetricAnalysis) bool {
	if len(metrics) == 0 {
		return false
	}

	for _, metric := range metrics {
		if metric.TestResult.BaselineSampleSize < 30 || metric.TestResult.CandidateSampleSize < 30 {
			return false
		}
	}

	return true
}

// calculateImprovement calculates the improvement percentage based on metric type
func calculateImprovement(metricType MetricType, result TestResult) float64 {
	switch metricType {
	case MetricTypeLatency, MetricTypeErrorRate, MetricTypeCost:
		// Lower is better
		return -result.RelativeImprovement
	case MetricTypeThroughput:
		// Higher is better
		return result.RelativeImprovement
	default:
		return result.RelativeImprovement
	}
}

// calculateStdDev calculates standard deviation
func calculateStdDev(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	mean := calculateMean(data)
	variance := calculateVariance(data, mean)
	return math.Sqrt(variance)
}

// min finds the minimum value in a slice
func min(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	minVal := data[0]
	for _, v := range data[1:] {
		if v < minVal {
			minVal = v
		}
	}
	return minVal
}

// max finds the maximum value in a slice
func max(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	maxVal := data[0]
	for _, v := range data[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}

// Types for integration

// MetricData contains the raw data for a metric
type MetricData struct {
	Type      MetricType
	Baseline  []float64
	Candidate []float64
}

// ExperimentAnalysis contains the complete analysis of an experiment
type ExperimentAnalysis struct {
	ExperimentID   string
	Name           string
	StartTime      time.Time
	AnalysisTime   time.Time
	Metrics        map[string]*MetricAnalysis
	Recommendation Recommendation
	Confidence     float64
	SufficientData bool
}

// MetricAnalysis contains the analysis results for a single metric
type MetricAnalysis struct {
	Name           string
	Type           MetricType
	TestResult     TestResult
	BaselineStats  MetricStats
	CandidateStats MetricStats
	Improvement    float64 // Percentage improvement (positive means better)
}

// MetricStats contains statistical summary of metric data
type MetricStats struct {
	Mean        float64
	StdDev      float64
	Min         float64
	Max         float64
	Count       int
	Percentiles map[string]float64
}

// Recommendation represents the analysis recommendation
type Recommendation string

const (
	RecommendationPromote  Recommendation = "PROMOTE"
	RecommendationReject   Recommendation = "REJECT"
	RecommendationContinue Recommendation = "CONTINUE"
	RecommendationNeutral  Recommendation = "NEUTRAL"
)

// ShouldAutoPromote determines if the experiment should be automatically promoted
func (ea *ExperimentAnalysis) ShouldAutoPromote() bool {
	return ea.Recommendation == RecommendationPromote && 
	       ea.Confidence > 0.9 && 
	       ea.SufficientData
}

// GetRiskLevel assesses the risk level of promoting the candidate
func (ea *ExperimentAnalysis) GetRiskLevel() string {
	maxRegression := 0.0
	
	for _, metric := range ea.Metrics {
		if metric.Improvement < maxRegression {
			maxRegression = metric.Improvement
		}
	}
	
	if maxRegression < -10 {
		return "HIGH"
	} else if maxRegression < -5 {
		return "MEDIUM"
	}
	
	return "LOW"
}

// GenerateReport creates a human-readable report of the analysis
func (ea *ExperimentAnalysis) GenerateReport() string {
	report := fmt.Sprintf("Experiment Analysis Report\n")
	report += fmt.Sprintf("========================\n")
	report += fmt.Sprintf("Experiment: %s (%s)\n", ea.Name, ea.ExperimentID)
	report += fmt.Sprintf("Analysis Time: %s\n", ea.AnalysisTime.Format(time.RFC3339))
	report += fmt.Sprintf("Duration: %s\n\n", ea.AnalysisTime.Sub(ea.StartTime).Round(time.Hour))
	
	report += fmt.Sprintf("Overall Results\n")
	report += fmt.Sprintf("--------------\n")
	report += fmt.Sprintf("Recommendation: %s\n", ea.Recommendation)
	report += fmt.Sprintf("Confidence: %.1f%%\n", ea.Confidence*100)
	report += fmt.Sprintf("Risk Level: %s\n", ea.GetRiskLevel())
	report += fmt.Sprintf("Sufficient Data: %v\n\n", ea.SufficientData)
	
	report += fmt.Sprintf("Metric Analysis\n")
	report += fmt.Sprintf("--------------\n")
	
	for name, metric := range ea.Metrics {
		report += fmt.Sprintf("\n%s (%s):\n", name, metric.Type)
		report += fmt.Sprintf("  Baseline:  %.2f (n=%d)\n", metric.BaselineStats.Mean, metric.BaselineStats.Count)
		report += fmt.Sprintf("  Candidate: %.2f (n=%d)\n", metric.CandidateStats.Mean, metric.CandidateStats.Count)
		report += fmt.Sprintf("  Change: %.1f%%\n", metric.Improvement)
		report += fmt.Sprintf("  P-value: %.4f\n", metric.TestResult.PValue)
		report += fmt.Sprintf("  Significant: %v\n", metric.TestResult.Significant)
		
		if metric.TestResult.Significant {
			report += fmt.Sprintf("  Effect Size: %.2f\n", metric.TestResult.EffectSize)
		}
	}
	
	return report
}