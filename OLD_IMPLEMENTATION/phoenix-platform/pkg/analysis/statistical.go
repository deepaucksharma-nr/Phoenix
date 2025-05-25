package analysis

import (
	"math"
	"sort"
)

// TestResult represents the outcome of a statistical test
type TestResult struct {
	// Test statistic value (e.g., t-statistic)
	Statistic float64
	
	// P-value of the test
	PValue float64
	
	// Whether the result is statistically significant
	Significant bool
	
	// Confidence level used (e.g., 0.95 for 95%)
	Confidence float64
	
	// Effect size (Cohen's d for t-test)
	EffectSize float64
	
	// Sample sizes
	BaselineSampleSize  int
	CandidateSampleSize int
	
	// Confidence intervals
	BaselineCI  ConfidenceInterval
	CandidateCI ConfidenceInterval
	
	// Relative improvement percentage
	RelativeImprovement float64
}

// ConfidenceInterval represents a confidence interval
type ConfidenceInterval struct {
	Lower float64
	Upper float64
	Mean  float64
}

// StatisticalAnalyzer performs statistical analysis on experiment results
type StatisticalAnalyzer interface {
	// Calculate p-value for A/B test results
	CalculatePValue(baseline, candidate []float64) float64
	
	// Determine statistical significance
	IsSignificant(pValue float64, alpha float64) bool
	
	// Calculate confidence intervals
	ConfidenceInterval(data []float64, confidence float64) (lower, upper float64)
	
	// Perform t-test for performance metrics
	TTest(baseline, candidate []float64) TestResult
	
	// Perform multiple comparison correction
	BonferroniCorrection(pValues []float64, alpha float64) []bool
	
	// Calculate minimum detectable effect
	MinimumDetectableEffect(baselineVariance float64, sampleSize int, alpha, power float64) float64
	
	// Check if sample size is sufficient
	IsSampleSizeSufficient(baseline, candidate []float64, minEffect float64) bool
}

// analyzer implements StatisticalAnalyzer
type analyzer struct {
	defaultAlpha      float64
	defaultConfidence float64
	minSampleSize     int
}

// NewStatisticalAnalyzer creates a new statistical analyzer
func NewStatisticalAnalyzer() StatisticalAnalyzer {
	return &analyzer{
		defaultAlpha:      0.05,
		defaultConfidence: 0.95,
		minSampleSize:     30,
	}
}

// CalculatePValue calculates the p-value using Welch's t-test
func (a *analyzer) CalculatePValue(baseline, candidate []float64) float64 {
	result := a.TTest(baseline, candidate)
	return result.PValue
}

// IsSignificant determines if a result is statistically significant
func (a *analyzer) IsSignificant(pValue float64, alpha float64) bool {
	if alpha <= 0 {
		alpha = a.defaultAlpha
	}
	return pValue < alpha
}

// ConfidenceInterval calculates the confidence interval for a dataset
func (a *analyzer) ConfidenceInterval(data []float64, confidence float64) (lower, upper float64) {
	if len(data) == 0 {
		return 0, 0
	}
	
	if confidence <= 0 || confidence >= 1 {
		confidence = a.defaultConfidence
	}
	
	mean := calculateMean(data)
	stdErr := calculateStandardError(data)
	
	// For large samples, use z-score; for small samples, use t-distribution
	// Here we use a simplified z-score approach
	zScore := inversNormalCDF((1 + confidence) / 2)
	
	margin := zScore * stdErr
	return mean - margin, mean + margin
}

