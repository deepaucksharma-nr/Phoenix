package grpc

// Temporary stub definitions until proto files are generated
// TODO: Remove this file once proto generation is set up

import (
	// "context"
	"time"

	"go.uber.org/zap"
	"github.com/phoenix-vnext/platform/projects/controller/internal/controller"
)

// StubAdapterServer is a temporary implementation that doesn't use proto
type StubAdapterServer struct {
	logger     *zap.Logger
	controller *controller.ExperimentController
}

// NewAdapterServer creates a new adapter server
func NewAdapterServer(logger *zap.Logger, controller *controller.ExperimentController) *StubAdapterServer {
	return &StubAdapterServer{
		logger:     logger,
		controller: controller,
	}
}

// Temporary experiment response
type ExperimentResponse struct {
	ID          string
	Name        string
	Description string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}