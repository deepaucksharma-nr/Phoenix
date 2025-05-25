// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/phoenix-vnext/platform/pkg/generator"
)

// TestConfigGeneratorIntegration tests the config generator service
func TestConfigGeneratorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zap.NewNop()
	
	// Create generator service
	generatorService := generator.NewService(logger, "https://github.com/phoenix/configs", "")

	t.Run("GenerateExperimentConfig", func(t *testing.T) {
		req := &generator.GenerateRequest{
			ExperimentID:      "test-exp-config-1",
			BaselinePipeline:  "process-baseline-v1",
			CandidatePipeline: "process-priority-filter-v1",
			TargetHosts:       []string{"test-node-1", "test-node-2"},
			Variables: map[string]string{
				"NEW_RELIC_API_KEY":      "test-api-key",
				"NEW_RELIC_OTLP_ENDPOINT": "https://otlp.nr-data.net:4317",
			},
			Duration: 10 * time.Minute,
		}

		ctx := context.Background()
		resp, err := generatorService.GenerateExperimentConfig(ctx, req)
		require.NoError(t, err)

		// Verify response structure
		assert.Equal(t, req.ExperimentID, resp.ExperimentID)
		assert.NotEmpty(t, resp.BaselineConfig)
		assert.NotEmpty(t, resp.CandidateConfig)
		assert.NotEmpty(t, resp.KubernetesManifests)
		assert.Equal(t, "experiment/test-exp-config-1", resp.GitBranch)

		// Verify baseline config is valid YAML
		var baselineConfig map[string]interface{}
		err = yaml.Unmarshal([]byte(resp.BaselineConfig), &baselineConfig)
		require.NoError(t, err, "Baseline config should be valid YAML")

		// Verify candidate config is valid YAML
		var candidateConfig map[string]interface{}
		err = yaml.Unmarshal([]byte(resp.CandidateConfig), &candidateConfig)
		require.NoError(t, err, "Candidate config should be valid YAML")

		// Verify basic structure of generated config
		assert.Contains(t, baselineConfig, "receivers")
		assert.Contains(t, baselineConfig, "processors")
		assert.Contains(t, baselineConfig, "exporters")
		assert.Contains(t, baselineConfig, "service")

		// Verify manifests were generated
		expectedManifests := []string{
			"namespace.yaml",
			"baseline-deployment.yaml",
			"candidate-deployment.yaml",
			"baseline-configmap.yaml",
			"candidate-configmap.yaml",
			"services.yaml",
			"network-policy.yaml",
		}

		for _, manifestName := range expectedManifests {
			assert.Contains(t, resp.KubernetesManifests, manifestName, 
				"Should contain manifest: %s", manifestName)
			assert.NotEmpty(t, resp.KubernetesManifests[manifestName],
				"Manifest %s should not be empty", manifestName)
		}

		// Verify namespace manifest contains experiment ID
		namespaceManifest := resp.KubernetesManifests["namespace.yaml"]
		assert.Contains(t, namespaceManifest, req.ExperimentID)

		// Verify deployment manifests contain proper labels
		baselineDeployment := resp.KubernetesManifests["baseline-deployment.yaml"]
		assert.Contains(t, baselineDeployment, "variant: baseline")
		assert.Contains(t, baselineDeployment, "phoenix.io/experiment-id: "+req.ExperimentID)

		candidateDeployment := resp.KubernetesManifests["candidate-deployment.yaml"]
		assert.Contains(t, candidateDeployment, "variant: candidate")
		assert.Contains(t, candidateDeployment, "phoenix.io/experiment-id: "+req.ExperimentID)
	})

	t.Run("GeneratePipelineConfig", func(t *testing.T) {
		ctx := context.Background()
		
		// Test basic pipeline config generation
		pipelineName := "process-filter-v1"
		params := map[string]interface{}{
			"filter_pattern": "process.*",
			"sampling_rate":  "10",
		}

		config, err := generatorService.GeneratePipelineConfig(ctx, pipelineName, params)
		require.NoError(t, err)
		assert.NotEmpty(t, config)

		// Verify it's valid YAML
		var configMap map[string]interface{}
		err = yaml.Unmarshal([]byte(config), &configMap)
		require.NoError(t, err)

		// Verify basic OTel structure
		assert.Contains(t, configMap, "receivers")
		assert.Contains(t, configMap, "processors")
		assert.Contains(t, configMap, "exporters")
		assert.Contains(t, configMap, "service")
	})

	t.Run("DifferentPipelineTypes", func(t *testing.T) {
		ctx := context.Background()
		
		testCases := []struct {
			name            string
			pipeline        string
			expectedProcessors []string
		}{
			{
				name:     "filter pipeline",
				pipeline: "process-filter-v1",
				expectedProcessors: []string{"batch", "filter"},
			},
			{
				name:     "aggregate pipeline",
				pipeline: "process-aggregate-v1",
				expectedProcessors: []string{"batch", "groupbyattrs"},
			},
			{
				name:     "sample pipeline",
				pipeline: "process-sample-v1",
				expectedProcessors: []string{"batch", "probabilistic_sampler"},
			},
			{
				name:     "combined pipeline",
				pipeline: "process-filter-aggregate-sample-v1",
				expectedProcessors: []string{"batch", "filter", "groupbyattrs", "probabilistic_sampler"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				config, err := generatorService.GeneratePipelineConfig(ctx, tc.pipeline, map[string]interface{}{})
				require.NoError(t, err)

				var configMap map[string]interface{}
				err = yaml.Unmarshal([]byte(config), &configMap)
				require.NoError(t, err)

				// Check that pipeline contains expected processors
				service := configMap["service"].(map[string]interface{})
				pipelines := service["pipelines"].(map[string]interface{})
				metrics := pipelines["metrics"].(map[string]interface{})
				processors := metrics["processors"].([]interface{})

				processorNames := make([]string, len(processors))
				for i, p := range processors {
					processorNames[i] = p.(string)
				}

				for _, expectedProcessor := range tc.expectedProcessors {
					assert.Contains(t, processorNames, expectedProcessor,
						"Pipeline %s should contain processor %s", tc.pipeline, expectedProcessor)
				}
			})
		}
	})

	t.Run("VariableSubstitution", func(t *testing.T) {
		ctx := context.Background()
		
		variables := map[string]string{
			"NEW_RELIC_API_KEY":       "test-key-123",
			"NEW_RELIC_OTLP_ENDPOINT": "https://test.endpoint.com:4317",
		}

		config, err := generatorService.GeneratePipelineConfig(ctx, "test-pipeline", variables)
		require.NoError(t, err)

		// Variables should be substituted in the config
		assert.Contains(t, config, "test-key-123")
		assert.Contains(t, config, "https://test.endpoint.com:4317")

		// Should not contain variable placeholders
		assert.NotContains(t, config, "${NEW_RELIC_API_KEY}")
		assert.NotContains(t, config, "${NEW_RELIC_OTLP_ENDPOINT}")
	})
}

