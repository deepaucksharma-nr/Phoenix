package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	commonModels "github.com/phoenix/platform/pkg/common/models"
	"github.com/phoenix/platform/pkg/models"
	commonstore "github.com/phoenix/platform/pkg/common/store"
	internalModels "github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/rs/zerolog/log"
)

// CompositeStore implements the full Store interface by combining different stores
type CompositeStore struct {
	postgresStore *commonstore.PostgresStore
	pipelineStore *PostgresPipelineDeploymentStore
}

// NewCompositeStore creates a new composite store
func NewCompositeStore(postgresStore *commonstore.PostgresStore, pipelineStore *PostgresPipelineDeploymentStore) Store {
	return &CompositeStore{
		postgresStore: postgresStore,
		pipelineStore: pipelineStore,
	}
}

// Experiment operations (delegated to internal models store)
func (s *CompositeStore) CreateExperiment(ctx context.Context, experiment *internalModels.Experiment) error {
	// Convert internal model to common model
	commonExp := &models.Experiment{
		ID:                experiment.ID,
		Name:              experiment.Name,
		Description:       experiment.Description,
		BaselinePipeline:  experiment.Config.BaselineTemplate.Name,
		CandidatePipeline: experiment.Config.CandidateTemplate.Name,
		Status:            experiment.Phase,
		TargetNodes:       experiment.Config.TargetHosts,
		CreatedAt:         experiment.CreatedAt,
		UpdatedAt:         experiment.UpdatedAt,
	}

	if experiment.ID == "" {
		commonExp.ID = fmt.Sprintf("exp-%d", time.Now().Unix())
		experiment.ID = commonExp.ID
	}

	return s.postgresStore.CreateExperiment(ctx, commonExp)
}

