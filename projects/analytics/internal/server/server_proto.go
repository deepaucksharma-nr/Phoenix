// +build proto

package server

import (
	"context"

	"github.com/phoenix/platform/pkg/analytics"
	"github.com/phoenix/platform/projects/analytics/api"
	"github.com/phoenix/platform/projects/analytics/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements the Analytics gRPC service
type Server struct {
	api.UnimplementedAnalyticsServer
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
	api.RegisterAnalyticsServer(grpcServer, s)
}

// AnalyzeMetrics implements the AnalyzeMetrics RPC method
func (s *Server) AnalyzeMetrics(ctx context.Context, req *api.AnalyzeMetricsRequest) (*api.AnalyzeMetricsResponse, error) {
	metrics := make([]analytics.Metric, len(req.Metrics))
	for i, m := range req.Metrics {
		metrics[i] = analytics.Metric{
			Timestamp: m.Timestamp.AsTime(),
			Name:      m.Name,
			Value:     m.Value,
			Labels:    m.Labels,
		}
	}

	result, err := s.service.AnalyzeMetrics(ctx, metrics)
	if err != nil {
		return nil, err
	}

	return &api.AnalyzeMetricsResponse{
		Result: &api.AnalysisResult{
			Timestamp: timestamppb.New(result.Timestamp),
			Summary:   result.Summary,
			Stats:     result.Stats,
		},
	}, nil
}

// DetectAnomalies implements the DetectAnomalies RPC method
func (s *Server) DetectAnomalies(ctx context.Context, req *api.DetectAnomaliesRequest) (*api.DetectAnomaliesResponse, error) {
	metrics := make([]analytics.Metric, len(req.Metrics))
	for i, m := range req.Metrics {
		metrics[i] = analytics.Metric{
			Timestamp: m.Timestamp.AsTime(),
			Name:      m.Name,
			Value:     m.Value,
			Labels:    m.Labels,
		}
	}

	anomalies, err := s.service.DetectAnomalies(ctx, metrics)
	if err != nil {
		return nil, err
	}

	response := &api.DetectAnomaliesResponse{
		Anomalies: make([]*api.Anomaly, len(anomalies)),
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
			Details: a.Details,
		}
	}

	return response, nil
}

// GenerateReport implements the GenerateReport RPC method
func (s *Server) GenerateReport(ctx context.Context, req *api.GenerateReportRequest) (*api.GenerateReportResponse, error) {
	analysis := &analytics.AnalysisResult{
		Timestamp: req.Analysis.Timestamp.AsTime(),
		Summary:   req.Analysis.Summary,
		Stats:     req.Analysis.Stats,
	}

	report, err := s.service.GenerateReport(ctx, analysis)
	if err != nil {
		return nil, err
	}

	return &api.GenerateReportResponse{
		Report: &api.Report{
			GeneratedAt:     timestamppb.New(report.GeneratedAt),
			Summary:         report.Summary,
			Recommendations: report.Recommendations,
		},
	}, nil
}
