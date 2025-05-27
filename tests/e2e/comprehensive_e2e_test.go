// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComprehensiveE2E performs a comprehensive end-to-end test of the Phoenix platform
func TestComprehensiveE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive E2E test in short mode")
	}

	// Setup test context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Get API URL from environment or use default
	apiURL := getEnvOrDefault("PHOENIX_API_URL", "http://localhost:8080")
	wsURL := "ws://localhost:8080/ws"
	
	t.Logf("🔗 Testing Phoenix API at %s", apiURL)
	
	// Wait for services to be ready
	waitForService(t, apiURL+"/health", 30*time.Second)

	// Run test suites
	t.Run("ExperimentLifecycle", func(t *testing.T) {
		testExperimentLifecycle(t, ctx, apiURL, wsURL)
	})

	t.Run("PipelineDeployment", func(t *testing.T) {
		testPipelineDeployment(t, ctx, apiURL)
	})

	t.Run("AgentIntegration", func(t *testing.T) {
		testAgentIntegration(t, ctx, apiURL)
	})

	t.Run("MetricsAndAnalysis", func(t *testing.T) {
		testMetricsAndAnalysis(t, ctx, apiURL)
	})

	t.Run("WebSocketRealTime", func(t *testing.T) {
		testWebSocketRealTime(t, ctx, wsURL)
	})

	t.Run("CostAnalysis", func(t *testing.T) {
		testCostAnalysis(t, ctx, apiURL)
	})

	t.Run("LoadAndStress", func(t *testing.T) {
		testLoadAndStress(t, ctx, apiURL)
	})
}

// testExperimentLifecycle tests the complete experiment workflow
func testExperimentLifecycle(t *testing.T, ctx context.Context, apiURL, wsURL string) {
	// Create experiment
	experiment := createExperiment(t, apiURL, map[string]interface{}{
		"name":        "E2E Test Experiment Lifecycle",
		"description": "Testing complete experiment lifecycle",
		"config": map[string]interface{}{
			"baseline_pipeline":  "process-baseline-v1",
			"candidate_pipeline": "process-adaptive-filter-v1",
			"duration":           "5m",
			"traffic_split": map[string]int{
				"baseline":  50,
				"candidate": 50,
			},
			"target_cardinality_reduction": 70,
		},
	})

	experimentID := experiment["id"].(string)
	t.Logf("✅ Created experiment: %s", experimentID)

	// Connect WebSocket for real-time updates
	ws := connectWebSocket(t, wsURL)
	defer ws.Close()

	// Subscribe to experiment updates
	subscribeToExperiment(t, ws, experimentID)

	// Start experiment
	startExperiment(t, apiURL, experimentID)

	// Monitor experiment progress
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorExperimentProgress(t, ws, experimentID)
	}()

	// Verify experiment transitions through states
	expectedStates := []string{"pending", "initializing", "running", "analyzing", "completed"}
	for _, expectedState := range expectedStates {
		waitForExperimentState(t, apiURL, experimentID, expectedState, 2*time.Minute)
		t.Logf("✅ Experiment reached state: %s", expectedState)
	}

	// Get final experiment results
	results := getExperimentResults(t, apiURL, experimentID)
	
	// Verify KPIs
	assert.GreaterOrEqual(t, results["cardinality_reduction"].(float64), 50.0)
	assert.GreaterOrEqual(t, results["cost_reduction"].(float64), 40.0)
	assert.GreaterOrEqual(t, results["data_accuracy"].(float64), 95.0)

	// Stop monitoring
	ws.Close()
	wg.Wait()

	// Cleanup: Delete experiment
	deleteExperiment(t, apiURL, experimentID)
}

