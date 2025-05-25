package simulator

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// PrometheusMetricsEmitter emits process metrics to Prometheus
type PrometheusMetricsEmitter struct {
	logger *zap.Logger
	port   int

	// Process metrics that simulate what the hostmetrics receiver would collect
	processCPUSeconds    *prometheus.CounterVec
	processMemoryBytes   *prometheus.GaugeVec
	processThreads       *prometheus.GaugeVec
	processOpenFDs       *prometheus.GaugeVec
	processStartTime     *prometheus.GaugeVec
	processUptime        *prometheus.GaugeVec
	
	// Additional metrics for monitoring the simulator itself
	simulatorInfo        prometheus.Gauge
	simulatorUptime      prometheus.Gauge
}

// NewPrometheusMetricsEmitter creates a new Prometheus metrics emitter
func NewPrometheusMetricsEmitter(logger *zap.Logger, port int) *PrometheusMetricsEmitter {
	emitter := &PrometheusMetricsEmitter{
		logger: logger,
		port:   port,

		// Define process metrics similar to what OpenTelemetry hostmetrics receiver collects
		processCPUSeconds: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "process",
				Name:      "cpu_seconds_total",
				Help:      "Total user and system CPU time spent in seconds",
			},
			[]string{"process_name", "pid", "priority", "host"},
		),

		processMemoryBytes: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "process",
				Name:      "memory_bytes",
				Help:      "Memory usage in bytes",
			},
			[]string{"process_name", "pid", "priority", "host", "type"},
		),

		processThreads: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "process",
				Name:      "threads",
				Help:      "Number of threads",
			},
			[]string{"process_name", "pid", "priority", "host"},
		),

		processOpenFDs: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "process",
				Name:      "open_fds",
				Help:      "Number of open file descriptors",
			},
			[]string{"process_name", "pid", "priority", "host"},
		),

		processStartTime: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "process",
				Name:      "start_time_seconds",
				Help:      "Start time of the process since unix epoch in seconds",
			},
			[]string{"process_name", "pid", "priority", "host"},
		),

		processUptime: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "process",
				Name:      "uptime_seconds",
				Help:      "Process uptime in seconds",
			},
			[]string{"process_name", "pid", "priority", "host"},
		),

		// Simulator info metrics
		simulatorInfo: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "phoenix_simulator",
				Name:      "info",
				Help:      "Information about the Phoenix process simulator",
			},
		),

		simulatorUptime: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "phoenix_simulator",
				Name:      "uptime_seconds",
				Help:      "Uptime of the simulator in seconds",
			},
		),
	}

	// Set simulator info
	emitter.simulatorInfo.Set(1)

	return emitter
}

// Start starts the metrics HTTP server
func (e *PrometheusMetricsEmitter) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", e.port),
		Handler: mux,
	}

	go func() {
		e.logger.Info("starting metrics server", zap.Int("port", e.port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			e.logger.Error("metrics server error", zap.Error(err))
		}
	}()

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			e.logger.Error("metrics server shutdown error", zap.Error(err))
		}
	}()

	return nil
}

// EmitProcessMetrics emits metrics for a single process
func (e *PrometheusMetricsEmitter) EmitProcessMetrics(hostname string, process *SimulatedProcess) {
	pidStr := fmt.Sprintf("%d", process.PID)
	
	// CPU metrics (convert percentage to seconds)
	cpuSeconds := process.CPUUsage / 100.0 * time.Since(process.StartTime).Seconds()
	e.processCPUSeconds.WithLabelValues(
		process.Name,
		pidStr,
		process.Priority,
		hostname,
	).Add(cpuSeconds)

	// Memory metrics (convert MB to bytes)
	memoryBytes := process.MemoryUsage * 1024 * 1024
	e.processMemoryBytes.WithLabelValues(
		process.Name,
		pidStr,
		process.Priority,
		hostname,
		"rss", // Resident Set Size
	).Set(memoryBytes)

	// Virtual memory (typically 2-3x RSS)
	virtualMemory := memoryBytes * 2.5
	e.processMemoryBytes.WithLabelValues(
		process.Name,
		pidStr,
		process.Priority,
		hostname,
		"vms", // Virtual Memory Size
	).Set(virtualMemory)

	// Thread count (simulated based on process type)
	threads := e.estimateThreadCount(process.Name)
	e.processThreads.WithLabelValues(
		process.Name,
		pidStr,
		process.Priority,
		hostname,
	).Set(float64(threads))

	// Open file descriptors (simulated)
	fds := e.estimateOpenFDs(process.Name)
	e.processOpenFDs.WithLabelValues(
		process.Name,
		pidStr,
		process.Priority,
		hostname,
	).Set(float64(fds))

	// Start time
	e.processStartTime.WithLabelValues(
		process.Name,
		pidStr,
		process.Priority,
		hostname,
	).Set(float64(process.StartTime.Unix()))

	// Uptime
	e.processUptime.WithLabelValues(
		process.Name,
		pidStr,
		process.Priority,
		hostname,
	).Set(time.Since(process.StartTime).Seconds())
}

// EmitSystemMetrics emits overall system metrics
func (e *PrometheusMetricsEmitter) EmitSystemMetrics(startTime time.Time, totalCPU, totalMemory float64, processCount int) {
	// Update simulator uptime
	e.simulatorUptime.Set(time.Since(startTime).Seconds())
}

// ClearProcessMetrics removes metrics for a stopped process
func (e *PrometheusMetricsEmitter) ClearProcessMetrics(hostname string, process *SimulatedProcess) {
	pidStr := fmt.Sprintf("%d", process.PID)
	
	// Delete all metrics for this process
	e.processCPUSeconds.DeleteLabelValues(process.Name, pidStr, process.Priority, hostname)
	e.processMemoryBytes.DeleteLabelValues(process.Name, pidStr, process.Priority, hostname, "rss")
	e.processMemoryBytes.DeleteLabelValues(process.Name, pidStr, process.Priority, hostname, "vms")
	e.processThreads.DeleteLabelValues(process.Name, pidStr, process.Priority, hostname)
	e.processOpenFDs.DeleteLabelValues(process.Name, pidStr, process.Priority, hostname)
	e.processStartTime.DeleteLabelValues(process.Name, pidStr, process.Priority, hostname)
	e.processUptime.DeleteLabelValues(process.Name, pidStr, process.Priority, hostname)
}

// Helper methods to estimate realistic values

func (e *PrometheusMetricsEmitter) estimateThreadCount(processName string) int {
	// Estimate thread count based on process type
	switch {
	case contains(processName, "nginx"):
		return 4 // Main + workers
	case contains(processName, "postgres"):
		return 10 // Connection handlers
	case contains(processName, "redis"):
		return 4 // Single-threaded with some IO threads
	case contains(processName, "java"):
		return 25 // JVM threads
	case contains(processName, "python"):
		return 2 // GIL limits threading
	case contains(processName, "node"):
		return 8 // Event loop + workers
	case contains(processName, "chrome"):
		return 15 // Multi-process architecture
	default:
		return 1
	}
}

func (e *PrometheusMetricsEmitter) estimateOpenFDs(processName string) int {
	// Estimate open file descriptors based on process type
	switch {
	case contains(processName, "nginx"):
		return 50 // Logs, sockets, config
	case contains(processName, "postgres"):
		return 200 // Database files, connections
	case contains(processName, "redis"):
		return 30 // AOF, RDB, sockets
	case contains(processName, "java"):
		return 150 // JARs, sockets, files
	case contains(processName, "node"):
		return 40 // Modules, sockets
	default:
		return 10 // Basic stdin/stdout/stderr + some files
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}