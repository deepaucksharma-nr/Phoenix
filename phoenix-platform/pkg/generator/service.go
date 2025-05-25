package generator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Service handles configuration generation for Phoenix experiments
type Service struct {
	logger         *zap.Logger
	gitRepoURL     string
	gitToken       string
	templateEngine *TemplateEngine
}

// NewService creates a new generator service
func NewService(logger *zap.Logger, gitRepoURL, gitToken string) *Service {
	// Initialize template engine
	templateEngine, err := NewTemplateEngine(logger, "pipelines/templates")
	if err != nil {
		logger.Error("failed to initialize template engine, using basic generation", zap.Error(err))
		// Continue without templates for backward compatibility
	}

	return &Service{
		logger:         logger,
		gitRepoURL:     gitRepoURL,
		gitToken:       gitToken,
		templateEngine: templateEngine,
	}
}

// GenerateExperimentConfig generates all configurations for an experiment
func (s *Service) GenerateExperimentConfig(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	s.logger.Info("generating experiment configuration",
		zap.String("experiment_id", req.ExperimentID),
		zap.String("baseline", req.BaselinePipeline),
		zap.String("candidate", req.CandidatePipeline),
	)

	// Generate OTel collector configurations
	baselineConfig, err := s.generateCollectorConfig(req.BaselinePipeline, req.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to generate baseline config: %w", err)
	}

	candidateConfig, err := s.generateCollectorConfig(req.CandidatePipeline, req.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to generate candidate config: %w", err)
	}

	// Generate Kubernetes manifests
	manifests, err := s.generateKubernetesManifests(req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate manifests: %w", err)
	}

	return &GenerateResponse{
		ExperimentID:        req.ExperimentID,
		BaselineConfig:      baselineConfig,
		CandidateConfig:     candidateConfig,
		KubernetesManifests: manifests,
		GitBranch:           fmt.Sprintf("experiment/%s", req.ExperimentID),
	}, nil
}

// GeneratePipelineConfig generates OTel collector configuration
func (s *Service) GeneratePipelineConfig(ctx context.Context, pipelineName string, params map[string]interface{}) (string, error) {
	variables := make(map[string]string)
	for k, v := range params {
		variables[k] = fmt.Sprintf("%v", v)
	}
	return s.generateCollectorConfig(pipelineName, variables)
}

// generateCollectorConfig generates OpenTelemetry collector configuration
func (s *Service) generateCollectorConfig(pipeline string, variables map[string]string) (string, error) {
	// Try to use template engine first
	if s.templateEngine != nil {
		// Check if this is a template name
		config, err := s.templateEngine.GenerateConfig(pipeline, variables)
		if err == nil {
			s.logger.Info("generated config from template",
				zap.String("template", pipeline),
				zap.Int("variables", len(variables)),
			)
			return config, nil
		}
		// If not found as template, continue with basic generation
		s.logger.Debug("template not found, using basic generation",
			zap.String("pipeline", pipeline),
			zap.Error(err),
		)
	}

	// Fallback to basic generation for backward compatibility
	// This handles cases where pipeline is a description rather than template name
	return s.generateBasicConfig(pipeline, variables)
}

