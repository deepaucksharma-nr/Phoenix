// +build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/phoenix-vnext/platform/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// PipelineDeploymentAPITestSuite tests the pipeline deployment API endpoints
type PipelineDeploymentAPITestSuite struct {
	suite.Suite
	apiURL    string
	authToken string
	db        *sql.DB
	client    *http.Client
}

func (s *PipelineDeploymentAPITestSuite) SetupSuite() {
	// API URL
	s.apiURL = os.Getenv("PHOENIX_API_URL")
	if s.apiURL == "" {
		s.apiURL = "http://localhost:8080"
	}

	// HTTP client
	s.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Connect to test database
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://phoenix:phoenix@localhost:5432/phoenix_test?sslmode=disable"
	}
	var err error
	s.db, err = sql.Open("postgres", dbURL)
	require.NoError(s.T(), err, "Failed to connect to test database")

	// Get auth token
	s.authToken = s.getAuthToken()

	// Clean up test data
	s.cleanupTestData()
}

func (s *PipelineDeploymentAPITestSuite) TearDownSuite() {
	if s.db != nil {
		s.cleanupTestData()
		s.db.Close()
	}
}

func (s *PipelineDeploymentAPITestSuite) getAuthToken() string {
	// Login to get token
	loginReq := map[string]string{
		"username": "test@example.com",
		"password": "testpass123",
	}
	body, _ := json.Marshal(loginReq)

	resp, err := s.client.Post(s.apiURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
	s.Require().NoError(err)
	defer resp.Body.Close()

	var loginResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	s.Require().NoError(err)

	return loginResp["token"].(string)
}

func (s *PipelineDeploymentAPITestSuite) cleanupTestData() {
	queries := []string{
		"DELETE FROM pipeline_deployment_history WHERE deployment_id IN (SELECT id FROM pipeline_deployments WHERE name LIKE 'api-test-%')",
		"DELETE FROM pipeline_deployments WHERE name LIKE 'api-test-%'",
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			s.T().Logf("Warning: cleanup query failed: %v", err)
		}
	}
}

func (s *PipelineDeploymentAPITestSuite) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, s.apiURL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.authToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return s.client.Do(req)
}

func (s *PipelineDeploymentAPITestSuite) TestCreateDeployment() {
	deployment := api.PipelineDeployment{
		Name:        fmt.Sprintf("api-test-deployment-%d", time.Now().Unix()),
		Namespace:   "test",
		Template:    "process-intelligent-v1",
		Config:      json.RawMessage(`{"sampling_rate": 0.1, "batch_size": 1000}`),
		Description: "API integration test deployment",
	}

	resp, err := s.makeRequest("POST", "/api/v1/pipeline-deployments", deployment)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusCreated, resp.StatusCode)

	var created api.PipelineDeployment
	err = json.NewDecoder(resp.Body).Decode(&created)
	s.Require().NoError(err)

	s.NotEmpty(created.ID)
	s.Equal(deployment.Name, created.Name)
	s.Equal(deployment.Namespace, created.Namespace)
	s.Equal("active", created.Status)
	s.NotZero(created.CreatedAt)

	// Verify in database
	var count int
	err = s.db.QueryRow("SELECT COUNT(*) FROM pipeline_deployments WHERE id = $1", created.ID).Scan(&count)
	s.Require().NoError(err)
	s.Equal(1, count)
}

func (s *PipelineDeploymentAPITestSuite) TestGetDeployment() {
	// Create a deployment first
	deployment := api.PipelineDeployment{
		Name:        fmt.Sprintf("api-test-get-%d", time.Now().Unix()),
		Namespace:   "test",
		Template:    "process-baseline-v1",
		Config:      json.RawMessage(`{}`),
		Description: "Test deployment for GET",
	}

	resp, err := s.makeRequest("POST", "/api/v1/pipeline-deployments", deployment)
	s.Require().NoError(err)
	defer resp.Body.Close()

	var created api.PipelineDeployment
	json.NewDecoder(resp.Body).Decode(&created)

	// Get the deployment
	resp, err = s.makeRequest("GET", "/api/v1/pipeline-deployments/"+created.ID, nil)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var retrieved api.PipelineDeployment
	err = json.NewDecoder(resp.Body).Decode(&retrieved)
	s.Require().NoError(err)

	s.Equal(created.ID, retrieved.ID)
	s.Equal(created.Name, retrieved.Name)
	s.Equal(created.Template, retrieved.Template)
}

func (s *PipelineDeploymentAPITestSuite) TestListDeployments() {
	// Create multiple deployments
	namespace := "test-list"
	for i := 0; i < 3; i++ {
		deployment := api.PipelineDeployment{
			Name:      fmt.Sprintf("api-test-list-%d-%d", i, time.Now().Unix()),
			Namespace: namespace,
			Template:  "process-baseline-v1",
			Config:    json.RawMessage(`{}`),
		}
		resp, err := s.makeRequest("POST", "/api/v1/pipeline-deployments", deployment)
		s.Require().NoError(err)
		resp.Body.Close()
	}

	// List deployments
	resp, err := s.makeRequest("GET", "/api/v1/pipeline-deployments?namespace="+namespace, nil)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var deployments []api.PipelineDeployment
	err = json.NewDecoder(resp.Body).Decode(&deployments)
	s.Require().NoError(err)

	s.GreaterOrEqual(len(deployments), 3)
}

