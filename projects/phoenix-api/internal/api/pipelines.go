package api

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/phoenix/platform/pkg/common/models"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// PipelineTemplate represents a pipeline template from the catalog
type PipelineTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Version     string                 `json:"version"`
	ConfigPath  string                 `json:"config_path"`
	Parameters  []TemplateParameter    `json:"parameters"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TemplateParameter represents a configurable parameter in a pipeline template
type TemplateParameter struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Required     bool        `json:"required"`
	Validation   interface{} `json:"validation,omitempty"`
}

// GET /api/v1/pipelines - List available pipeline templates
func (s *Server) handleListPipelines(w http.ResponseWriter, r *http.Request) {
	// Get catalog path from config or environment
	catalogPath := os.Getenv("PHOENIX_PIPELINE_CATALOG_PATH")
	if catalogPath == "" {
		catalogPath = "/app/configs/pipelines/catalog"
	}

	templates, err := s.loadPipelineTemplates(catalogPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load pipeline templates")
		respondError(w, http.StatusInternalServerError, "Failed to load pipeline templates")
		return
	}

	respondJSON(w, http.StatusOK, templates)
}

// GET /api/v1/pipelines/{id} - Get pipeline template details
func (s *Server) handleGetPipeline(w http.ResponseWriter, r *http.Request) {
	pipelineID := chi.URLParam(r, "id")

	// Get catalog path from config or environment
	catalogPath := os.Getenv("PHOENIX_PIPELINE_CATALOG_PATH")
	if catalogPath == "" {
		catalogPath = "/app/configs/pipelines/catalog"
	}

	templates, err := s.loadPipelineTemplates(catalogPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load pipeline templates")
		respondError(w, http.StatusInternalServerError, "Failed to load pipeline templates")
		return
	}

	// Find the requested template
	for _, template := range templates {
		if template.ID == pipelineID {
			// Load full configuration if requested
			if r.URL.Query().Get("include_config") == "true" {
				config, err := s.loadPipelineConfig(template.ConfigPath)
				if err != nil {
					log.Error().Err(err).Str("pipeline_id", pipelineID).Msg("Failed to load pipeline config")
					respondError(w, http.StatusInternalServerError, "Failed to load pipeline configuration")
					return
				}
				template.Metadata["config"] = config
			}
			respondJSON(w, http.StatusOK, template)
			return
		}
	}

	respondError(w, http.StatusNotFound, "Pipeline template not found")
}

// GET /api/v1/pipelines/status - Get aggregated pipeline deployment status
func (s *Server) handleGetPipelineStatus(w http.ResponseWriter, r *http.Request) {
	// Get status from pipeline deployment service
	ctx := r.Context()
	
	// Query parameters for filtering
	namespace := r.URL.Query().Get("namespace")
	status := r.URL.Query().Get("status")
	
	// Get deployed pipelines from store
	req := &models.ListDeploymentsRequest{
		Namespace: namespace,
		Status:    status,
		PageSize:  100, // Default page size
	}
	
	deployments, total, err := s.store.ListDeployments(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list pipeline deployments")
		respondError(w, http.StatusInternalServerError, "Failed to get pipeline status")
		return
	}

	// Aggregate status
	statusSummary := map[string]interface{}{
		"total_deployments": total,
		"deployments_by_status": map[string]int{
			"active":   0,
			"pending":  0,
			"failed":   0,
			"updating": 0,
		},
		"deployments_by_pipeline": make(map[string]int),
		"last_updated": time.Now(),
	}

	// Count deployments by status and pipeline
	for _, deployment := range deployments {
		if deployment.Status != "" {
			if count, ok := statusSummary["deployments_by_status"].(map[string]int)[deployment.Status]; ok {
				statusSummary["deployments_by_status"].(map[string]int)[deployment.Status] = count + 1
			}
		}
		
		if deployment.PipelineName != "" {
			pipelineCount := statusSummary["deployments_by_pipeline"].(map[string]int)
			pipelineCount[deployment.PipelineName]++
		}
	}

	respondJSON(w, http.StatusOK, statusSummary)
}

// loadPipelineTemplates loads pipeline templates from the catalog directory
func (s *Server) loadPipelineTemplates(catalogPath string) ([]*PipelineTemplate, error) {
	var templates []*PipelineTemplate

	// Walk through catalog directory
	err := filepath.Walk(catalogPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Process only YAML files
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Extract template info from file
		template, err := s.parsePipelineTemplate(path, catalogPath)
		if err != nil {
			log.Warn().Err(err).Str("file", path).Msg("Failed to parse pipeline template")
			return nil // Continue with other files
		}

		templates = append(templates, template)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return templates, nil
}

// parsePipelineTemplate parses a pipeline template file
func (s *Server) parsePipelineTemplate(filePath, basePath string) (*PipelineTemplate, error) {
	// Read file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Parse YAML to extract metadata
	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Extract relative path
	relPath, _ := filepath.Rel(basePath, filePath)
	
	// Extract category from directory structure
	parts := strings.Split(filepath.Dir(relPath), string(os.PathSeparator))
	category := "general"
	if len(parts) > 0 && parts[0] != "." {
		category = parts[0]
	}

	// Extract template name from filename
	basename := filepath.Base(filePath)
	name := strings.TrimSuffix(basename, filepath.Ext(basename))
	
	// Create template ID
	id := strings.ReplaceAll(name, "-", "_")

	// Extract description from comments
	description := extractDescription(string(data))

	// Extract version if present
	version := "v1"
	if strings.Contains(name, "-v") {
		parts := strings.Split(name, "-v")
		if len(parts) > 1 {
			version = "v" + parts[len(parts)-1]
		}
	}

	// Extract parameters from processor configurations
	parameters := extractParameters(config)

	template := &PipelineTemplate{
		ID:          id,
		Name:        name,
		Description: description,
		Category:    category,
		Version:     version,
		ConfigPath:  filePath,
		Parameters:  parameters,
		Metadata: map[string]interface{}{
			"category": category,
		},
	}

	return template, nil
}

// extractDescription extracts description from YAML comments
func extractDescription(content string) string {
	lines := strings.Split(content, "\n")
	var description []string
	inHeader := true
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Skip first comment line (usually just the title)
		if inHeader && strings.HasPrefix(trimmed, "#") && len(description) == 0 {
			continue
		}
		
		// Collect comment lines as description
		if strings.HasPrefix(trimmed, "#") && inHeader {
			cleaned := strings.TrimPrefix(trimmed, "#")
			cleaned = strings.TrimSpace(cleaned)
			if cleaned != "" {
				description = append(description, cleaned)
			}
		} else if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			// Stop at first non-comment line
			break
		}
	}
	
	return strings.Join(description, " ")
}

// extractParameters extracts configurable parameters from the pipeline config
func extractParameters(config map[string]interface{}) []TemplateParameter {
	var parameters []TemplateParameter

	// Look for processors section
	if processors, ok := config["processors"].(map[string]interface{}); ok {
		// Check for phoenix processor configurations
		for processorName, processorConfig := range processors {
			if strings.HasPrefix(processorName, "phoenix/") {
				if cfg, ok := processorConfig.(map[string]interface{}); ok {
					// Extract parameters from processor config
					if baseThresholds, ok := cfg["base_thresholds"].(map[string]interface{}); ok {
						for key, value := range baseThresholds {
							param := TemplateParameter{
								Name:         "base_threshold_" + key,
								Description:  "Base threshold for " + key,
								Type:         "number",
								DefaultValue: value,
								Required:     false,
							}
							parameters = append(parameters, param)
						}
					}
					
					// Extract other configurable fields
					if maxCardinality, ok := cfg["max_cardinality"]; ok {
						param := TemplateParameter{
							Name:         "max_cardinality",
							Description:  "Maximum number of unique values to track",
							Type:         "integer",
							DefaultValue: maxCardinality,
							Required:     false,
						}
						parameters = append(parameters, param)
					}
				}
			}
		}
	}

	// Add common parameters
	parameters = append(parameters, []TemplateParameter{
		{
			Name:         "collection_interval",
			Description:  "How often to collect metrics",
			Type:         "duration",
			DefaultValue: "10s",
			Required:     false,
		},
		{
			Name:         "batch_size",
			Description:  "Number of metrics to batch before sending",
			Type:         "integer",
			DefaultValue: 1000,
			Required:     false,
		},
		{
			Name:         "memory_limit_mib",
			Description:  "Memory limit for the collector in MiB",
			Type:         "integer",
			DefaultValue: 512,
			Required:     false,
		},
	}...)

	return parameters
}

// loadPipelineConfig loads the full pipeline configuration
func (s *Server) loadPipelineConfig(configPath string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}