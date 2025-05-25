package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	
	"github.com/phoenix/platform/pkg/clients/generator"
	generatorv1 "github.com/phoenix/platform/api/proto/v1/generator"
	commonv1 "github.com/phoenix/platform/api/proto/v1/common"
)

// GeneratorHandler handles HTTP requests for configuration generation
type GeneratorHandler struct {
	logger *zap.Logger
	client *generator.Client
}

// NewGeneratorHandler creates a new generator handler
func NewGeneratorHandler(logger *zap.Logger, client *generator.Client) *GeneratorHandler {
	return &GeneratorHandler{
		logger: logger,
		client: client,
	}
}

// HandleGenerate handles /api/v1/generate
func (h *GeneratorHandler) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.generateConfiguration(w, r)
}

// HandleTemplates handles /api/v1/templates
func (h *GeneratorHandler) HandleTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.listTemplates(w, r)
}

// HandleTemplateByID handles /api/v1/templates/{id}
func (h *GeneratorHandler) HandleTemplateByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/templates/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Template ID required", http.StatusBadRequest)
		return
	}
	id := parts[0]

	switch r.Method {
	case http.MethodGet:
		h.getTemplate(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *GeneratorHandler) generateConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second) // Longer timeout for generation
	defer cancel()

	var req struct {
		PipelineID           string                      `json:"pipeline_id"`
		Pipeline             *commonv1.Pipeline          `json:"pipeline,omitempty"`
		Goals                []generatorGoal             `json:"goals"`
		Constraints          []generatorConstraint       `json:"constraints"`
		PreferredTemplateIDs []string                    `json:"preferred_template_ids"`
		ExcludedTemplateIDs  []string                    `json:"excluded_template_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("generating configuration", zap.String("pipeline_id", req.PipelineID))

	// Convert to generator request
	genReq := &generator.GenerateConfigRequest{
		PipelineID:           req.PipelineID,
		Pipeline:             req.Pipeline,
		PreferredTemplateIDs: req.PreferredTemplateIDs,
		ExcludedTemplateIDs:  req.ExcludedTemplateIDs,
	}

	// Convert goals
	for _, goal := range req.Goals {
		genReq.Goals = append(genReq.Goals, generator.OptimizationGoal{
			Type:        goal.Type,
			TargetValue: goal.TargetValue,
			Priority:    goal.Priority,
		})
	}

	// Convert constraints
	for _, constraint := range req.Constraints {
		genReq.Constraints = append(genReq.Constraints, generator.Constraint{
			Type:     constraint.Type,
			Metric:   constraint.Metric,
			MinValue: constraint.MinValue,
			MaxValue: constraint.MaxValue,
		})
	}

	response, err := h.client.GenerateConfiguration(ctx, genReq)
	if err != nil {
		h.logger.Error("failed to generate configuration", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to generate configuration: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *GeneratorHandler) listTemplates(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	h.logger.Debug("listing templates")

	// Parse query parameters
	categories := parseCategories(r.URL.Query().Get("categories"))
	labels := parseLabels(r.URL.Query().Get("labels"))

	templates, _, err := h.client.ListTemplates(ctx, categories, labels, 50, "")
	if err != nil {
		h.logger.Error("failed to list templates", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to list templates: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to REST response
	response := map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *GeneratorHandler) getTemplate(w http.ResponseWriter, r *http.Request, id string) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	h.logger.Debug("getting template", zap.String("id", id))

	template, err := h.client.GetTemplate(ctx, id)
	if err != nil {
		h.logger.Error("failed to get template", zap.String("id", id), zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to get template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(template); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// Helper types for JSON parsing

type generatorGoal struct {
	Type        generatorv1.GoalType `json:"type"`
	TargetValue float64              `json:"target_value"`
	Priority    int32                `json:"priority"`
}

type generatorConstraint struct {
	Type     generatorv1.ConstraintType `json:"type"`
	Metric   string                     `json:"metric"`
	MinValue float64                    `json:"min_value"`
	MaxValue float64                    `json:"max_value"`
}

// Helper functions

func parseCategories(categoriesStr string) []string {
	if categoriesStr == "" {
		return nil
	}

	categories := strings.Split(categoriesStr, ",")
	var result []string
	
	for _, category := range categories {
		trimmed := strings.TrimSpace(category)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}