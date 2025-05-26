package framework

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// TestReport represents the structure of an acceptance test report
type TestReport struct {
	TestName    string                 `json:"test_name"`
	TestID      string                 `json:"test_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Duration    time.Duration          `json:"duration"`
	Passed      bool                   `json:"passed"`
	Results     map[string]interface{} `json:"results"`
	Errors      []string               `json:"errors,omitempty"`
	Environment map[string]string      `json:"environment"`
	KPIs        map[string]KPIResult   `json:"kpis,omitempty"`
}

// KPIResult represents a KPI measurement result
type KPIResult struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Target      interface{} `json:"target"`
	Unit        string      `json:"unit"`
	Passed      bool        `json:"passed"`
	Description string      `json:"description"`
}

// Save saves the report to a file
func (r *TestReport) Save(filepath string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	
	// Ensure directory exists
	dir := "reports"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}
	
	// Write file
	fullPath := fmt.Sprintf("%s/%s", dir, filepath)
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}
	
	return nil
}

// TestSummary provides a summary of multiple test reports
type TestSummary struct {
	TotalTests   int                      `json:"total_tests"`
	PassedTests  int                      `json:"passed_tests"`
	FailedTests  int                      `json:"failed_tests"`
	TotalDuration time.Duration           `json:"total_duration"`
	StartTime    time.Time               `json:"start_time"`
	EndTime      time.Time               `json:"end_time"`
	KPISummary   map[string]KPISummary   `json:"kpi_summary"`
	TestResults  []TestResult            `json:"test_results"`
}

// KPISummary summarizes KPI results across tests
type KPISummary struct {
	TotalMeasurements int     `json:"total_measurements"`
	PassedMeasurements int    `json:"passed_measurements"`
	AverageValue      float64 `json:"average_value"`
	MinValue          float64 `json:"min_value"`
	MaxValue          float64 `json:"max_value"`
}

// TestResult provides a summary of a single test
type TestResult struct {
	TestName  string        `json:"test_name"`
	TestID    string        `json:"test_id"`
	Passed    bool          `json:"passed"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
	Errors    []string      `json:"errors,omitempty"`
}