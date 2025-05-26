package pipeline

import (
	"testing"
)

func TestValidator_ValidateYAML(t *testing.T) {
	validator := NewValidator(true)

	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errors  int
		warnings int
	}{
		{
			name: "valid minimal config",
			yaml: `
receivers:
  hostmetrics:
    scrapers:
      process:
        include:
          match_type: regexp
          names: [".*"]

processors:
  memory_limiter:
    limit_mib: 512
  phoenix/topk:
    metric_name: process.cpu.utilization
    top_k: 10

exporters:
  prometheus:
    endpoint: "0.0.0.0:8889"

service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: [memory_limiter, phoenix/topk]
      exporters: [prometheus]`,
			wantErr: false,
			errors:  0,
			warnings: 0,
		},
		{
			name: "missing required sections",
			yaml: `
receivers:
  hostmetrics:
    scrapers:
      process: {}`,
			wantErr: false,
			errors:  3, // missing processors, exporters, service
			warnings: 0,
		},
		{
			name: "missing process scraper",
			yaml: `
receivers:
  hostmetrics:
    scrapers:
      cpu: {}

processors:
  batch: {}

exporters:
  logging: {}

service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: [batch]
      exporters: [logging]`,
			wantErr: false,
			errors:  1, // missing process scraper
			warnings: 2, // no memory_limiter, no phoenix processor
		},
		{
			name: "invalid phoenix processor config",
			yaml: `
receivers:
  hostmetrics:
    scrapers:
      process: {}

processors:
  phoenix/topk:
    # missing required fields

exporters:
  prometheus: {}

service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: [phoenix/topk]
      exporters: [prometheus]`,
			wantErr: false,
			errors:  2, // missing metric_name and top_k
			warnings: 1, // no memory_limiter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateYAML([]byte(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if err != nil {
				return
			}

			if len(result.Errors) != tt.errors {
				t.Errorf("Expected %d errors, got %d", tt.errors, len(result.Errors))
				for _, e := range result.Errors {
					t.Logf("Error: %s - %s", e.Path, e.Message)
				}
			}

			if len(result.Warnings) != tt.warnings {
				t.Errorf("Expected %d warnings, got %d", tt.warnings, len(result.Warnings))
				for _, w := range result.Warnings {
					t.Logf("Warning: %s - %s", w.Path, w.Message)
				}
			}
		})
	}
}

func TestValidator_validateTopKProcessor(t *testing.T) {
	validator := NewValidator(true)
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	tests := []struct {
		name   string
		config map[string]interface{}
		errors int
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"metric_name": "process.cpu.utilization",
				"top_k":       10,
			},
			errors: 0,
		},
		{
			name: "missing metric_name",
			config: map[string]interface{}{
				"top_k": 10,
			},
			errors: 1,
		},
		{
			name: "missing top_k",
			config: map[string]interface{}{
				"metric_name": "process.cpu.utilization",
			},
			errors: 1,
		},
		{
			name: "invalid top_k",
			config: map[string]interface{}{
				"metric_name": "process.cpu.utilization",
				"top_k":       0,
			},
			errors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result.Errors = []ValidationError{}
			validator.validateTopKProcessor(tt.config, result)
			
			if len(result.Errors) != tt.errors {
				t.Errorf("Expected %d errors, got %d", tt.errors, len(result.Errors))
			}
		})
	}
}