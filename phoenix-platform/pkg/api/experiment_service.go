package api

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/phoenix/platform/pkg/api/v1"
	"github.com/phoenix/platform/pkg/generator"
	"github.com/phoenix/platform/pkg/store"
)

// ExperimentService handles experiment-related operations
type ExperimentService struct {
	pb.UnimplementedExperimentServiceServer
	store     store.Store
	generator *generator.Service
	logger    *zap.Logger
}

// NewExperimentService creates a new experiment service
func NewExperimentService(store store.Store, generator *generator.Service, logger *zap.Logger) *ExperimentService {
	return &ExperimentService{
		store:     store,
		generator: generator,
		logger:    logger,
	}
}

// CreateExperiment creates a new experiment
func (s *ExperimentService) CreateExperiment(ctx context.Context, req *pb.CreateExperimentRequest) (*pb.CreateExperimentResponse, error) {
	s.logger.Info("creating experiment", zap.String("name", req.Name))
	
	// TODO: Implement experiment creation
	experiment := &pb.Experiment{
		Id:                "exp-123",
		Name:              req.Name,
		Description:       req.Description,
		BaselinePipeline:  req.BaselinePipeline,
		CandidatePipeline: req.CandidatePipeline,
		Status:            "pending",
		TargetNodes:       req.TargetNodes,
	}
	
	return &pb.CreateExperimentResponse{
		Experiment: experiment,
	}, nil
}

// GetExperiment retrieves an experiment by ID
func (s *ExperimentService) GetExperiment(ctx context.Context, req *pb.GetExperimentRequest) (*pb.GetExperimentResponse, error) {
	s.logger.Info("getting experiment", zap.String("id", req.Id))
	
	// TODO: Implement experiment retrieval
	experiment := &pb.Experiment{
		Id:     req.Id,
		Name:   "test-experiment",
		Status: "running",
	}
	
	return &pb.GetExperimentResponse{
		Experiment: experiment,
	}, nil
}

// ListExperiments lists all experiments
func (s *ExperimentService) ListExperiments(ctx context.Context, req *pb.ListExperimentsRequest) (*pb.ListExperimentsResponse, error) {
	s.logger.Info("listing experiments", zap.Int32("page_size", req.PageSize))
	
	// TODO: Implement experiment listing
	return &pb.ListExperimentsResponse{
		Experiments: []*pb.Experiment{},
	}, nil
}

// UpdateExperiment updates an experiment
func (s *ExperimentService) UpdateExperiment(ctx context.Context, req *pb.UpdateExperimentRequest) (*pb.Experiment, error) {
	if req.Experiment == nil {
		return nil, status.Error(codes.InvalidArgument, "experiment is required")
	}
	
	s.logger.Info("updating experiment", zap.String("id", req.Experiment.Id))
	
	// TODO: Implement experiment update
	return req.Experiment, nil
}

// DeleteExperiment deletes an experiment
func (s *ExperimentService) DeleteExperiment(ctx context.Context, req *pb.DeleteExperimentRequest) (*pb.DeleteExperimentResponse, error) {
	s.logger.Info("deleting experiment", zap.String("id", req.Id))
	
	// TODO: Implement experiment deletion
	return &pb.DeleteExperimentResponse{}, nil
}

// GetExperimentStatus gets the status of an experiment
func (s *ExperimentService) GetExperimentStatus(ctx context.Context, req *pb.GetExperimentStatusRequest) (*pb.ExperimentStatus, error) {
	s.logger.Info("getting experiment status", zap.String("id", req.Id))
	
	// TODO: Implement status retrieval
	return &pb.ExperimentStatus{
		Status:  "running",
		Message: "Experiment is running",
	}, nil
}