func (s *CompositeStore) GetExperiment(ctx context.Context, experimentID string) (*internalModels.Experiment, error) {
	commonExp, err := s.postgresStore.GetExperiment(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	// Convert common model to internal model
	// Convert map[string]string to []string for target hosts
	targetHosts := make([]string, 0, len(commonExp.TargetNodes))
	for _, host := range commonExp.TargetNodes {
		targetHosts = append(targetHosts, host)
	}

	return &internalModels.Experiment{
		ID:          commonExp.ID,
		Name:        commonExp.Name,
		Description: commonExp.Description,
		Phase:       commonExp.Status,
		Config: internalModels.ExperimentConfig{
			TargetHosts: targetHosts,
			BaselineTemplate: internalModels.PipelineTemplate{
				Name: commonExp.BaselinePipeline,
			},
			CandidateTemplate: internalModels.PipelineTemplate{
				Name: commonExp.CandidatePipeline,
			},
		},
		Status:    internalModels.ExperimentStatus{},
		Metadata:  map[string]interface{}{},
		CreatedAt: commonExp.CreatedAt,
		UpdatedAt: commonExp.UpdatedAt,
	}, nil
}

func (s *CompositeStore) ListExperiments(ctx context.Context) ([]*internalModels.Experiment, error) {
	commonExps, err := s.postgresStore.ListExperiments(ctx, 100, 0)
	if err != nil {
		return nil, err
	}

	experiments := make([]*internalModels.Experiment, 0, len(commonExps))
	for _, commonExp := range commonExps {
		// Convert map[string]string to []string for target hosts
		targetHosts := make([]string, 0, len(commonExp.TargetNodes))
		for _, host := range commonExp.TargetNodes {
			targetHosts = append(targetHosts, host)
		}

		experiments = append(experiments, &internalModels.Experiment{
			ID:          commonExp.ID,
			Name:        commonExp.Name,
			Description: commonExp.Description,
			Phase:       commonExp.Status,
			Config: internalModels.ExperimentConfig{
				TargetHosts: targetHosts,
				BaselineTemplate: internalModels.PipelineTemplate{
					Name: commonExp.BaselinePipeline,
				},
				CandidateTemplate: internalModels.PipelineTemplate{
					Name: commonExp.CandidatePipeline,
				},
			},
			Status:    internalModels.ExperimentStatus{},
			Metadata:  map[string]interface{}{},
			CreatedAt: commonExp.CreatedAt,
			UpdatedAt: commonExp.UpdatedAt,
		})
	}

	return experiments, nil
}

func (s *CompositeStore) UpdateExperiment(ctx context.Context, experiment *internalModels.Experiment) error {
	commonExp := &models.Experiment{
		ID:                experiment.ID,
		Name:              experiment.Name,
		Description:       experiment.Description,
		BaselinePipeline:  experiment.Config.BaselineTemplate.Name,
		CandidatePipeline: experiment.Config.CandidateTemplate.Name,
		Status:            experiment.Phase,
		TargetNodes:       experiment.Config.TargetHosts,
		CreatedAt:         experiment.CreatedAt,
		UpdatedAt:         time.Now(),
	}

	return s.postgresStore.UpdateExperiment(ctx, commonExp)
}

func (s *CompositeStore) UpdateExperimentPhase(ctx context.Context, experimentID string, phase string) error {
	exp, err := s.postgresStore.GetExperiment(ctx, experimentID)
	if err != nil {
		return err
	}

	exp.Status = phase
	exp.UpdatedAt = time.Now()

	return s.postgresStore.UpdateExperiment(ctx, exp)
}

func (s *CompositeStore) DeleteExperiment(ctx context.Context, experimentID string) error {
	return s.postgresStore.DeleteExperiment(ctx, experimentID)
}

// Pipeline deployment operations (delegated to pipeline store)
func (s *CompositeStore) CreateDeployment(ctx context.Context, deployment *commonModels.PipelineDeployment) error {
	return s.pipelineStore.CreateDeployment(ctx, deployment)
}

func (s *CompositeStore) GetDeployment(ctx context.Context, deploymentID string) (*commonModels.PipelineDeployment, error) {
	return s.pipelineStore.GetDeployment(ctx, deploymentID)
}

func (s *CompositeStore) ListDeployments(ctx context.Context, req *commonModels.ListDeploymentsRequest) ([]*commonModels.PipelineDeployment, int, error) {
	return s.pipelineStore.ListDeployments(ctx, req)
}

func (s *CompositeStore) UpdateDeployment(ctx context.Context, deploymentID string, update *commonModels.UpdateDeploymentRequest) error {
	return s.pipelineStore.UpdateDeployment(ctx, deploymentID, update)
}

func (s *CompositeStore) DeleteDeployment(ctx context.Context, deploymentID string) error {
	return s.pipelineStore.DeleteDeployment(ctx, deploymentID)
}

func (s *CompositeStore) UpdateDeploymentMetrics(ctx context.Context, deploymentID string, metrics *commonModels.DeploymentMetrics) error {
	return s.pipelineStore.UpdateDeploymentMetrics(ctx, deploymentID, metrics)
}

// Task and Agent operations are implemented in all_methods.go

// Event operations
func (s *CompositeStore) CreateExperimentEvent(ctx context.Context, event *internalModels.ExperimentEvent) error {
	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO experiment_events (
			experiment_id, event_type, phase, message, metadata
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err = s.pipelineStore.db.DB().QueryRowContext(ctx, query,
		event.ExperimentID, event.EventType, event.Phase,
		event.Message, string(metadataJSON),
	).Scan(&event.ID, &event.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create experiment event: %w", err)
	}

	return nil
}

func (s *CompositeStore) ListExperimentEvents(ctx context.Context, experimentID string) ([]*internalModels.ExperimentEvent, error) {
	query := `
		SELECT id, experiment_id, event_type, phase, message, metadata, created_at
		FROM experiment_events
		WHERE experiment_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := s.pipelineStore.db.DB().QueryContext(ctx, query, experimentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list experiment events: %w", err)
	}
	defer rows.Close()

	var events []*internalModels.ExperimentEvent

	for rows.Next() {
		var event internalModels.ExperimentEvent
		var metadataJSON string

		err := rows.Scan(
			&event.ID, &event.ExperimentID, &event.EventType,
			&event.Phase, &event.Message, &metadataJSON, &event.CreatedAt,
		)

		if err != nil {
			log.Error().Err(err).Msg("Failed to scan event row")
			continue
		}

		// Unmarshal metadata
		if err := json.Unmarshal([]byte(metadataJSON), &event.Metadata); err != nil {
			event.Metadata = make(map[string]interface{})
		}

		events = append(events, &event)
	}

	return events, nil
}

// UI-specific operations (TODO: Implement)
func (s *CompositeStore) GetMetricCostFlow(ctx context.Context) (*MetricCostFlow, error) {
	// TODO: Implement
	return &MetricCostFlow{
		TotalCostPerMinute: 0,
		TopMetrics:         []MetricCostDetail{},
		ByService:          map[string]float64{},
		ByNamespace:        map[string]float64{},
		LastUpdated:        time.Now(),
	}, nil
}

func (s *CompositeStore) GetCardinalityBreakdown(ctx context.Context, namespace, service string) (*CardinalityBreakdown, error) {
	// TODO: Implement
	return &CardinalityBreakdown{
		TotalCardinality: 0,
		ByMetric:         map[string]int64{},
		ByLabel:          map[string]int64{},
		TopContributors:  []CardinalityContributor{},
		Timestamp:        time.Now(),
	}, nil
}

func (s *CompositeStore) GetPipelineTemplates(ctx context.Context) ([]*PipelineTemplate, error) {
	// TODO: Implement - this should load from a templates table or config files
	// For now, return some default templates
	return []*PipelineTemplate{
		{
			ID:          "process-baseline-v1",
			Name:        "Process Baseline",
			Description: "Standard process metrics collection",
			ConfigURL:   "/configs/pipelines/process-baseline.yaml",
			Variables: map[string]string{
				"sampling_rate":  "1.0",
				"include_labels": "true",
			},
		},
		{
			ID:          "process-topk-v1",
			Name:        "Process TopK",
			Description: "Optimized process metrics with TopK filtering",
			ConfigURL:   "/configs/pipelines/process-topk.yaml",
			Variables: map[string]string{
				"top_k":         "10",
				"sampling_rate": "0.1",
			},
		},
		{
			ID:          "process-adaptive-v1",
			Name:        "Process Adaptive",
			Description: "Adaptive filtering based on metric importance",
			ConfigURL:   "/configs/pipelines/process-adaptive.yaml",
			Variables: map[string]string{
				"threshold":     "0.8",
				"learning_rate": "0.01",
			},
		},
	}, nil
}

func (s *CompositeStore) GetCostAnalytics(ctx context.Context, period string) (*CostAnalytics, error) {
	// TODO: Implement
	return &CostAnalytics{
		Period:               period,
		TotalCost:            0,
		TotalSavings:         0,
		SavingsPercent:       0,
		CostTrend:            []CostDataPoint{},
		SavingsByPipeline:    map[string]float64{},
		SavingsByService:     map[string]float64{},
		TopCostDrivers:       []CostDriver{},
		ProjectedMonthlyCost: 0,
		ProjectedSavings:     0,
	}, nil
}

// CacheMetric is implemented in agent_store.go