// testPipelineDeployment tests pipeline deployment functionality
func testPipelineDeployment(t *testing.T, ctx context.Context, apiURL string) {
	// List available pipelines
	pipelines := listPipelines(t, apiURL)
	assert.GreaterOrEqual(t, len(pipelines), 5, "Should have at least 5 pipeline templates")

	// Create a deployment
	deployment := createPipelineDeployment(t, apiURL, map[string]interface{}{
		"pipeline_id": pipelines[0]["id"],
		"name":        "E2E Test Deployment",
		"variant":     "test",
		"config": map[string]interface{}{
			"target_nodes": []string{"test-node-1", "test-node-2"},
			"variables": map[string]string{
				"METRICS_ENDPOINT": "http://metrics.test:9090",
			},
		},
	})

	deploymentID := deployment["id"].(string)
	t.Logf("✅ Created deployment: %s", deploymentID)

	// Validate deployment
	validateResult := validatePipelineDeployment(t, apiURL, deploymentID)
	assert.Equal(t, "valid", validateResult["status"])

	// Get deployment config
	config := getDeploymentConfig(t, apiURL, deploymentID)
	assert.Contains(t, config, "receivers:")
	assert.Contains(t, config, "processors:")
	assert.Contains(t, config, "exporters:")

	// Update deployment status
	updateDeploymentStatus(t, apiURL, deploymentID, "deployed")

	// List deployments
	deployments := listDeployments(t, apiURL)
	found := false
	for _, d := range deployments {
		if d["id"] == deploymentID {
			found = true
			assert.Equal(t, "deployed", d["status"])
			break
		}
	}
	assert.True(t, found, "Deployment should be in list")

	// Rollback deployment
	rollbackDeployment(t, apiURL, deploymentID)
}

// testAgentIntegration tests agent task polling and execution
func testAgentIntegration(t *testing.T, ctx context.Context, apiURL string) {
	hostID := "e2e-test-agent-1"

	// Simulate agent heartbeat
	sendAgentHeartbeat(t, apiURL, hostID)

	// Poll for tasks
	tasks := pollAgentTasks(t, apiURL, hostID)
	if len(tasks) > 0 {
		taskID := tasks[0]["id"].(string)
		t.Logf("✅ Received task: %s", taskID)

		// Update task status
		updateTaskStatus(t, apiURL, taskID, "running")
		
		// Simulate task execution
		time.Sleep(2 * time.Second)
		
		// Complete task
		updateTaskStatus(t, apiURL, taskID, "completed")
	}

	// Send metrics
	sendAgentMetrics(t, apiURL, hostID, []map[string]interface{}{
		{
			"name":      "agent.cpu.percent",
			"value":     45.5,
			"timestamp": time.Now().Unix(),
			"labels": map[string]string{
				"host_id": hostID,
			},
		},
		{
			"name":      "agent.memory.used_bytes",
			"value":     2147483648, // 2GB
			"timestamp": time.Now().Unix(),
			"labels": map[string]string{
				"host_id": hostID,
			},
		},
	})

	// Verify agent appears in fleet status
	fleetStatus := getFleetStatus(t, apiURL)
	found := false
	for _, agent := range fleetStatus["agents"].([]interface{}) {
		if agent.(map[string]interface{})["host_id"] == hostID {
			found = true
			break
		}
	}
	assert.True(t, found, "Agent should appear in fleet status")
}

// testMetricsAndAnalysis tests metrics collection and analysis
func testMetricsAndAnalysis(t *testing.T, ctx context.Context, apiURL string) {
	// Create an experiment for metrics testing
	experiment := createExperiment(t, apiURL, map[string]interface{}{
		"name":        "E2E Metrics Test",
		"description": "Testing metrics and analysis",
		"config": map[string]interface{}{
			"baseline_pipeline":  "process-baseline-v1",
			"candidate_pipeline": "process-topk-v1",
			"duration":           "2m",
		},
	})

	experimentID := experiment["id"].(string)
	
	// Start experiment
	startExperiment(t, apiURL, experimentID)
	
	// Wait for running state
	waitForExperimentState(t, apiURL, experimentID, "running", 1*time.Minute)
	
	// Send test metrics
	sendTestMetrics(t, apiURL, experimentID)
	
	// Wait a bit for metrics to be processed
	time.Sleep(10 * time.Second)
	
	// Get experiment metrics
	metrics := getExperimentMetrics(t, apiURL, experimentID)
	assert.NotEmpty(t, metrics["baseline"])
	assert.NotEmpty(t, metrics["candidate"])
	
	// Get KPI analysis
	analysis := getExperimentAnalysis(t, apiURL, experimentID)
	assert.NotNil(t, analysis["cardinality_reduction"])
	assert.NotNil(t, analysis["cost_reduction"])
	assert.NotNil(t, analysis["cpu_usage"])
	assert.NotNil(t, analysis["memory_usage"])
	
	// Stop experiment
	stopExperiment(t, apiURL, experimentID)
}

