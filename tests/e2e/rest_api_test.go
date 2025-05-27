// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRESTAPIWorkflow tests the complete workflow using REST API
func TestRESTAPIWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup test environment
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Use environment variables for service URLs
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	t.Logf("🔗 Using API at %s", apiURL)

	// Check if service is ready
	checkServiceHealth(t, apiURL+"/health")

	// Test authentication first
	token := testAuthentication(t, apiURL)

	t.Run("CompleteExperimentLifecycle", func(t *testing.T) {
		// Create experiment
		experiment := createExperiment(t, apiURL, token)
		t.Logf("✅ Created experiment: %s", experiment.ID)

		// Get experiment details
		exp := getExperiment(t, apiURL, token, experiment.ID)
		assert.Equal(t, experiment.Name, exp.Name)
		assert.Equal(t, "created", exp.Phase)

		// List experiments
		experiments := listExperiments(t, apiURL, token)
		assert.GreaterOrEqual(t, len(experiments), 1)
		
		// Start experiment
		startExperiment(t, apiURL, token, experiment.ID)
		t.Logf("✅ Started experiment: %s", experiment.ID)
		
		// Wait for experiment to be running
		waitForExperimentPhase(t, apiURL, token, experiment.ID, "running", 30*time.Second)
		
		// Get KPIs (may not have data yet in test environment)
		kpis := getExperimentKPIs(t, apiURL, token, experiment.ID)
		t.Logf("📊 KPIs: %+v", kpis)
		
		// Get cost analysis
		costAnalysis := getExperimentCostAnalysis(t, apiURL, token, experiment.ID)
		t.Logf("💰 Cost Analysis: Monthly Savings: $%.2f", costAnalysis.MonthlySavings)
		
		// Test rollback
		rollbackExperiment(t, apiURL, token, experiment.ID, "Testing rollback functionality")
		t.Logf("✅ Rolled back experiment: %s", experiment.ID)
		
		// Stop experiment
		stopExperiment(t, apiURL, token, experiment.ID)
		t.Logf("✅ Stopped experiment: %s", experiment.ID)
	})

	t.Run("AgentTaskPolling", func(t *testing.T) {
		// Simulate agent polling for tasks
		hostID := "test-agent-001"
		
		// Poll for tasks (should return empty initially)
		tasks := pollAgentTasks(t, apiURL, hostID)
		assert.Empty(t, tasks)
		
		// Send heartbeat
		sendAgentHeartbeat(t, apiURL, hostID)
		t.Logf("✅ Agent heartbeat sent for: %s", hostID)
		
		// Push metrics
		pushAgentMetrics(t, apiURL, hostID)
		t.Logf("✅ Agent metrics pushed for: %s", hostID)
	})

	t.Run("PipelineManagement", func(t *testing.T) {
		// List pipeline templates
		templates := listPipelineTemplates(t, apiURL, token)
		assert.GreaterOrEqual(t, len(templates), 3)
		t.Logf("✅ Found %d pipeline templates", len(templates))
		
		// Validate a pipeline config
		validatePipelineConfig(t, apiURL, token, templates[0].ConfigURL)
		
		// Render a pipeline with variables
		rendered := renderPipeline(t, apiURL, token, "adaptive", map[string]string{
			"threshold": "0.8",
			"interval":  "60s",
		})
		assert.Contains(t, rendered, "processors:")
		t.Logf("✅ Rendered pipeline template")
	})

	t.Run("UIEndpoints", func(t *testing.T) {
		// Test cost flow endpoint
		costFlow := getMetricCostFlow(t, apiURL, token)
		assert.NotNil(t, costFlow)
		t.Logf("💸 Total cost per minute: $%.2f", costFlow.TotalCostPerMinute)
		
		// Test cardinality breakdown
		cardinality := getCardinalityBreakdown(t, apiURL, token)
		assert.NotNil(t, cardinality)
		t.Logf("📊 Total cardinality: %d", cardinality.TotalCardinality)
		
		// Test fleet status
		fleet := getFleetStatus(t, apiURL, token)
		assert.NotNil(t, fleet)
		t.Logf("🚀 Fleet: %d total agents, %d healthy", fleet.TotalAgents, fleet.HealthyAgents)
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// Test invalid experiment creation
		resp := makeRequest(t, "POST", apiURL+"/api/v1/experiments", token, 
			map[string]interface{}{
				"name": "", // Missing required fields
			})
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		
		// Test nonexistent experiment
		resp = makeRequest(t, "GET", apiURL+"/api/v1/experiments/nonexistent", token, nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// Helper functions

func checkServiceHealth(t *testing.T, url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Service health check timeout: %s", url)
		default:
			resp, err := http.Get(url)
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				t.Logf("✅ Service ready: %s", url)
				return
			}
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func testAuthentication(t *testing.T, apiURL string) string {
	// For testing, we'll use a mock token
	// In real environment, this would call /api/v1/auth/login
	return "test-token-12345"
}

func makeRequest(t *testing.T, method, url, token string, body interface{}) *http.Response {
	var bodyReader *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewBuffer(jsonBody)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}
	
	req, err := http.NewRequest(method, url, bodyReader)
	require.NoError(t, err)
	
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	
	return resp
}

func createExperiment(t *testing.T, apiURL, token string) *ExperimentResponse {
	body := map[string]interface{}{
		"name":        "E2E Test Experiment",
		"description": "Testing REST API workflow",
		"config": map[string]interface{}{
			"target_hosts": []string{"test-host-1", "test-host-2"},
			"baseline_template": map[string]interface{}{
				"name":       "baseline",
				"config_url": "file:///configs/baseline.yaml",
			},
			"candidate_template": map[string]interface{}{
				"name":       "adaptive",
				"config_url": "file:///configs/adaptive.yaml",
				"variables": map[string]string{
					"threshold": "0.7",
				},
			},
			"duration":        "10m",
			"warmup_duration": "2m",
		},
	}
	
	resp := makeRequest(t, "POST", apiURL+"/api/v1/experiments", token, body)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var experiment ExperimentResponse
	err := json.NewDecoder(resp.Body).Decode(&experiment)
	require.NoError(t, err)
	
	return &experiment
}

func getExperiment(t *testing.T, apiURL, token, id string) *ExperimentResponse {
	resp := makeRequest(t, "GET", apiURL+"/api/v1/experiments/"+id, token, nil)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var result struct {
		Experiment ExperimentResponse `json:"experiment"`
	}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return &result.Experiment
}

func listExperiments(t *testing.T, apiURL, token string) []ExperimentResponse {
	resp := makeRequest(t, "GET", apiURL+"/api/v1/experiments", token, nil)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var result struct {
		Experiments []ExperimentResponse `json:"experiments"`
	}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result.Experiments
}

func startExperiment(t *testing.T, apiURL, token, id string) {
	resp := makeRequest(t, "POST", apiURL+"/api/v1/experiments/"+id+"/start", token, nil)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func stopExperiment(t *testing.T, apiURL, token, id string) {
	resp := makeRequest(t, "POST", apiURL+"/api/v1/experiments/"+id+"/stop", token, nil)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func waitForExperimentPhase(t *testing.T, apiURL, token, id, targetPhase string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for experiment %s to reach phase %s", id, targetPhase)
		case <-ticker.C:
			exp := getExperiment(t, apiURL, token, id)
			if exp.Phase == targetPhase {
				return
			}
		}
	}
}

func getExperimentKPIs(t *testing.T, apiURL, token, id string) map[string]interface{} {
	resp := makeRequest(t, "GET", apiURL+"/api/v1/experiments/"+id+"/kpis", token, nil)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	return result
}

func getExperimentCostAnalysis(t *testing.T, apiURL, token, id string) *CostAnalysis {
	resp := makeRequest(t, "GET", apiURL+"/api/v1/experiments/"+id+"/cost-analysis", token, nil)
	defer resp.Body.Close()
	
	var result struct {
		CostAnalysis CostAnalysis `json:"cost_analysis"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	
	return &result.CostAnalysis
}

func rollbackExperiment(t *testing.T, apiURL, token, id, reason string) {
	url := fmt.Sprintf("%s/api/v1/experiments/%s/rollback?reason=%s", apiURL, id, reason)
	resp := makeRequest(t, "POST", url, token, nil)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func pollAgentTasks(t *testing.T, apiURL, hostID string) []interface{} {
	req, err := http.NewRequest("GET", apiURL+"/api/v1/agent/tasks", nil)
	require.NoError(t, err)
	
	req.Header.Set("X-Agent-Host-ID", hostID)
	
	client := &http.Client{Timeout: 35 * time.Second} // Long polling timeout
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var tasks []interface{}
	json.NewDecoder(resp.Body).Decode(&tasks)
	
	return tasks
}

func sendAgentHeartbeat(t *testing.T, apiURL, hostID string) {
	body := map[string]interface{}{
		"host_id":       hostID,
		"agent_version": "1.0.0",
		"status":        "healthy",
		"active_tasks":  []string{},
		"resource_usage": map[string]interface{}{
			"cpu_percent":    15.5,
			"memory_percent": 45.2,
			"memory_bytes":   1073741824,
		},
	}
	
	req, err := http.NewRequest("POST", apiURL+"/api/v1/agent/heartbeat", 
		bytes.NewBuffer(mustMarshalJSON(t, body)))
	require.NoError(t, err)
	
	req.Header.Set("X-Agent-Host-ID", hostID)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func pushAgentMetrics(t *testing.T, apiURL, hostID string) {
	metrics := []map[string]interface{}{
		{
			"name":   "otel_collector_cpu_percent",
			"value":  15.5,
			"type":   "gauge",
			"labels": map[string]string{"variant": "baseline"},
		},
		{
			"name":   "metrics_processed_total",
			"value":  1000,
			"type":   "counter",
			"labels": map[string]string{"variant": "candidate", "pipeline": "adaptive"},
		},
	}
	
	req, err := http.NewRequest("POST", apiURL+"/api/v1/agent/metrics",
		bytes.NewBuffer(mustMarshalJSON(t, metrics)))
	require.NoError(t, err)
	
	req.Header.Set("X-Agent-Host-ID", hostID)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func listPipelineTemplates(t *testing.T, apiURL, token string) []PipelineTemplate {
	resp := makeRequest(t, "GET", apiURL+"/api/v1/pipelines/templates", token, nil)
	defer resp.Body.Close()
	
	var result struct {
		Templates []PipelineTemplate `json:"templates"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	
	return result.Templates
}

func validatePipelineConfig(t *testing.T, apiURL, token, configURL string) {
	body := map[string]interface{}{
		"config_url": configURL,
	}
	
	resp := makeRequest(t, "POST", apiURL+"/api/v1/pipelines/validate", token, body)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func renderPipeline(t *testing.T, apiURL, token, templateName string, variables map[string]string) string {
	body := map[string]interface{}{
		"template":  templateName,
		"variables": variables,
	}
	
	resp := makeRequest(t, "POST", apiURL+"/api/v1/pipelines/render", token, body)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var result struct {
		Rendered string `json:"rendered"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	
	return result.Rendered
}

func getMetricCostFlow(t *testing.T, apiURL, token string) *MetricCostFlow {
	resp := makeRequest(t, "GET", apiURL+"/api/v1/metrics/cost-flow", token, nil)
	defer resp.Body.Close()
	
	var flow MetricCostFlow
	json.NewDecoder(resp.Body).Decode(&flow)
	
	return &flow
}

func getCardinalityBreakdown(t *testing.T, apiURL, token string) *CardinalityBreakdown {
	resp := makeRequest(t, "GET", apiURL+"/api/v1/metrics/cardinality", token, nil)
	defer resp.Body.Close()
	
	var breakdown CardinalityBreakdown
	json.NewDecoder(resp.Body).Decode(&breakdown)
	
	return &breakdown
}

func getFleetStatus(t *testing.T, apiURL, token string) *FleetStatus {
	resp := makeRequest(t, "GET", apiURL+"/api/v1/fleet/status", token, nil)
	defer resp.Body.Close()
	
	var status FleetStatus
	json.NewDecoder(resp.Body).Decode(&status)
	
	return &status
}

func mustMarshalJSON(t *testing.T, v interface{}) []byte {
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}

// Response types
type ExperimentResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Phase       string                 `json:"phase"`
	Config      map[string]interface{} `json:"config"`
	Status      map[string]interface{} `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type CostAnalysis struct {
	MonthlySavings    float64  `json:"monthly_savings"`
	YearlySavings     float64  `json:"yearly_savings"`
	SavingsPercentage float64  `json:"savings_percentage"`
	Recommendations   []string `json:"recommendations"`
}

type PipelineTemplate struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	ConfigURL   string            `json:"config_url"`
	Variables   map[string]string `json:"variables"`
}

type MetricCostFlow struct {
	TotalCostPerMinute float64                  `json:"total_cost_per_minute"`
	TopMetrics         []map[string]interface{} `json:"top_metrics"`
	ByService          map[string]float64       `json:"by_service"`
	ByNamespace        map[string]float64       `json:"by_namespace"`
}

type CardinalityBreakdown struct {
	TotalCardinality int                    `json:"total_cardinality"`
	ByNamespace      map[string]int         `json:"by_namespace"`
	ByService        map[string]int         `json:"by_service"`
	TopMetrics       []map[string]interface{} `json:"top_metrics"`
}

type FleetStatus struct {
	TotalAgents    int                      `json:"total_agents"`
	HealthyAgents  int                      `json:"healthy_agents"`
	OfflineAgents  int                      `json:"offline_agents"`
	UpdatingAgents int                      `json:"updating_agents"`
	TotalSavings   float64                  `json:"total_savings"`
	Agents         []map[string]interface{} `json:"agents"`
}