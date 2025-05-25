package generator

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func TestGenerateCollectorConfig(t *testing.T) {
	logger := zap.NewNop()
	service := NewService(logger, "https://github.com/test/repo", "test-token")

	tests := []struct {
		name      string
		pipeline  string
		variables map[string]string
		validate  func(t *testing.T, config string)
	}{
		{
			name:     "basic pipeline generation",
			pipeline: "process-baseline-v1",
			variables: map[string]string{
				"NEW_RELIC_API_KEY":       "test-key-123",
				"NEW_RELIC_OTLP_ENDPOINT": "https://otlp.nr-data.net:4317",
			},
			validate: func(t *testing.T, config string) {
				assert.Contains(t, config, "receivers:")
				assert.Contains(t, config, "hostmetrics:")
				assert.Contains(t, config, "processors:")
				assert.Contains(t, config, "exporters:")
				assert.Contains(t, config, "service:")
				assert.Contains(t, config, "test-key-123")
				assert.Contains(t, config, "https://otlp.nr-data.net:4317")
				
				// Validate it's proper YAML
				var parsed map[string]interface{}
				err := yaml.Unmarshal([]byte(config), &parsed)
				require.NoError(t, err)
				
				// Check required sections
				assert.Contains(t, parsed, "receivers")
				assert.Contains(t, parsed, "processors")
				assert.Contains(t, parsed, "exporters")
				assert.Contains(t, parsed, "service")
			},
		},
		{
			name:     "filter pipeline",
			pipeline: "process-priority-filter",
			variables: map[string]string{
				"NEW_RELIC_API_KEY": "test-key-456",
			},
			validate: func(t *testing.T, config string) {
				assert.Contains(t, config, "filter:")
				assert.Contains(t, config, "test-key-456")
				
				// Should contain filter processor
				var parsed map[string]interface{}
				err := yaml.Unmarshal([]byte(config), &parsed)
				require.NoError(t, err)
				
				processors, ok := parsed["processors"].(map[string]interface{})
				require.True(t, ok)
				
				// Should have some filter-related processor
				found := false
				for key := range processors {
					if strings.Contains(key, "filter") {
						found = true
						break
					}
				}
				assert.True(t, found, "Should contain a filter processor")
			},
		},
		{
			name:     "aggregate pipeline",
			pipeline: "process-aggregate-v1",
			variables: map[string]string{},
			validate: func(t *testing.T, config string) {
				assert.Contains(t, config, "groupbyattrs")
				
				var parsed map[string]interface{}
				err := yaml.Unmarshal([]byte(config), &parsed)
				require.NoError(t, err)
				
				processors, ok := parsed["processors"].(map[string]interface{})
				require.True(t, ok)
				
				// Should have groupbyattrs processor
				found := false
				for key := range processors {
					if strings.Contains(key, "groupbyattrs") {
						found = true
						break
					}
				}
				assert.True(t, found, "Should contain a groupbyattrs processor")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := service.generateCollectorConfig(tt.pipeline, tt.variables)
			require.NoError(t, err)
			assert.NotEmpty(t, config)
			
			tt.validate(t, config)
		})
	}
}