// testWebSocketRealTime tests real-time WebSocket functionality
func testWebSocketRealTime(t *testing.T, ctx context.Context, wsURL string) {
	// Connect to WebSocket
	ws := connectWebSocket(t, wsURL)
	defer ws.Close()

	// Test various subscriptions
	subscriptions := []map[string]interface{}{
		{
			"action": "subscribe",
			"topic":  "experiments",
		},
		{
			"action": "subscribe", 
			"topic":  "metrics",
		},
		{
			"action": "subscribe",
			"topic":  "agents",
		},
	}

	for _, sub := range subscriptions {
		err := ws.WriteJSON(sub)
		require.NoError(t, err)
		t.Logf("✅ Subscribed to topic: %s", sub["topic"])
	}

	// Read messages for 5 seconds
	done := make(chan bool)
	messageCount := 0
	
	go func() {
		for {
			var msg map[string]interface{}
			err := ws.ReadJSON(&msg)
			if err != nil {
				close(done)
				return
			}
			messageCount++
			t.Logf("📨 Received WebSocket message: %v", msg["type"])
		}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}

	assert.Greater(t, messageCount, 0, "Should have received some WebSocket messages")
}

// testCostAnalysis tests cost calculation and analysis features
func testCostAnalysis(t *testing.T, ctx context.Context, apiURL string) {
	// Get real-time cost flow
	costFlow := getCostFlow(t, apiURL)
	assert.NotNil(t, costFlow["total_cost_per_minute"])
	assert.NotNil(t, costFlow["top_metrics"])
	assert.NotNil(t, costFlow["by_service"])
	
	// Get cardinality trends
	trends := getCardinalityTrends(t, apiURL)
	assert.NotEmpty(t, trends["data"])
	
	// Create experiment and calculate cost savings
	experiment := createExperiment(t, apiURL, map[string]interface{}{
		"name":        "E2E Cost Analysis Test",
		"description": "Testing cost analysis",
		"config": map[string]interface{}{
			"baseline_pipeline":  "process-baseline-v1",
			"candidate_pipeline": "process-adaptive-filter-v1",
			"duration":           "1m",
		},
	})
	
	experimentID := experiment["id"].(string)
	
	// Get cost analysis for experiment
	costAnalysis := getExperimentCostAnalysis(t, apiURL, experimentID)
	assert.NotNil(t, costAnalysis["baseline_cost"])
	assert.NotNil(t, costAnalysis["candidate_cost"])
	assert.NotNil(t, costAnalysis["monthly_savings"])
	assert.NotNil(t, costAnalysis["yearly_savings"])
	assert.NotNil(t, costAnalysis["recommendations"])
}

// testLoadAndStress performs load and stress testing
func testLoadAndStress(t *testing.T, ctx context.Context, apiURL string) {
	// Create multiple experiments concurrently
	var wg sync.WaitGroup
	experimentCount := 10
	
	for i := 0; i < experimentCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			experiment := createExperiment(t, apiURL, map[string]interface{}{
				"name":        fmt.Sprintf("Load Test Experiment %d", index),
				"description": "Load testing",
				"config": map[string]interface{}{
					"baseline_pipeline":  "process-baseline-v1",
					"candidate_pipeline": "process-topk-v1",
					"duration":           "30s",
				},
			})
			
			t.Logf("✅ Created load test experiment %d: %s", index, experiment["id"])
		}(i)
	}
	
	wg.Wait()
	
	// Send high volume of metrics
	metricsBatch := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		metricsBatch[i] = map[string]interface{}{
			"name":      fmt.Sprintf("load_test_metric_%d", i%100),
			"value":     float64(i),
			"timestamp": time.Now().Unix(),
			"labels": map[string]string{
				"test":      "load",
				"iteration": fmt.Sprintf("%d", i),
			},
		}
	}
	
	// Send metrics in parallel
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(batch int) {
			defer wg.Done()
			sendAgentMetrics(t, apiURL, fmt.Sprintf("load-test-agent-%d", batch), metricsBatch[batch*100:(batch+1)*100])
		}(i)
	}
	
	wg.Wait()
	
	// Verify system is still responsive
	health := checkHealth(t, apiURL)
	assert.Equal(t, "healthy", health["status"])
}

