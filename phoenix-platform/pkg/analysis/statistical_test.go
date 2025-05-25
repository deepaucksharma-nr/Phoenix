package analysis

import (
	"math"
	"testing"
)

func TestNewStatisticalAnalyzer(t *testing.T) {
	analyzer := NewStatisticalAnalyzer()
	if analyzer == nil {
		t.Fatal("NewStatisticalAnalyzer returned nil")
	}
}

func TestCalculateMean(t *testing.T) {
	tests := []struct {
		name     string
		data     []float64
		expected float64
	}{
		{
			name:     "empty data",
			data:     []float64{},
			expected: 0,
		},
		{
			name:     "single value",
			data:     []float64{5.0},
			expected: 5.0,
		},
		{
			name:     "multiple values",
			data:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 3.0,
		},
		{
			name:     "negative values",
			data:     []float64{-5.0, -3.0, -1.0, 1.0, 3.0, 5.0},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMean(tt.data)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("calculateMean() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTTest(t *testing.T) {
	analyzer := NewStatisticalAnalyzer()

	// Test with sufficient sample sizes
	baseline := []float64{
		100, 102, 98, 103, 101, 99, 100, 102, 101, 98,
		100, 102, 98, 103, 101, 99, 100, 102, 101, 98,
		100, 102, 98, 103, 101, 99, 100, 102, 101, 98,
		100, 102, 98, 103, 101, 99, 100, 102, 101, 98,
	}

	// Candidate with 10% improvement
	candidate := make([]float64, len(baseline))
	for i, v := range baseline {
		candidate[i] = v * 0.9 // 10% reduction (improvement for latency)
	}

	result := analyzer.TTest(baseline, candidate)

	// Verify basic properties
	if result.BaselineSampleSize != len(baseline) {
		t.Errorf("Expected baseline sample size %d, got %d", len(baseline), result.BaselineSampleSize)
	}

	if result.CandidateSampleSize != len(candidate) {
		t.Errorf("Expected candidate sample size %d, got %d", len(candidate), result.CandidateSampleSize)
	}

	// With 10% improvement, we expect significant results
	if !result.Significant {
		t.Error("Expected significant result for 10% improvement")
	}

	// Relative improvement should be around -10% (negative because lower is better)
	expectedImprovement := -10.0
	if math.Abs(result.RelativeImprovement-expectedImprovement) > 1.0 {
		t.Errorf("Expected relative improvement around %v%%, got %v%%", expectedImprovement, result.RelativeImprovement)
	}
}

func TestTTest_InsufficientData(t *testing.T) {
	analyzer := NewStatisticalAnalyzer()

	// Test with insufficient sample size
	baseline := []float64{100, 102, 98, 103, 101}
	candidate := []float64{90, 92, 88, 93, 91}

	result := analyzer.TTest(baseline, candidate)

	// Should not be marked as significant due to insufficient data
	if result.Significant {
		t.Error("Should not be significant with insufficient sample size")
	}
}

func TestBonferroniCorrection(t *testing.T) {
	analyzer := NewStatisticalAnalyzer()

	// Test with multiple p-values
	pValues := []float64{0.01, 0.03, 0.04, 0.06, 0.001}
	alpha := 0.05

	results := analyzer.BonferroniCorrection(pValues, alpha)

	// With 5 comparisons, adjusted alpha = 0.05/5 = 0.01
	// Only p-values < 0.01 should be significant
	expected := []bool{false, false, false, false, true} // Only 0.001 < 0.01

	for i, result := range results {
		if result != expected[i] {
			t.Errorf("Test %d: expected %v, got %v", i, expected[i], result)
		}
	}
}

func TestConfidenceInterval(t *testing.T) {
	analyzer := NewStatisticalAnalyzer()

	data := []float64{
		100, 102, 98, 103, 101, 99, 100, 102, 101, 98,
		100, 102, 98, 103, 101, 99, 100, 102, 101, 98,
		100, 102, 98, 103, 101, 99, 100, 102, 101, 98,
	}

	lower, upper := analyzer.ConfidenceInterval(data, 0.95)

	mean := calculateMean(data)

	// Confidence interval should contain the mean
	if lower > mean || upper < mean {
		t.Errorf("Confidence interval [%v, %v] does not contain mean %v", lower, upper, mean)
	}

	// Interval should be reasonable (not too wide)
	width := upper - lower
	if width > 5.0 { // Assuming reasonable variance
		t.Errorf("Confidence interval too wide: %v", width)
	}
}

func TestAnalyzeExperiment(t *testing.T) {
	// Create test data
	metricsData := map[string]struct{ Baseline, Candidate []float64 }{
		"latency": {
			Baseline: []float64{
				100, 102, 98, 103, 101, 99, 100, 102, 101, 98,
				100, 102, 98, 103, 101, 99, 100, 102, 101, 98,
				100, 102, 98, 103, 101, 99, 100, 102, 101, 98,
				100, 102, 98, 103, 101, 99, 100, 102, 101, 98,
			},
			Candidate: make([]float64, 40),
		},
		"error_rate": {
			Baseline: []float64{
				0.05, 0.06, 0.04, 0.05, 0.06, 0.05, 0.04, 0.05, 0.06, 0.05,
				0.05, 0.06, 0.04, 0.05, 0.06, 0.05, 0.04, 0.05, 0.06, 0.05,
				0.05, 0.06, 0.04, 0.05, 0.06, 0.05, 0.04, 0.05, 0.06, 0.05,
				0.05, 0.06, 0.04, 0.05, 0.06, 0.05, 0.04, 0.05, 0.06, 0.05,
			},
			Candidate: make([]float64, 40),
		},
	}

	// Set candidate values with improvements
	for i := range metricsData["latency"].Candidate {
		metricsData["latency"].Candidate[i] = metricsData["latency"].Baseline[i] * 0.9 // 10% improvement
	}
	for i := range metricsData["error_rate"].Candidate {
		metricsData["error_rate"].Candidate[i] = metricsData["error_rate"].Baseline[i] * 0.8 // 20% improvement
	}

	result := AnalyzeExperiment("test-experiment-1", metricsData)

	// Verify result structure
	if result.ExperimentID != "test-experiment-1" {
		t.Errorf("Expected experiment ID 'test-experiment-1', got '%s'", result.ExperimentID)
	}

	if len(result.Metrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(result.Metrics))
	}

	// Check summary
	if result.Summary.MetricsAnalyzed != 2 {
		t.Errorf("Expected 2 metrics analyzed, got %d", result.Summary.MetricsAnalyzed)
	}

	if !result.Summary.SufficientData {
		t.Error("Expected sufficient data")
	}

	// With improvements, we should get a positive recommendation
	if result.Summary.Recommendation != "Promote candidate - significant improvements detected" {
		t.Errorf("Unexpected recommendation: %s", result.Summary.Recommendation)
	}

	if result.Summary.RiskLevel != "low" {
		t.Errorf("Expected low risk level, got %s", result.Summary.RiskLevel)
	}
}

func TestAnalyzeMetricByType(t *testing.T) {
	baseline := []float64{
		100, 102, 98, 103, 101, 99, 100, 102, 101, 98,
		100, 102, 98, 103, 101, 99, 100, 102, 101, 98,
		100, 102, 98, 103, 101, 99, 100, 102, 101, 98,
		100, 102, 98, 103, 101, 99, 100, 102, 101, 98,
	}

	tests := []struct {
		name         string
		metricType   MetricType
		improvement  float64
		expectNegative bool
	}{
		{
			name:         "latency improvement",
			metricType:   MetricTypeLatency,
			improvement:  0.9, // 10% reduction
			expectNegative: true,
		},
		{
			name:         "throughput improvement",
			metricType:   MetricTypeThroughput,
			improvement:  1.1, // 10% increase
			expectNegative: false,
		},
		{
			name:         "error rate improvement",
			metricType:   MetricTypeErrorRate,
			improvement:  0.8, // 20% reduction
			expectNegative: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := make([]float64, len(baseline))
			for i, v := range baseline {
				candidate[i] = v * tt.improvement
			}

			result := AnalyzeMetricByType(tt.metricType, baseline, candidate)

			if tt.expectNegative && result.RelativeImprovement > 0 {
				t.Errorf("Expected negative improvement for %s, got %v", tt.metricType, result.RelativeImprovement)
			}
			if !tt.expectNegative && result.RelativeImprovement < 0 {
				t.Errorf("Expected positive improvement for %s, got %v", tt.metricType, result.RelativeImprovement)
			}
		})
	}
}

func TestCalculateSampleSize(t *testing.T) {
	tests := []struct {
		name       string
		effectSize float64
		alpha      float64
		power      float64
		variance   float64
		minExpected int
	}{
		{
			name:       "standard parameters",
			effectSize: 5.0,
			alpha:      0.05,
			power:      0.8,
			variance:   100.0,
			minExpected: 50,
		},
		{
			name:       "higher power requirement",
			effectSize: 5.0,
			alpha:      0.05,
			power:      0.95,
			variance:   100.0,
			minExpected: 80,
		},
		{
			name:       "smaller effect size",
			effectSize: 2.0,
			alpha:      0.05,
			power:      0.8,
			variance:   100.0,
			minExpected: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := CalculateSampleSize(tt.effectSize, tt.alpha, tt.power, tt.variance)
			if n < tt.minExpected {
				t.Errorf("Expected sample size >= %d, got %d", tt.minExpected, n)
			}
		})
	}
}

func TestPercentile(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	tests := []struct {
		percentile float64
		expected   float64
		tolerance  float64
	}{
		{0, 1, 0.1},
		{25, 3.25, 0.5},
		{50, 5.5, 0.5},
		{75, 7.75, 0.5},
		{100, 10, 0.1},
	}

	for _, tt := range tests {
		result := Percentile(data, tt.percentile)
		if math.Abs(result-tt.expected) > tt.tolerance {
			t.Errorf("Percentile(%v) = %v, want %v (±%v)", tt.percentile, result, tt.expected, tt.tolerance)
		}
	}
}