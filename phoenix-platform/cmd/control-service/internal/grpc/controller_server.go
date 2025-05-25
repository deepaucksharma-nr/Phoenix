package grpc

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	controllerv1 "github.com/phoenix-platform/api/proto/v1/controller"
	commonv1 "github.com/phoenix-platform/api/proto/v1/common"
)

// ControllerServer implements the controller service for traffic management and drift detection
type ControllerServer struct {
	controllerv1.UnimplementedControllerServiceServer
	logger *zap.Logger
	
	// In-memory stores for demo purposes
	// In production, these would be backed by a database
	controlSignals   map[string]*controllerv1.ControlSignal
	controlLoops     map[string]*controllerv1.ControlLoopStatus
	driftDetections  map[string][]*controllerv1.DriftDetection
}

// NewControllerServer creates a new controller gRPC server
func NewControllerServer(logger *zap.Logger) *ControllerServer {
	return &ControllerServer{
		logger:          logger,
		controlSignals:  make(map[string]*controllerv1.ControlSignal),
		controlLoops:    make(map[string]*controllerv1.ControlLoopStatus),
		driftDetections: make(map[string][]*controllerv1.DriftDetection),
	}
}

// ExecuteControlSignal executes a control signal
func (s *ControllerServer) ExecuteControlSignal(ctx context.Context, req *controllerv1.ExecuteControlSignalRequest) (*controllerv1.ExecuteControlSignalResponse, error) {
	if req.Signal == nil {
		return nil, status.Error(codes.InvalidArgument, "control signal is required")
	}

	if req.Signal.ExperimentId == "" {
		return nil, status.Error(codes.InvalidArgument, "experiment ID is required")
	}

	s.logger.Info("executing control signal",
		zap.String("experiment_id", req.Signal.ExperimentId),
		zap.String("type", req.Signal.Type.String()),
		zap.Bool("dry_run", req.DryRun),
	)

	// Validate the signal
	validationErrors := s.validateControlSignal(req.Signal)
	if len(validationErrors) > 0 {
		return &controllerv1.ExecuteControlSignalResponse{
			Signal:           req.Signal,
			ValidationErrors: validationErrors,
		}, nil
	}

	// Generate signal ID if not provided
	if req.Signal.Id == "" {
		req.Signal.Id = fmt.Sprintf("signal-%d", time.Now().UnixNano())
	}

	// Set timestamps
	req.Signal.CreatedAt = timestamppb.Now()
	req.Signal.Status = controllerv1.SignalStatus_SIGNAL_STATUS_PENDING

	if req.DryRun {
		req.Signal.Status = controllerv1.SignalStatus_SIGNAL_STATUS_COMPLETED
		req.Signal.ExecutedAt = timestamppb.Now()
		
		s.logger.Info("dry run completed", zap.String("signal_id", req.Signal.Id))
		
		return &controllerv1.ExecuteControlSignalResponse{
			Signal:           req.Signal,
			ValidationErrors: []*controllerv1.ValidationError{},
		}, nil
	}

	// Execute the signal
	err := s.executeSignal(ctx, req.Signal)
	if err != nil {
		req.Signal.Status = controllerv1.SignalStatus_SIGNAL_STATUS_FAILED
		req.Signal.StatusMessage = fmt.Sprintf("Execution failed: %v", err)
		
		s.logger.Error("failed to execute signal", zap.String("signal_id", req.Signal.Id), zap.Error(err))
		
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to execute signal: %v", err))
	}

	req.Signal.Status = controllerv1.SignalStatus_SIGNAL_STATUS_COMPLETED
	req.Signal.ExecutedAt = timestamppb.Now()

	// Store the signal
	s.controlSignals[req.Signal.Id] = req.Signal

	s.logger.Info("control signal executed successfully", zap.String("signal_id", req.Signal.Id))

	return &controllerv1.ExecuteControlSignalResponse{
		Signal:           req.Signal,
		ValidationErrors: []*controllerv1.ValidationError{},
	}, nil
}

// GetControlLoopStatus gets the control loop status for an experiment
func (s *ControllerServer) GetControlLoopStatus(ctx context.Context, req *controllerv1.GetControlLoopStatusRequest) (*controllerv1.GetControlLoopStatusResponse, error) {
	if req.ExperimentId == "" {
		return nil, status.Error(codes.InvalidArgument, "experiment ID is required")
	}

	s.logger.Debug("getting control loop status", zap.String("experiment_id", req.ExperimentId))

	// Get or create control loop status
	status, exists := s.controlLoops[req.ExperimentId]
	if !exists {
		status = &controllerv1.ControlLoopStatus{
			ExperimentId:        req.ExperimentId,
			Active:              true,
			LastEvaluation:     timestamppb.Now(),
			EvaluationInterval: &time.Duration{Seconds: 30},
			CurrentTrafficSplit: 10, // Default 10% to candidate
			ActiveSignals:      []*controllerv1.ActiveSignal{},
			Health:             commonv1.HealthStatus_HEALTH_STATUS_HEALTHY,
			ComponentHealth:    []*commonv1.ComponentHealth{},
		}
		s.controlLoops[req.ExperimentId] = status
	}

	return &controllerv1.GetControlLoopStatusResponse{
		Status: status,
	}, nil
}

