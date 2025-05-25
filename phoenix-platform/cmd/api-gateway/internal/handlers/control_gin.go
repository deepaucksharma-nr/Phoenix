package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	
	controllerv1 "github.com/phoenix-platform/api/phoenix/controller/v1"
)

// ControlHandler handles HTTP requests for control operations
type ControlHandler struct {
	logger *zap.Logger
	client controllerv1.ControllerServiceClient
}

// NewControlHandler creates a new control handler
func NewControlHandler(logger *zap.Logger, client controllerv1.ControllerServiceClient) *ControlHandler {
	return &ControlHandler{
		logger: logger,
		client: client,
	}
}

// ApplyControlSignalRequest represents the HTTP request for applying control signal
type ApplyControlSignalRequest struct {
	ExperimentID string                 `json:"experiment_id" binding:"required"`
	Type         string                 `json:"type" binding:"required"`
	Parameters   map[string]interface{} `json:"parameters"`
	Reason       string                 `json:"reason"`
}

// ApplyControlSignal handles POST /control/signals
func (h *ControlHandler) ApplyControlSignal(c *gin.Context) {
	var req ApplyControlSignalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"details": err.Error(),
		})
		return
	}

	// Convert parameters to protobuf struct
	params := make(map[string]*structpb.Value)
	for k, v := range req.Parameters {
		val, err := structpb.NewValue(v)
		if err != nil {
			h.logger.Error("failed to convert parameter", zap.String("key", k), zap.Error(err))
			continue
		}
		params[k] = val
	}

	// Map string type to enum
	signalType := controllerv1.SignalType_SIGNAL_TYPE_UNSPECIFIED
	switch req.Type {
	case "traffic_split":
		signalType = controllerv1.SignalType_SIGNAL_TYPE_TRAFFIC_SPLIT
	case "rollback":
		signalType = controllerv1.SignalType_SIGNAL_TYPE_ROLLBACK
	case "config_update":
		signalType = controllerv1.SignalType_SIGNAL_TYPE_CONFIG_UPDATE
	case "pause":
		signalType = controllerv1.SignalType_SIGNAL_TYPE_PAUSE
	case "resume":
		signalType = controllerv1.SignalType_SIGNAL_TYPE_RESUME
	}

	grpcReq := &controllerv1.ApplyControlSignalRequest{
		ExperimentId: req.ExperimentID,
		Signal: &controllerv1.ControlSignal{
			Type:       signalType,
			Parameters: params,
			Reason:     req.Reason,
		},
	}

	resp, err := h.client.ApplyControlSignal(c.Request.Context(), grpcReq)
	if err != nil {
		h.logger.Error("failed to apply control signal", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to apply control signal",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"signal_id": resp.SignalId,
		"status":    resp.Status.String(),
		"message":   resp.Message,
	})
}

// GetControlSignal handles GET /control/signals/:id
func (h *ControlHandler) GetControlSignal(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "signal ID required",
		})
		return
	}

	req := &controllerv1.GetControlSignalRequest{
		SignalId: id,
	}

	resp, err := h.client.GetControlSignal(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("failed to get control signal", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "control signal not found",
		})
		return
	}

	c.JSON(http.StatusOK, resp.Signal)
}

// ListControlSignals handles GET /control/experiments/:id/signals
func (h *ControlHandler) ListControlSignals(c *gin.Context) {
	experimentID := c.Param("id")
	if experimentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "experiment ID required",
		})
		return
	}

	pageSize := c.DefaultQuery("page_size", "50")
	pageToken := c.Query("page_token")

	req := &controllerv1.ListControlSignalsRequest{
		ExperimentId: experimentID,
		PageSize:     parseInt32(pageSize, 50),
		PageToken:    pageToken,
	}

	resp, err := h.client.ListControlSignals(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("failed to list control signals", zap.String("experiment_id", experimentID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list control signals",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"signals":         resp.Signals,
		"next_page_token": resp.NextPageToken,
	})
}

// GetDriftReport handles GET /control/experiments/:id/drift
func (h *ControlHandler) GetDriftReport(c *gin.Context) {
	experimentID := c.Param("id")
	if experimentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "experiment ID required",
		})
		return
	}

	req := &controllerv1.GetDriftReportRequest{
		ExperimentId: experimentID,
	}

	resp, err := h.client.GetDriftReport(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("failed to get drift report", zap.String("experiment_id", experimentID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get drift report",
		})
		return
	}

	c.JSON(http.StatusOK, resp.Report)
}

// ControlStatus is a stub method for backward compatibility
func (h *ControlHandler) HandleControlStatus(c *gin.Context) {
	experimentID := c.Param("id")
	if experimentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "experiment ID required",
		})
		return
	}

	// This is a stub - in real implementation, this might aggregate status
	c.JSON(http.StatusOK, gin.H{
		"experiment_id": experimentID,
		"status":        "active",
		"message":       "Control system operational",
	})
}