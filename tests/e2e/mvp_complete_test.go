package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMVPCompleteFlow validates the entire Phoenix MVP workflow
func TestMVPCompleteFlow(t *testing.T) {
	// Skip if not in integration mode
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()
	apiURL := getAPIURL()
	
	// Phase 1: Setup and connectivity
	t.Run("Setup", func(t *testing.T) {
		// Verify API is healthy
		resp, err := http.Get(apiURL + "/health")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
		
		// Verify at least one agent is connected
		agents := listAgents(t, apiURL)
		require.NotEmpty(t, agents, "No agents connected")
		t.Logf("Found %d agents", len(agents))
	})

	// Phase 2: Pipeline deployment and rollback
	var deploymentID string
	t.Run("PipelineLifecycle", func(t *testing.T) {
		// Deploy a baseline pipeline
		deployment := deployPipeline(t, apiURL, "baseline", []string{"localhost"})
		deploymentID = deployment["id"].(string)
		
		// Wait for deployment to be active
		waitForDeploymentStatus(t, apiURL, deploymentID, "active", 30*time.Second)
		
		// Verify pipeline is running on agent
		status := getDeploymentStatus(t, apiURL, deploymentID)
		assert.Equal(t, "active", status["status"])
		
		// Rollback the pipeline
		rollbackPipeline(t, apiURL, deploymentID)
		
		// Verify rollback completed
		waitForDeploymentStatus(t, apiURL, deploymentID, "rolled_back", 30*time.Second)
	})

	// Phase 3: Experiment with WebSocket monitoring
	var experimentID string
	t.Run("ExperimentLifecycle", func(t *testing.T) {
		// Connect to WebSocket for real-time events
		ws := connectWebSocket(t, apiURL)
		defer ws.Close()
		
		// Channel to collect events
		events := make(chan map[string]interface{}, 10)
		go collectWebSocketEvents(ws, events)
		
		// Create experiment
		experiment := createExperiment(t, apiURL, map[string]interface{}{
			"name":               "E2E Test Experiment",
			"description":        "Automated E2E validation",
			"baseline_template":  "baseline",
			"candidate_template": "topk",
			"target_hosts":       []string{"localhost"},
			"load_profile":       "normal",
			"duration":           "30s",
		})
		experimentID = experiment["id"].(string)
		
		// Verify experiment_created event
		expectEvent(t, events, "experiment_created", 5*time.Second)
		
		// Start experiment
		startExperiment(t, apiURL, experimentID)
		
		// Verify experiment_started event
		expectEvent(t, events, "experiment_started", 5*time.Second)
		
		// Wait for running status
		waitForExperimentPhase(t, apiURL, experimentID, "running", 30*time.Second)
		
		// Verify metrics are being collected
		time.Sleep(10 * time.Second)
		metrics := getExperimentMetrics(t, apiURL, experimentID)
		assert.NotNil(t, metrics["baseline"])
		assert.NotNil(t, metrics["candidate"])
		
		// Wait for completion
		waitForExperimentPhase(t, apiURL, experimentID, "completed", 60*time.Second)
		
		// Verify experiment_completed event
		expectEvent(t, events, "experiment_completed", 5*time.Second)
		
		// Check results
		experiment = getExperiment(t, apiURL, experimentID)
		results := experiment["results"].(map[string]interface{})
		
		assert.NotNil(t, results)
		assert.Contains(t, results, "cardinality_reduction")
		assert.Contains(t, results, "cost_reduction")
		
		cardinalityReduction := results["cardinality_reduction"].(float64)
		assert.Greater(t, cardinalityReduction, 0.0, "Expected some cardinality reduction")
		
		t.Logf("Experiment completed with %.2f%% cardinality reduction", cardinalityReduction)
	})

	// Phase 4: Stop mid-flight test
	t.Run("ExperimentStopMidFlight", func(t *testing.T) {
		// Create and start experiment
		experiment := createExperiment(t, apiURL, map[string]interface{}{
			"name":               "Stop Test Experiment",
			"baseline_template":  "baseline",
			"candidate_template": "adaptive",
			"target_hosts":       []string{"localhost"},
			"duration":           "5m", // Long duration
		})
		
		expID := experiment["id"].(string)
		startExperiment(t, apiURL, expID)
		
		// Wait for running
		waitForExperimentPhase(t, apiURL, expID, "running", 30*time.Second)
		
		// Stop after 5 seconds
		time.Sleep(5 * time.Second)
		stopExperiment(t, apiURL, expID)
		
		// Verify stopped
		waitForExperimentPhase(t, apiURL, expID, "stopped", 30*time.Second)
		
		// Verify no orphan processes (via agent metrics)
		time.Sleep(5 * time.Second)
		agents := listAgents(t, apiURL)
		for _, agent := range agents {
			metrics := agent["metrics"].(map[string]interface{})
			activeProcesses := metrics["active_processes"].(float64)
			assert.Equal(t, 0.0, activeProcesses, "Agent has orphan processes")
		}
	})

	// Phase 5: Error handling
	t.Run("ErrorHandling", func(t *testing.T) {
		// Try to create experiment with invalid template
		_, err := createExperimentRaw(apiURL, map[string]interface{}{
			"name":               "Invalid Experiment",
			"baseline_template":  "non-existent",
			"candidate_template": "also-invalid",
			"target_hosts":       []string{"localhost"},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "400") // Bad request
		
		// Try to start non-existent experiment
		err = startExperimentRaw(apiURL, "invalid-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404") // Not found
	})

	// Phase 6: Cost and KPI validation
	t.Run("MetricsAccuracy", func(t *testing.T) {
		// Get KPIs from completed experiment
		kpis := getExperimentKPIs(t, apiURL, experimentID)
		assert.NotEmpty(t, kpis)
		
		// Get cost analysis
		costAnalysis := getExperimentCostAnalysis(t, apiURL, experimentID)
		assert.NotEmpty(t, costAnalysis)
		
		// Verify cost calculations make sense
		monthlySavings := costAnalysis["monthly_savings"].(float64)
		savingsPercent := costAnalysis["savings_percentage"].(float64)
		
		assert.GreaterOrEqual(t, savingsPercent, 0.0)
		assert.LessOrEqual(t, savingsPercent, 100.0)
		
		if savingsPercent > 0 {
			assert.Greater(t, monthlySavings, 0.0)
		}
		
		t.Logf("Cost analysis: %.2f%% savings ($%.2f/month)", savingsPercent, monthlySavings)
	})
}

// Helper functions

func getAPIURL() string {
	url := os.Getenv("PHOENIX_API_URL")
	if url == "" {
		return "http://localhost:8080"
	}
	return url
}

func listAgents(t *testing.T, apiURL string) []map[string]interface{} {
	resp, err := http.Get(apiURL + "/api/v1/agents")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var agents []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&agents))
	return agents
}

