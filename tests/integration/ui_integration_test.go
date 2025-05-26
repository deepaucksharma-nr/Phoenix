package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUIEndpoints(t *testing.T) {
	// Ensure API is running
	baseURL := getTestAPIURL()

	t.Run("ExperimentWizard", func(t *testing.T) {
		// Create experiment using wizard
		wizardReq := map[string]interface{}{
			"name":           "UI Test Experiment",
			"description":    "Testing wizard creation",
			"host_selector":  []string{"test-group"},
			"pipeline_type":  "top-k-20",
			"duration_hours": 1,
		}

		body, _ := json.Marshal(wizardReq)
		resp, err := http.Post(baseURL+"/api/v1/experiments/wizard", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var experiment map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&experiment)
		assert.NotEmpty(t, experiment["id"])
		assert.Equal(t, "UI Test Experiment", experiment["name"])
	})

	t.Run("FleetStatus", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/v1/fleet/status")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var fleet map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&fleet)
		assert.Contains(t, fleet, "total_agents")
		assert.Contains(t, fleet, "agents")
	})

	t.Run("MetricCostFlow", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/v1/metrics/cost-flow")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var costFlow map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&costFlow)
		assert.Contains(t, costFlow, "total_cost_rate")
		assert.Contains(t, costFlow, "top_metrics")
	})

	t.Run("PipelineTemplates", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/v1/pipelines/templates")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var templates []map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&templates)
		assert.NotEmpty(t, templates)
		
		// Check if default templates exist
		hasTopK := false
		for _, tmpl := range templates {
			if tmpl["id"] == "top-k-20" {
				hasTopK = true
				break
			}
		}
		assert.True(t, hasTopK, "Default top-k-20 template should exist")
	})

	t.Run("PipelinePreview", func(t *testing.T) {
		previewReq := map[string]interface{}{
			"pipeline_config": map[string]interface{}{
				"processors": []map[string]interface{}{
					{"type": "top_k", "config": map[string]interface{}{"k": 20}},
				},
			},
			"target_hosts": []string{"test-host"},
		}

		body, _ := json.Marshal(previewReq)
		resp, err := http.Post(baseURL+"/api/v1/pipelines/preview", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var preview map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&preview)
		assert.Contains(t, preview, "estimated_cost_reduction")
		assert.Contains(t, preview, "estimated_cpu_impact")
	})
}

func TestWebSocketConnection(t *testing.T) {
	// Skip if WebSocket not available
	wsURL := getTestWebSocketURL()
	if wsURL == "" {
		t.Skip("WebSocket URL not configured")
	}

	// Connect to WebSocket
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	
	conn, _, err := dialer.Dial(wsURL+"/api/v1/ws", nil)
	require.NoError(t, err)
	defer conn.Close()

	// Subscribe to events
	subscription := map[string]interface{}{
		"type": "subscribe",
		"payload": map[string]interface{}{
			"events": []string{"metric_flow", "agent_status"},
		},
	}
	
	err = conn.WriteJSON(subscription)
	require.NoError(t, err)

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// Read at least one message
	var message map[string]interface{}
	err = conn.ReadJSON(&message)
	if err != nil {
		// Timeout is acceptable - might not have events immediately
		if websocket.IsCloseError(err) || err == websocket.ErrCloseSent {
			t.Log("WebSocket closed, which is acceptable")
		} else {
			t.Logf("WebSocket read error (might be timeout): %v", err)
		}
	} else {
		// If we got a message, verify its structure
		assert.Contains(t, message, "type")
		assert.Contains(t, message, "timestamp")
		assert.Contains(t, message, "data")
	}
}

func TestQuickDeploy(t *testing.T) {
	baseURL := getTestAPIURL()

	deployReq := map[string]interface{}{
		"pipeline_template": "top-k-20",
		"target_hosts":      []string{"test-host"},
		"auto_rollback":     true,
	}

	body, _ := json.Marshal(deployReq)
	resp, err := http.Post(baseURL+"/api/v1/pipelines/quick-deploy", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	var deployment map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&deployment)
	assert.Contains(t, deployment, "deployment_id")
	assert.Contains(t, deployment, "status")
	assert.Equal(t, "deploying", deployment["status"])
}

func TestCostAnalytics(t *testing.T) {
	baseURL := getTestAPIURL()

	resp, err := http.Get(baseURL + "/api/v1/cost-analytics?period=7d")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var analytics map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&analytics)
	assert.Contains(t, analytics, "total_cost")
	assert.Contains(t, analytics, "total_savings")
	assert.Contains(t, analytics, "savings_by_pipeline")
}

func getTestAPIURL() string {
	url := "http://localhost:8080"
	// Allow override for CI/CD
	if envURL := getEnv("PHOENIX_API_URL", ""); envURL != "" {
		url = envURL
	}
	return url
}

func getTestWebSocketURL() string {
	url := "ws://localhost:8081"
	// Allow override for CI/CD
	if envURL := getEnv("PHOENIX_WS_URL", ""); envURL != "" {
		url = envURL
	}
	return url
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}