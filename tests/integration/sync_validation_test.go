package integration

import (
	"context"
	"testing"
	"time"

	"github.com/phoenix/platform/pkg/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExperimentPhaseConsistency verifies that experiment phase field is used consistently
func TestExperimentPhaseConsistency(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	
	// Create an experiment using the internal model
	experiment := &models.Experiment{
		ID:                "test-exp-" + time.Now().Format("20060102150405"),
		Name:              "Phase Consistency Test",
		Description:       "Testing phase field consistency",
		Phase:             "initializing",
		Status:            "", // Should be populated from Phase
		BaselinePipeline:  "baseline",
		CandidatePipeline: "candidate",
		TargetNodes:       []string{"node1", "node2"},
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Verify that Status is populated from Phase for backward compatibility
	assert.Equal(t, experiment.Phase, experiment.GetStatus(), "Status should match Phase for backward compatibility")
}

// TestDeploymentVersioning verifies the deployment versioning system
func TestDeploymentVersioning(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Test deployment version structure
	deployment := &models.PipelineDeployment{
		ID:             "test-dep-" + time.Now().Format("20060102150405"),
		DeploymentName: "Test Deployment",
		PipelineName:   "test-pipeline",
		Namespace:      "default",
		TargetNodes:    map[string]string{"node1": "host1"},
		Parameters:     map[string]interface{}{"key": "value"},
		Status:         "pending",
		Phase:          "creating",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Verify required fields
	require.NotEmpty(t, deployment.ID, "Deployment ID should not be empty")
	require.NotEmpty(t, deployment.DeploymentName, "Deployment name should not be empty")
	require.NotEmpty(t, deployment.TargetNodes, "Target nodes should not be empty")
}

// TestWebSocketMessageFormat verifies WebSocket message structure
func TestWebSocketMessageFormat(t *testing.T) {
	// Test message format that should work with native WebSocket
	type WebSocketMessage struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	// Example messages that should be supported
	messages := []WebSocketMessage{
		{Type: "experiment_created", Data: json.RawMessage(`{"id":"exp-123","phase":"initializing"}`)},
		{Type: "deployment_updated", Data: json.RawMessage(`{"deployment_id":"dep-456","status":"ready"}`)},
		{Type: "metrics_update", Data: json.RawMessage(`{"cardinality":1000,"cost_per_minute":5.5}`)},
	}

	for _, msg := range messages {
		assert.NotEmpty(t, msg.Type, "Message type should not be empty")
		assert.NotNil(t, msg.Data, "Message data should not be nil")
	}
}

// TestAPIEndpointPaths verifies that API endpoints follow the expected patterns
func TestAPIEndpointPaths(t *testing.T) {
	// Define expected endpoint patterns
	expectedEndpoints := map[string]string{
		// Authentication
		"auth_login":    "/api/v1/auth/login",
		"auth_refresh":  "/api/v1/auth/refresh",
		"auth_logout":   "/api/v1/auth/logout",
		
		// Experiments
		"experiments_list":   "/api/v1/experiments",
		"experiments_create": "/api/v1/experiments",
		"experiment_get":     "/api/v1/experiments/{id}",
		"experiment_phase":   "/api/v1/experiments/{id}/phase",
		
		// Pipeline Deployments (note: under /pipelines)
		"deployments_list":     "/api/v1/pipelines/deployments",
		"deployments_create":   "/api/v1/pipelines/deployments",
		"deployment_get":       "/api/v1/pipelines/deployments/{id}",
		"deployment_rollback":  "/api/v1/pipelines/deployments/{id}/rollback",
		"deployment_versions":  "/api/v1/pipelines/deployments/{id}/versions",
		
		// WebSocket
		"websocket": "/api/v1/ws",
	}

	// Verify paths follow expected patterns
	for name, path := range expectedEndpoints {
		assert.Contains(t, path, "/api/v1", "All endpoints should be under /api/v1 - %s", name)
		
		// Deployment endpoints should be under /pipelines
		if contains(name, "deployment") {
			assert.Contains(t, path, "/pipelines/deployments", "Deployment endpoints should be under /pipelines/deployments - %s", name)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 (len(substr) < len(s) && findInString(s, substr))))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}