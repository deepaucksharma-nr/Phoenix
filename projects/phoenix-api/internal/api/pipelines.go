package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/phoenix/platform/pkg/common/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/services"
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

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pipelines": templates,
		"total":     len(templates),
	})
}

// GET /api/v1/pipelines/{id} - Get pipeline template details
func (s *Server) handleGetPipeline(w http.ResponseWriter, r *http.Request) {
	pipelineID := chi.URLParam(r, "id")

	// Get catalog path from config or environment
	catalogPath := os.Getenv("PHOENIX_PIPELINE_CATALOG_PATH")
	if catalogPath == "" {
		catalogPath = "/app/configs/pipelines/catalog"
	}

	// Look for the template file in known categories
	var template *PipelineTemplate
	categories := []string{"process", "infra", "app"}

	for _, category := range categories {
		templatePath := filepath.Join(catalogPath, category, pipelineID+".yaml")
		if info, err := os.Stat(templatePath); err == nil && !info.IsDir() {
			// Load the template
			data, err := os.ReadFile(templatePath)
			if err != nil {
				log.Error().Err(err).Str("path", templatePath).Msg("Failed to read template file")
				continue
			}

			// Parse the template to extract metadata
			var config map[string]interface{}
			if err := yaml.Unmarshal(data, &config); err != nil {
				log.Error().Err(err).Str("path", templatePath).Msg("Failed to parse template YAML")
				continue
			}

			template = &PipelineTemplate{
				ID:          pipelineID,
				Name:        pipelineID,
				Category:    category,
				ConfigPath:  templatePath,
				Version:     "1.0.0",
				Description: fmt.Sprintf("%s pipeline template", strings.Title(category)),
				Metadata:    config,
			}

			// Extract description from metadata if available
			if metadata, ok := config["metadata"].(map[string]interface{}); ok {
				if desc, ok := metadata["description"].(string); ok {
					template.Description = desc
				}
				if name, ok := metadata["name"].(string); ok {
					template.Name = name
				}
			}

			break
		}
	}

	if template == nil {
		respondError(w, http.StatusNotFound, "Pipeline template not found")
		return
	}

	respondJSON(w, http.StatusOK, template)
}