func TestGenerateExperimentConfig(t *testing.T) {
	logger := zap.NewNop()
	service := NewService(logger, "https://github.com/test/repo", "test-token")

	req := &GenerateRequest{
		ExperimentID:      "test-exp-123",
		BaselinePipeline:  "process-baseline-v1",
		CandidatePipeline: "process-priority-filter-v1",
		TargetHosts:       []string{"host1", "host2"},
		Variables: map[string]string{
			"NEW_RELIC_API_KEY":       "test-key-789",
			"NEW_RELIC_OTLP_ENDPOINT": "https://otlp.nr-data.net:4317",
		},
	}

	resp, err := service.GenerateExperimentConfig(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Validate response structure
	assert.Equal(t, req.ExperimentID, resp.ExperimentID)
	assert.NotEmpty(t, resp.BaselineConfig)
	assert.NotEmpty(t, resp.CandidateConfig)
	assert.NotEmpty(t, resp.KubernetesManifests)
	assert.Equal(t, "experiment/test-exp-123", resp.GitBranch)

	// Validate baseline config
	var baselineYAML map[string]interface{}
	err = yaml.Unmarshal([]byte(resp.BaselineConfig), &baselineYAML)
	require.NoError(t, err)
	assert.Contains(t, baselineYAML, "receivers")
	assert.Contains(t, baselineYAML, "exporters")

	// Validate candidate config  
	var candidateYAML map[string]interface{}
	err = yaml.Unmarshal([]byte(resp.CandidateConfig), &candidateYAML)
	require.NoError(t, err)
	assert.Contains(t, candidateYAML, "receivers")
	assert.Contains(t, candidateYAML, "exporters")

	// Validate Kubernetes manifests
	assert.Contains(t, resp.KubernetesManifests, "namespace.yaml")
	assert.Contains(t, resp.KubernetesManifests, "baseline-deployment.yaml")
	assert.Contains(t, resp.KubernetesManifests, "candidate-deployment.yaml")
	assert.Contains(t, resp.KubernetesManifests, "baseline-configmap.yaml")
	assert.Contains(t, resp.KubernetesManifests, "candidate-configmap.yaml")

	// Check namespace manifest contains experiment ID
	namespaceManifest := resp.KubernetesManifests["namespace.yaml"]
	assert.Contains(t, namespaceManifest, "phoenix-experiment-test-exp-123")
	assert.Contains(t, namespaceManifest, "phoenix.io/experiment-id: test-exp-123")

	// Check deployment manifests
	baselineDeployment := resp.KubernetesManifests["baseline-deployment.yaml"]
	assert.Contains(t, baselineDeployment, "otel-collector-baseline")
	assert.Contains(t, baselineDeployment, "variant: baseline")

	candidateDeployment := resp.KubernetesManifests["candidate-deployment.yaml"]
	assert.Contains(t, candidateDeployment, "otel-collector-candidate")
	assert.Contains(t, candidateDeployment, "variant: candidate")
}

func TestGenerateProcessors(t *testing.T) {
	logger := zap.NewNop()
	service := NewService(logger, "https://github.com/test/repo", "test-token")

	tests := []struct {
		name     string
		pipeline string
		expected []string
	}{
		{
			name:     "basic pipeline",
			pipeline: "basic",
			expected: []string{"batch"},
		},
		{
			name:     "filter pipeline",
			pipeline: "priority-filter",
			expected: []string{"batch", "filter"},
		},
		{
			name:     "aggregate pipeline",
			pipeline: "group-aggregate",
			expected: []string{"batch", "groupbyattrs"},
		},
		{
			name:     "sampling pipeline",
			pipeline: "sample-reduce",
			expected: []string{"batch", "probabilistic_sampler"},
		},
		{
			name:     "complex pipeline",
			pipeline: "filter-aggregate-sample",
			expected: []string{"batch", "filter", "groupbyattrs", "probabilistic_sampler"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processors := service.generateProcessors(tt.pipeline)
			processorNames := service.getProcessorNames(tt.pipeline)

			// Check that expected processors are present
			for _, expectedName := range tt.expected {
				found := false
				for processorName := range processors {
					if strings.Contains(processorName, expectedName) || processorName == expectedName {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected processor %s not found", expectedName)

				// Check processor names list
				assert.Contains(t, processorNames, expectedName)
			}

			// Batch processor should always be present
			assert.Contains(t, processorNames, "batch")
		})
	}
}

func TestGenerateKubernetesManifests(t *testing.T) {
	logger := zap.NewNop()
	service := NewService(logger, "https://github.com/test/repo", "test-token")

	req := &GenerateRequest{
		ExperimentID:      "k8s-test-456",
		BaselinePipeline:  "baseline",
		CandidatePipeline: "candidate",
		TargetHosts:       []string{"node1", "node2", "node3"},
	}

	manifests, err := service.generateKubernetesManifests(req)
	require.NoError(t, err)
	assert.NotEmpty(t, manifests)

	// Check all expected manifests are present
	expectedManifests := []string{
		"namespace.yaml",
		"baseline-deployment.yaml", 
		"candidate-deployment.yaml",
		"baseline-configmap.yaml",
		"candidate-configmap.yaml",
		"services.yaml",
		"network-policy.yaml",
	}

	for _, expected := range expectedManifests {
		assert.Contains(t, manifests, expected, "Missing manifest: %s", expected)
		assert.NotEmpty(t, manifests[expected], "Empty manifest: %s", expected)
	}

	// Validate namespace manifest
	namespace := manifests["namespace.yaml"]
	assert.Contains(t, namespace, "phoenix-experiment-k8s-test-456")
	assert.Contains(t, namespace, "phoenix.io/experiment-id: k8s-test-456")

	// Validate deployment manifests have correct structure
	baselineDeployment := manifests["baseline-deployment.yaml"]
	assert.Contains(t, baselineDeployment, "kind: Deployment")
	assert.Contains(t, baselineDeployment, "otel-collector-baseline")
	assert.Contains(t, baselineDeployment, "variant: baseline")
	assert.Contains(t, baselineDeployment, "otel/opentelemetry-collector")

	candidateDeployment := manifests["candidate-deployment.yaml"]
	assert.Contains(t, candidateDeployment, "kind: Deployment")
	assert.Contains(t, candidateDeployment, "otel-collector-candidate")
	assert.Contains(t, candidateDeployment, "variant: candidate")

	// Validate services manifest
	services := manifests["services.yaml"]
	assert.Contains(t, services, "otel-collector-baseline")
	assert.Contains(t, services, "otel-collector-candidate")
	assert.Contains(t, services, "port: 8888")
	assert.Contains(t, services, "port: 8889")

	// Validate network policy
	networkPolicy := manifests["network-policy.yaml"]
	assert.Contains(t, networkPolicy, "kind: NetworkPolicy")
	assert.Contains(t, networkPolicy, "otel-collector-network-policy")
}

func TestListTemplates(t *testing.T) {
	logger := zap.NewNop()
	service := NewService(logger, "https://github.com/test/repo", "test-token")

	templates := service.ListTemplates()
	// Should return empty list if template engine is not initialized with real templates
	// This is expected behavior for the current implementation
	assert.NotNil(t, templates)
}

func TestCreateGitPR(t *testing.T) {
	logger := zap.NewNop()
	service := NewService(logger, "https://github.com/test/repo", "test-token")

	// This is a stub implementation, so it should not error
	files := map[string]string{
		"test.yaml":   "test: value",
		"config.yaml": "config: data",
	}

	err := service.CreateGitPR(context.Background(), "test-branch", files)
	assert.NoError(t, err) // Current implementation is a stub, so should not error
}