// TTest performs Welch's t-test for unequal variances
func (a *analyzer) TTest(baseline, candidate []float64) TestResult {
	result := TestResult{
		BaselineSampleSize:  len(baseline),
		CandidateSampleSize: len(candidate),
		Confidence:          a.defaultConfidence,
	}
	
	// Check minimum sample size
	if len(baseline) < a.minSampleSize || len(candidate) < a.minSampleSize {
		return result
	}
	
	// Calculate means
	baselineMean := calculateMean(baseline)
	candidateMean := calculateMean(candidate)
	
	// Calculate variances
	baselineVar := calculateVariance(baseline, baselineMean)
	candidateVar := calculateVariance(candidate, candidateMean)
	
	// Calculate standard errors
	baselineSE := math.Sqrt(baselineVar / float64(len(baseline)))
	candidateSE := math.Sqrt(candidateVar / float64(len(candidate)))
	
	// Calculate t-statistic
	pooledSE := math.Sqrt(baselineSE*baselineSE + candidateSE*candidateSE)
	if pooledSE > 0 {
		result.Statistic = (candidateMean - baselineMean) / pooledSE
	}
	
	// Calculate degrees of freedom (Welch-Satterthwaite equation)
	df := calculateWelchDF(baseline, candidate, baselineVar, candidateVar)
	
	// Calculate p-value (two-tailed test)
	result.PValue = calculatePValueFromT(result.Statistic, df)
	
	// Determine significance
	result.Significant = a.IsSignificant(result.PValue, a.defaultAlpha)
	
	// Calculate effect size (Cohen's d)
	pooledSD := math.Sqrt((baselineVar + candidateVar) / 2)
	if pooledSD > 0 {
		result.EffectSize = (candidateMean - baselineMean) / pooledSD
	}
	
	// Calculate confidence intervals
	baselineLower, baselineUpper := a.ConfidenceInterval(baseline, a.defaultConfidence)
	candidateLower, candidateUpper := a.ConfidenceInterval(candidate, a.defaultConfidence)
	
	result.BaselineCI = ConfidenceInterval{
		Lower: baselineLower,
		Upper: baselineUpper,
		Mean:  baselineMean,
	}
	
	result.CandidateCI = ConfidenceInterval{
		Lower: candidateLower,
		Upper: candidateUpper,
		Mean:  candidateMean,
	}
	
	// Calculate relative improvement
	if baselineMean != 0 {
		result.RelativeImprovement = ((candidateMean - baselineMean) / math.Abs(baselineMean)) * 100
	}
	
	return result
}

// BonferroniCorrection applies Bonferroni correction for multiple comparisons
func (a *analyzer) BonferroniCorrection(pValues []float64, alpha float64) []bool {
	if alpha <= 0 {
		alpha = a.defaultAlpha
	}
	
	n := len(pValues)
	if n == 0 {
		return []bool{}
	}
	
	// Adjusted alpha
	adjustedAlpha := alpha / float64(n)
	
	results := make([]bool, n)
	for i, pValue := range pValues {
		results[i] = pValue < adjustedAlpha
	}
	
	return results
}

// MinimumDetectableEffect calculates the minimum detectable effect size
func (a *analyzer) MinimumDetectableEffect(baselineVariance float64, sampleSize int, alpha, power float64) float64 {
	if alpha <= 0 {
		alpha = a.defaultAlpha
	}
	if power <= 0 || power >= 1 {
		power = 0.8 // Default 80% power
	}
	
	// Z-scores for alpha (two-tailed) and power
	zAlpha := inversNormalCDF(1 - alpha/2)
	zPower := inversNormalCDF(power)
	
	// Minimum detectable effect
	mde := (zAlpha + zPower) * math.Sqrt(2*baselineVariance/float64(sampleSize))
	
	return mde
}

// IsSampleSizeSufficient checks if the sample size is sufficient for detecting the minimum effect
func (a *analyzer) IsSampleSizeSufficient(baseline, candidate []float64, minEffect float64) bool {
	if len(baseline) < a.minSampleSize || len(candidate) < a.minSampleSize {
		return false
	}
	
	baselineVar := calculateVariance(baseline, calculateMean(baseline))
	n := len(baseline)
	
	// Calculate minimum detectable effect for current sample size
	mde := a.MinimumDetectableEffect(baselineVar, n, a.defaultAlpha, 0.8)
	
	return mde <= minEffect
}

// Helper functions

func calculateMean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func calculateVariance(data []float64, mean float64) float64 {
	if len(data) <= 1 {
		return 0
	}
	
	sumSquares := 0.0
	for _, v := range data {
		diff := v - mean
		sumSquares += diff * diff
	}
	
	return sumSquares / float64(len(data)-1)
}