// GET /api/v1/pipelines/{id}/config - Get pipeline configuration by name
func (s *Server) handleGetPipelineConfigByName(w http.ResponseWriter, r *http.Request) {
	pipelineID := chi.URLParam(r, "id")
	
	// Try to load the pipeline template directly from catalog
	catalogPath := os.Getenv("PHOENIX_PIPELINE_CATALOG_PATH")
	if catalogPath == "" {
		catalogPath = "/app/configs/pipelines/catalog"
	}
	
	// Look for the pipeline template file
	var templatePath string
	categories := []string{"process", "infra", "app"}
	
	for _, category := range categories {
		path := filepath.Join(catalogPath, category, pipelineID+".yaml")
		if _, err := os.Stat(path); err == nil {
			templatePath = path
			break
		}
	}
	
	if templatePath == "" {
		respondError(w, http.StatusNotFound, "Pipeline template not found")
		return
	}
	
	// Read the template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		log.Error().Err(err).Str("path", templatePath).Msg("Failed to read pipeline template")
		respondError(w, http.StatusInternalServerError, "Failed to read pipeline template")
		return
	}
	
	// Return the YAML content directly
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// GET /api/v1/pipelines/status - Get aggregated pipeline status
func (s *Server) handleGetPipelineStatus(w http.ResponseWriter, r *http.Request) {
	// Get deployment statistics
	deployments, _, err := s.store.ListDeployments(r.Context(), &models.ListDeploymentsRequest{
		PageSize: 100,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to get deployments")
		respondError(w, http.StatusInternalServerError, "Failed to get pipeline status")
		return
	}

	// Count by status
	statusCounts := map[string]int{
		"ready":     0,
		"deploying": 0,
		"failed":    0,
		"stopped":   0,
	}

	for _, d := range deployments {
		if count, exists := statusCounts[d.Status]; exists {
			statusCounts[d.Status] = count + 1
		} else {
			statusCounts["unknown"] = statusCounts["unknown"] + 1
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"total":   len(deployments),
		"status":  statusCounts,
		"updated": time.Now(),
	})
}

// POST /api/v1/pipelines/validate - Validate a pipeline configuration
func (s *Server) handleValidatePipeline(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Config map[string]interface{} `json:"config"`
		YAML   string                 `json:"yaml"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// If YAML is provided, parse it
	if req.YAML != "" {
		var config map[string]interface{}
		if err := yaml.Unmarshal([]byte(req.YAML), &config); err != nil {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"valid": false,
				"error": fmt.Sprintf("Invalid YAML: %v", err),
			})
			return
		}
		req.Config = config
	}

	// Validate the pipeline configuration structure
	if req.Config == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"valid": false,
			"error": "No configuration provided",
		})
		return
	}

	// Convert to PipelineConfig for validation
	config := &services.PipelineConfig{
		Receivers:  make(map[string]interface{}),
		Processors: []services.ProcessorConfig{},
		Exporters:  make(map[string]interface{}),
		Service: services.ServiceConfig{
			Pipelines: make(map[string]services.PipelineService),
		},
	}

	// Copy receivers directly
	if receivers, ok := req.Config["receivers"].(map[string]interface{}); ok {
		config.Receivers = receivers
	}

	// Parse processors
	if processors, ok := req.Config["processors"].(map[string]interface{}); ok {
		for name, processor := range processors {
			if p, ok := processor.(map[string]interface{}); ok {
				pc := services.ProcessorConfig{
					Type: name,
				}

				// Handle different processor types
				switch {
				case strings.HasPrefix(name, "memory_limiter"):
					if limit, ok := p["limit_mib"].(float64); ok {
						pc.Limit = int(limit)
					}
					if checkInterval, ok := p["check_interval"].(string); ok {
						pc.CheckInterval = checkInterval
					}

				case strings.HasPrefix(name, "batch"):
					if timeout, ok := p["timeout"].(string); ok {
						pc.Timeout = timeout
					}
					if sendBatchSize, ok := p["send_batch_size"].(float64); ok {
						pc.SendBatchSize = int(sendBatchSize)
					}

				case name == "phoenix_adaptive_filter":
					// Custom processor
					if af, ok := p["adaptive_filter"].(map[string]interface{}); ok {
						adaptiveFilter := make(map[string]interface{})
						
						if enabled, ok := af["enabled"].(bool); ok {
							adaptiveFilter["enabled"] = enabled
						}
						if thresholds, ok := af["thresholds"].(map[string]interface{}); ok {
							adaptiveFilter["thresholds"] = thresholds
						}
						if rules, ok := af["rules"].([]interface{}); ok {
							adaptiveFilter["rules"] = rules
						}
						
						pc.Config = map[string]interface{}{
							"adaptive_filter": adaptiveFilter,
						}
					}

				case name == "phoenix_topk":
					// TopK processor
					if topk, ok := p["topk"].(map[string]interface{}); ok {
						topkConfig := make(map[string]interface{})
						
						if k, ok := topk["k"].(float64); ok {
							topkConfig["k"] = int(k)
						}
						if windowSize, ok := topk["window_size"].(string); ok {
							topkConfig["window_size"] = windowSize
						}
						if dimensions, ok := topk["dimensions"].([]interface{}); ok {
							topkConfig["dimensions"] = dimensions
						}
						
						pc.Config = map[string]interface{}{
							"topk": topkConfig,
						}
					}

				default:
					// Store raw config for unknown processors
					pc.Config = p
				}

				config.Processors = append(config.Processors, pc)
			}
		}
	}

	// Parse exporters
	if exporters, ok := req.Config["exporters"].(map[string]interface{}); ok {
		config.Exporters = exporters
	}

	// Parse service
	if service, ok := req.Config["service"].(map[string]interface{}); ok {
		if pipelines, ok := service["pipelines"].(map[string]interface{}); ok {
			for name, pipeline := range pipelines {
				if p, ok := pipeline.(map[string]interface{}); ok {
					ps := services.PipelineService{}

					// Parse receivers
					if receivers, ok := p["receivers"].([]interface{}); ok {
						for _, r := range receivers {
							if recv, ok := r.(string); ok {
								ps.Receivers = append(ps.Receivers, recv)
							}
						}
					}

					// Parse processors
					if processors, ok := p["processors"].([]interface{}); ok {
						for _, proc := range processors {
							if p, ok := proc.(string); ok {
								ps.Processors = append(ps.Processors, p)
							}
						}
					}

					// Parse exporters
					if exporters, ok := p["exporters"].([]interface{}); ok {
						for _, exp := range exporters {
							if e, ok := exp.(string); ok {
								ps.Exporters = append(ps.Exporters, e)
							}
						}
					}

					config.Service.Pipelines[name] = ps
				}
			}
		}
	}

	// Validate the configuration
	if err := s.templateRenderer.ValidatePipelineConfig(config); err != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"valid": true,
		"message": "Pipeline configuration is valid",
	})
}

// POST /api/v1/pipelines/render - Render a pipeline template with parameters
func (s *Server) handleRenderPipeline(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Template     string                 `json:"template"`
		ExperimentID string                 `json:"experiment_id"`
		Variant      string                 `json:"variant"`
		HostID       string                 `json:"host_id"`
		Parameters   map[string]interface{} `json:"parameters"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Template == "" {
		respondError(w, http.StatusBadRequest, "Template name is required")
		return
	}

	// Create template data
	templateData := services.TemplateData{
		ExperimentID: req.ExperimentID,
		Variant:      req.Variant,
		HostID:       req.HostID,
		Config:       req.Parameters,
	}

	// Default values
	if templateData.Variant == "" {
		templateData.Variant = "candidate"
	}

	// Render the template
	rendered, err := s.templateRenderer.RenderTemplate(r.Context(), req.Template, templateData)
	if err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Failed to render template: %v", err))
		return
	}

	// Parse the rendered YAML to validate it
	var config map[string]interface{}
	if err := yaml.Unmarshal([]byte(rendered), &config); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Rendered template is not valid YAML: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"rendered": rendered,
		"config":   config,
		"template": req.Template,
	})
}

