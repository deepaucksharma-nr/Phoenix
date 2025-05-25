package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	
	experimentv1 "github.com/phoenix/platform/api/phoenix/experiment/v1"
)

// ExperimentHandler handles HTTP requests for experiments
type ExperimentHandler struct {
	logger *zap.Logger
	client experimentv1.ExperimentServiceClient
}

// NewExperimentHandler creates a new experiment handler
func NewExperimentHandler(logger *zap.Logger, client experimentv1.ExperimentServiceClient) *ExperimentHandler {
	return &ExperimentHandler{
		logger: logger,
		client: client,
	}
}

// CreateExperimentRequest represents the HTTP request for creating an experiment
type CreateExperimentRequest struct {
	Name              string            `json:"name" binding:"required"`
	Description       string            `json:"description"`
	BaselinePipeline  string            `json:"baseline_pipeline" binding:"required"`
	CandidatePipeline string            `json:"candidate_pipeline" binding:"required"`
	TargetNodes       map[string]string `json:"target_nodes"`
	Config            *ExperimentConfig `json:"config"`
}

// ExperimentConfig for HTTP requests
type ExperimentConfig struct {
	DurationSeconds int64 `json:"duration_seconds"`
	TrafficSplit    struct {
		BaselinePercentage  int32 `json:"baseline_percentage"`
		CandidatePercentage int32 `json:"candidate_percentage"`
	} `json:"traffic_split"`
}

// ListExperiments handles GET /experiments
func (h *ExperimentHandler) ListExperiments(c *gin.Context) {
	pageSize := c.DefaultQuery("page_size", "50")
	pageToken := c.Query("page_token")

	req := &experimentv1.ListExperimentsRequest{
		PageSize:  parseInt32(pageSize, 50),
		PageToken: pageToken,
	}

	resp, err := h.client.ListExperiments(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("failed to list experiments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list experiments",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"experiments":     resp.Experiments,
		"next_page_token": resp.NextPageToken,
	})
}

// CreateExperiment handles POST /experiments
func (h *ExperimentHandler) CreateExperiment(c *gin.Context) {
	var req CreateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"details": err.Error(),
		})
		return
	}

	// Convert to gRPC request
	grpcReq := &experimentv1.CreateExperimentRequest{
		Name:              req.Name,
		Description:       req.Description,
		BaselinePipeline:  req.BaselinePipeline,
		CandidatePipeline: req.CandidatePipeline,
		TargetNodes:       req.TargetNodes,
	}

	// Add config if provided
	if req.Config != nil {
		grpcReq.Config = &experimentv1.ExperimentConfig{
			DurationSeconds: req.Config.DurationSeconds,
			TrafficSplit: &experimentv1.TrafficSplit{
				BaselinePercentage:  req.Config.TrafficSplit.BaselinePercentage,
				CandidatePercentage: req.Config.TrafficSplit.CandidatePercentage,
			},
		}
	}

	resp, err := h.client.CreateExperiment(c.Request.Context(), grpcReq)
	if err != nil {
		h.logger.Error("failed to create experiment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create experiment",
		})
		return
	}

	c.JSON(http.StatusCreated, resp.Experiment)
}

// GetExperiment handles GET /experiments/:id
func (h *ExperimentHandler) GetExperiment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "experiment ID required",
		})
		return
	}

	req := &experimentv1.GetExperimentRequest{
		Id: id,
	}

	resp, err := h.client.GetExperiment(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("failed to get experiment", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "experiment not found",
		})
		return
	}

	c.JSON(http.StatusOK, resp.Experiment)
}

// UpdateExperiment handles PUT /experiments/:id
func (h *ExperimentHandler) UpdateExperiment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "experiment ID required",
		})
		return
	}

	var req CreateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"details": err.Error(),
		})
		return
	}

	// Convert to gRPC request
	grpcReq := &experimentv1.UpdateExperimentRequest{
		Experiment: &experimentv1.Experiment{
			Id:          id,
			Name:        req.Name,
			Description: req.Description,
		},
	}

	resp, err := h.client.UpdateExperiment(c.Request.Context(), grpcReq)
	if err != nil {
		h.logger.Error("failed to update experiment", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update experiment",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteExperiment handles DELETE /experiments/:id
func (h *ExperimentHandler) DeleteExperiment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "experiment ID required",
		})
		return
	}

	req := &experimentv1.DeleteExperimentRequest{
		Id: id,
	}

	_, err := h.client.DeleteExperiment(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("failed to delete experiment", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete experiment",
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetExperimentStatus handles GET /experiments/:id/status
func (h *ExperimentHandler) GetExperimentStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "experiment ID required",
		})
		return
	}

	req := &experimentv1.GetExperimentStatusRequest{
		Id: id,
	}

	resp, err := h.client.GetExperimentStatus(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("failed to get experiment status", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "experiment not found",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// StartExperiment handles POST /experiments/:id/start
func (h *ExperimentHandler) StartExperiment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "experiment ID required",
		})
		return
	}

	// In a real implementation, this would call a StartExperiment RPC
	h.logger.Info("starting experiment", zap.String("id", id))
	
	c.JSON(http.StatusOK, gin.H{
		"message": "experiment started",
		"id":      id,
	})
}

// StopExperiment handles POST /experiments/:id/stop
func (h *ExperimentHandler) StopExperiment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "experiment ID required",
		})
		return
	}

	// In a real implementation, this would call a StopExperiment RPC
	h.logger.Info("stopping experiment", zap.String("id", id))
	
	c.JSON(http.StatusOK, gin.H{
		"message": "experiment stopped",
		"id":      id,
	})
}

// Helper function to parse int32
func parseInt32(s string, defaultValue int32) int32 {
	var val int32
	_, _ = fmt.Sscanf(s, "%d", &val)
	if val <= 0 {
		return defaultValue
	}
	return val
}