// +build !proto

package server

import (
	"github.com/phoenix/platform/projects/anomaly-detector/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Server implements a stub anomaly detector server
type Server struct {
	service *service.Service
	logger  *zap.Logger
}

// New creates a new stub server
func New(svc *service.Service, logger *zap.Logger) *Server {
	return &Server{
		service: svc,
		logger:  logger,
	}
}

// Register registers the server with gRPC (no-op for stub)
func (s *Server) Register(grpcServer *grpc.Server) {
	s.logger.Info("Anomaly detector server running without proto support")
}