// Helper function to load pipeline templates from catalog
func (s *Server) loadPipelineTemplates(catalogPath string) ([]PipelineTemplate, error) {
	var templates []PipelineTemplate

	// Define known categories
	categories := []string{"process", "infra", "app"}

	for _, category := range categories {
		categoryPath := filepath.Join(catalogPath, category)
		
		// Check if category directory exists
		if info, err := os.Stat(categoryPath); err != nil || !info.IsDir() {
			continue
		}

		// Read all YAML files in the category
		files, err := os.ReadDir(categoryPath)
		if err != nil {
			log.Warn().Err(err).Str("category", category).Msg("Failed to read category directory")
			continue
		}

		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
				continue
			}

			// Create template entry
			templateID := strings.TrimSuffix(file.Name(), ".yaml")
			template := PipelineTemplate{
				ID:          templateID,
				Name:        templateID,
				Category:    category,
				Version:     "1.0.0",
				ConfigPath:  filepath.Join(categoryPath, file.Name()),
				Description: fmt.Sprintf("%s pipeline template", strings.Title(category)),
				Metadata:    make(map[string]interface{}),
			}

			// Try to read and parse the template for metadata
			data, err := os.ReadFile(template.ConfigPath)
			if err == nil {
				var config map[string]interface{}
				if err := yaml.Unmarshal(data, &config); err == nil {
					// Extract metadata if available
					if metadata, ok := config["metadata"].(map[string]interface{}); ok {
						if desc, ok := metadata["description"].(string); ok {
							template.Description = desc
						}
						if name, ok := metadata["name"].(string); ok {
							template.Name = name
						}
						template.Metadata = metadata
					}
				}
			}

			templates = append(templates, template)
		}
	}

	return templates, nil
}