package services

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/google/uuid"

	pb "github.com/phoenix/platform/packages/contracts/proto/v1"
	"github.com/phoenix/platform/packages/go-common/models"
	"github.com/phoenix/platform/packages/go-common/store"
)

// ExperimentService handles experiment-related operations
type ExperimentService struct {
	pb.UnimplementedExperimentServiceServer
	store     store.Store
	logger    *zap.Logger
}

// NewExperimentService creates a new experiment service
func NewExperimentService(store store.Store, generator interface{}, logger *zap.Logger) *ExperimentService {
	return &ExperimentService{
		store:     store,
		logger:    logger,
	}
}

// CreateExperiment creates a new experiment
func (s *ExperimentService) CreateExperiment(ctx context.Context, req *pb.CreateExperimentRequest) (*pb.CreateExperimentResponse, error) {
	// Validate request
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "experiment name is required")
	}
	if req.BaselinePipeline == "" {
		return nil, status.Error(codes.InvalidArgument, "baseline pipeline is required")
	}
	if req.CandidatePipeline == "" {
		return nil, status.Error(codes.InvalidArgument, "candidate pipeline is required")
	}
	if len(req.TargetNodes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one target node is required")
	}

	s.logger.Info("creating experiment", 
		zap.String("name", req.Name),
		zap.String("baseline", req.BaselinePipeline),
		zap.String("candidate", req.CandidatePipeline),
	)
	
	// Create experiment model
	now := time.Now()
	experiment := &models.Experiment{
		ID:                uuid.New().String(),
		Name:              req.Name,
		Description:       req.Description,
		BaselinePipeline:  req.BaselinePipeline,
		CandidatePipeline: req.CandidatePipeline,
		Status:            models.ExperimentStatusPending,
		TargetNodes:       req.TargetNodes,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	
	// Store experiment
	if err := s.store.CreateExperiment(ctx, experiment); err != nil {
		s.logger.Error("failed to store experiment", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create experiment")
	}
	
	// Convert to proto
	protoExp := &pb.Experiment{
		Id:                experiment.ID,
		Name:              experiment.Name,
		Description:       experiment.Description,
		BaselinePipeline:  experiment.BaselinePipeline,
		CandidatePipeline: experiment.CandidatePipeline,
		Status:            experiment.Status,
		TargetNodes:       experiment.TargetNodes,
		CreatedAt:         experiment.CreatedAt.Unix(),
		UpdatedAt:         experiment.UpdatedAt.Unix(),
	}
	
	s.logger.Info("experiment created successfully", zap.String("id", experiment.ID))
	
	return &pb.CreateExperimentResponse{
		Experiment: protoExp,
	}, nil
}

// GetExperiment retrieves an experiment by ID
func (s *ExperimentService) GetExperiment(ctx context.Context, req *pb.GetExperimentRequest) (*pb.GetExperimentResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "experiment ID is required")
	}

	s.logger.Info("getting experiment", zap.String("id", req.Id))
	
	// Retrieve from store
	experiment, err := s.store.GetExperiment(ctx, req.Id)
	if err != nil {
		s.logger.Error("failed to get experiment", zap.String("id", req.Id), zap.Error(err))
		return nil, status.Error(codes.NotFound, "experiment not found")
	}
	
	// Convert to proto
	protoExp := &pb.Experiment{
		Id:                experiment.ID,
		Name:              experiment.Name,
		Description:       experiment.Description,
		BaselinePipeline:  experiment.BaselinePipeline,
		CandidatePipeline: experiment.CandidatePipeline,
		Status:            experiment.Status,
		TargetNodes:       experiment.TargetNodes,
		CreatedAt:         experiment.CreatedAt.Unix(),
		UpdatedAt:         experiment.UpdatedAt.Unix(),
	}
	
	if experiment.StartedAt != nil {
		protoExp.StartedAt = experiment.StartedAt.Unix()
	}
	if experiment.CompletedAt != nil {
		protoExp.CompletedAt = experiment.CompletedAt.Unix()
	}
	
	return &pb.GetExperimentResponse{
		Experiment: protoExp,
	}, nil
}

// ListExperiments lists all experiments
func (s *ExperimentService) ListExperiments(ctx context.Context, req *pb.ListExperimentsRequest) (*pb.ListExperimentsResponse, error) {
	s.logger.Info("listing experiments", zap.Int32("page_size", req.PageSize))
	
	// Set default page size
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 50 // Default page size
	}
	
	// List from store  
	experiments, err := s.store.ListExperiments(ctx, int(pageSize), 0) // limit, offset
	if err != nil {
		s.logger.Error("failed to list experiments", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list experiments")
	}
	
	// Convert to proto
	var protoExps []*pb.Experiment
	for _, exp := range experiments {
		protoExp := &pb.Experiment{
			Id:                exp.ID,
			Name:              exp.Name,
			Description:       exp.Description,
			BaselinePipeline:  exp.BaselinePipeline,
			CandidatePipeline: exp.CandidatePipeline,
			Status:            exp.Status,
			TargetNodes:       exp.TargetNodes,
			CreatedAt:         exp.CreatedAt.Unix(),
			UpdatedAt:         exp.UpdatedAt.Unix(),
		}
		
		if exp.StartedAt != nil {
			protoExp.StartedAt = exp.StartedAt.Unix()
		}
		if exp.CompletedAt != nil {
			protoExp.CompletedAt = exp.CompletedAt.Unix()
		}
		
		protoExps = append(protoExps, protoExp)
	}
	
	return &pb.ListExperimentsResponse{
		Experiments: protoExps,
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