// +build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// CLIIntegrationTestSuite tests the Phoenix CLI against a running API server
type CLIIntegrationTestSuite struct {
	suite.Suite
	apiURL     string
	cliPath    string
	configPath string
	db         *sql.DB
	authToken  string
}

func (s *CLIIntegrationTestSuite) SetupSuite() {
	// Check if API server is running
	s.apiURL = os.Getenv("PHOENIX_API_URL")
	if s.apiURL == "" {
		s.apiURL = "http://localhost:8080"
	}

	// Build CLI binary
	s.cliPath = filepath.Join(s.T().TempDir(), "phoenix")
	cmd := exec.Command("go", "build", "-o", s.cliPath, "../../cmd/phoenix-cli")
	err := cmd.Run()
	require.NoError(s.T(), err, "Failed to build CLI")

	// Setup test config directory
	s.configPath = filepath.Join(s.T().TempDir(), ".phoenix")
	err = os.MkdirAll(s.configPath, 0755)
	require.NoError(s.T(), err)

	// Connect to test database
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://phoenix:phoenix@localhost:5432/phoenix_test?sslmode=disable"
	}
	s.db, err = sql.Open("postgres", dbURL)
	require.NoError(s.T(), err, "Failed to connect to test database")

	// Clean up test data
	s.cleanupTestData()
}

func (s *CLIIntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		s.cleanupTestData()
		s.db.Close()
	}
}

func (s *CLIIntegrationTestSuite) cleanupTestData() {
	// Clean up test experiments and deployments
	queries := []string{
		"DELETE FROM experiment_metrics WHERE experiment_id IN (SELECT id FROM experiments WHERE name LIKE 'cli-test-%')",
		"DELETE FROM experiments WHERE name LIKE 'cli-test-%'",
		"DELETE FROM pipeline_deployment_history WHERE deployment_id IN (SELECT id FROM pipeline_deployments WHERE name LIKE 'cli-test-%')",
		"DELETE FROM pipeline_deployments WHERE name LIKE 'cli-test-%'",
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			s.T().Logf("Warning: cleanup query failed: %v", err)
		}
	}
}

