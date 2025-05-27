package services

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"gopkg.in/yaml.v3"
)

// PipelineTemplateRenderer handles rendering of pipeline templates
type PipelineTemplateRenderer struct {
	templates map[string]*template.Template
}

// TemplateData represents the data passed to pipeline templates
type TemplateData struct {
	ExperimentID string
	Variant      string
	HostID       string
	Config       map[string]interface{}
	Metrics      *models.KPIResult
	Thresholds   map[string]float64
}

// ProcessorConfig represents a processor configuration
type ProcessorConfig struct {
	Name   string                 `yaml:"name"`
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

// PipelineConfig represents a complete pipeline configuration
type PipelineConfig struct {
	Receivers  map[string]interface{} `yaml:"receivers"`
	Processors []ProcessorConfig      `yaml:"processors"`
	Exporters  map[string]interface{} `yaml:"exporters"`
	Service    ServiceConfig          `yaml:"service"`
}

// ServiceConfig represents the service configuration section
type ServiceConfig struct {
	Pipelines map[string]PipelineService `yaml:"pipelines"`
}

// PipelineService represents a pipeline service configuration
type PipelineService struct {
	Receivers  []string `yaml:"receivers"`
	Processors []string `yaml:"processors"`
	Exporters  []string `yaml:"exporters"`
}

func NewPipelineTemplateRenderer() *PipelineTemplateRenderer {
	return &PipelineTemplateRenderer{
		templates: make(map[string]*template.Template),
	}
}

// LoadTemplate loads a pipeline template from string
func (ptr *PipelineTemplateRenderer) LoadTemplate(name, templateStr string) error {
	tmpl, err := template.New(name).Funcs(sprig.TxtFuncMap()).Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	ptr.templates[name] = tmpl
	return nil
}

// RenderTemplate renders a pipeline template with the given data
func (ptr *PipelineTemplateRenderer) RenderTemplate(ctx context.Context, templateName string, data TemplateData) (string, error) {
	tmpl, exists := ptr.templates[templateName]
	if !exists {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// GenerateOptimizedPipeline generates an optimized pipeline based on KPIs
func (ptr *PipelineTemplateRenderer) GenerateOptimizedPipeline(ctx context.Context, experiment *models.Experiment, kpis *models.KPIResult) (*PipelineConfig, error) {
	config := &PipelineConfig{
		Receivers: map[string]interface{}{
			"otlp": map[string]interface{}{
				"protocols": map[string]interface{}{
					"grpc": map[string]interface{}{
						"endpoint": "0.0.0.0:4317",
					},
					"http": map[string]interface{}{
						"endpoint": "0.0.0.0:4318",
					},
				},
			},
		},
		Processors: []ProcessorConfig{},
		Exporters: map[string]interface{}{
			"prometheus": map[string]interface{}{
				"endpoint": "0.0.0.0:8889",
				"namespace": "phoenix",
				"const_labels": map[string]string{
					"experiment_id": experiment.ID,
				},
			},
		},
		Service: ServiceConfig{
			Pipelines: map[string]PipelineService{
				"metrics": {
					Receivers:  []string{"otlp"},
					Processors: []string{},
					Exporters:  []string{"prometheus"},
				},
			},
		},
	}

	// Add processors based on experiment configuration and KPIs
	processors := ptr.selectProcessors(experiment, kpis)
	config.Processors = processors

	// Update service pipeline with processor names
	processorNames := make([]string, len(processors))
	for i, p := range processors {
		processorNames[i] = p.Name
	}
	// Get the pipeline, update it, and put it back
	pipeline := config.Service.Pipelines["metrics"]
	pipeline.Processors = processorNames
	config.Service.Pipelines["metrics"] = pipeline

	return config, nil
}

// selectProcessors selects appropriate processors based on experiment and KPIs
func (ptr *PipelineTemplateRenderer) selectProcessors(experiment *models.Experiment, kpis *models.KPIResult) []ProcessorConfig {
	processors := []ProcessorConfig{}

	// Always add batch processor for performance
	processors = append(processors, ProcessorConfig{
		Name: "batch",
		Type: "batch",
		Config: map[string]interface{}{
			"timeout":   "1s",
			"send_batch_size": 1024,
		},
	})

	// Add memory limiter to prevent OOM
	processors = append(processors, ProcessorConfig{
		Name: "memory_limiter",
		Type: "memory_limiter",
		Config: map[string]interface{}{
			"check_interval": "1s",
			"limit_mib":      512,
			"spike_limit_mib": 128,
		},
	})

	// Add filtering based on experiment metadata
	if experiment.Metadata != nil {
		// Check for TopK processor
		if topkConfig, ok := experiment.Metadata["topk"]; ok {
			if config, ok := topkConfig.(map[string]interface{}); ok {
				processors = append(processors, ProcessorConfig{
					Name: "topk",
					Type: "topk",
					Config: config,
				})
			}
		}

		// Check for Adaptive Filter processor
		if afConfig, ok := experiment.Metadata["adaptive_filter"]; ok {
			if config, ok := afConfig.(map[string]interface{}); ok {
				processors = append(processors, ProcessorConfig{
					Name: "adaptive_filter",
					Type: "adaptive_filter",
					Config: config,
				})
			}
		}

		// Check for custom filtering rules
		if filterConfig, ok := experiment.Metadata["filter"]; ok {
			if config, ok := filterConfig.(map[string]interface{}); ok {
				processors = append(processors, ProcessorConfig{
					Name: "filter",
					Type: "filter",
					Config: config,
				})
			}
		}
	}

	// Add resource processor for labeling
	processors = append(processors, ProcessorConfig{
		Name: "resource",
		Type: "resource",
		Config: map[string]interface{}{
			"attributes": []map[string]interface{}{
				{
					"key":    "experiment_id",
					"value":  experiment.ID,
					"action": "insert",
				},
				{
					"key":    "variant",
					"value":  "candidate",
					"action": "insert",
				},
			},
		},
	})

	return processors
}

// RenderPipelineYAML renders a pipeline configuration to YAML
func (ptr *PipelineTemplateRenderer) RenderPipelineYAML(config *PipelineConfig) (string, error) {
	// Convert to a format suitable for YAML marshaling
	yamlConfig := map[string]interface{}{
		"receivers":  config.Receivers,
		"processors": map[string]interface{}{},
		"exporters":  config.Exporters,
		"service":    config.Service,
	}

	// Convert processors array to map
	for _, proc := range config.Processors {
		yamlConfig["processors"].(map[string]interface{})[proc.Name] = proc.Config
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(yamlConfig); err != nil {
		return "", fmt.Errorf("failed to encode pipeline config: %w", err)
	}

	return buf.String(), nil
}

// ValidatePipelineConfig validates a pipeline configuration
func (ptr *PipelineTemplateRenderer) ValidatePipelineConfig(config *PipelineConfig) error {
	// Check required components
	if len(config.Receivers) == 0 {
		return fmt.Errorf("pipeline must have at least one receiver")
	}

	if len(config.Exporters) == 0 {
		return fmt.Errorf("pipeline must have at least one exporter")
	}

	if len(config.Service.Pipelines) == 0 {
		return fmt.Errorf("pipeline must have at least one service pipeline")
	}

	// Validate service pipelines
	for name, pipeline := range config.Service.Pipelines {
		if len(pipeline.Receivers) == 0 {
			return fmt.Errorf("pipeline %s must have at least one receiver", name)
		}

		if len(pipeline.Exporters) == 0 {
			return fmt.Errorf("pipeline %s must have at least one exporter", name)
		}

		// Check that referenced components exist
		for _, receiver := range pipeline.Receivers {
			if _, exists := config.Receivers[receiver]; !exists {
				return fmt.Errorf("pipeline %s references undefined receiver: %s", name, receiver)
			}
		}

		for _, processor := range pipeline.Processors {
			found := false
			for _, p := range config.Processors {
				if p.Name == processor {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("pipeline %s references undefined processor: %s", name, processor)
			}
		}

		for _, exporter := range pipeline.Exporters {
			if _, exists := config.Exporters[exporter]; !exists {
				return fmt.Errorf("pipeline %s references undefined exporter: %s", name, exporter)
			}
		}
	}

	return nil
}

// GetBuiltinTemplates returns a map of built-in pipeline templates
func (ptr *PipelineTemplateRenderer) GetBuiltinTemplates() map[string]string {
	return map[string]string{
		"baseline": baselinePipelineTemplate,
		"topk":     topkPipelineTemplate,
		"adaptive": adaptiveFilterPipelineTemplate,
		"hybrid":   hybridPipelineTemplate,
	}
}

// Built-in pipeline templates
const baselinePipelineTemplate = `
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
  
  memory_limiter:
    check_interval: 1s
    limit_mib: 512
    spike_limit_mib: 128
  
  resource:
    attributes:
      - key: experiment_id
        value: "{{ .ExperimentID }}"
        action: insert
      - key: variant
        value: "{{ .Variant }}"
        action: insert
      - key: host_id
        value: "{{ .HostID }}"
        action: insert

exporters:
  prometheus:
    endpoint: 0.0.0.0:8889
    namespace: phoenix
    const_labels:
      experiment_id: "{{ .ExperimentID }}"
      variant: "{{ .Variant }}"

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [batch, memory_limiter, resource]
      exporters: [prometheus]
`

const topkPipelineTemplate = `
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
  
  memory_limiter:
    check_interval: 1s
    limit_mib: 512
    spike_limit_mib: 128
  
  topk:
    k: {{ .Config.k | default 100 }}
    metric_names:
      {{- range .Config.metric_names }}
      - {{ . }}
      {{- end }}
    group_by_keys:
      {{- range .Config.group_by_keys }}
      - {{ . }}
      {{- end }}
  
  resource:
    attributes:
      - key: experiment_id
        value: "{{ .ExperimentID }}"
        action: insert
      - key: variant
        value: "{{ .Variant }}"
        action: insert

exporters:
  prometheus:
    endpoint: 0.0.0.0:8889
    namespace: phoenix
    const_labels:
      experiment_id: "{{ .ExperimentID }}"
      variant: "{{ .Variant }}"

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [batch, memory_limiter, topk, resource]
      exporters: [prometheus]
`

const adaptiveFilterPipelineTemplate = `
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
  
  memory_limiter:
    check_interval: 1s
    limit_mib: 512
    spike_limit_mib: 128
  
  adaptive_filter:
    threshold: {{ .Config.threshold | default 0.9 }}
    min_cardinality: {{ .Config.min_cardinality | default 100 }}
    retention_period: {{ .Config.retention_period | default "5m" }}
    critical_metrics:
      {{- range .Config.critical_metrics }}
      - {{ . }}
      {{- end }}
  
  resource:
    attributes:
      - key: experiment_id
        value: "{{ .ExperimentID }}"
        action: insert
      - key: variant
        value: "{{ .Variant }}"
        action: insert

exporters:
  prometheus:
    endpoint: 0.0.0.0:8889
    namespace: phoenix
    const_labels:
      experiment_id: "{{ .ExperimentID }}"
      variant: "{{ .Variant }}"

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [batch, memory_limiter, adaptive_filter, resource]
      exporters: [prometheus]
`

const hybridPipelineTemplate = `
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
  
  memory_limiter:
    check_interval: 1s
    limit_mib: 512
    spike_limit_mib: 128
  
  filter:
    metrics:
      include:
        match_type: regexp
        metric_names:
          {{- range .Config.include_patterns }}
          - {{ . }}
          {{- end }}
      exclude:
        match_type: regexp
        metric_names:
          {{- range .Config.exclude_patterns }}
          - {{ . }}
          {{- end }}
  
  topk:
    k: {{ .Config.topk.k | default 50 }}
    metric_names:
      {{- range .Config.topk.metric_names }}
      - {{ . }}
      {{- end }}
  
  resource:
    attributes:
      - key: experiment_id
        value: "{{ .ExperimentID }}"
        action: insert
      - key: variant
        value: "{{ .Variant }}"
        action: insert

exporters:
  prometheus:
    endpoint: 0.0.0.0:8889
    namespace: phoenix
    const_labels:
      experiment_id: "{{ .ExperimentID }}"
      variant: "{{ .Variant }}"

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [batch, memory_limiter, filter, topk, resource]
      exporters: [prometheus]
`