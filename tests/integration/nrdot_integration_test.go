package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNRDOTParameterFlow tests the complete parameter flow for NRDOT integration
func TestNRDOTParameterFlow(t *testing.T) {
	ctx := context.Background()
	
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test environment
	apiURL := setupTestEnvironment(t)

	// Test 1: Create experiment with NRDOT parameters
	t.Run("CreateExperimentWithNRDOT", func(t *testing.T) {
		experiment := map[string]interface{}{
			"name":        "NRDOT Integration Test",
			"description": "Test NRDOT parameter flow",
			"parameters": map[string]interface{}{
				"baseline_pipeline":      "baseline",
				"candidate_pipeline":     "nrdot-cardinality",
				"collector_type":        "nrdot",
				"nr_license_key":        "test-license-key",
				"nr_otlp_endpoint":      "test.endpoint:4317",
				"max_cardinality":       5000,
				"reduction_percentage":  70,
			},
		}

		resp, err := createExperiment(apiURL, experiment)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var created map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&created)
		require.NoError(t, err)
		
		// Verify NRDOT parameters are stored in metadata
		metadata := created["metadata"].(map[string]interface{})
		assert.Equal(t, "nrdot", metadata["collector_type"])
		assert.Equal(t, "test-license-key", metadata["nr_license_key"])
		assert.Equal(t, "test.endpoint:4317", metadata["nr_otlp_endpoint"])
		assert.Equal(t, float64(5000), metadata["max_cardinality"])
		assert.Equal(t, float64(70), metadata["reduction_percentage"])
	})

	// Test 2: Verify task contains NRDOT parameters
	t.Run("TaskContainsNRDOTParams", func(t *testing.T) {
		// Create and start experiment
		expID := createAndStartExperiment(t, apiURL, "nrdot-cardinality")
		
		// Poll for task as agent
		task := pollForTask(t, apiURL, "test-agent-001")
		require.NotNil(t, task)
		
		// Verify task config contains NRDOT parameters
		config := task["config"].(map[string]interface{})
		assert.Equal(t, "nrdot", config["collector_type"])
		assert.Equal(t, "test-license-key", config["nr_license_key"])
		assert.Equal(t, "test.endpoint:4317", config["nr_otlp_endpoint"])
		assert.Equal(t, float64(5000), config["max_cardinality"])
		assert.Equal(t, float64(70), config["reduction_percentage"])
	})

	// Test 3: Validate pipeline template rendering with NRDOT
	t.Run("NRDOTPipelineTemplateRendering", func(t *testing.T) {
		payload := map[string]interface{}{
			"template": "nrdot-cardinality",
			"parameters": map[string]interface{}{
				"nr_license_key":        "test-key",
				"nr_otlp_endpoint":      "test.endpoint:4317",
				"max_cardinality":       10000,
				"reduction_percentage":  80,
			},
		}

		resp, err := renderPipeline(apiURL, payload)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		config := result["config"].(string)
		assert.Contains(t, config, "otlp/newrelic")
		assert.Contains(t, config, "newrelic/cardinality")
		assert.Contains(t, config, "api-key: test-key")
		assert.Contains(t, config, "endpoint: test.endpoint:4317")
	})

	// Test 4: CLI-style experiment creation
	t.Run("CLIStyleExperimentCreation", func(t *testing.T) {
		// Simulate CLI request format
		payload := map[string]interface{}{
			"name": "CLI NRDOT Test",
			"baseline_pipeline": "baseline",
			"candidate_pipeline": "nrdot-cardinality",
			"parameters": map[string]interface{}{
				"use_nrdot":             true,
				"nr_license_key":        "cli-test-key",
				"nr_otlp_endpoint":      "cli.endpoint:4317",
				"max_cardinality":       3000,
				"reduction_percentage":  60,
			},
		}

		resp, err := createExperiment(apiURL, payload)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var created map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&created)
		require.NoError(t, err)

		// Verify parameters are properly stored
		metadata := created["metadata"].(map[string]interface{})
		assert.Equal(t, true, metadata["use_nrdot"])
		assert.Equal(t, "cli-test-key", metadata["nr_license_key"])
	})

	// Test 5: Agent heartbeat with NRDOT info
	t.Run("AgentHeartbeatWithNRDOT", func(t *testing.T) {
		heartbeat := map[string]interface{}{
			"status": "healthy",
			"collector_info": map[string]interface{}{
				"type":    "nrdot",
				"version": "1.0.0",
				"status":  "running",
			},
		}

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/agent/heartbeat", apiURL), jsonBody(heartbeat))
		req.Header.Set("X-Agent-Host-ID", "test-agent-nrdot")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestNRDOTCollectorValidation tests NRDOT-specific validations
func TestNRDOTCollectorValidation(t *testing.T) {
	ctx := context.Background()
	
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	apiURL := setupTestEnvironment(t)

	t.Run("RequireLicenseKeyForNRDOT", func(t *testing.T) {
		// Try to create experiment without license key
		experiment := map[string]interface{}{
			"name": "Invalid NRDOT Test",
			"parameters": map[string]interface{}{
				"baseline_pipeline":  "baseline",
				"candidate_pipeline": "nrdot-cardinality",
				"collector_type":     "nrdot",
				// Missing nr_license_key
			},
		}

		resp, err := createExperiment(apiURL, experiment)
		require.NoError(t, err)
		
		// Should succeed at creation but fail when agent tries to start collector
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("ValidateNRDOTEndpoint", func(t *testing.T) {
		validEndpoints := []string{
			"otlp.nr-data.net:4317",
			"otlp.eu01.nr-data.net:4317",
			"custom.endpoint:4317",
			"internal-collector:4317",
		}

		for _, endpoint := range validEndpoints {
			t.Run(endpoint, func(t *testing.T) {
				payload := map[string]interface{}{
					"template": "nrdot-baseline",
					"parameters": map[string]interface{}{
						"nr_license_key":   "test-key",
						"nr_otlp_endpoint": endpoint,
					},
				}

				resp, err := renderPipeline(apiURL, payload)
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			})
		}
	})
}

// TestNRDOTMetricsFlow tests metrics reporting with NRDOT
func TestNRDOTMetricsFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	apiURL := setupTestEnvironment(t)

	t.Run("ReportMetricsWithNRDOT", func(t *testing.T) {
		// Create experiment
		expID := createAndStartExperiment(t, apiURL, "nrdot-cardinality")

		// Report metrics as NRDOT agent
		metrics := map[string]interface{}{
			"experiment_id": expID,
			"variant":       "candidate",
			"metrics": map[string]interface{}{
				"cardinality_before":     10000,
				"cardinality_after":      3000,
				"reduction_percentage":   70,
				"metrics_per_second":     5000,
				"dropped_metrics":        7000,
				"collector_type":         "nrdot",
			},
		}

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/metrics", apiURL), jsonBody(metrics))
		req.Header.Set("X-Agent-Host-ID", "test-agent-nrdot")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify metrics are stored
		metricsResp, err := http.Get(fmt.Sprintf("%s/api/v1/experiments/%s/metrics", apiURL, expID))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, metricsResp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(metricsResp.Body).Decode(&result)
		require.NoError(t, err)

		// Verify NRDOT metrics are present
		candidateMetrics := result["candidate"].(map[string]interface{})
		assert.Equal(t, float64(70), candidateMetrics["reduction_percentage"])
		assert.Equal(t, "nrdot", candidateMetrics["collector_type"])
	})
}

// Helper function to create and start an experiment with NRDOT
func createAndStartExperiment(t *testing.T, apiURL, pipeline string) string {
	experiment := map[string]interface{}{
		"name": fmt.Sprintf("Test NRDOT %d", time.Now().Unix()),
		"parameters": map[string]interface{}{
			"baseline_pipeline":     "baseline",
			"candidate_pipeline":    pipeline,
			"collector_type":       "nrdot",
			"nr_license_key":       "test-key",
			"nr_otlp_endpoint":     "test.endpoint:4317",
			"max_cardinality":      5000,
			"reduction_percentage": 70,
		},
	}

	resp, err := createExperiment(apiURL, experiment)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&created)
	require.NoError(t, err)

	expID := created["id"].(string)

	// Start the experiment
	startResp, err := http.Post(
		fmt.Sprintf("%s/api/v1/experiments/%s/start", apiURL, expID),
		"application/json",
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, startResp.StatusCode)

	return expID
}

// Helper function to render a pipeline
func renderPipeline(apiURL string, payload interface{}) (*http.Response, error) {
	return http.Post(
		fmt.Sprintf("%s/api/v1/pipelines/render", apiURL),
		"application/json",
		jsonBody(payload),
	)
}