func calculateStandardError(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	
	mean := calculateMean(data)
	variance := calculateVariance(data, mean)
	return math.Sqrt(variance / float64(len(data)))
}

func calculateWelchDF(baseline, candidate []float64, baselineVar, candidateVar float64) float64 {
	n1 := float64(len(baseline))
	n2 := float64(len(candidate))
	
	s1Squared := baselineVar
	s2Squared := candidateVar
	
	numerator := math.Pow(s1Squared/n1+s2Squared/n2, 2)
	denominator := math.Pow(s1Squared/n1, 2)/(n1-1) + math.Pow(s2Squared/n2, 2)/(n2-1)
	
	if denominator == 0 {
		return 0
	}
	
	return numerator / denominator
}

// Simplified p-value calculation from t-statistic
func calculatePValueFromT(t float64, df float64) float64 {
	// This is a simplified approximation
	// In production, use a proper t-distribution CDF
	
	// Convert to standard normal for approximation (valid for df > 30)
	z := math.Abs(t)
	
	// Approximate two-tailed p-value
	if df > 30 {
		return 2 * (1 - normalCDF(z))
	}
	
	// For smaller df, apply correction
	correction := math.Sqrt(df / (df - 2))
	z = z / correction
	
	return 2 * (1 - normalCDF(z))
}

// Simplified normal CDF
func normalCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt(2)))
}

// Simplified inverse normal CDF
func inversNormalCDF(p float64) float64 {
	// Approximation using the inverse error function
	return math.Sqrt(2) * inverseErf(2*p-1)
}

// Simplified inverse error function
func inverseErf(x float64) float64 {
	// Approximation for inverse erf
	a := 0.147
	sign := 1.0
	if x < 0 {
		sign = -1.0
		x = -x
	}
	
	ln := math.Log(1 - x*x)
	t1 := 2/(math.Pi*a) + ln/2
	t2 := ln / a
	
	return sign * math.Sqrt(math.Sqrt(t1*t1-t2) - t1)
}

// AnalysisResult represents aggregated analysis results for an experiment
type AnalysisResult struct {
	ExperimentID string
	Metrics      map[string]TestResult
	Summary      Summary
	Timestamp    int64
}

// Summary provides an overall summary of the experiment
type Summary struct {
	// Overall recommendation
	Recommendation string
	
	// Confidence in the recommendation (0-1)
	Confidence float64
	
	// Number of metrics analyzed
	MetricsAnalyzed int
	
	// Number of statistically significant results
	SignificantResults int
	
	// Whether the experiment has sufficient data
	SufficientData bool
	
	// Risk assessment
	RiskLevel string // "low", "medium", "high"
}

// AnalyzeExperiment performs comprehensive analysis on experiment metrics
func AnalyzeExperiment(experimentID string, metricsData map[string]struct{ Baseline, Candidate []float64 }) AnalysisResult {
	analyzer := NewStatisticalAnalyzer()
	
	result := AnalysisResult{
		ExperimentID: experimentID,
		Metrics:      make(map[string]TestResult),
		Timestamp:    getCurrentTimestamp(),
	}
	
	significantCount := 0
	sufficientData := true
	
	// Analyze each metric
	for metricName, data := range metricsData {
		testResult := analyzer.TTest(data.Baseline, data.Candidate)
		result.Metrics[metricName] = testResult
		
		if testResult.Significant {
			significantCount++
		}
		
		if len(data.Baseline) < 30 || len(data.Candidate) < 30 {
			sufficientData = false
		}
	}
	
	// Generate summary
	result.Summary = Summary{
		MetricsAnalyzed:    len(metricsData),
		SignificantResults: significantCount,
		SufficientData:     sufficientData,
	}
	
	// Determine recommendation
	result.Summary.Recommendation = generateRecommendation(result.Metrics, sufficientData)
	result.Summary.Confidence = calculateConfidence(result.Metrics, sufficientData)
	result.Summary.RiskLevel = assessRisk(result.Metrics)
	
	return result
}