func (s *PipelineDeploymentAPITestSuite) TestUpdateDeployment() {
	// Create a deployment
	deployment := api.PipelineDeployment{
		Name:      fmt.Sprintf("api-test-update-%d", time.Now().Unix()),
		Namespace: "test",
		Template:  "process-intelligent-v1",
		Config:    json.RawMessage(`{"sampling_rate": 0.1}`),
	}

	resp, err := s.makeRequest("POST", "/api/v1/pipeline-deployments", deployment)
	s.Require().NoError(err)
	defer resp.Body.Close()

	var created api.PipelineDeployment
	json.NewDecoder(resp.Body).Decode(&created)

	// Update the deployment
	update := api.PipelineDeploymentUpdate{
		Config: json.RawMessage(`{"sampling_rate": 0.05, "batch_size": 2000}`),
		Reason: "Optimizing for lower volume",
	}

	resp, err = s.makeRequest("PATCH", "/api/v1/pipeline-deployments/"+created.ID, update)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	// Verify update
	resp, err = s.makeRequest("GET", "/api/v1/pipeline-deployments/"+created.ID, nil)
	s.Require().NoError(err)
	defer resp.Body.Close()

	var updated api.PipelineDeployment
	json.NewDecoder(resp.Body).Decode(&updated)

	var config map[string]interface{}
	json.Unmarshal(updated.Config, &config)
	s.Equal(0.05, config["sampling_rate"])
	s.Equal(float64(2000), config["batch_size"])
}

func (s *PipelineDeploymentAPITestSuite) TestUpdateDeploymentStatus() {
	// Create a deployment
	deployment := api.PipelineDeployment{
		Name:      fmt.Sprintf("api-test-status-%d", time.Now().Unix()),
		Namespace: "test",
		Template:  "process-baseline-v1",
		Config:    json.RawMessage(`{}`),
	}

	resp, err := s.makeRequest("POST", "/api/v1/pipeline-deployments", deployment)
	s.Require().NoError(err)
	defer resp.Body.Close()

	var created api.PipelineDeployment
	json.NewDecoder(resp.Body).Decode(&created)

	// Update status
	statusUpdate := map[string]string{
		"status": "suspended",
		"reason": "Maintenance window",
	}

	resp, err = s.makeRequest("PUT", "/api/v1/pipeline-deployments/"+created.ID+"/status", statusUpdate)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	// Verify status update
	resp, err = s.makeRequest("GET", "/api/v1/pipeline-deployments/"+created.ID, nil)
	s.Require().NoError(err)
	defer resp.Body.Close()

	var updated api.PipelineDeployment
	json.NewDecoder(resp.Body).Decode(&updated)
	s.Equal("suspended", updated.Status)
}

func (s *PipelineDeploymentAPITestSuite) TestDeploymentHistory() {
	// Create and update a deployment
	deployment := api.PipelineDeployment{
		Name:      fmt.Sprintf("api-test-history-%d", time.Now().Unix()),
		Namespace: "test",
		Template:  "process-intelligent-v1",
		Config:    json.RawMessage(`{"sampling_rate": 0.1}`),
	}

	resp, err := s.makeRequest("POST", "/api/v1/pipeline-deployments", deployment)
	s.Require().NoError(err)
	defer resp.Body.Close()

	var created api.PipelineDeployment
	json.NewDecoder(resp.Body).Decode(&created)

	// Update config
	update := api.PipelineDeploymentUpdate{
		Config: json.RawMessage(`{"sampling_rate": 0.05}`),
		Reason: "First update",
	}
	resp, err = s.makeRequest("PATCH", "/api/v1/pipeline-deployments/"+created.ID, update)
	s.Require().NoError(err)
	resp.Body.Close()

	// Update status
	statusUpdate := map[string]string{
		"status": "suspended",
		"reason": "Second update",
	}
	resp, err = s.makeRequest("PUT", "/api/v1/pipeline-deployments/"+created.ID+"/status", statusUpdate)
	s.Require().NoError(err)
	resp.Body.Close()

	// Get history
	resp, err = s.makeRequest("GET", "/api/v1/pipeline-deployments/"+created.ID+"/history", nil)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var history []api.PipelineDeploymentHistory
	err = json.NewDecoder(resp.Body).Decode(&history)
	s.Require().NoError(err)

	s.GreaterOrEqual(len(history), 3) // create + 2 updates
	
	// Verify history entries
	actions := make(map[string]bool)
	for _, h := range history {
		actions[h.Action] = true
	}
	s.True(actions["create"])
	s.True(actions["update"])
	s.True(actions["status_change"])
}

