package grpc

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/phoenix-vnext/platform/projects/controller/internal/controller"
	pb "github.com/phoenix-vnext/platform/packages/contracts/proto/v1"
)

// AdapterServer implements the gRPC experiment service using the proto definitions
type AdapterServer struct {
	pb.UnimplementedExperimentServiceServer
	logger     *zap.Logger
	controller *controller.ExperimentController
}

// NewAdapterServer creates a new gRPC adapter server
func NewAdapterServer(logger *zap.Logger, controller *controller.ExperimentController) *AdapterServer {
	return &AdapterServer{
		logger:     logger,
		controller: controller,
	}
}

// CreateExperiment handles experiment creation requests
func (s *AdapterServer) CreateExperiment(ctx context.Context, req *pb.CreateExperimentRequest) (*pb.CreateExperimentResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	s.logger.Info("creating experiment via adapter", zap.String("name", req.Name))

	// For now, return a simple response with a basic experiment
	exp := &pb.Experiment{
		Id:          "exp-" + time.Now().Format("20060102-150405"),
		Name:        req.Name,
		Description: req.Description,
		Status:      "pending",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	return &pb.CreateExperimentResponse{
		Experiment: exp,
	}, nil
}

// GetExperiment retrieves an experiment by ID
func (s *AdapterServer) GetExperiment(ctx context.Context, req *pb.GetExperimentRequest) (*pb.GetExperimentResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Return a minimal experiment
	exp := &pb.Experiment{
		Id:          req.Id,
		Name:        "Test Experiment",
		Description: "Experiment managed by controller",
		Status:      "pending",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	return &pb.GetExperimentResponse{
		Experiment: exp,
	}, nil
}

// ListExperiments retrieves a list of experiments
func (s *AdapterServer) ListExperiments(ctx context.Context, req *pb.ListExperimentsRequest) (*pb.ListExperimentsResponse, error) {
	return &pb.ListExperimentsResponse{
		Experiments: []*pb.Experiment{},
	}, nil
}

// UpdateExperiment updates an experiment
func (s *AdapterServer) UpdateExperiment(ctx context.Context, req *pb.UpdateExperimentRequest) (*pb.Experiment, error) {
	if req.Experiment == nil || req.Experiment.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "experiment with id is required")
	}

	// For now, just return the experiment from the request
	return req.Experiment, nil
}

// DeleteExperiment deletes an experiment
func (s *AdapterServer) DeleteExperiment(ctx context.Context, req *pb.DeleteExperimentRequest) (*pb.DeleteExperimentResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	return &pb.DeleteExperimentResponse{}, nil
}

// GetExperimentStatus gets the status of an experiment
func (s *AdapterServer) GetExperimentStatus(ctx context.Context, req *pb.GetExperimentStatusRequest) (*pb.ExperimentStatus, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	return &pb.ExperimentStatus{
		Status:  "pending",
		Message: "Experiment is pending",
	}, nil
}