// TestConfigGeneratorHTTPAPI tests the HTTP API endpoints
func TestConfigGeneratorHTTPAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zap.NewNop()
	generatorService := generator.NewService(logger, "https://github.com/phoenix/configs", "")

	t.Run("GenerateConfigEndpoint", func(t *testing.T) {
		// Create HTTP handler
		handler := createGenerateConfigHandler(logger, generatorService)

		// Create test request
		req := generator.GenerateRequest{
			ExperimentID:      "test-http-1",
			BaselinePipeline:  "process-baseline-v1",
			CandidatePipeline: "process-priority-filter-v1",
			TargetHosts:       []string{"test-node-1"},
			Variables: map[string]string{
				"NEW_RELIC_API_KEY": "test-api-key",
			},
			Duration: 5 * time.Minute,
		}

		reqBody, err := json.Marshal(req)
		require.NoError(t, err)

		// Create HTTP request
		httpReq := httptest.NewRequest("POST", "/api/v1/generate", bytes.NewReader(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		// Create response recorder
		recorder := httptest.NewRecorder()

		// Execute request
		handler.ServeHTTP(recorder, httpReq)

		// Verify response
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		// Parse response
		var resp generator.GenerateResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		require.NoError(t, err)

		// Verify response content
		assert.Equal(t, req.ExperimentID, resp.ExperimentID)
		assert.NotEmpty(t, resp.BaselineConfig)
		assert.NotEmpty(t, resp.CandidateConfig)
		assert.NotEmpty(t, resp.KubernetesManifests)
	})

	t.Run("InvalidRequestMethod", func(t *testing.T) {
		handler := createGenerateConfigHandler(logger, generatorService)

		// Test with GET request (should be POST)
		httpReq := httptest.NewRequest("GET", "/api/v1/generate", nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
	})

	t.Run("InvalidRequestBody", func(t *testing.T) {
		handler := createGenerateConfigHandler(logger, generatorService)

		// Test with invalid JSON
		httpReq := httptest.NewRequest("POST", "/api/v1/generate", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, httpReq)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("ListTemplatesEndpoint", func(t *testing.T) {
		// Create HTTP handler for templates
		handler := createListTemplatesHandler(logger, generatorService)

		// Create HTTP request
		httpReq := httptest.NewRequest("GET", "/api/v1/templates", nil)
		recorder := httptest.NewRecorder()

		// Execute request
		handler.ServeHTTP(recorder, httpReq)

		// Verify response
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		// Parse response
		var resp struct {
			Templates []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"templates"`
		}
		err := json.Unmarshal(recorder.Body.Bytes(), &resp)
		require.NoError(t, err)

		// Verify templates are returned (may be empty list if no templates configured)
		assert.NotNil(t, resp.Templates)
	})
}

// TestConfigGeneratorErrorHandling tests error scenarios
func TestConfigGeneratorErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zap.NewNop()
	generatorService := generator.NewService(logger, "https://github.com/phoenix/configs", "")

	t.Run("EmptyExperimentID", func(t *testing.T) {
		req := &generator.GenerateRequest{
			ExperimentID:      "", // Empty ID
			BaselinePipeline:  "process-baseline-v1",
			CandidatePipeline: "process-priority-filter-v1",
			TargetHosts:       []string{"test-node-1"},
		}

		ctx := context.Background()
		resp, err := generatorService.GenerateExperimentConfig(ctx, req)
		
		// Should handle gracefully and generate config with empty ID
		// (In a real implementation, you might want to validate and return an error)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("EmptyTargetHosts", func(t *testing.T) {
		req := &generator.GenerateRequest{
			ExperimentID:      "test-empty-hosts",
			BaselinePipeline:  "process-baseline-v1",
			CandidatePipeline: "process-priority-filter-v1",
			TargetHosts:       []string{}, // Empty hosts
		}

		ctx := context.Background()
		resp, err := generatorService.GenerateExperimentConfig(ctx, req)
		
		// Should handle gracefully
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

// Helper functions

func createGenerateConfigHandler(logger *zap.Logger, generatorService *generator.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req generator.GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		resp, err := generatorService.GenerateExperimentConfig(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func createListTemplatesHandler(logger *zap.Logger, generatorService *generator.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		templates := generatorService.ListTemplates()
		
		type templateResponse struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		
		response := struct {
			Templates []templateResponse `json:"templates"`
		}{
			Templates: make([]templateResponse, len(templates)),
		}
		
		for i, tmpl := range templates {
			response.Templates[i] = templateResponse{
				Name:        tmpl.Name,
				Description: tmpl.Description,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}