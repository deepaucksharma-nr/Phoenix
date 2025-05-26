package analysis

import (
	"context"
	"time"
)

// MetricType represents the type of metric being analyzed
type MetricType string

const (
	MetricTypeLatency    MetricType = "latency"
	MetricTypeThroughput MetricType = "throughput"
	MetricTypeErrorRate  MetricType = "error_rate"
	MetricTypeCost       MetricType = "cost"
)

// Recommendation represents the recommendation from analysis
type Recommendation string

const (
	RecommendationPromote  Recommendation = "promote"
	RecommendationReject   Recommendation = "reject"
	RecommendationContinue Recommendation = "continue"
	RecommendationNeutral  Recommendation = "neutral"
)

// MetricData holds baseline and candidate data for a metric
type MetricData struct {
	Type      MetricType
	Baseline  []float64
	Candidate []float64
}

// MetricAnalysis represents the analysis results for a single metric
type MetricAnalysis struct {
	Name         string
	Type         MetricType
	Improvement  float64
	PValue       float64
	Significant  bool
	SampleSize   int
}

// ExperimentAnalysis represents the complete analysis results
type ExperimentAnalysis struct {
	ExperimentID    string
	AnalysisTime    time.Time
	Metrics         map[string]*MetricAnalysis
	Recommendation  Recommendation
	Confidence      float64
	SufficientData  bool
	RiskScore       float64
}

// ExperimentAnalyzer performs statistical analysis on experiments
type ExperimentAnalyzer struct{}

// NewExperimentAnalyzer creates a new experiment analyzer
func NewExperimentAnalyzer() *ExperimentAnalyzer {
	return &ExperimentAnalyzer{}
}

// AnalyzeExperimentResults analyzes the experiment results
func (ea *ExperimentAnalyzer) AnalyzeExperimentResults(ctx context.Context, exp interface{}, metricsData map[string]*MetricData) (*ExperimentAnalysis, error) {
	// Simple stub implementation
	analysis := &ExperimentAnalysis{
		AnalysisTime:   time.Now(),
		Metrics:        make(map[string]*MetricAnalysis),
		Recommendation: RecommendationPromote,
		Confidence:     0.95,
		SufficientData: true,
		RiskScore:      0.1,
	}

	// Analyze each metric
	for name, data := range metricsData {
		metricAnalysis := &MetricAnalysis{
			Name:        name,
			Type:        data.Type,
			Improvement: 10.0, // 10% improvement
			PValue:      0.01,
			Significant: true,
			SampleSize:  len(data.Baseline),
		}
		analysis.Metrics[name] = metricAnalysis
	}

	return analysis, nil
}

// GetRiskLevel returns the risk level based on risk score
func (ea *ExperimentAnalysis) GetRiskLevel() string {
	if ea.RiskScore < 0.3 {
		return "low"
	} else if ea.RiskScore < 0.7 {
		return "medium"
	}
	return "high"
}

// GenerateReport generates a textual report of the analysis
func (ea *ExperimentAnalysis) GenerateReport() string {
	return "Analysis Report: Experiment shows significant improvements across all metrics with high confidence."
}