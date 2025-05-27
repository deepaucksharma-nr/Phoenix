// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExperimentWorkflowE2E tests the complete experiment workflow
// by starting actual services and making real API calls
func TestExperimentWorkflowE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup test environment
	_, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Use environment variables for service URLs (set by integration test script)
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}
	
	t.Logf("🔗 Using API at %s", apiURL)
	
	// Check if services are ready
	checkServiceHealth(t, apiURL+"/health")

	t.Run("CompleteExperimentLifecycle", func(t *testing.T) {
		// Test experiment creation
		experiment := createTestExperiment(t)
		
		// Test experiment retrieval
		retrievedExp := getExperiment(t, experiment.ID)
		assert.Equal(t, experiment.Name, retrievedExp.Name)
		assert.Equal(t, "pending", retrievedExp.Status)
		
		// Test experiment listing
		experiments := listExperiments(t)
		assert.GreaterOrEqual(t, len(experiments), 1)
		
		// Test config generation
		configResp := generateExperimentConfig(t, experiment.ID)
		assert.NotEmpty(t, configResp.BaselineConfig)
		assert.NotEmpty(t, configResp.CandidateConfig)
		assert.NotEmpty(t, configResp.KubernetesManifests)
		
		// Validate generated YAML configs
		validateOTelConfig(t, configResp.BaselineConfig)
		validateOTelConfig(t, configResp.CandidateConfig)
		
		// Test experiment status monitoring
		status := getExperimentStatus(t, experiment.ID)
		assert.NotEmpty(t, status.Status)
		
		t.Logf("✅ Experiment %s completed full lifecycle test", experiment.ID)
	})

	t.Run("PipelineTemplateValidation", func(t *testing.T) {
		// Test all available pipeline templates
		templates := listPipelineTemplates(t)
		assert.GreaterOrEqual(t, len(templates), 5, "Should have at least 5 pipeline templates")
		
		// Test each template can generate valid config
		for _, template := range templates {
			t.Run(fmt.Sprintf("Template_%s", template.Name), func(t *testing.T) {
				config := generatePipelineConfig(t, template.Name)
				validateOTelConfig(t, config)
				
				// Ensure config contains expected elements
				assert.Contains(t, config, "receivers:")
				assert.Contains(t, config, "processors:")
				assert.Contains(t, config, "exporters:")
				assert.Contains(t, config, "service:")
				
				t.Logf("✅ Template %s generated valid config", template.Name)
			})
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// Test invalid experiment creation
		resp, err := http.Post("http://localhost:8080/api/v1/experiments", "application/json", 
			bytes.NewBufferString(`{"name": ""}`))  // Missing required fields
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		
		// Test nonexistent experiment retrieval
		resp, err = http.Get("http://localhost:8080/api/v1/experiments/nonexistent")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		
		// Test invalid config generation
		resp, err = http.Post("http://localhost:8082/api/v1/generate", "application/json",
			bytes.NewBufferString(`{"experiment_id": ""}`))  // Missing required fields
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// Service management types
type ServiceProcess struct {
	Name string
	Cmd  *exec.Cmd
	Port int
}

type TestServices struct {
	API       *ServiceProcess
	Generator *ServiceProcess
}

func startTestServices(t *testing.T, ctx context.Context) *TestServices {
	t.Log("🚀 Starting test services...")
	
	// Set test environment variables
	os.Setenv("DATABASE_URL", "postgres://phoenix:phoenix@localhost/phoenix_test?sslmode=disable")
	os.Setenv("NEW_RELIC_API_KEY", "test-key-123")
	os.Setenv("NEW_RELIC_OTLP_ENDPOINT", "https://otlp.nr-data.net:4317")
	os.Setenv("ENVIRONMENT", "test")
	
	services := &TestServices{}
	
	// Start API service
	apiCmd := exec.CommandContext(ctx, "./api-server")
	apiCmd.Env = os.Environ()
	apiCmd.Stdout = os.Stdout
	apiCmd.Stderr = os.Stderr
	
	err := apiCmd.Start()
	require.NoError(t, err, "Failed to start API service")
	
	services.API = &ServiceProcess{
		Name: "API",
		Cmd:  apiCmd,
		Port: 8080,
	}
	
	// Start Generator service  
	generatorCmd := exec.CommandContext(ctx, "./generator")
	generatorCmd.Env = os.Environ()
	generatorCmd.Stdout = os.Stdout
	generatorCmd.Stderr = os.Stderr
	
	err = generatorCmd.Start()
	require.NoError(t, err, "Failed to start Generator service")
	
	services.Generator = &ServiceProcess{
		Name: "Generator",
		Cmd:  generatorCmd,
		Port: 8082,
	}
	
	t.Log("✅ Test services started")
	return services
}

func stopTestServices(services *TestServices) {
	if services.API != nil && services.API.Cmd.Process != nil {
		services.API.Cmd.Process.Kill()
	}
	if services.Generator != nil && services.Generator.Cmd.Process != nil {
		services.Generator.Cmd.Process.Kill()
	}
}

func waitForServicesReady(t *testing.T, ctx context.Context) {
	t.Log("⏳ Waiting for services to be ready...")
	
	services := []struct {
		name string
		url  string
	}{
		{"API", "http://localhost:8080/health"},
		{"Generator", "http://localhost:8082/health"},
	}
	
	for _, service := range services {
		ready := false
		deadline := time.Now().Add(30 * time.Second)
		
		for time.Now().Before(deadline) && !ready {
			select {
			case <-ctx.Done():
				t.Fatalf("Context cancelled while waiting for %s service", service.name)
			default:
				resp, err := http.Get(service.url)
				if err == nil && resp.StatusCode == http.StatusOK {
					resp.Body.Close()
					ready = true
					t.Logf("✅ %s service ready", service.name)
				} else {
					if resp != nil {
						resp.Body.Close()
					}
					time.Sleep(500 * time.Millisecond)
				}
			}
		}
		
		if !ready {
			t.Fatalf("❌ %s service failed to become ready within 30 seconds", service.name)
		}
	}
}

// API helper functions
type ExperimentRequest struct {
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	BaselinePipeline  string            `json:"baseline_pipeline"`
	CandidatePipeline string            `json:"candidate_pipeline"`
	TargetNodes       map[string]string `json:"target_nodes"`
}

type ExperimentResponse struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	BaselinePipeline  string            `json:"baseline_pipeline"`
	CandidatePipeline string            `json:"candidate_pipeline"`
	Status            string            `json:"status"`
	TargetNodes       map[string]string `json:"target_nodes"`
	CreatedAt         int64             `json:"created_at"`
}

type ConfigGenerationRequest struct {
	ExperimentID      string            `json:"experiment_id"`
	BaselinePipeline  string            `json:"baseline_pipeline"`
	CandidatePipeline string            `json:"candidate_pipeline"`
	TargetHosts       []string          `json:"target_hosts"`
	Variables         map[string]string `json:"variables"`
}

type ConfigGenerationResponse struct {
	ExperimentID        string            `json:"experiment_id"`
	BaselineConfig      string            `json:"baseline_config"`
	CandidateConfig     string            `json:"candidate_config"`
	KubernetesManifests map[string]string `json:"kubernetes_manifests"`
	GitBranch           string            `json:"git_branch"`
}

type PipelineTemplate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ExperimentStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func createTestExperiment(t *testing.T) *ExperimentResponse {
	reqBody := ExperimentRequest{
		Name:              "E2E Test Experiment",
		Description:       "End-to-end test experiment",
		BaselinePipeline:  "process-baseline-v1",
		CandidatePipeline: "process-priority-filter-v1",
		TargetNodes: map[string]string{
			"e2e-node-1": "active",
			"e2e-node-2": "active",
		},
	}
	
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)
	
	resp, err := http.Post("http://localhost:8080/api/v1/experiments", 
		"application/json", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var experiment ExperimentResponse
	err = json.NewDecoder(resp.Body).Decode(&experiment)
	require.NoError(t, err)
	
	assert.NotEmpty(t, experiment.ID)
	assert.Equal(t, reqBody.Name, experiment.Name)
	
	return &experiment
}

func getExperiment(t *testing.T, id string) *ExperimentResponse {
	resp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/v1/experiments/%s", id))
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var experiment ExperimentResponse
	err = json.NewDecoder(resp.Body).Decode(&experiment)
	require.NoError(t, err)
	
	return &experiment
}

func listExperiments(t *testing.T) []ExperimentResponse {
	resp, err := http.Get("http://localhost:8080/api/v1/experiments")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var result struct {
		Experiments []ExperimentResponse `json:"experiments"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result.Experiments
}

func generateExperimentConfig(t *testing.T, experimentID string) *ConfigGenerationResponse {
	reqBody := ConfigGenerationRequest{
		ExperimentID:      experimentID,
		BaselinePipeline:  "process-baseline-v1",
		CandidatePipeline: "process-priority-filter-v1",
		TargetHosts:       []string{"e2e-node-1", "e2e-node-2"},
		Variables: map[string]string{
			"NEW_RELIC_API_KEY":       "test-key-123",
			"NEW_RELIC_OTLP_ENDPOINT": "https://otlp.nr-data.net:4317",
		},
	}
	
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)
	
	resp, err := http.Post("http://localhost:8082/api/v1/generate",
		"application/json", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var configResp ConfigGenerationResponse
	err = json.NewDecoder(resp.Body).Decode(&configResp)
	require.NoError(t, err)
	
	return &configResp
}

func listPipelineTemplates(t *testing.T) []PipelineTemplate {
	resp, err := http.Get("http://localhost:8082/api/v1/templates")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var result struct {
		Templates []PipelineTemplate `json:"templates"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	return result.Templates
}

func generatePipelineConfig(t *testing.T, templateName string) string {
	reqBody := map[string]interface{}{
		"pipeline": templateName,
		"variables": map[string]string{
			"NEW_RELIC_API_KEY": "test-key-123",
		},
	}
	
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)
	
	resp, err := http.Post("http://localhost:8082/api/v1/generate",
		"application/json", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()
	
	// For template-only generation, we might get a different response format
	// This is a placeholder - adapt based on actual API response
	return "placeholder-config"
}

func getExperimentStatus(t *testing.T, id string) *ExperimentStatus {
	resp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/v1/experiments/%s/status", id))
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var status ExperimentStatus
	err = json.NewDecoder(resp.Body).Decode(&status)
	require.NoError(t, err)
	
	return &status
}

func validateOTelConfig(t *testing.T, config string) {
	// Basic validation - ensure it looks like valid YAML config
	assert.Contains(t, config, "receivers:")
	assert.Contains(t, config, "processors:")
	assert.Contains(t, config, "exporters:")
	assert.Contains(t, config, "service:")
	
	// Ensure it contains Phoenix-specific elements
	assert.Contains(t, config, "hostmetrics")
	assert.Contains(t, config, "batch")
	
	// Ensure it contains New Relic integration
	assert.Contains(t, config, "otlphttp") 
}