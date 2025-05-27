package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	internalModels "github.com/phoenix/platform/projects/phoenix-api/internal/models"
	internalStore "github.com/phoenix/platform/projects/phoenix-api/internal/store"
)

// ExperimentService handles experiment-related operations
type ExperimentService struct {
	store  internalStore.Store
	logger *zap.Logger
	wsHub  interface{} // WebSocket hub interface
}

// NewExperimentService creates a new experiment service
func NewExperimentService(store internalStore.Store, generator interface{}, logger *zap.Logger) *ExperimentService {
	return &ExperimentService{
		store:  store,
		logger: logger,
	}
}

// CreateExperiment creates a new experiment (temporary implementation without proto)
func (s *ExperimentService) CreateExperiment(ctx context.Context, name, description, baselinePipeline, candidatePipeline string, targetNodes map[string]string) (*internalModels.Experiment, error) {
	s.logger.Info("creating experiment",
		zap.String("name", name),
		zap.String("baseline", baselinePipeline),
		zap.String("candidate", candidatePipeline))

	// Convert map to array for target hosts
	targetHosts := make([]string, 0, len(targetNodes))
	for _, host := range targetNodes {
		targetHosts = append(targetHosts, host)
	}

	// Create experiment model
	experiment := &internalModels.Experiment{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Phase:       internalModels.PhasePending,
		Config: internalModels.ExperimentConfig{
			TargetHosts: targetHosts,
			BaselineTemplate: internalModels.PipelineTemplate{
				Name: baselinePipeline,
			},
			CandidateTemplate: internalModels.PipelineTemplate{
				Name: candidatePipeline,
			},
		},
		Status:    internalModels.ExperimentStatus{},
		Metadata:  map[string]interface{}{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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
func (s *ExperimentService) GetExperiment(ctx context.Context, id string) (*internalModels.Experiment, error) {
	return s.store.GetExperiment(ctx, id)
}

// ListExperiments lists all experiments
func (s *ExperimentService) ListExperiments(ctx context.Context) ([]*internalModels.Experiment, error) {
	return s.store.ListExperiments(ctx)
}

// UpdateExperimentStatus updates the status of an experiment
func (s *ExperimentService) UpdateExperimentStatus(ctx context.Context, id string, phase string) error {
	experiment, err := s.store.GetExperiment(ctx, id)
	if err != nil {
		return err
	}

	experiment.Phase = phase
	experiment.UpdatedAt = time.Now()

	if phase == internalModels.PhaseRunning && experiment.Status.StartTime == nil {
		now := time.Now()
		experiment.Status.StartTime = &now
	}

	if (phase == internalModels.PhaseCompleted || phase == internalModels.PhaseFailed || phase == internalModels.PhaseStopped) && experiment.Status.EndTime == nil {
		now := time.Now()
		experiment.Status.EndTime = &now
	}

	return s.store.UpdateExperiment(ctx, experiment)
}

// DeleteExperiment deletes an experiment
func (s *ExperimentService) DeleteExperiment(ctx context.Context, id string) error {
	// For now, we'll just mark it as deleted by updating status
	return s.UpdateExperimentStatus(ctx, id, internalModels.PhaseDeleted)
}

// SetWebSocketHub sets the WebSocket hub for broadcasting updates
func (s *ExperimentService) SetWebSocketHub(hub interface{}) {
	s.wsHub = hub
}
