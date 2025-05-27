package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAPIClient holds the API client for testing
var mockAPIClient *client.APIClient

// Initialize experiment commands for testing
var (
	experimentListCmd = &cobra.Command{
		Use:  "list",
		RunE: runListExperiments,
	}
	experimentCreateCmd = &cobra.Command{
		Use:  "create",
		RunE: runCreateExperiment,
	}
	experimentStatusCmd = &cobra.Command{
		Use:  "status",
		RunE: runExperimentStatus,
	}
	experimentMetricsCmd = &cobra.Command{
		Use:  "metrics",
		RunE: runExperimentMetrics,
	}
)

func init() {
	experimentCmd.AddCommand(experimentListCmd)
	experimentCmd.AddCommand(experimentCreateCmd)
	experimentCmd.AddCommand(experimentStatusCmd)
	experimentCmd.AddCommand(experimentMetricsCmd)
}

// RunWithClient executes a command with a specific API client for testing
func RunWithClient(cmd *cobra.Command, output *cobra.Command, apiClient *client.APIClient, args []string) error {
	mockAPIClient = apiClient
	return cmd.RunE(output, args)
}

func TestExperimentListCommand(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/experiments", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := client.ListExperimentsResponse{
			Experiments: []client.Experiment{
				{
					ID:        "exp-1",
					Name:      "test-experiment-1",
					Status:    "running",
					CreatedAt: time.Now().Add(-1 * time.Hour),
				},
				{
					ID:        "exp-2",
					Name:      "test-experiment-2",
					Status:    "completed",
					CreatedAt: time.Now().Add(-2 * time.Hour),
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create command with output buffer
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Set flags
	experimentListCmd.Flags().String("namespace", "", "")
	experimentListCmd.Flags().String("output", "", "")
	experimentListCmd.Flags().Set("namespace", "default")
	experimentListCmd.Flags().Set("output", "table")

	// Execute command
	apiClient := client.NewAPIClient(server.URL, "test-token")
	RunWithClient(experimentListCmd, cmd, apiClient, []string{})

	// Check output
	output := buf.String()
	assert.Contains(t, output, "exp-1")
	assert.Contains(t, output, "test-experiment-1")
	assert.Contains(t, output, "running")
	assert.Contains(t, output, "exp-2")
	assert.Contains(t, output, "test-experiment-2")
	assert.Contains(t, output, "completed")
}

func TestExperimentCreateCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/experiments", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req client.CreateExperimentRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		response := client.Experiment{
			ID:                "exp-123",
			Name:              req.Name,
			Description:       req.Description,
			BaselinePipeline:  req.BaselinePipeline,
			CandidatePipeline: req.CandidatePipeline,
			Status:            "created",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create command with output buffer
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Set required flags
	experimentCreateCmd.Flags().String("name", "", "")
	experimentCreateCmd.Flags().String("namespace", "", "")
	experimentCreateCmd.Flags().String("pipeline-a", "", "")
	experimentCreateCmd.Flags().String("pipeline-b", "", "")
	experimentCreateCmd.Flags().String("traffic-split", "", "")
	experimentCreateCmd.Flags().String("duration", "", "")
	experimentCreateCmd.Flags().String("selector", "", "")
	experimentCreateCmd.Flags().Set("name", "test-experiment")
	experimentCreateCmd.Flags().Set("namespace", "default")
	experimentCreateCmd.Flags().Set("pipeline-a", "baseline")
	experimentCreateCmd.Flags().Set("pipeline-b", "optimized")
	experimentCreateCmd.Flags().Set("traffic-split", "50/50")
	experimentCreateCmd.Flags().Set("duration", "1h")
	experimentCreateCmd.Flags().Set("selector", "app=test")

	// Execute command
	apiClient := client.NewAPIClient(server.URL, "test-token")
	RunWithClient(experimentCreateCmd, cmd, apiClient, []string{})

	// Check output
	output := buf.String()
	assert.Contains(t, output, "Experiment created successfully")
	assert.Contains(t, output, "exp-123")
}

func TestExperimentStatusCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/experiments/exp-123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		startedAt := time.Now().Add(-25 * time.Minute)
		response := client.Experiment{
			ID:                "exp-123",
			Name:              "test-experiment",
			BaselinePipeline:  "process-baseline-v1",
			CandidatePipeline: "process-optimized-v1",
			Status:            "running",
			CreatedAt:         time.Now().Add(-30 * time.Minute),
			UpdatedAt:         time.Now(),
			StartedAt:         &startedAt,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create command with output buffer
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute command
	apiClient := client.NewAPIClient(server.URL, "test-token")
	RunWithClient(experimentStatusCmd, cmd, apiClient, []string{"exp-123"})

	// Check output
	output := buf.String()
	assert.Contains(t, output, "exp-123")
	assert.Contains(t, output, "test-experiment")
	assert.Contains(t, output, "running")
}

func TestExperimentMetricsCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/experiments/exp-123/metrics", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := client.ExperimentMetrics{
			ExperimentID: "exp-123",
			Summary: client.MetricsSummary{
				CostReductionPercent:    35.5,
				DataLossPercent:         0.8,
				ProgressPercent:         75,
				EstimatedMonthlySavings: 1500.50,
				DataProcessedGB:         1024.5,
				ActiveCollectors:        10,
			},
			PipelineA: client.PipelineMetrics{
				DataPointsPerSecond: 10000,
				BytesPerSecond:      1048576,
				ErrorRate:           0.001,
				P50Latency:          10,
				P95Latency:          25,
				P99Latency:          50,
				Timestamp:           time.Now(),
			},
			PipelineB: client.PipelineMetrics{
				DataPointsPerSecond: 6500,
				BytesPerSecond:      682000,
				ErrorRate:           0.0008,
				P50Latency:          8,
				P95Latency:          20,
				P99Latency:          40,
				Timestamp:           time.Now(),
			},
			Timestamp: time.Now(),
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create command with output buffer
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute command
	apiClient := client.NewAPIClient(server.URL, "test-token")
	RunWithClient(experimentMetricsCmd, cmd, apiClient, []string{"exp-123"})

	// Check output contains key metrics
	output := buf.String()
	assert.Contains(t, output, "Summary Metrics")
	assert.Contains(t, output, "35.50%")    // Cost reduction
	assert.Contains(t, output, "0.80%")     // Data loss
	assert.Contains(t, output, "$1,500.50") // Monthly savings
	assert.Contains(t, output, "Pipeline Comparison")
	assert.Contains(t, output, "10,000") // Pipeline A data points
	assert.Contains(t, output, "6,500")  // Pipeline B data points
}

// validateTrafficSplit validates traffic split parameters
func validateTrafficSplit(split string) (a int, b int, err error) {
	_, err = fmt.Sscanf(split, "%d/%d", &a, &b)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid traffic split format, expected format: A/B (e.g., 50/50)")
	}

	if a < 0 || b < 0 {
		return 0, 0, fmt.Errorf("traffic split percentages cannot be negative")
	}

	if a > 100 || b > 100 {
		return 0, 0, fmt.Errorf("traffic split percentages cannot exceed 100")
	}

	if a+b != 100 {
		return 0, 0, fmt.Errorf("traffic split percentages must add up to 100")
	}

	return a, b, nil
}
