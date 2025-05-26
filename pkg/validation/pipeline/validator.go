package pipeline

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Validator validates OTel collector pipeline configurations
type Validator struct {
	strict bool
}

// NewValidator creates a new pipeline validator
func NewValidator(strict bool) *Validator {
	return &Validator{
		strict: strict,
	}
}

// ValidationResult contains the results of validation
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationWarning
}

// ValidationError represents a validation error
type ValidationError struct {
	Path    string
	Message string
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Path    string
	Message string
}

// ValidateYAML validates a pipeline configuration from YAML bytes
func (v *Validator) ValidateYAML(content []byte) (*ValidationResult, error) {
	var config map[string]interface{}
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("invalid YAML syntax: %w", err)
	}

	return v.Validate(config)
}

// Validate validates a pipeline configuration
func (v *Validator) Validate(config map[string]interface{}) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	// Validate required top-level sections
	v.validateRequiredSections(config, result)

	// Validate service section
	if service, ok := config["service"].(map[string]interface{}); ok {
		v.validateService(service, result)
	}

	// Validate receivers
	if receivers, ok := config["receivers"].(map[string]interface{}); ok {
		v.validateReceivers(receivers, result)
	}

	// Validate processors
	if processors, ok := config["processors"].(map[string]interface{}); ok {
		v.validateProcessors(processors, result)
	}

	// Validate exporters
	if exporters, ok := config["exporters"].(map[string]interface{}); ok {
		v.validateExporters(exporters, result)
	}

	// Validate extensions if present
	if extensions, ok := config["extensions"].(map[string]interface{}); ok {
		v.validateExtensions(extensions, result)
	}

	// Set valid flag based on errors
	result.Valid = len(result.Errors) == 0

	return result, nil
}

// validateRequiredSections checks for required top-level sections
func (v *Validator) validateRequiredSections(config map[string]interface{}, result *ValidationResult) {
	requiredSections := []string{"receivers", "processors", "exporters", "service"}
	
	for _, section := range requiredSections {
		if _, ok := config[section]; !ok {
			result.Errors = append(result.Errors, ValidationError{
				Path:    section,
				Message: fmt.Sprintf("missing required section: %s", section),
			})
		}
	}
}

// validateService validates the service section
func (v *Validator) validateService(service map[string]interface{}, result *ValidationResult) {
	// Check for pipelines
	pipelines, ok := service["pipelines"].(map[string]interface{})
	if !ok {
		result.Errors = append(result.Errors, ValidationError{
			Path:    "service.pipelines",
			Message: "missing or invalid pipelines section",
		})
		return
	}

	// For process metrics, we need a metrics pipeline
	if _, ok := pipelines["metrics"]; !ok {
		result.Errors = append(result.Errors, ValidationError{
			Path:    "service.pipelines.metrics",
			Message: "missing metrics pipeline (required for process metrics)",
		})
	} else {
		// Validate metrics pipeline
		if metricsPipeline, ok := pipelines["metrics"].(map[string]interface{}); ok {
			v.validatePipeline("metrics", metricsPipeline, result)
		}
	}

	// Validate telemetry section if present
	if telemetry, ok := service["telemetry"].(map[string]interface{}); ok {
		v.validateTelemetry(telemetry, result)
	}
}

// validatePipeline validates a single pipeline configuration
func (v *Validator) validatePipeline(name string, pipeline map[string]interface{}, result *ValidationResult) {
	components := []string{"receivers", "processors", "exporters"}
	
	for _, component := range components {
		if val, ok := pipeline[component]; ok {
			switch v := val.(type) {
			case []interface{}:
				if len(v) == 0 {
					result.Warnings = append(result.Warnings, ValidationWarning{
						Path:    fmt.Sprintf("service.pipelines.%s.%s", name, component),
						Message: fmt.Sprintf("empty %s list", component),
					})
				}
			case nil:
				result.Errors = append(result.Errors, ValidationError{
					Path:    fmt.Sprintf("service.pipelines.%s.%s", name, component),
					Message: fmt.Sprintf("null %s", component),
				})
			}
		} else {
			result.Errors = append(result.Errors, ValidationError{
				Path:    fmt.Sprintf("service.pipelines.%s", name),
				Message: fmt.Sprintf("missing %s", component),
			})
		}
	}
}

// validateReceivers validates the receivers section
func (v *Validator) validateReceivers(receivers map[string]interface{}, result *ValidationResult) {
	if len(receivers) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Path:    "receivers",
			Message: "no receivers defined",
		})
		return
	}

	// Check for hostmetrics receiver (required for process metrics)
	if _, ok := receivers["hostmetrics"]; !ok {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Path:    "receivers",
			Message: "missing 'hostmetrics' receiver - required for process metrics collection",
		})
	} else {
		// Validate hostmetrics configuration
		if hostmetrics, ok := receivers["hostmetrics"].(map[string]interface{}); ok {
			v.validateHostmetricsReceiver(hostmetrics, result)
		}
	}
}

// validateHostmetricsReceiver validates the hostmetrics receiver configuration
func (v *Validator) validateHostmetricsReceiver(hostmetrics map[string]interface{}, result *ValidationResult) {
	// Check for scrapers
	scrapers, ok := hostmetrics["scrapers"].(map[string]interface{})
	if !ok {
		result.Errors = append(result.Errors, ValidationError{
			Path:    "receivers.hostmetrics.scrapers",
			Message: "missing scrapers configuration",
		})
		return
	}

	// Check for process scraper
	if _, ok := scrapers["process"]; !ok {
		result.Errors = append(result.Errors, ValidationError{
			Path:    "receivers.hostmetrics.scrapers",
			Message: "missing 'process' scraper - required for process metrics",
		})
	}
}

