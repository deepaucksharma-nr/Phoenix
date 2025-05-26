package service

import (
	"context"
	"testing"
	"time"

	"github.com/phoenix/platform/pkg/analytics"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAnomalyDetector(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := New(logger)
	ctx := context.Background()

	t.Run("Detect", func(t *testing.T) {
		metrics := []analytics.Metric{
			{
				Timestamp: time.Now(),
				Name:      "test_metric",
				Value:     42.0,
				Labels:    map[string]string{"test": "value"},
			},
		}

		anomalies, err := service.Detect(ctx, metrics, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, anomalies)
		assert.Equal(t, "spike", anomalies[0].Type)
		assert.Equal(t, "high", anomalies[0].Severity)
	})

	t.Run("ConfigureDetection", func(t *testing.T) {
		config := Config{
			Threshold:  0.99,
			WindowSize: 200,
			Algorithms: []string{"zscore"},
			Parameters: map[string]string{
				"zscore_threshold": "4.0",
			},
		}

		err := service.ConfigureDetection(ctx, config)
		assert.NoError(t, err)

		// Verify configuration
		currentConfig, _, err := service.GetDetectionStatus(ctx)
		assert.NoError(t, err)
		assert.Equal(t, config.Threshold, currentConfig.Threshold)
		assert.Equal(t, config.WindowSize, currentConfig.WindowSize)
		assert.Equal(t, config.Algorithms, currentConfig.Algorithms)
		assert.Equal(t, config.Parameters, currentConfig.Parameters)
	})

	t.Run("GetDetectionStatus", func(t *testing.T) {
		config, stats, err := service.GetDetectionStatus(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "last_detection")
		assert.Contains(t, stats, "metrics_processed")
		assert.Contains(t, stats, "anomalies_found")
	})
}
