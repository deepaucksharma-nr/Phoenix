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
}

// NewExperimentService creates a new experiment service
func NewExperimentService(store store.Store, generator interface{}, logger *zap.Logger) *ExperimentService {
	return &ExperimentService{
		store:     store,
		logger:    logger,
	}
}

// CreateExperiment creates a new experiment (temporary implementation without proto)
func (s *ExperimentService) CreateExperiment(ctx context.Context, name, description, baselinePipeline, candidatePipeline string, targetNodes []string, trafficPercentage float64) (*models.Experiment, error) {
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
		TrafficPercentage: trafficPercentage,
		State:             models.ExperimentStatePending,
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
	return s.store.ListExperiments(ctx)
}

// UpdateExperimentState updates the state of an experiment
func (s *ExperimentService) UpdateExperimentState(ctx context.Context, id string, state models.ExperimentState) error {
	return s.store.UpdateExperimentState(ctx, id, state)
}

// DeleteExperiment deletes an experiment
func (s *ExperimentService) DeleteExperiment(ctx context.Context, id string) error {
	return s.store.DeleteExperiment(ctx, id)
}