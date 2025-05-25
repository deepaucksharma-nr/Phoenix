package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// APIRequestsTotal tracks total API requests
	APIRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "phoenix_api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// APIRequestDuration tracks API request duration
	APIRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "phoenix_api_request_duration_seconds",
			Help:    "API request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// ExperimentsActive tracks active experiments
	ExperimentsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "phoenix_experiments_active",
			Help: "Number of active experiments",
		},
	)
)

// InitMetrics initializes metrics
func InitMetrics() {
	// Metrics are automatically registered with promauto
}