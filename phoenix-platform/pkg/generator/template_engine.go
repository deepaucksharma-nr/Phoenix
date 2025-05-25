package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// TemplateEngine handles loading and processing of pipeline templates
type TemplateEngine struct {
	logger       *zap.Logger
	templatePath string
	templates    map[string]*PipelineTemplate
}

// PipelineTemplate represents a loaded pipeline template
type PipelineTemplate struct {
	Name        string
	Description string
	Content     string
	Config      map[string]interface{}
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine(logger *zap.Logger, templatePath string) (*TemplateEngine, error) {
	if templatePath == "" {
		templatePath = "pipelines/templates"
	}

	engine := &TemplateEngine{
		logger:       logger,
		templatePath: templatePath,
		templates:    make(map[string]*PipelineTemplate),
	}

	// Load all templates at initialization
	if err := engine.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return engine, nil
}

// loadTemplates loads all pipeline templates from the templates directory
func (te *TemplateEngine) loadTemplates() error {
	te.logger.Info("loading pipeline templates", zap.String("path", te.templatePath))

	entries, err := os.ReadDir(te.templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		templateName := strings.TrimSuffix(entry.Name(), ".yaml")
		templatePath := filepath.Join(te.templatePath, entry.Name())

		content, err := os.ReadFile(templatePath)
		if err != nil {
			te.logger.Error("failed to read template file",
				zap.String("file", templatePath),
				zap.Error(err),
			)
			continue
		}

		// Parse YAML to validate structure
		var config map[string]interface{}
		if err := yaml.Unmarshal(content, &config); err != nil {
			te.logger.Error("failed to parse template YAML",
				zap.String("file", templatePath),
				zap.Error(err),
			)
			continue
		}

		template := &PipelineTemplate{
			Name:        templateName,
			Description: te.extractDescription(config),
			Content:     string(content),
			Config:      config,
		}

		te.templates[templateName] = template
		te.logger.Info("loaded pipeline template",
			zap.String("name", templateName),
			zap.String("description", template.Description),
		)
	}

	te.logger.Info("loaded pipeline templates",
		zap.Int("count", len(te.templates)),
		zap.String("templates", strings.Join(te.getTemplateNames(), ", ")),
	)

	return nil
}

// getTemplateNames returns a list of loaded template names
func (te *TemplateEngine) getTemplateNames() []string {
	names := make([]string, 0, len(te.templates))
	for name := range te.templates {
		names = append(names, name)
	}
	return names
}

// extractDescription extracts description from template config if available
func (te *TemplateEngine) extractDescription(config map[string]interface{}) string {
	// Look for description in metadata or as a comment
	if metadata, ok := config["metadata"].(map[string]interface{}); ok {
		if desc, ok := metadata["description"].(string); ok {
			return desc
		}
	}
	return ""
}

// GetTemplate returns a pipeline template by name
func (te *TemplateEngine) GetTemplate(name string) (*PipelineTemplate, error) {
	template, ok := te.templates[name]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", name)
	}
	return template, nil
}

// ListTemplates returns all available templates
func (te *TemplateEngine) ListTemplates() []*PipelineTemplate {
	templates := make([]*PipelineTemplate, 0, len(te.templates))
	for _, template := range te.templates {
		templates = append(templates, template)
	}
	return templates
}

// GenerateConfig generates a configuration from a template with variables
func (te *TemplateEngine) GenerateConfig(templateName string, variables map[string]string) (string, error) {
	pipelineTemplate, err := te.GetTemplate(templateName)
	if err != nil {
		return "", err
	}

	// Create a text template
	tmpl, err := template.New(templateName).Parse(pipelineTemplate.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Prepare template data with environment variable style substitution
	templateData := make(map[string]string)
	for k, v := range variables {
		templateData[k] = v
	}

	// Add default values if not provided
	te.addDefaultValues(templateData)

	// First pass: Execute Go template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Second pass: Replace ${VAR} style placeholders
	result := buf.String()
	for k, v := range templateData {
		result = strings.ReplaceAll(result, fmt.Sprintf("${%s}", k), v)
	}

	// Validate the generated YAML
	var validationConfig map[string]interface{}
	if err := yaml.Unmarshal([]byte(result), &validationConfig); err != nil {
		return "", fmt.Errorf("generated invalid YAML: %w", err)
	}

	return result, nil
}

// addDefaultValues adds default values for common variables
func (te *TemplateEngine) addDefaultValues(variables map[string]string) {
	defaults := map[string]string{
		"NEW_RELIC_OTLP_ENDPOINT": "https://otlp.nr-data.net",
		"NODE_NAME":               "${NODE_NAME}",
		"PHOENIX_EXPERIMENT_ID":   "${PHOENIX_EXPERIMENT_ID}",
		"PHOENIX_VARIANT":         "${PHOENIX_VARIANT}",
	}

	for k, v := range defaults {
		if _, exists := variables[k]; !exists {
			variables[k] = v
		}
	}
}

// MergeConfigs merges multiple configurations into one
func (te *TemplateEngine) MergeConfigs(configs ...map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, config := range configs {
		for k, v := range config {
			result[k] = v
		}
	}

	return result, nil
}

// ValidateConfig validates an OpenTelemetry configuration
func (te *TemplateEngine) ValidateConfig(config string) error {
	var configMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(config), &configMap); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	// Check required sections
	requiredSections := []string{"receivers", "processors", "exporters", "service"}
	for _, section := range requiredSections {
		if _, ok := configMap[section]; !ok {
			return fmt.Errorf("missing required section: %s", section)
		}
	}

	// Validate service pipelines
	if service, ok := configMap["service"].(map[string]interface{}); ok {
		if pipelines, ok := service["pipelines"].(map[string]interface{}); ok {
			for pipelineName, pipeline := range pipelines {
				if err := te.validatePipeline(pipelineName, pipeline); err != nil {
					return fmt.Errorf("invalid pipeline %s: %w", pipelineName, err)
				}
			}
		} else {
			return fmt.Errorf("service must contain pipelines")
		}
	}

	return nil
}

// validatePipeline validates a single pipeline configuration
func (te *TemplateEngine) validatePipeline(name string, pipeline interface{}) error {
	pipelineMap, ok := pipeline.(map[string]interface{})
	if !ok {
		return fmt.Errorf("pipeline must be a map")
	}

	// Check required fields
	requiredFields := []string{"receivers", "exporters"}
	for _, field := range requiredFields {
		if _, ok := pipelineMap[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	return nil
}