// generateBasicConfig generates a basic configuration when no template is available
func (s *Service) generateBasicConfig(pipeline string, variables map[string]string) (string, error) {
	// Generate YAML configuration
	collectorConfig := map[string]interface{}{
		"receivers": map[string]interface{}{
			"hostmetrics": map[string]interface{}{
				"collection_interval": "30s",
				"scrapers": map[string]interface{}{
					"process": map[string]interface{}{
						"include": map[string]interface{}{
							"match_type": "regexp",
							"names":      []string{".*"},
						},
					},
				},
			},
		},
		"processors": s.generateProcessors(pipeline),
		"exporters": map[string]interface{}{
			"otlphttp": map[string]interface{}{
				"endpoint": "${NEW_RELIC_OTLP_ENDPOINT}",
				"headers": map[string]string{
					"api-key": "${NEW_RELIC_API_KEY}",
				},
			},
			"prometheus": map[string]interface{}{
				"endpoint": "0.0.0.0:8888",
			},
		},
		"service": map[string]interface{}{
			"pipelines": map[string]interface{}{
				"metrics": map[string]interface{}{
					"receivers":  []string{"hostmetrics"},
					"processors": s.getProcessorNames(pipeline),
					"exporters":  []string{"otlphttp", "prometheus"},
				},
			},
		},
	}

	yamlData, err := yaml.Marshal(collectorConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	// Apply variables
	result := string(yamlData)
	for k, v := range variables {
		result = strings.ReplaceAll(result, fmt.Sprintf("${%s}", k), v)
	}

	return result, nil
}

// generateProcessors generates processor configuration based on pipeline
func (s *Service) generateProcessors(pipeline string) map[string]interface{} {
	processors := make(map[string]interface{})

	// Always include batch processor
	processors["batch"] = map[string]interface{}{
		"timeout":         "10s",
		"send_batch_size": 1024,
	}

	// Add pipeline-specific processors
	if strings.Contains(pipeline, "filter") {
		processors["filter"] = map[string]interface{}{
			"metrics": map[string]interface{}{
				"include": map[string]interface{}{
					"match_type":   "regexp",
					"metric_names": []string{"process_.*"},
				},
			},
		}
	}

	if strings.Contains(pipeline, "aggregate") {
		processors["groupbyattrs"] = map[string]interface{}{
			"keys": []string{"process.name", "host.name"},
		}
	}

	if strings.Contains(pipeline, "sample") {
		processors["probabilistic_sampler"] = map[string]interface{}{
			"sampling_percentage": 10,
		}
	}

	return processors
}

// getProcessorNames returns the list of processor names for the pipeline
func (s *Service) getProcessorNames(pipeline string) []string {
	names := []string{"batch"}

	if strings.Contains(pipeline, "filter") {
		names = append(names, "filter")
	}
	if strings.Contains(pipeline, "aggregate") {
		names = append(names, "groupbyattrs")
	}
	if strings.Contains(pipeline, "sample") {
		names = append(names, "probabilistic_sampler")
	}

	return names
}

// parsePipelineTemplate parses and returns the base pipeline configuration
func (s *Service) parsePipelineTemplate(pipeline string) string {
	// In a real implementation, this would parse the visual pipeline
	// For now, return the pipeline name as-is
	return pipeline
}

// generateKubernetesManifests generates K8s manifests for the experiment
func (s *Service) generateKubernetesManifests(req *GenerateRequest) (map[string]string, error) {
	manifests := make(map[string]string)

	// Generate namespace
	namespace := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: phoenix-experiment-%s
  labels:
    phoenix.io/experiment-id: %s
`, req.ExperimentID, req.ExperimentID)
	manifests["namespace.yaml"] = namespace

	// Generate baseline collector deployment
	baselineDeployment := s.generateCollectorDeployment("baseline", req.ExperimentID, req.TargetHosts)
	manifests["baseline-deployment.yaml"] = baselineDeployment

	// Generate candidate collector deployment
	candidateDeployment := s.generateCollectorDeployment("candidate", req.ExperimentID, req.TargetHosts)
	manifests["candidate-deployment.yaml"] = candidateDeployment

	// Generate ConfigMaps
	baselineConfigMap := s.generateConfigMap("baseline", req.ExperimentID, req.BaselinePipeline)
	manifests["baseline-configmap.yaml"] = baselineConfigMap

	candidateConfigMap := s.generateConfigMap("candidate", req.ExperimentID, req.CandidatePipeline)
	manifests["candidate-configmap.yaml"] = candidateConfigMap

	// Generate services
	services := s.generateServices(req.ExperimentID)
	manifests["services.yaml"] = services

	// Generate NetworkPolicy
	networkPolicy := s.generateNetworkPolicy(req.ExperimentID)
	manifests["network-policy.yaml"] = networkPolicy

	return manifests, nil
}

// generateConfigMap generates a ConfigMap for collector configuration
func (s *Service) generateConfigMap(variant string, experimentID string, configData string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: otel-collector-config-%s
  namespace: phoenix-experiment-%s
data:
  config.yaml: |
%s
`, variant, experimentID, s.indentString(configData, "    "))
}

