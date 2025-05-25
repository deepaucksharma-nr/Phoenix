package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/phoenix/platform/cmd/phoenix-cli/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
					Namespace: "default",
					Status:    "running",
					CreatedAt: time.Now().Add(-1 * time.Hour),
				},
				{
					ID:        "exp-2",
					Name:      "test-experiment-2",
					Namespace: "default",
					Status:    "completed",
					CreatedAt: time.Now().Add(-2 * time.Hour),
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override API URL for testing
	apiURL = server.URL

	// Create command with output buffer
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Set flags
	experimentListCmd.Flags().Set("namespace", "default")
	experimentListCmd.Flags().Set("output", "table")

	// Mock getAPIClient to return client with test server URL
	oldGetAPIClient := getAPIClient
	getAPIClient = func() (*client.APIClient, error) {
		return client.NewAPIClient(server.URL, "test-token"), nil
	}
	defer func() { getAPIClient = oldGetAPIClient }()

	// Execute command
	experimentListCmd.Run(cmd, []string{})

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
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/experiments", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req client.CreateExperimentRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Validate request
		assert.Equal(t, "test-experiment", req.Name)
		assert.Equal(t, "default", req.Namespace)
		assert.Equal(t, "baseline", req.PipelineA)
		assert.Equal(t, "optimized", req.PipelineB)
		assert.Equal(t, "50/50", req.TrafficSplit)

		response := client.Experiment{
			ID:        "exp-123",
			Name:      req.Name,
			Namespace: req.Namespace,
			Status:    "created",
			CreatedAt: time.Now(),
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
	experimentCreateCmd.Flags().Set("name", "test-experiment")
	experimentCreateCmd.Flags().Set("namespace", "default")
	experimentCreateCmd.Flags().Set("pipeline-a", "baseline")
	experimentCreateCmd.Flags().Set("pipeline-b", "optimized")
	experimentCreateCmd.Flags().Set("traffic-split", "50/50")
	experimentCreateCmd.Flags().Set("duration", "1h")
	experimentCreateCmd.Flags().Set("selector", "app=test")

	// Mock getAPIClient
	oldGetAPIClient := getAPIClient
	getAPIClient = func() (*client.APIClient, error) {
		return client.NewAPIClient(server.URL, "test-token"), nil
	}
	defer func() { getAPIClient = oldGetAPIClient }()

	// Execute command
	experimentCreateCmd.Run(cmd, []string{})

	// Check output
	output := buf.String()
	assert.Contains(t, output, "Experiment created successfully")
	assert.Contains(t, output, "exp-123")
}

func TestExperimentStatusCommand(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/experiments/exp-123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := client.Experiment{
			ID:        "exp-123",
			Name:      "test-experiment",
			Namespace: "default",
			Status:    "running",
			CreatedAt: time.Now().Add(-30 * time.Minute),
			StartedAt: time.Now().Add(-25 * time.Minute),
			PipelineA: client.PipelineInfo{
				Name:     "baseline",
				Template: "process-baseline-v1",
			},
			PipelineB: client.PipelineInfo{
				Name:     "optimized",
				Template: "process-optimized-v1",
			},
			TrafficSplit: client.TrafficSplit{
				PipelineA: 50,
				PipelineB: 50,
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

	// Mock getAPIClient
	oldGetAPIClient := getAPIClient
	getAPIClient = func() (*client.APIClient, error) {
		return client.NewAPIClient(server.URL, "test-token"), nil
	}
	defer func() { getAPIClient = oldGetAPIClient }()

	// Execute command
	experimentStatusCmd.Run(cmd, []string{"exp-123"})

	// Check output
	output := buf.String()
	assert.Contains(t, output, "exp-123")
	assert.Contains(t, output, "test-experiment")
	assert.Contains(t, output, "running")
	assert.Contains(t, output, "baseline")
	assert.Contains(t, output, "optimized")
	assert.Contains(t, output, "50/50")
}

func TestExperimentMetricsCommand(t *testing.T) {
	// Mock server
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
			},
			PipelineB: client.PipelineMetrics{
				DataPointsPerSecond: 6500,
				BytesPerSecond:      682000,
				ErrorRate:           0.0008,
				P50Latency:          8,
				P95Latency:          20,
				P99Latency:          40,
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

	// Mock getAPIClient
	oldGetAPIClient := getAPIClient
	getAPIClient = func() (*client.APIClient, error) {
		return client.NewAPIClient(server.URL, "test-token"), nil
	}
	defer func() { getAPIClient = oldGetAPIClient }()

	// Execute command
	experimentMetricsCmd.Run(cmd, []string{"exp-123"})

	// Check output contains key metrics
	output := buf.String()
	assert.Contains(t, output, "Summary Metrics")
	assert.Contains(t, output, "35.50%") // Cost reduction
	assert.Contains(t, output, "0.80%")  // Data loss
	assert.Contains(t, output, "$1,500.50") // Monthly savings
	assert.Contains(t, output, "Pipeline Comparison")
	assert.Contains(t, output, "10,000") // Pipeline A data points
	assert.Contains(t, output, "6,500")  // Pipeline B data points
}

func TestValidateTrafficSplit(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedA   int
		expectedB   int
		expectError bool
	}{
		{
			name:        "valid 50/50 split",
			input:       "50/50",
			expectedA:   50,
			expectedB:   50,
			expectError: false,
		},
		{
			name:        "valid 80/20 split",
			input:       "80/20",
			expectedA:   80,
			expectedB:   20,
			expectError: false,
		},
		{
			name:        "invalid format",
			input:       "50-50",
			expectError: true,
		},
		{
			name:        "not adding to 100",
			input:       "60/30",
			expectError: true,
		},
		{
			name:        "negative values",
			input:       "-50/150",
			expectError: true,
		},
		{
			name:        "over 100",
			input:       "101/0",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b, err := validateTrafficSplit(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedA, a)
				assert.Equal(t, tt.expectedB, b)
			}
		})
	}
}