package service

import (
	"context"
	"testing"
	"time"

	"github.com/phoenix/platform/pkg/analytics"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAnalyticsService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := New(logger)
	ctx := context.Background()

	t.Run("AnalyzeMetrics", func(t *testing.T) {
		metrics := []analytics.Metric{
			{
				Timestamp: time.Now(),
				Name:      "test_metric",
				Value:     42.0,
				Labels:    map[string]string{"test": "value"},
			},
		}

		result, err := service.AnalyzeMetrics(ctx, metrics)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Analysis completed", result.Summary)
	})

	t.Run("DetectAnomalies", func(t *testing.T) {
		metrics := []analytics.Metric{
			{
				Timestamp: time.Now(),
				Name:      "test_metric",
				Value:     42.0,
				Labels:    map[string]string{"test": "value"},
			},
		}

		anomalies, err := service.DetectAnomalies(ctx, metrics)
		assert.NoError(t, err)
		assert.NotEmpty(t, anomalies)
		assert.Equal(t, "spike", anomalies[0].Type)
		assert.Equal(t, "high", anomalies[0].Severity)
	})

	t.Run("GenerateReport", func(t *testing.T) {
		analysis := &analytics.AnalysisResult{
			Timestamp: time.Now(),
			Summary:   "Test analysis",
		}

		report, err := service.GenerateReport(ctx, analysis)
		assert.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, "Report generated", report.Summary)
	})
}
