package grpc

import (
	// "context"
	// "fmt"
	// "time"

	// "github.com/google/uuid"
	"go.uber.org/zap"
	// "google.golang.org/grpc/codes"
	// "google.golang.org/grpc/status"

	"github.com/phoenix/platform/projects/controller/internal/controller"
	// pb "github.com/phoenix/platform/pkg/contracts/proto/v1"
)

// SimpleExperimentServer implements a basic experiment service
type SimpleExperimentServer struct {
	// pb.UnimplementedExperimentServiceServer
	logger     *zap.Logger
	controller *controller.ExperimentController
}

// NewSimpleExperimentServer creates a new simple experiment gRPC server
func NewSimpleExperimentServer(logger *zap.Logger, controller *controller.ExperimentController) *SimpleExperimentServer {
	return &SimpleExperimentServer{
		logger:     logger,
		controller: controller,
	}
}

// CreateExperiment creates a new experiment
// func (s *SimpleExperimentServer) CreateExperiment(ctx context.Context, req *pb.CreateExperimentRequest) (*pb.CreateExperimentResponse, error) {
/*
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "experiment name is required")
	}

	if req.BaselinePipeline == "" {
		return nil, status.Error(codes.InvalidArgument, "baseline pipeline is required")
	}

	if req.CandidatePipeline == "" {
		return nil, status.Error(codes.InvalidArgument, "candidate pipeline is required")
	}

	s.logger.Info("creating experiment",
		zap.String("name", req.Name),
		zap.String("baseline_pipeline", req.BaselinePipeline),
		zap.String("candidate_pipeline", req.CandidatePipeline),
	)

	// Convert target nodes from map to slice if needed
	var targetNodes []string
	if req.TargetNodes != nil {
		for node := range req.TargetNodes {
			targetNodes = append(targetNodes, node)
		}
	}

	// Convert to domain model
	experiment := &controller.Experiment{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Phase:       controller.ExperimentPhasePending,
		Config: controller.ExperimentConfig{
			BaselinePipeline:  req.BaselinePipeline,
			CandidatePipeline: req.CandidatePipeline,
			TargetHosts:       targetNodes,
			Duration:          24 * time.Hour, // Default duration
		},
	}

	// Create experiment via controller
	err := s.controller.CreateExperiment(ctx, experiment)
	if err != nil {
		s.logger.Error("failed to create experiment", zap.Error(err))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create experiment: %v", err))
	}

	// Get the created experiment (CreateExperiment only returns error)
	createdExp := experiment

	// Convert back to proto response
	protoExp := &pb.Experiment{
		Id:          createdExp.ID,
		Name:        createdExp.Name,
		Description: createdExp.Description,
		Status:      string(createdExp.Phase),
	}

	s.logger.Info("experiment created successfully", zap.String("id", protoExp.Id))

	return &pb.CreateExperimentResponse{
		Experiment: protoExp,
	}, nil
}
*/

// GetExperiment retrieves an experiment by ID
// func (s *SimpleExperimentServer) GetExperiment(ctx context.Context, req *pb.GetExperimentRequest) (*pb.GetExperimentResponse, error) {
/*
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "experiment ID is required")
	}

	s.logger.Debug("getting experiment", zap.String("id", req.Id))

	experiment, err := s.controller.GetExperiment(ctx, req.Id)
	if err != nil {
		s.logger.Error("failed to get experiment", zap.String("id", req.Id), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to retrieve experiment")
	}

	protoExp := &pb.Experiment{
		Id:          experiment.ID,
		Name:        experiment.Name,
		Description: experiment.Description,
	}

	return &pb.GetExperimentResponse{
		Experiment: protoExp,
	}, nil
}
*/

// ListExperiments lists experiments with optional filters
// func (s *SimpleExperimentServer) ListExperiments(ctx context.Context, req *pb.ListExperimentsRequest) (*pb.ListExperimentsResponse, error) {
/*
	s.logger.Debug("listing experiments")

	// For now, use basic listing without complex filters
	filter := controller.ExperimentFilter{
		Limit: 50,
	}
	experiments, err := s.controller.ListExperiments(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list experiments", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list experiments")
	}

	var protoExps []*pb.Experiment
	for _, exp := range experiments {
		protoExp := &pb.Experiment{
			Id:          exp.ID,
			Name:        exp.Name,
			Description: exp.Description,
			Status:      string(exp.Phase),
		}
		protoExps = append(protoExps, protoExp)
	}

	s.logger.Debug("listed experiments", zap.Int("count", len(protoExps)))

	return &pb.ListExperimentsResponse{
		Experiments: protoExps,
	}, nil
}
*/

// UpdateExperiment updates an experiment
// func (s *SimpleExperimentServer) UpdateExperiment(ctx context.Context, req *pb.UpdateExperimentRequest) (*pb.Experiment, error) {
/*
	if req.Experiment == nil {
		return nil, status.Error(codes.InvalidArgument, "experiment is required")
	}

	if req.Experiment.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "experiment ID is required")
	}

	s.logger.Info("updating experiment", zap.String("id", req.Experiment.Id))

	// Get existing experiment
	experiment, err := s.controller.GetExperiment(ctx, req.Experiment.Id)
	if err != nil {
		s.logger.Error("failed to get experiment for update", zap.String("id", req.Experiment.Id), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to retrieve experiment")
	}

	// Update fields if provided
	if req.Experiment.Name != "" {
		experiment.Name = req.Experiment.Name
	}
	if req.Experiment.Description != "" {
		experiment.Description = req.Experiment.Description
	}

	// Note: The controller doesn't have an UpdateExperiment method, only UpdateExperimentPhase
	// For now, just return the experiment (this would need proper implementation)
	
	protoExp := &pb.Experiment{
		Id:          experiment.ID,
		Name:        experiment.Name,
		Description: experiment.Description,
		Status:      string(experiment.Phase),
	}

	s.logger.Info("experiment updated successfully", zap.String("id", req.Experiment.Id))

	return protoExp, nil
}
*/

// DeleteExperiment deletes an experiment
// func (s *SimpleExperimentServer) DeleteExperiment(ctx context.Context, req *pb.DeleteExperimentRequest) (*pb.DeleteExperimentResponse, error) {
/*
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "experiment ID is required")
	}

	s.logger.Info("deleting experiment", zap.String("id", req.Id))

	// Note: DeleteExperiment method not implemented in controller yet
	// For now, just log the operation
	s.logger.Info("experiment delete operation logged", zap.String("id", req.Id))

	return &pb.DeleteExperimentResponse{}, nil
}
*/

// GetExperimentStatus gets the status of an experiment
// func (s *SimpleExperimentServer) GetExperimentStatus(ctx context.Context, req *pb.GetExperimentStatusRequest) (*pb.ExperimentStatus, error) {
/*
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "experiment ID is required")
	}

	s.logger.Debug("getting experiment status", zap.String("id", req.Id))

	experiment, err := s.controller.GetExperiment(ctx, req.Id)
	if err != nil {
		s.logger.Error("failed to get experiment status", zap.String("id", req.Id), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to retrieve experiment status")
	}

	status := &pb.ExperimentStatus{
		Status:  string(experiment.Phase),
		Message: fmt.Sprintf("Experiment %s is in %s phase", experiment.ID, experiment.Phase),
	}

	return status, nil
}
*/