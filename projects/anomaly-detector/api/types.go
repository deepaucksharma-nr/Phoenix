// Package api contains the generated protobuf types
// This is a placeholder until protoc is available
package api

import (
	"context"
	"google.golang.org/grpc"
)

// UnimplementedAnomalyDetectorServer is a placeholder for the generated server
type UnimplementedAnomalyDetectorServer struct{}

// DetectAnomaliesRequest placeholder
type DetectAnomaliesRequest struct {
	WindowId  string
	Metrics   map[string]float64
	Timestamp int64
}

// DetectAnomaliesResponse placeholder  
type DetectAnomaliesResponse struct {
	Anomalies []*Anomaly
}

// Anomaly placeholder
type Anomaly struct {
	MetricName string
	Score      float64
	Severity   string
	Message    string
}

// AnomalyDetectorServer placeholder interface
type AnomalyDetectorServer interface {
	DetectAnomalies(context.Context, *DetectAnomaliesRequest) (*DetectAnomaliesResponse, error)
}

// RegisterAnomalyDetectorServer placeholder
func RegisterAnomalyDetectorServer(s *grpc.Server, srv AnomalyDetectorServer) {
	// Placeholder implementation
}