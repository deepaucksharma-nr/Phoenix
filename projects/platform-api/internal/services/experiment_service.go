package services

import (
	"context"
	"time"

	"go.uber.org/zap"
	"github.com/google/uuid"

	"github.com/phoenix-vnext/platform/packages/go-common/models"
	"github.com/phoenix-vnext/platform/packages/go-common/store"
)

// ExperimentService handles experiment-related operations
type ExperimentService struct {
	store     store.Store
	logger    *zap.Logger
	wsHub     interface{} // WebSocket hub interface
}

// NewExperimentService creates a new experiment service
func NewExperimentService(store store.Store, generator interface{}, logger *zap.Logger) *ExperimentService {
	return &ExperimentService{
		store:     store,
		logger:    logger,
	}
}

// CreateExperiment creates a new experiment (temporary implementation without proto)
func (s *ExperimentService) CreateExperiment(ctx context.Context, name, description, baselinePipeline, candidatePipeline string, targetNodes map[string]string) (*models.Experiment, error) {
	s.logger.Info("creating experiment", 
		zap.String("name", name),
		zap.String("baseline", baselinePipeline),
		zap.String("candidate", candidatePipeline))

	// Create experiment model
	experiment := &models.Experiment{
		ID:                uuid.New().String(),
		Name:              name,
		Description:       description,
		BaselinePipeline:  baselinePipeline,
		CandidatePipeline: candidatePipeline,
		TargetNodes:       targetNodes,
		Status:            models.ExperimentStatusPending,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Save to store
	if err := s.store.CreateExperiment(ctx, experiment); err != nil {
		s.logger.Error("failed to create experiment", zap.Error(err))
		return nil, err
	}

	s.logger.Info("experiment created successfully", zap.String("id", experiment.ID))
	return experiment, nil
}

// GetExperiment retrieves an experiment by ID
func (s *ExperimentService) GetExperiment(ctx context.Context, id string) (*models.Experiment, error) {
	return s.store.GetExperiment(ctx, id)
}

// ListExperiments lists all experiments
func (s *ExperimentService) ListExperiments(ctx context.Context) ([]*models.Experiment, error) {
	// For now, use default pagination
	return s.store.ListExperiments(ctx, 100, 0)
}

// UpdateExperimentStatus updates the status of an experiment
func (s *ExperimentService) UpdateExperimentStatus(ctx context.Context, id string, status string) error {
	experiment, err := s.store.GetExperiment(ctx, id)
	if err != nil {
		return err
	}
	
	experiment.Status = status
	experiment.UpdatedAt = time.Now()
	
	if status == models.ExperimentStatusRunning && experiment.StartedAt == nil {
		now := time.Now()
		experiment.StartedAt = &now
	}
	
	if (status == models.ExperimentStatusCompleted || status == models.ExperimentStatusFailed || status == models.ExperimentStatusStopped) && experiment.CompletedAt == nil {
		now := time.Now()
		experiment.CompletedAt = &now
	}
	
	return s.store.UpdateExperiment(ctx, experiment)
}

// DeleteExperiment deletes an experiment
func (s *ExperimentService) DeleteExperiment(ctx context.Context, id string) error {
	// For now, we'll just mark it as deleted by updating status
	return s.UpdateExperimentStatus(ctx, id, "deleted")
}

// SetWebSocketHub sets the WebSocket hub for broadcasting updates
func (s *ExperimentService) SetWebSocketHub(hub interface{}) {
	s.wsHub = hub
}