func generateRecommendation(metrics map[string]TestResult, sufficientData bool) string {
	if !sufficientData {
		return "Continue experiment - insufficient data for decision"
	}
	
	// Count improvements and regressions
	improvements := 0
	regressions := 0
	
	for _, result := range metrics {
		if result.Significant {
			if result.RelativeImprovement > 0 {
				improvements++
			} else {
				regressions++
			}
		}
	}
	
	if regressions > 0 {
		return "Do not promote - significant regressions detected"
	}
	
	if improvements > 0 {
		return "Promote candidate - significant improvements detected"
	}
	
	return "No significant difference - decision based on business factors"
}

func calculateConfidence(metrics map[string]TestResult, sufficientData bool) float64 {
	if !sufficientData {
		return 0.3
	}
	
	// Base confidence on effect sizes and p-values
	totalConfidence := 0.0
	count := 0
	
	for _, result := range metrics {
		if result.PValue < 0.001 {
			totalConfidence += 0.95
		} else if result.PValue < 0.01 {
			totalConfidence += 0.85
		} else if result.PValue < 0.05 {
			totalConfidence += 0.75
		} else {
			totalConfidence += 0.5
		}
		count++
	}
	
	if count > 0 {
		return totalConfidence / float64(count)
	}
	
	return 0.5
}

func assessRisk(metrics map[string]TestResult) string {
	// Look for large effect sizes in critical metrics
	maxNegativeEffect := 0.0
	
	for _, result := range metrics {
		if result.EffectSize < maxNegativeEffect {
			maxNegativeEffect = result.EffectSize
		}
	}
	
	if maxNegativeEffect < -0.8 {
		return "high"
	} else if maxNegativeEffect < -0.5 {
		return "medium"
	}
	
	return "low"
}

func getCurrentTimestamp() int64 {
	// In production, use time.Now().Unix()
	return 1706313600 // Placeholder
}

// MetricType represents the type of metric being analyzed
type MetricType string

const (
	MetricTypeLatency    MetricType = "latency"
	MetricTypeThroughput MetricType = "throughput"
	MetricTypeErrorRate  MetricType = "error_rate"
	MetricTypeCost       MetricType = "cost"
)

// AnalyzeMetricByType performs metric-specific analysis
func AnalyzeMetricByType(metricType MetricType, baseline, candidate []float64) TestResult {
	analyzer := NewStatisticalAnalyzer()
	
	// For latency and error rate, lower is better
	// For throughput, higher is better
	// Adjust the data accordingly for consistent interpretation
	
	switch metricType {
	case MetricTypeLatency, MetricTypeErrorRate, MetricTypeCost:
		// Lower is better - no adjustment needed
		return analyzer.TTest(baseline, candidate)
		
	case MetricTypeThroughput:
		// Higher is better - negate values for consistent interpretation
		negBaseline := make([]float64, len(baseline))
		negCandidate := make([]float64, len(candidate))
		
		for i, v := range baseline {
			negBaseline[i] = -v
		}
		for i, v := range candidate {
			negCandidate[i] = -v
		}
		
		result := analyzer.TTest(negBaseline, negCandidate)
		// Adjust the relative improvement sign back
		result.RelativeImprovement = -result.RelativeImprovement
		return result
		
	default:
		return analyzer.TTest(baseline, candidate)
	}
}

// CalculateSampleSize calculates required sample size for desired power
func CalculateSampleSize(effectSize, alpha, power, variance float64) int {
	if alpha <= 0 {
		alpha = 0.05
	}
	if power <= 0 || power >= 1 {
		power = 0.8
	}
	
	// Z-scores
	zAlpha := inversNormalCDF(1 - alpha/2)
	zPower := inversNormalCDF(power)
	
	// Sample size formula
	n := 2 * variance * math.Pow(zAlpha+zPower, 2) / math.Pow(effectSize, 2)
	
	return int(math.Ceil(n))
}

// Percentile calculates the percentile value from a dataset
func Percentile(data []float64, p float64) float64 {
	if len(data) == 0 || p < 0 || p > 100 {
		return 0
	}
	
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)
	
	index := p / 100 * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	
	if lower == upper {
		return sorted[lower]
	}
	
	// Linear interpolation
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}