// validateProcessors validates the processors section
func (v *Validator) validateProcessors(processors map[string]interface{}, result *ValidationResult) {
	// Check for memory_limiter (recommended)
	if _, ok := processors["memory_limiter"]; !ok {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Path:    "processors",
			Message: "missing 'memory_limiter' processor - recommended for production",
		})
	}

	// Check for Phoenix-specific processors
	phoenixProcessors := []string{"phoenix/topk", "phoenix/adaptive_filter", "phoenix/sampling"}
	hasPhoenixProcessor := false
	
	for _, proc := range phoenixProcessors {
		if _, ok := processors[proc]; ok {
			hasPhoenixProcessor = true
			// Validate Phoenix processor configuration
			v.validatePhoenixProcessor(proc, processors[proc], result)
			break
		}
	}

	if !hasPhoenixProcessor && v.strict {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Path:    "processors",
			Message: "no Phoenix-specific processors found - pipeline may not reduce cardinality",
		})
	}
}

// validatePhoenixProcessor validates Phoenix-specific processor configuration
func (v *Validator) validatePhoenixProcessor(name string, processor interface{}, result *ValidationResult) {
	procMap, ok := processor.(map[string]interface{})
	if !ok {
		return
	}

	switch name {
	case "phoenix/topk":
		v.validateTopKProcessor(procMap, result)
	case "phoenix/adaptive_filter":
		v.validateAdaptiveFilterProcessor(procMap, result)
	case "phoenix/sampling":
		v.validateSamplingProcessor(procMap, result)
	}
}

// validateTopKProcessor validates the top-k processor configuration
func (v *Validator) validateTopKProcessor(config map[string]interface{}, result *ValidationResult) {
	// Check for required fields
	if _, ok := config["metric_name"]; !ok {
		result.Errors = append(result.Errors, ValidationError{
			Path:    "processors.phoenix/topk",
			Message: "missing required field 'metric_name'",
		})
	}

	if topK, ok := config["top_k"]; ok {
		if k, ok := topK.(int); ok {
			if k <= 0 {
				result.Errors = append(result.Errors, ValidationError{
					Path:    "processors.phoenix/topk.top_k",
					Message: "top_k must be greater than 0",
				})
			}
		}
	} else {
		result.Errors = append(result.Errors, ValidationError{
			Path:    "processors.phoenix/topk",
			Message: "missing required field 'top_k'",
		})
	}
}

// validateAdaptiveFilterProcessor validates the adaptive filter processor configuration
func (v *Validator) validateAdaptiveFilterProcessor(config map[string]interface{}, result *ValidationResult) {
	// Check for base_thresholds
	if _, ok := config["base_thresholds"]; !ok {
		result.Errors = append(result.Errors, ValidationError{
			Path:    "processors.phoenix/adaptive_filter",
			Message: "missing required field 'base_thresholds'",
		})
	}
}

// validateSamplingProcessor validates the sampling processor configuration
func (v *Validator) validateSamplingProcessor(config map[string]interface{}, result *ValidationResult) {
	// Check for sampling_rate
	if rate, ok := config["sampling_rate"]; ok {
		if r, ok := rate.(float64); ok {
			if r <= 0 || r > 1 {
				result.Errors = append(result.Errors, ValidationError{
					Path:    "processors.phoenix/sampling.sampling_rate",
					Message: "sampling_rate must be between 0 and 1",
				})
			}
		}
	}
}

// validateExporters validates the exporters section
func (v *Validator) validateExporters(exporters map[string]interface{}, result *ValidationResult) {
	if len(exporters) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Path:    "exporters",
			Message: "no exporters defined",
		})
		return
	}

	// Check for at least one metrics exporter
	hasMetricsExporter := false
	for name := range exporters {
		if strings.Contains(name, "prometheus") || strings.Contains(name, "otlp") {
			hasMetricsExporter = true
			break
		}
	}

	if !hasMetricsExporter {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Path:    "exporters",
			Message: "no metrics exporter found (prometheus or otlp recommended)",
		})
	}
}

// validateExtensions validates the extensions section
func (v *Validator) validateExtensions(extensions map[string]interface{}, result *ValidationResult) {
	// Check for health_check extension (recommended)
	if _, ok := extensions["health_check"]; !ok && v.strict {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Path:    "extensions",
			Message: "missing 'health_check' extension - recommended for production",
		})
	}
}

// validateTelemetry validates the telemetry section
func (v *Validator) validateTelemetry(telemetry map[string]interface{}, result *ValidationResult) {
	// Basic validation for telemetry configuration
	if logs, ok := telemetry["logs"].(map[string]interface{}); ok {
		if level, ok := logs["level"].(string); ok {
			validLevels := []string{"debug", "info", "warn", "error"}
			isValid := false
			for _, validLevel := range validLevels {
				if level == validLevel {
					isValid = true
					break
				}
			}
			if !isValid {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Path:    "service.telemetry.logs.level",
					Message: fmt.Sprintf("invalid log level '%s'", level),
				})
			}
		}
	}
}