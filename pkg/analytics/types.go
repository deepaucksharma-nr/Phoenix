package analytics

import "time"

// Metric represents a single metric data point
type Metric struct {
	Timestamp time.Time
	Name      string
	Value     float64
	Labels    map[string]string
}

// AnalysisResult contains the results of metric analysis
type AnalysisResult struct {
	Timestamp time.Time
	Summary   string
	Metrics   []Metric
	Stats     map[string]float64
}

// Anomaly represents a detected anomaly in the metrics
type Anomaly struct {
	Timestamp time.Time
	Type      string
	Severity  string
	Metric    Metric
	Details   map[string]interface{}
}

// Report represents a comprehensive analysis report
type Report struct {
	GeneratedAt     time.Time
	Summary         string
	Analysis        *AnalysisResult
	Anomalies       []Anomaly
	Recommendations []string
}
