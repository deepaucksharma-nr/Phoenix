// +build proto

package server

import (
	"context"

	"github.com/phoenix/platform/pkg/analytics"
	"github.com/phoenix/platform/projects/anomaly-detector/api"
	"github.com/phoenix/platform/projects/anomaly-detector/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements the AnomalyDetector gRPC service
type Server struct {
	api.UnimplementedAnomalyDetectorServer
	service *service.Service
	logger  *zap.Logger
}

// New creates a new gRPC server
func New(svc *service.Service, logger *zap.Logger) *Server {
	return &Server{
		service: svc,
		logger:  logger,
	}
}

// Register registers the server with the gRPC server
func (s *Server) Register(grpcServer *grpc.Server) {
	api.RegisterAnomalyDetectorServer(grpcServer, s)
}

// Detect implements the Detect RPC method
func (s *Server) Detect(ctx context.Context, req *api.DetectRequest) (*api.DetectResponse, error) {
	metrics := make([]analytics.Metric, len(req.Metrics))
	for i, m := range req.Metrics {
		metrics[i] = analytics.Metric{
			Timestamp: m.Timestamp.AsTime(),
			Name:      m.Name,
			Value:     m.Value,
			Labels:    m.Labels,
		}
	}

	var config *service.Config
	if req.Config != nil {
		config = &service.Config{
			Threshold:  req.Config.Threshold,
			WindowSize: int(req.Config.WindowSize),
			Algorithms: req.Config.Algorithms,
			Parameters: req.Config.Parameters,
		}
	}

	anomalies, err := s.service.Detect(ctx, metrics, config)
	if err != nil {
		return nil, err
	}

	response := &api.DetectResponse{
		Anomalies: make([]*api.Anomaly, len(anomalies)),
		Status:    "success",
		Stats:     make(map[string]float64),
	}

	for i, a := range anomalies {
		response.Anomalies[i] = &api.Anomaly{
			Timestamp: timestamppb.New(a.Timestamp),
			Type:      a.Type,
			Severity:  a.Severity,
			Metric: &api.Metric{
				Timestamp: timestamppb.New(a.Metric.Timestamp),
				Name:      a.Metric.Name,
				Value:     a.Metric.Value,
				Labels:    a.Metric.Labels,
			},
			Details:    make(map[string]string),
			Confidence: 0.95, // TODO: Calculate actual confidence
		}

		// Convert details to string map
		for k, v := range a.Details {
			if str, ok := v.(string); ok {
				response.Anomalies[i].Details[k] = str
			}
		}
	}

	return response, nil
}

// ConfigureDetection implements the ConfigureDetection RPC method
func (s *Server) ConfigureDetection(ctx context.Context, req *api.ConfigureRequest) (*api.ConfigureResponse, error) {
	config := service.Config{
		Threshold:  req.Config.Threshold,
		WindowSize: int(req.Config.WindowSize),
		Algorithms: req.Config.Algorithms,
		Parameters: req.Config.Parameters,
	}

	err := s.service.ConfigureDetection(ctx, config)
	if err != nil {
		return &api.ConfigureResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &api.ConfigureResponse{
		Success: true,
		Message: "Configuration updated successfully",
	}, nil
}

// GetDetectionStatus implements the GetDetectionStatus RPC method
func (s *Server) GetDetectionStatus(ctx context.Context, req *api.StatusRequest) (*api.StatusResponse, error) {
	config, stats, err := s.service.GetDetectionStatus(ctx)
	if err != nil {
		return nil, err
	}

	return &api.StatusResponse{
		Active: true,
		CurrentConfig: &api.DetectionConfig{
			Threshold:  config.Threshold,
			WindowSize: int32(config.WindowSize),
			Algorithms: config.Algorithms,
			Parameters: config.Parameters,
		},
		Stats: stats,
	}, nil
}