func (s *CLIIntegrationTestSuite) runCLI(args ...string) (string, error) {
	cmd := exec.Command(s.cliPath, args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("HOME=%s", filepath.Dir(s.configPath)),
		fmt.Sprintf("PHOENIX_API_URL=%s", s.apiURL),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("CLI error: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	return stdout.String(), nil
}

func (s *CLIIntegrationTestSuite) TestAuthenticationFlow() {
	// Test login
	output, err := s.runCLI("auth", "login", "--username", "test@example.com", "--password", "testpass123")
	s.Require().NoError(err)
	s.Contains(output, "Successfully logged in")

	// Test auth status
	output, err = s.runCLI("auth", "status")
	s.Require().NoError(err)
	s.Contains(output, "Authenticated as: test@example.com")

	// Save token for other tests
	configFile := filepath.Join(s.configPath, "config.yaml")
	data, err := os.ReadFile(configFile)
	s.Require().NoError(err)
	s.NotEmpty(data)

	// Test logout
	output, err = s.runCLI("auth", "logout")
	s.Require().NoError(err)
	s.Contains(output, "Successfully logged out")

	// Verify logout
	output, err = s.runCLI("auth", "status")
	s.Error(err)
	s.Contains(output, "Not authenticated")

	// Login again for subsequent tests
	_, err = s.runCLI("auth", "login", "--username", "test@example.com", "--password", "testpass123")
	s.Require().NoError(err)
}

func (s *CLIIntegrationTestSuite) TestExperimentLifecycle() {
	// Create experiment
	expName := fmt.Sprintf("cli-test-exp-%d", time.Now().Unix())
	output, err := s.runCLI("experiment", "create",
		"--name", expName,
		"--namespace", "test",
		"--pipeline-a", "process-baseline-v1",
		"--pipeline-b", "process-intelligent-v1",
		"--traffic-split", "50/50",
		"--duration", "30m",
		"--selector", "app=test-service",
		"--output", "json",
	)
	s.Require().NoError(err)

	var createResp map[string]interface{}
	err = json.Unmarshal([]byte(output), &createResp)
	s.Require().NoError(err)
	expID := createResp["id"].(string)
	s.NotEmpty(expID)

	// List experiments
	output, err = s.runCLI("experiment", "list", "--namespace", "test")
	s.Require().NoError(err)
	s.Contains(output, expName)

	// Get experiment status
	output, err = s.runCLI("experiment", "status", expID)
	s.Require().NoError(err)
	s.Contains(output, expName)
	s.Contains(output, "created")

	// Start experiment
	output, err = s.runCLI("experiment", "start", expID)
	s.Require().NoError(err)
	s.Contains(output, "started successfully")

	// Wait a bit for status update
	time.Sleep(2 * time.Second)

	// Check status again
	output, err = s.runCLI("experiment", "status", expID, "--output", "json")
	s.Require().NoError(err)
	var statusResp map[string]interface{}
	err = json.Unmarshal([]byte(output), &statusResp)
	s.Require().NoError(err)
	s.Equal("running", statusResp["status"])

	// Get metrics
	output, err = s.runCLI("experiment", "metrics", expID)
	s.Require().NoError(err)
	s.Contains(output, "Summary Metrics")

	// Stop experiment
	output, err = s.runCLI("experiment", "stop", expID, "--reason", "Integration test complete")
	s.Require().NoError(err)
	s.Contains(output, "stopped successfully")

	// Export configuration
	output, err = s.runCLI("experiment", "export", expID)
	s.Require().NoError(err)
	s.Contains(output, "apiVersion:")
	s.Contains(output, expName)
}

func (s *CLIIntegrationTestSuite) TestPipelineDeploymentFlow() {
	// Create deployment
	depName := fmt.Sprintf("cli-test-dep-%d", time.Now().Unix())
	output, err := s.runCLI("pipeline", "deploy",
		"--name", depName,
		"--namespace", "test",
		"--template", "process-intelligent-v1",
		"--description", "CLI integration test deployment",
		"--config-override", `{"sampling_rate": 0.1}`,
		"--output", "json",
	)
	s.Require().NoError(err)

	var deployResp map[string]interface{}
	err = json.Unmarshal([]byte(output), &deployResp)
	s.Require().NoError(err)
	depID := deployResp["id"].(string)
	s.NotEmpty(depID)

	// List deployments
	output, err = s.runCLI("pipeline", "deployments", "list", "--namespace", "test")
	s.Require().NoError(err)
	s.Contains(output, depName)

	// Get deployment status
	output, err = s.runCLI("pipeline", "deployment", "status", depID)
	s.Require().NoError(err)
	s.Contains(output, depName)
	s.Contains(output, "active")

	// Update deployment
	output, err = s.runCLI("pipeline", "deployment", "update", depID,
		"--config-override", `{"sampling_rate": 0.05}`,
		"--reason", "Reduce sampling rate for test",
	)
	s.Require().NoError(err)
	s.Contains(output, "updated successfully")

	// Get deployment history
	output, err = s.runCLI("pipeline", "deployment", "history", depID)
	s.Require().NoError(err)
	s.Contains(output, "Reduce sampling rate")

	// Export deployment
	output, err = s.runCLI("pipeline", "deployment", "export", depID)
	s.Require().NoError(err)
	s.Contains(output, "deployment:")
	s.Contains(output, depName)
}

func (s *CLIIntegrationTestSuite) TestErrorHandling() {
	// Test invalid experiment ID
	output, err := s.runCLI("experiment", "status", "invalid-exp-id")
	s.Error(err)
	s.Contains(output, "not found")

	// Test invalid traffic split
	output, err = s.runCLI("experiment", "create",
		"--name", "invalid-test",
		"--namespace", "test",
		"--pipeline-a", "baseline",
		"--pipeline-b", "optimized",
		"--traffic-split", "60/60", // Invalid: doesn't sum to 100
		"--duration", "30m",
		"--selector", "app=test",
	)
	s.Error(err)
	s.Contains(output, "must sum to 100")

	// Test missing required flags
	output, err = s.runCLI("experiment", "create", "--name", "incomplete-test")
	s.Error(err)
	s.Contains(output, "required flag")
}

func (s *CLIIntegrationTestSuite) TestConfigManagement() {
	// Set config value
	output, err := s.runCLI("config", "set", "default_namespace", "test-namespace")
	s.Require().NoError(err)
	s.Contains(output, "Configuration updated")

	// Get config value
	output, err = s.runCLI("config", "get", "default_namespace")
	s.Require().NoError(err)
	s.Contains(output, "test-namespace")

	// List all config
	output, err = s.runCLI("config", "list")
	s.Require().NoError(err)
	s.Contains(output, "default_namespace")
	s.Contains(output, "test-namespace")

	// Reset config
	output, err = s.runCLI("config", "reset")
	s.Require().NoError(err)
	s.Contains(output, "Configuration reset")

	// Verify reset
	output, err = s.runCLI("config", "get", "default_namespace")
	s.Require().NoError(err)
	s.NotContains(output, "test-namespace")
}

func (s *CLIIntegrationTestSuite) TestOutputFormats() {
	// Create test experiment
	expName := fmt.Sprintf("cli-test-output-%d", time.Now().Unix())
	output, err := s.runCLI("experiment", "create",
		"--name", expName,
		"--namespace", "test",
		"--pipeline-a", "process-baseline-v1",
		"--pipeline-b", "process-intelligent-v1",
		"--traffic-split", "50/50",
		"--duration", "30m",
		"--selector", "app=test-service",
		"--output", "json",
	)
	s.Require().NoError(err)

	// Test JSON output
	var jsonResp map[string]interface{}
	err = json.Unmarshal([]byte(output), &jsonResp)
	s.Require().NoError(err)
	s.Equal(expName, jsonResp["name"])

	// Test YAML output
	output, err = s.runCLI("experiment", "list", "--namespace", "test", "--output", "yaml")
	s.Require().NoError(err)
	s.Contains(output, "experiments:")
	s.Contains(output, expName)

	// Test table output (default)
	output, err = s.runCLI("experiment", "list", "--namespace", "test")
	s.Require().NoError(err)
	s.Contains(output, "ID")
	s.Contains(output, "NAME")
	s.Contains(output, "STATUS")
}

func TestCLIIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Check if API server is available
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Skip("API server not available, skipping integration tests")
	}
	resp.Body.Close()

	suite.Run(t, new(CLIIntegrationTestSuite))
}