func deployPipeline(t *testing.T, apiURL string, template string, hosts []string) map[string]interface{} {
	payload := map[string]interface{}{
		"name":         fmt.Sprintf("E2E Pipeline %s", time.Now().Format("15:04:05")),
		"template":     template,
		"target_hosts": hosts,
	}
	
	resp := postJSON(t, apiURL+"/api/v1/deployments", payload)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	
	var deployment map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deployment))
	resp.Body.Close()
	
	return deployment
}

func waitForDeploymentStatus(t *testing.T, apiURL, deploymentID, expectedStatus string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		status := getDeploymentStatus(t, apiURL, deploymentID)
		if status["status"] == expectedStatus {
			return
		}
		time.Sleep(1 * time.Second)
	}
	
	t.Fatalf("Deployment %s did not reach status %s within %v", deploymentID, expectedStatus, timeout)
}

func connectWebSocket(t *testing.T, apiURL string) *websocket.Conn {
	wsURL := "ws" + apiURL[4:] + "/api/v1/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	return ws
}

func collectWebSocketEvents(ws *websocket.Conn, events chan<- map[string]interface{}) {
	for {
		var event map[string]interface{}
		err := ws.ReadJSON(&event)
		if err != nil {
			close(events)
			return
		}
		events <- event
	}
}

func expectEvent(t *testing.T, events <-chan map[string]interface{}, eventType string, timeout time.Duration) {
	select {
	case event := <-events:
		assert.Equal(t, eventType, event["type"])
	case <-time.After(timeout):
		t.Fatalf("Did not receive %s event within %v", eventType, timeout)
	}
}

// Additional helper functions would be implemented similarly...
// (createExperiment, startExperiment, getExperimentMetrics, etc.)