// ListControlSignals lists control signals for an experiment
func (s *ControllerServer) ListControlSignals(ctx context.Context, req *controllerv1.ListControlSignalsRequest) (*controllerv1.ListControlSignalsResponse, error) {
	if req.ExperimentId == "" {
		return nil, status.Error(codes.InvalidArgument, "experiment ID is required")
	}

	s.logger.Debug("listing control signals", zap.String("experiment_id", req.ExperimentId))

	var signals []*controllerv1.ControlSignal
	for _, signal := range s.controlSignals {
		if signal.ExperimentId == req.ExperimentId {
			// Apply status filter if provided
			if len(req.Statuses) > 0 {
				matchesStatus := false
				for _, statusFilter := range req.Statuses {
					if signal.Status == statusFilter {
						matchesStatus = true
						break
					}
				}
				if !matchesStatus {
					continue
				}
			}

			// Apply time range filter if provided
			if req.StartTime != nil && signal.CreatedAt != nil {
				if signal.CreatedAt.AsTime().Before(req.StartTime.AsTime()) {
					continue
				}
			}
			if req.EndTime != nil && signal.CreatedAt != nil {
				if signal.CreatedAt.AsTime().After(req.EndTime.AsTime()) {
					continue
				}
			}

			signals = append(signals, signal)
		}
	}

	// Handle pagination
	pageSize := int32(50) // default
	if req.Pagination != nil && req.Pagination.PageSize > 0 {
		pageSize = req.Pagination.PageSize
	}

	// Simple pagination
	totalItems := int32(len(signals))
	var paginatedSignals []*controllerv1.ControlSignal
	
	if totalItems > pageSize {
		paginatedSignals = signals[:pageSize]
	} else {
		paginatedSignals = signals
	}

	response := &controllerv1.ListControlSignalsResponse{
		Signals: paginatedSignals,
	}

	if totalItems > 0 {
		response.Pagination = &commonv1.PaginationResponse{
			NextPageToken: "", // No pagination implemented yet
			TotalItems:    totalItems,
		}
	}

	s.logger.Debug("listed control signals", zap.Int("count", len(paginatedSignals)))

	return response, nil
}

// EnableDriftDetection enables drift detection for an experiment
func (s *ControllerServer) EnableDriftDetection(ctx context.Context, req *controllerv1.EnableDriftDetectionRequest) (*controllerv1.EnableDriftDetectionResponse, error) {
	if req.ExperimentId == "" {
		return nil, status.Error(codes.InvalidArgument, "experiment ID is required")
	}

	if req.Config == nil {
		return nil, status.Error(codes.InvalidArgument, "drift detection config is required")
	}

	s.logger.Info("enabling drift detection", 
		zap.String("experiment_id", req.ExperimentId),
		zap.Bool("auto_remediate", req.Config.AutoRemediate),
	)

	// Store the configuration (in production, this would be persisted)
	// For now, we'll just log that it's enabled
	s.logger.Info("drift detection enabled successfully", zap.String("experiment_id", req.ExperimentId))

	return &controllerv1.EnableDriftDetectionResponse{
		Enabled: true,
	}, nil
}

// GetDriftDetections gets drift detections for an experiment
func (s *ControllerServer) GetDriftDetections(ctx context.Context, req *controllerv1.GetDriftDetectionsRequest) (*controllerv1.GetDriftDetectionsResponse, error) {
	if req.ExperimentId == "" {
		return nil, status.Error(codes.InvalidArgument, "experiment ID is required")
	}

	s.logger.Debug("getting drift detections", zap.String("experiment_id", req.ExperimentId))

	// Get stored drift detections
	detections, exists := s.driftDetections[req.ExperimentId]
	if !exists {
		detections = []*controllerv1.DriftDetection{}
	}

	// Apply time range filter if provided
	var filteredDetections []*controllerv1.DriftDetection
	for _, detection := range detections {
		if req.StartTime != nil && detection.DetectedAt != nil {
			if detection.DetectedAt.AsTime().Before(req.StartTime.AsTime()) {
				continue
			}
		}
		if req.EndTime != nil && detection.DetectedAt != nil {
			if detection.DetectedAt.AsTime().After(req.EndTime.AsTime()) {
				continue
			}
		}
		filteredDetections = append(filteredDetections, detection)
	}

	return &controllerv1.GetDriftDetectionsResponse{
		Detections: filteredDetections,
	}, nil
}