func (s *PipelineDeploymentAPITestSuite) TestRollbackDeployment() {
	// Create and update a deployment
	deployment := api.PipelineDeployment{
		Name:      fmt.Sprintf("api-test-rollback-%d", time.Now().Unix()),
		Namespace: "test",
		Template:  "process-intelligent-v1",
		Config:    json.RawMessage(`{"sampling_rate": 0.1, "version": "v1"}`),
	}

	resp, err := s.makeRequest("POST", "/api/v1/pipeline-deployments", deployment)
	s.Require().NoError(err)
	defer resp.Body.Close()

	var created api.PipelineDeployment
	json.NewDecoder(resp.Body).Decode(&created)

	// Get initial history entry
	resp, err = s.makeRequest("GET", "/api/v1/pipeline-deployments/"+created.ID+"/history", nil)
	s.Require().NoError(err)
	var history []api.PipelineDeploymentHistory
	json.NewDecoder(resp.Body).Decode(&history)
	resp.Body.Close()
	initialHistoryID := history[0].ID

	// Update config
	update := api.PipelineDeploymentUpdate{
		Config: json.RawMessage(`{"sampling_rate": 0.05, "version": "v2"}`),
		Reason: "Update to v2",
	}
	resp, err = s.makeRequest("PATCH", "/api/v1/pipeline-deployments/"+created.ID, update)
	s.Require().NoError(err)
	resp.Body.Close()

	// Rollback to initial version
	rollbackReq := map[string]string{
		"history_id": initialHistoryID,
		"reason":     "Rollback to v1",
	}
	resp, err = s.makeRequest("POST", "/api/v1/pipeline-deployments/"+created.ID+"/rollback", rollbackReq)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	// Verify rollback
	resp, err = s.makeRequest("GET", "/api/v1/pipeline-deployments/"+created.ID, nil)
	s.Require().NoError(err)
	defer resp.Body.Close()

	var rolledBack api.PipelineDeployment
	json.NewDecoder(resp.Body).Decode(&rolledBack)

	var config map[string]interface{}
	json.Unmarshal(rolledBack.Config, &config)
	s.Equal(0.1, config["sampling_rate"])
	s.Equal("v1", config["version"])
}

func (s *PipelineDeploymentAPITestSuite) TestDeleteDeployment() {
	// Create a deployment
	deployment := api.PipelineDeployment{
		Name:      fmt.Sprintf("api-test-delete-%d", time.Now().Unix()),
		Namespace: "test",
		Template:  "process-baseline-v1",
		Config:    json.RawMessage(`{}`),
	}

	resp, err := s.makeRequest("POST", "/api/v1/pipeline-deployments", deployment)
	s.Require().NoError(err)
	defer resp.Body.Close()

	var created api.PipelineDeployment
	json.NewDecoder(resp.Body).Decode(&created)

	// Delete the deployment
	resp, err = s.makeRequest("DELETE", "/api/v1/pipeline-deployments/"+created.ID, nil)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify deletion
	resp, err = s.makeRequest("GET", "/api/v1/pipeline-deployments/"+created.ID, nil)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *PipelineDeploymentAPITestSuite) TestExportDeployment() {
	// Create a deployment with some history
	deployment := api.PipelineDeployment{
		Name:        fmt.Sprintf("api-test-export-%d", time.Now().Unix()),
		Namespace:   "test",
		Template:    "process-intelligent-v1",
		Config:      json.RawMessage(`{"sampling_rate": 0.1}`),
		Description: "Export test deployment",
	}

	resp, err := s.makeRequest("POST", "/api/v1/pipeline-deployments", deployment)
	s.Require().NoError(err)
	defer resp.Body.Close()

	var created api.PipelineDeployment
	json.NewDecoder(resp.Body).Decode(&created)

	// Export the deployment
	resp, err = s.makeRequest("GET", "/api/v1/pipeline-deployments/"+created.ID+"/export", nil)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var export api.PipelineDeploymentExport
	err = json.NewDecoder(resp.Body).Decode(&export)
	s.Require().NoError(err)

	s.Equal(created.ID, export.Deployment.ID)
	s.NotZero(export.ExportedAt)
	s.NotNil(export.History)
}

func (s *PipelineDeploymentAPITestSuite) TestErrorHandling() {
	// Test 404 for non-existent deployment
	resp, err := s.makeRequest("GET", "/api/v1/pipeline-deployments/non-existent-id", nil)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Equal(http.StatusNotFound, resp.StatusCode)

	// Test invalid JSON
	resp, err = s.makeRequest("POST", "/api/v1/pipeline-deployments", "invalid json")
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Equal(http.StatusBadRequest, resp.StatusCode)

	// Test missing required fields
	invalidDeployment := map[string]string{
		"name": "missing-fields",
		// Missing namespace, template
	}
	resp, err = s.makeRequest("POST", "/api/v1/pipeline-deployments", invalidDeployment)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Equal(http.StatusBadRequest, resp.StatusCode)

	// Test unauthorized (no token)
	req, _ := http.NewRequest("GET", s.apiURL+"/api/v1/pipeline-deployments", nil)
	resp, err = s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func TestPipelineDeploymentAPITestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Check if API server is available
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Skip("API server not available, skipping integration tests")
	}
	resp.Body.Close()

	suite.Run(t, new(PipelineDeploymentAPITestSuite))
}