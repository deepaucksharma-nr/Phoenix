// Package api contains the generated protobuf types
// This is a placeholder until protoc is available
package api

import (
	"context"
	"google.golang.org/grpc"
)

// UnimplementedAnalyticsServiceServer is a placeholder for the generated server
type UnimplementedAnalyticsServiceServer struct{}

// AnalyzeMetricsRequest placeholder
type AnalyzeMetricsRequest struct {
	ExperimentId string
	WindowId     string
	Metrics      map[string]float64
}

// AnalyzeMetricsResponse placeholder
type AnalyzeMetricsResponse struct {
	Analysis *Analysis
}

// Analysis placeholder
type Analysis struct {
	Summary     string
	Insights    []string
	Anomalies   int32
	Performance float64
}

// AnalyticsServiceServer placeholder interface
type AnalyticsServiceServer interface {
	AnalyzeMetrics(context.Context, *AnalyzeMetricsRequest) (*AnalyzeMetricsResponse, error)
}

// RegisterAnalyticsServiceServer placeholder
func RegisterAnalyticsServiceServer(s *grpc.Server, srv AnalyticsServiceServer) {
	// Placeholder implementation
}