// StreamControlSignals streams control signals (placeholder implementation)
func (s *ControllerServer) StreamControlSignals(req *controllerv1.GetControlLoopStatusRequest, stream controllerv1.ControllerService_StreamControlSignalsServer) error {
	if req.ExperimentId == "" {
		return status.Error(codes.InvalidArgument, "experiment ID is required")
	}

	s.logger.Info("starting control signals stream", zap.String("experiment_id", req.ExperimentId))

	// For demo purposes, send any existing signals and then close
	for _, signal := range s.controlSignals {
		if signal.ExperimentId == req.ExperimentId {
			if err := stream.Send(signal); err != nil {
				s.logger.Error("failed to send signal", zap.Error(err))
				return err
			}
		}
	}

	s.logger.Info("control signals stream completed", zap.String("experiment_id", req.ExperimentId))
	return nil
}

// StreamDriftDetections streams drift detections (placeholder implementation)
func (s *ControllerServer) StreamDriftDetections(req *controllerv1.GetControlLoopStatusRequest, stream controllerv1.ControllerService_StreamDriftDetectionsServer) error {
	if req.ExperimentId == "" {
		return status.Error(codes.InvalidArgument, "experiment ID is required")
	}

	s.logger.Info("starting drift detections stream", zap.String("experiment_id", req.ExperimentId))

	// For demo purposes, send any existing detections and then close
	if detections, exists := s.driftDetections[req.ExperimentId]; exists {
		for _, detection := range detections {
			if err := stream.Send(detection); err != nil {
				s.logger.Error("failed to send detection", zap.Error(err))
				return err
			}
		}
	}

	s.logger.Info("drift detections stream completed", zap.String("experiment_id", req.ExperimentId))
	return nil
}

// Helper methods

func (s *ControllerServer) validateControlSignal(signal *controllerv1.ControlSignal) []*controllerv1.ValidationError {
	var errors []*controllerv1.ValidationError

	if signal.Type == controllerv1.SignalType_SIGNAL_TYPE_UNSPECIFIED {
		errors = append(errors, &controllerv1.ValidationError{
			Field:   "type",
			Message: "signal type is required",
		})
	}

	switch action := signal.Action.(type) {
	case *controllerv1.ControlSignal_TrafficSplit:
		if action.TrafficSplit.CandidatePercentage < 0 || action.TrafficSplit.CandidatePercentage > 100 {
			errors = append(errors, &controllerv1.ValidationError{
				Field:   "traffic_split.candidate_percentage",
				Message: "percentage must be between 0 and 100",
			})
		}
	case *controllerv1.ControlSignal_PipelineState:
		if action.PipelineState.PipelineId == "" {
			errors = append(errors, &controllerv1.ValidationError{
				Field:   "pipeline_state.pipeline_id",
				Message: "pipeline ID is required",
			})
		}
	case *controllerv1.ControlSignal_Rollback:
		if action.Rollback.TargetPipelineId == "" {
			errors = append(errors, &controllerv1.ValidationError{
				Field:   "rollback.target_pipeline_id",
				Message: "target pipeline ID is required",
			})
		}
	case *controllerv1.ControlSignal_ConfigUpdate:
		if action.ConfigUpdate.PipelineId == "" {
			errors = append(errors, &controllerv1.ValidationError{
				Field:   "config_update.pipeline_id",
				Message: "pipeline ID is required",
			})
		}
	}

	return errors
}

func (s *ControllerServer) executeSignal(ctx context.Context, signal *controllerv1.ControlSignal) error {
	// In a real implementation, this would:
	// 1. Update traffic routing for traffic split signals
	// 2. Update pipeline configurations for config update signals
	// 3. Trigger rollbacks for rollback signals
	// 4. Update pipeline states for state change signals

	s.logger.Info("executing signal",
		zap.String("signal_id", signal.Id),
		zap.String("type", signal.Type.String()),
	)

	// Simulate execution time
	time.Sleep(100 * time.Millisecond)

	// For demo purposes, just log the action
	switch action := signal.Action.(type) {
	case *controllerv1.ControlSignal_TrafficSplit:
		s.logger.Info("adjusting traffic split",
			zap.String("baseline", action.TrafficSplit.BaselinePipelineId),
			zap.String("candidate", action.TrafficSplit.CandidatePipelineId),
			zap.Int32("percentage", action.TrafficSplit.CandidatePercentage),
		)
	case *controllerv1.ControlSignal_PipelineState:
		s.logger.Info("updating pipeline state",
			zap.String("pipeline", action.PipelineState.PipelineId),
			zap.String("state", action.PipelineState.TargetState.String()),
		)
	case *controllerv1.ControlSignal_Rollback:
		s.logger.Info("performing rollback",
			zap.String("from", action.Rollback.TargetPipelineId),
			zap.String("to", action.Rollback.RollbackToPipelineId),
		)
	case *controllerv1.ControlSignal_ConfigUpdate:
		s.logger.Info("updating configuration",
			zap.String("pipeline", action.ConfigUpdate.PipelineId),
			zap.Int("changes", len(action.ConfigUpdate.Changes)),
		)
	}

	return nil
}