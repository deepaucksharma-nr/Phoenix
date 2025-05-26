//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/phoenix/platform/pkg/common/analysis"
)

// TestExperimentAnalysis verifies the experiment analysis workflow
func TestExperimentAnalysis(t *testing.T) {
	analyzer := analysis.NewExperimentAnalyzer()

	metrics := map[string]*analysis.MetricData{
		"latency": {
			Type:      analysis.MetricTypeLatency,
			Baseline:  []float64{120, 110, 115},
			Candidate: []float64{90, 95, 92},
		},
		"error_rate": {
			Type:      analysis.MetricTypeErrorRate,
			Baseline:  []float64{0.02, 0.015, 0.018},
			Candidate: []float64{0.01, 0.012, 0.011},
		},
	}

	result, err := analyzer.AnalyzeExperimentResults(context.Background(), nil, metrics)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, analysis.RecommendationPromote, result.Recommendation)
	assert.True(t, result.SufficientData)
	assert.Contains(t, result.Metrics, "latency")
	assert.Contains(t, result.Metrics, "error_rate")
}
