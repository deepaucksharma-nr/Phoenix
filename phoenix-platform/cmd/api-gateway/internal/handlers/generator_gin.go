package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	
	generatorv1 "github.com/phoenix/platform/api/phoenix/generator/v1"
)

// GeneratorHandler handles HTTP requests for configuration generation
type GeneratorHandler struct {
	logger *zap.Logger
	client generatorv1.GeneratorServiceClient
}

// NewGeneratorHandler creates a new generator handler
func NewGeneratorHandler(logger *zap.Logger, client generatorv1.GeneratorServiceClient) *GeneratorHandler {
	return &GeneratorHandler{
		logger: logger,
		client: client,
	}
}

// GenerateConfigRequest represents the HTTP request for generating config
type GenerateConfigRequest struct {
	ExperimentID string            `json:"experiment_id" binding:"required"`
	Template     string            `json:"template" binding:"required"`
	Parameters   map[string]string `json:"parameters"`
}

// ValidateConfigRequest represents the HTTP request for validating config
type ValidateConfigRequest struct {
	Configuration string `json:"configuration" binding:"required"`
}

// TemplateRequest represents template creation/update request
type TemplateRequest struct {
	Name             string            `json:"name" binding:"required"`
	Description      string            `json:"description"`
	Content          string            `json:"content" binding:"required"`
	DefaultParameters map[string]string `json:"default_parameters"`
}

// GenerateConfig handles POST /generator/generate
func (h *GeneratorHandler) GenerateConfig(c *gin.Context) {
	var req GenerateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"details": err.Error(),
		})
		return
	}

	grpcReq := &generatorv1.GenerateConfigurationRequest{
		ExperimentId: req.ExperimentID,
		Template:     req.Template,
		Parameters:   req.Parameters,
	}

	resp, err := h.client.GenerateConfiguration(c.Request.Context(), grpcReq)
	if err != nil {
		h.logger.Error("failed to generate configuration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate configuration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"config_id":     resp.ConfigId,
		"configuration": resp.Configuration,
		"version":       resp.Version,
		"generated_at":  resp.GeneratedAt,
	})
}

// ValidateConfig handles POST /generator/validate
func (h *GeneratorHandler) ValidateConfig(c *gin.Context) {
	var req ValidateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"details": err.Error(),
		})
		return
	}

	grpcReq := &generatorv1.ValidateConfigurationRequest{
		Configuration: req.Configuration,
	}

	resp, err := h.client.ValidateConfiguration(c.Request.Context(), grpcReq)
	if err != nil {
		h.logger.Error("failed to validate configuration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to validate configuration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":  resp.Valid,
		"errors": resp.Errors,
	})
}

// ListTemplates handles GET /templates
func (h *GeneratorHandler) ListTemplates(c *gin.Context) {
	pageSize := c.DefaultQuery("page_size", "50")
	pageToken := c.Query("page_token")

	req := &generatorv1.ListTemplatesRequest{
		PageSize:  parseInt32(pageSize, 50),
		PageToken: pageToken,
	}

	resp, err := h.client.ListTemplates(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("failed to list templates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list templates",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates":       resp.Templates,
		"next_page_token": resp.NextPageToken,
	})
}

// CreateTemplate handles POST /templates
func (h *GeneratorHandler) CreateTemplate(c *gin.Context) {
	var req TemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"details": err.Error(),
		})
		return
	}

	grpcReq := &generatorv1.CreateTemplateRequest{
		Template: &generatorv1.Template{
			Name:               req.Name,
			Description:        req.Description,
			Content:            req.Content,
			DefaultParameters:  req.DefaultParameters,
		},
	}

	resp, err := h.client.CreateTemplate(c.Request.Context(), grpcReq)
	if err != nil {
		h.logger.Error("failed to create template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create template",
		})
		return
	}

	c.JSON(http.StatusCreated, resp.Template)
}

// GetTemplate handles GET /templates/:name
func (h *GeneratorHandler) GetTemplate(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "template name required",
		})
		return
	}

	req := &generatorv1.GetTemplateRequest{
		Name: name,
	}

	resp, err := h.client.GetTemplate(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("failed to get template", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "template not found",
		})
		return
	}

	c.JSON(http.StatusOK, resp.Template)
}

// UpdateTemplate handles PUT /templates/:name
func (h *GeneratorHandler) UpdateTemplate(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "template name required",
		})
		return
	}

	var req TemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"details": err.Error(),
		})
		return
	}

	grpcReq := &generatorv1.UpdateTemplateRequest{
		Name: name,
		Template: &generatorv1.Template{
			Name:               req.Name,
			Description:        req.Description,
			Content:            req.Content,
			DefaultParameters:  req.DefaultParameters,
		},
	}

	resp, err := h.client.UpdateTemplate(c.Request.Context(), grpcReq)
	if err != nil {
		h.logger.Error("failed to update template", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update template",
		})
		return
	}

	c.JSON(http.StatusOK, resp.Template)
}

// DeleteTemplate handles DELETE /templates/:name
func (h *GeneratorHandler) DeleteTemplate(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "template name required",
		})
		return
	}

	req := &generatorv1.DeleteTemplateRequest{
		Name: name,
	}

	_, err := h.client.DeleteTemplate(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("failed to delete template", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete template",
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}