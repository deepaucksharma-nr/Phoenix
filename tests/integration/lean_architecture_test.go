//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// LeanArchitectureTestSuite tests the Phoenix lean architecture end-to-end
type LeanArchitectureTestSuite struct {
	suite.Suite
	apiURL        string
	agentHostID   string
	experimentID  string
}

func (s *LeanArchitectureTestSuite) SetupSuite() {
	s.apiURL = os.Getenv("PHOENIX_API_URL")
	if s.apiURL == "" {
		s.apiURL = "http://localhost:8080"
	}
	s.agentHostID = "test-host-001"
}

func (s *LeanArchitectureTestSuite) TearDownSuite() {
	// Clean up any created experiments
	if s.experimentID != "" {
		s.stopExperiment(s.experimentID)
	}
}

func TestLeanArchitectureSuite(t *testing.T) {
	suite.Run(t, new(LeanArchitectureTestSuite))
}

// Test the complete flow: API -> Agent -> Tasks -> Metrics
func (s *LeanArchitectureTestSuite) TestCompleteExperimentFlow() {
	// Step 1: Create an experiment
	s.T().Log("Creating experiment...")
	experiment := s.createExperiment()
	s.experimentID = experiment["id"].(string)
	
	// Step 2: Start the experiment
	s.T().Log("Starting experiment...")
	s.startExperiment(s.experimentID)
	
	// Step 3: Simulate agent polling for tasks
	s.T().Log("Agent polling for tasks...")
	tasks := s.agentGetTasks()
	s.Require().NotEmpty(tasks, "Agent should receive tasks")
	
	// Verify we got both baseline and candidate tasks
	var baselineTask, candidateTask map[string]interface{}
	for _, task := range tasks {
		taskMap := task.(map[string]interface{})
		config := taskMap["config"].(map[string]interface{})
		if config["variant"] == "baseline" {
			baselineTask = taskMap
		} else if config["variant"] == "candidate" {
			candidateTask = taskMap
		}
	}
	
	s.Require().NotNil(baselineTask, "Should have baseline task")
	s.Require().NotNil(candidateTask, "Should have candidate task")
	
	// Step 4: Update task status to running
	s.T().Log("Updating task status to running...")
	s.updateTaskStatus(baselineTask["id"].(string), "running")
	s.updateTaskStatus(candidateTask["id"].(string), "running")
	
	// Step 5: Send agent heartbeat
	s.T().Log("Sending agent heartbeat...")
	s.sendHeartbeat()
	
	// Step 6: Simulate metrics push
	s.T().Log("Pushing metrics...")
	s.pushMetrics()
	
	// Step 7: Complete tasks
	s.T().Log("Completing tasks...")
	s.updateTaskStatus(baselineTask["id"].(string), "completed")
	s.updateTaskStatus(candidateTask["id"].(string), "completed")
	
	// Step 8: Check experiment status
	s.T().Log("Checking experiment status...")
	s.Eventually(func() bool {
		exp := s.getExperiment(s.experimentID)
		phase := exp["phase"].(string)
		s.T().Logf("Experiment phase: %s", phase)
		return phase == "running" || phase == "monitoring"
	}, 30*time.Second, 1*time.Second, "Experiment should be in running/monitoring phase")
	
	// Step 9: Calculate KPIs
	s.T().Log("Calculating KPIs...")
	kpis := s.calculateKPIs(s.experimentID)
	s.Require().NotNil(kpis)
	s.T().Logf("KPIs: %+v", kpis)
	
	// Step 10: Stop experiment
	s.T().Log("Stopping experiment...")
	s.stopExperiment(s.experimentID)
}

// Test agent task polling with long-polling
func (s *LeanArchitectureTestSuite) TestAgentLongPolling() {
	// Start timing
	start := time.Now()
	
	// Poll with no tasks available (should wait)
	tasks := s.agentGetTasks()
	elapsed := time.Since(start)
	
	// Should have waited close to timeout (allowing some margin)
	s.Assert().True(elapsed >= 25*time.Second, "Long polling should wait when no tasks available")
	s.Assert().Empty(tasks, "No tasks should be returned")
}

// Test task retry on failure
func (s *LeanArchitectureTestSuite) TestTaskRetryMechanism() {
	// Create and start experiment
	experiment := s.createExperiment()
	s.experimentID = experiment["id"].(string)
	s.startExperiment(s.experimentID)
	
	// Get tasks
	tasks := s.agentGetTasks()
	s.Require().NotEmpty(tasks)
	
	taskID := tasks[0].(map[string]interface{})["id"].(string)
	
	// Mark task as failed
	s.updateTaskStatusWithError(taskID, "failed", "Simulated failure")
	
	// Poll again - should get retry task
	time.Sleep(2 * time.Second)
	retryTasks := s.agentGetTasks()
	s.Require().NotEmpty(retryTasks, "Should receive retry task")
	
	// Verify retry count increased
	retryTask := retryTasks[0].(map[string]interface{})
	s.Assert().Equal(1, int(retryTask["retry_count"].(float64)), "Retry count should be 1")
}

// Helper methods

