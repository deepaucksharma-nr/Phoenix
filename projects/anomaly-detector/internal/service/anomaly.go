package service

import (
	"context"
	"sync"
	"time"

	"github.com/phoenix/platform/pkg/analytics"
	"go.uber.org/zap"
)

// Config represents the configuration for anomaly detection
type Config struct {
	Threshold  float64
	WindowSize int
	Algorithms []string
	Parameters map[string]string
}

// Service represents the anomaly detector service
type Service struct {
	logger *zap.Logger
	config Config
	mu     sync.RWMutex
	stats  map[string]string
}

// New creates a new anomaly detector service
func New(logger *zap.Logger) *Service {
	return &Service{
		logger: logger,
		config: Config{
			Threshold:  0.95,
			WindowSize: 100,
			Algorithms: []string{"zscore", "mad"},
			Parameters: map[string]string{
				"zscore_threshold": "3.0",
				"mad_threshold":    "3.5",
			},
		},
		stats: make(map[string]string),
	}
}

// Detect detects anomalies in the given metrics
func (s *Service) Detect(ctx context.Context, metrics []analytics.Metric, config *Config) ([]analytics.Anomaly, error) {
	s.logger.Info("Detecting anomalies", zap.Int("count", len(metrics)))

	// Use provided config or default
	detectionConfig := s.config
	if config != nil {
		detectionConfig = *config
	}

	// Simple anomaly detection based on threshold
	anomalies := make([]analytics.Anomaly, 0)
	for _, metric := range metrics {
		if metric.Value > detectionConfig.Threshold {
			anomalies = append(anomalies, analytics.Anomaly{
				Timestamp: time.Now(),
				Type:      "threshold_exceeded",
				Severity:  "high",
				Metric:    metric,
				Details: map[string]interface{}{
					"threshold": detectionConfig.Threshold,
					"value":     metric.Value,
				},
			})
		}
	}

	// Update stats
	s.mu.Lock()
	s.stats["last_detection"] = time.Now().Format(time.RFC3339)
	s.stats["metrics_processed"] = string(len(metrics))
	s.stats["anomalies_found"] = string(len(anomalies))
	s.mu.Unlock()

	return anomalies, nil
}

// ConfigureDetection configures the anomaly detection parameters
func (s *Service) ConfigureDetection(ctx context.Context, config Config) error {
	s.logger.Info("Configuring anomaly detection", zap.Any("config", config))

	s.mu.Lock()
	s.config = config
	s.mu.Unlock()

	return nil
}

// GetDetectionStatus returns the current status of anomaly detection
func (s *Service) GetDetectionStatus(ctx context.Context) (Config, map[string]string, error) {
	s.logger.Info("Getting detection status")

	s.mu.RLock()
	config := s.config
	stats := make(map[string]string)
	for k, v := range s.stats {
		stats[k] = v
	}
	s.mu.RUnlock()

	return config, stats, nil
}