// indentString indents each line of a string
func (s *Service) indentString(str string, indent string) string {
	lines := strings.Split(str, "\n")
	for i := range lines {
		if lines[i] != "" {
			lines[i] = indent + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}

// generateCollectorDeployment generates a deployment manifest for a collector
func (s *Service) generateCollectorDeployment(variant string, experimentID string, targetHosts []string) string {
	return fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: otel-collector-%s
  namespace: phoenix-experiment-%s
  labels:
    app: otel-collector
    variant: %s
    phoenix.io/experiment-id: %s
spec:
  replicas: 1
  selector:
    matchLabels:
      app: otel-collector
      variant: %s
  template:
    metadata:
      labels:
        app: otel-collector
        variant: %s
    spec:
      containers:
      - name: otel-collector
        image: otel/opentelemetry-collector:0.88.0
        args:
          - --config=/etc/otel-collector-config.yaml
        ports:
        - containerPort: 8888
          name: metrics
        - containerPort: 8889
          name: prometheus
        volumeMounts:
        - name: config
          mountPath: /etc/otel-collector-config.yaml
          subPath: config.yaml
        env:
        - name: OTLP_ENDPOINT
          valueFrom:
            secretKeyRef:
              name: phoenix-secrets
              key: otlp-endpoint
        - name: NEW_RELIC_API_KEY
          valueFrom:
            secretKeyRef:
              name: phoenix-secrets
              key: new-relic-api-key
      volumes:
      - name: config
        configMap:
          name: otel-collector-config-%s
`, variant, experimentID, variant, experimentID, variant, variant, variant)
}

// generateServices generates service manifests
func (s *Service) generateServices(experimentID string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: otel-collector-baseline
  namespace: phoenix-experiment-%s
spec:
  selector:
    app: otel-collector
    variant: baseline
  ports:
  - name: metrics
    port: 8888
  - name: prometheus
    port: 8889
---
apiVersion: v1
kind: Service
metadata:
  name: otel-collector-candidate
  namespace: phoenix-experiment-%s
spec:
  selector:
    app: otel-collector
    variant: candidate
  ports:
  - name: metrics
    port: 8888
  - name: prometheus
    port: 8889
`, experimentID, experimentID)
}

// generateNetworkPolicy generates network policy for the experiment
func (s *Service) generateNetworkPolicy(experimentID string) string {
	return fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: otel-collector-network-policy
  namespace: phoenix-experiment-%s
spec:
  podSelector:
    matchLabels:
      app: otel-collector
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: prometheus
    ports:
    - protocol: TCP
      port: 8889
  egress:
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: TCP
      port: 9100
`, experimentID)
}

// ListTemplates returns all available pipeline templates
func (s *Service) ListTemplates() []*PipelineTemplate {
	if s.templateEngine == nil {
		return []*PipelineTemplate{}
	}
	return s.templateEngine.ListTemplates()
}

// GetTemplate returns a specific pipeline template
func (s *Service) GetTemplate(name string) (*PipelineTemplate, error) {
	if s.templateEngine == nil {
		return nil, fmt.Errorf("template engine not initialized")
	}
	return s.templateEngine.GetTemplate(name)
}

// CreateGitPR creates a pull request with generated configs
func (s *Service) CreateGitPR(ctx context.Context, branch string, files map[string]string) error {
	// In a real implementation, this would:
	// 1. Clone the repository
	// 2. Create a new branch
	// 3. Write all files
	// 4. Commit and push
	// 5. Create a pull request
	s.logger.Info("creating git pull request",
		zap.String("branch", branch),
		zap.Int("file_count", len(files)),
	)
	return nil
}

// Request/Response types
type GenerateRequest struct {
	ExperimentID      string            `json:"experiment_id"`
	BaselinePipeline  string            `json:"baseline_pipeline"`
	CandidatePipeline string            `json:"candidate_pipeline"`
	TargetHosts       []string          `json:"target_hosts"`
	Variables         map[string]string `json:"variables"`
	Duration          time.Duration     `json:"duration"`
}

type GenerateResponse struct {
	ExperimentID        string            `json:"experiment_id"`
	BaselineConfig      string            `json:"baseline_config"`
	CandidateConfig     string            `json:"candidate_config"`
	KubernetesManifests map[string]string `json:"kubernetes_manifests"`
	GitBranch           string            `json:"git_branch"`
}