// Helper functions

func waitForService(t *testing.T, healthURL string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		resp, err := http.Get(healthURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			t.Logf("✅ Service is ready")
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	
	t.Fatalf("Service did not become ready within %v", timeout)
}

func createExperiment(t *testing.T, apiURL string, experiment map[string]interface{}) map[string]interface{} {
	body, _ := json.Marshal(experiment)
	resp, err := http.Post(apiURL+"/api/v1/experiments", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result
}

func startExperiment(t *testing.T, apiURL, experimentID string) {
	resp, err := http.Post(apiURL+"/api/v1/experiments/"+experimentID+"/start", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func stopExperiment(t *testing.T, apiURL, experimentID string) {
	resp, err := http.Post(apiURL+"/api/v1/experiments/"+experimentID+"/stop", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func deleteExperiment(t *testing.T, apiURL, experimentID string) {
	req, _ := http.NewRequest("DELETE", apiURL+"/api/v1/experiments/"+experimentID, nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func waitForExperimentState(t *testing.T, apiURL, experimentID, targetState string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		resp, err := http.Get(apiURL + "/api/v1/experiments/" + experimentID)
		require.NoError(t, err)
		
		var experiment map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&experiment)
		resp.Body.Close()
		require.NoError(t, err)
		
		if experiment["status"] == targetState {
			return
		}
		
		time.Sleep(2 * time.Second)
	}
	
	t.Fatalf("Experiment did not reach state %s within %v", targetState, timeout)
}

func getExperimentResults(t *testing.T, apiURL, experimentID string) map[string]interface{} {
	resp, err := http.Get(apiURL + "/api/v1/experiments/" + experimentID + "/analysis")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var results map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	require.NoError(t, err)
	
	return results
}

func connectWebSocket(t *testing.T, wsURL string) *websocket.Conn {
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	return ws
}

func subscribeToExperiment(t *testing.T, ws *websocket.Conn, experimentID string) {
	msg := map[string]interface{}{
		"action": "subscribe",
		"topic":  "experiment",
		"id":     experimentID,
	}
	err := ws.WriteJSON(msg)
	require.NoError(t, err)
}

func monitorExperimentProgress(t *testing.T, ws *websocket.Conn, experimentID string) {
	for {
		var msg map[string]interface{}
		err := ws.ReadJSON(&msg)
		if err != nil {
			return
		}
		
		if msg["experiment_id"] == experimentID {
			t.Logf("📊 Experiment update: %v", msg)
		}
	}
}

func listPipelines(t *testing.T, apiURL string) []map[string]interface{} {
	resp, err := http.Get(apiURL + "/api/v1/pipelines")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result["pipelines"].([]map[string]interface{})
}

func createPipelineDeployment(t *testing.T, apiURL string, deployment map[string]interface{}) map[string]interface{} {
	body, _ := json.Marshal(deployment)
	resp, err := http.Post(apiURL+"/api/v1/pipelines/deployments", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result
}

func validatePipelineDeployment(t *testing.T, apiURL, deploymentID string) map[string]interface{} {
	resp, err := http.Post(apiURL+"/api/v1/pipelines/deployments/"+deploymentID+"/validate", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result
}

func getDeploymentConfig(t *testing.T, apiURL, deploymentID string) string {
	resp, err := http.Get(apiURL + "/api/v1/pipelines/deployments/" + deploymentID + "/config")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result["config"].(string)
}

func updateDeploymentStatus(t *testing.T, apiURL, deploymentID, status string) {
	body, _ := json.Marshal(map[string]string{"status": status})
	req, _ := http.NewRequest("PATCH", apiURL+"/api/v1/pipelines/deployments/"+deploymentID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func listDeployments(t *testing.T, apiURL string) []map[string]interface{} {
	resp, err := http.Get(apiURL + "/api/v1/pipelines/deployments")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result["deployments"].([]map[string]interface{})
}

func rollbackDeployment(t *testing.T, apiURL, deploymentID string) {
	resp, err := http.Post(apiURL+"/api/v1/pipelines/deployments/"+deploymentID+"/rollback", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func sendAgentHeartbeat(t *testing.T, apiURL, hostID string) {
	body, _ := json.Marshal(map[string]interface{}{
		"host_id":    hostID,
		"version":    "1.0.0",
		"status":     "healthy",
		"cpu_cores":  4,
		"memory_gb":  8,
		"disk_gb":    100,
		"os":         "linux",
		"arch":       "amd64",
	})
	
	req, _ := http.NewRequest("POST", apiURL+"/api/v1/agent/heartbeat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Host-ID", hostID)
	
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func pollAgentTasks(t *testing.T, apiURL, hostID string) []map[string]interface{} {
	req, _ := http.NewRequest("GET", apiURL+"/api/v1/agent/tasks", nil)
	req.Header.Set("X-Agent-Host-ID", hostID)
	
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	if tasks, ok := result["tasks"].([]interface{}); ok {
		var taskList []map[string]interface{}
		for _, task := range tasks {
			taskList = append(taskList, task.(map[string]interface{}))
		}
		return taskList
	}
	
	return []map[string]interface{}{}
}

func updateTaskStatus(t *testing.T, apiURL, taskID, status string) {
	body, _ := json.Marshal(map[string]string{
		"status": status,
	})
	
	req, _ := http.NewRequest("PATCH", apiURL+"/api/v1/agent/tasks/"+taskID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func sendAgentMetrics(t *testing.T, apiURL, hostID string, metrics []map[string]interface{}) {
	body, _ := json.Marshal(map[string]interface{}{
		"metrics": metrics,
	})
	
	req, _ := http.NewRequest("POST", apiURL+"/api/v1/agent/metrics", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Host-ID", hostID)
	
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func getFleetStatus(t *testing.T, apiURL string) map[string]interface{} {
	resp, err := http.Get(apiURL + "/api/v1/fleet/status")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result
}

func sendTestMetrics(t *testing.T, apiURL, experimentID string) {
	// Send baseline metrics
	baselineMetrics := []map[string]interface{}{}
	for i := 0; i < 100; i++ {
		baselineMetrics = append(baselineMetrics, map[string]interface{}{
			"name":      fmt.Sprintf("test_metric_%d", i),
			"value":     float64(i * 10),
			"timestamp": time.Now().Unix(),
			"labels": map[string]string{
				"experiment_id": experimentID,
				"variant":       "baseline",
				"service":       fmt.Sprintf("service_%d", i%10),
			},
		})
	}
	sendAgentMetrics(t, apiURL, "test-baseline-agent", baselineMetrics)
	
	// Send candidate metrics (70% less)
	candidateMetrics := []map[string]interface{}{}
	for i := 0; i < 30; i++ {
		candidateMetrics = append(candidateMetrics, map[string]interface{}{
			"name":      fmt.Sprintf("test_metric_%d", i),
			"value":     float64(i * 10),
			"timestamp": time.Now().Unix(),
			"labels": map[string]string{
				"experiment_id": experimentID,
				"variant":       "candidate",
				"service":       fmt.Sprintf("service_%d", i%3),
			},
		})
	}
	sendAgentMetrics(t, apiURL, "test-candidate-agent", candidateMetrics)
}

func getExperimentMetrics(t *testing.T, apiURL, experimentID string) map[string]interface{} {
	resp, err := http.Get(apiURL + "/api/v1/experiments/" + experimentID + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result
}

func getExperimentAnalysis(t *testing.T, apiURL, experimentID string) map[string]interface{} {
	resp, err := http.Get(apiURL + "/api/v1/experiments/" + experimentID + "/analysis")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result
}

func getCostFlow(t *testing.T, apiURL string) map[string]interface{} {
	resp, err := http.Get(apiURL + "/api/v1/cost-flow")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result
}

func getCardinalityTrends(t *testing.T, apiURL string) map[string]interface{} {
	resp, err := http.Get(apiURL + "/api/v1/analytics/cardinality/trends?duration=1h")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result
}

func getExperimentCostAnalysis(t *testing.T, apiURL, experimentID string) map[string]interface{} {
	resp, err := http.Get(apiURL + "/api/v1/experiments/" + experimentID + "/cost-analysis")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result
}

func checkHealth(t *testing.T, apiURL string) map[string]interface{} {
	resp, err := http.Get(apiURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result
}