func (s *LeanArchitectureTestSuite) createExperiment() map[string]interface{} {
	payload := map[string]interface{}{
		"name":        "Integration Test Experiment",
		"description": "Testing lean architecture",
		"config": map[string]interface{}{
			"target_hosts": []string{s.agentHostID},
			"baseline_template": map[string]interface{}{
				"url": "http://config-server/baseline.yaml",
				"variables": map[string]string{
					"BATCH_SIZE": "1000",
				},
			},
			"candidate_template": map[string]interface{}{
				"url": "http://config-server/candidate.yaml",
				"variables": map[string]string{
					"BATCH_SIZE":     "500",
					"CPU_THRESHOLD":  "0.05",
				},
			},
			"duration":        "5m",
			"warmup_duration": "30s",
		},
	}
	
	resp := s.makeAPIRequest("POST", "/api/v1/experiments", payload)
	s.Require().Equal(201, resp.StatusCode)
	
	var experiment map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&experiment)
	return experiment
}

func (s *LeanArchitectureTestSuite) startExperiment(id string) {
	resp := s.makeAPIRequest("POST", fmt.Sprintf("/api/v1/experiments/%s/start", id), nil)
	s.Require().Equal(202, resp.StatusCode)
}

func (s *LeanArchitectureTestSuite) stopExperiment(id string) {
	resp := s.makeAPIRequest("POST", fmt.Sprintf("/api/v1/experiments/%s/stop", id), nil)
	s.Require().Equal(202, resp.StatusCode)
}

func (s *LeanArchitectureTestSuite) getExperiment(id string) map[string]interface{} {
	resp := s.makeAPIRequest("GET", fmt.Sprintf("/api/v1/experiments/%s", id), nil)
	s.Require().Equal(200, resp.StatusCode)
	
	var experiment map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&experiment)
	return experiment
}

func (s *LeanArchitectureTestSuite) agentGetTasks() []interface{} {
	req, _ := http.NewRequest("GET", s.apiURL+"/api/v1/agent/tasks", nil)
	req.Header.Set("X-Agent-Host-ID", s.agentHostID)
	
	client := &http.Client{Timeout: 35 * time.Second}
	resp, err := client.Do(req)
	s.Require().NoError(err)
	s.Require().Equal(200, resp.StatusCode)
	
	var tasks []interface{}
	json.NewDecoder(resp.Body).Decode(&tasks)
	return tasks
}

func (s *LeanArchitectureTestSuite) updateTaskStatus(taskID, status string) {
	payload := map[string]interface{}{
		"status": status,
	}
	
	req, _ := http.NewRequest("POST", 
		fmt.Sprintf("%s/api/v1/agent/tasks/%s/status", s.apiURL, taskID), 
		nil)
	req.Header.Set("X-Agent-Host-ID", s.agentHostID)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	s.Require().NoError(err)
	s.Require().Equal(204, resp.StatusCode)
}

func (s *LeanArchitectureTestSuite) updateTaskStatusWithError(taskID, status, errorMsg string) {
	payload := map[string]interface{}{
		"status":        status,
		"error_message": errorMsg,
	}
	
	s.makeAgentRequest("POST", 
		fmt.Sprintf("/api/v1/agent/tasks/%s/status", taskID), 
		payload)
}

func (s *LeanArchitectureTestSuite) sendHeartbeat() {
	payload := map[string]interface{}{
		"agent_version": "1.0.0",
		"status":        "healthy",
		"active_tasks":  []string{},
		"resource_usage": map[string]interface{}{
			"cpu_percent":    25.5,
			"memory_percent": 40.2,
			"memory_bytes":   1073741824,
		},
	}
	
	resp := s.makeAgentRequest("POST", "/api/v1/agent/heartbeat", payload)
	s.Require().Equal(204, resp.StatusCode)
}

func (s *LeanArchitectureTestSuite) pushMetrics() {
	payload := map[string]interface{}{
		"timestamp": time.Now(),
		"metrics": []map[string]interface{}{
			{
				"name":      "http.request.count",
				"value":     100,
				"timestamp": time.Now().Unix(),
				"labels": map[string]string{
					"experiment_id": s.experimentID,
					"variant":       "baseline",
				},
			},
			{
				"name":      "http.request.count",
				"value":     50,
				"timestamp": time.Now().Unix(),
				"labels": map[string]string{
					"experiment_id": s.experimentID,
					"variant":       "candidate",
				},
			},
		},
	}
	
	resp := s.makeAgentRequest("POST", "/api/v1/agent/metrics", payload)
	s.Require().Equal(202, resp.StatusCode)
}

func (s *LeanArchitectureTestSuite) calculateKPIs(experimentID string) map[string]interface{} {
	payload := map[string]interface{}{
		"duration": "5m",
	}
	
	resp := s.makeAPIRequest("POST", 
		fmt.Sprintf("/api/v1/experiments/%s/kpis", experimentID), 
		payload)
	s.Require().Equal(200, resp.StatusCode)
	
	var kpis map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&kpis)
	return kpis
}

func (s *LeanArchitectureTestSuite) makeAPIRequest(method, path string, payload interface{}) *http.Response {
	var req *http.Request
	var err error
	
	if payload != nil {
		data, _ := json.Marshal(payload)
		req, err = http.NewRequest(method, s.apiURL+path, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, s.apiURL+path, nil)
	}
	
	s.Require().NoError(err)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	s.Require().NoError(err)
	
	return resp
}

func (s *LeanArchitectureTestSuite) makeAgentRequest(method, path string, payload interface{}) *http.Response {
	resp := s.makeAPIRequest(method, path, payload)
	resp.Request.Header.Set("X-Agent-Host-ID", s.agentHostID)
	return resp
}