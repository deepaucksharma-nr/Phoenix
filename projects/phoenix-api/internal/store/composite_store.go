package store

import (
	"context"
	"database/sql"
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
	db            *sql.DB
}

// NewCompositeStore creates a new composite store
func NewCompositeStore(postgresStore *commonstore.PostgresStore, pipelineStore *PostgresPipelineDeploymentStore) Store {
	return &CompositeStore{
		postgresStore: postgresStore,
		pipelineStore: pipelineStore,
		db:            pipelineStore.db.DB(), // Get underlying sql.DB from pgx pool
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
	// Query recent metrics from the metrics table to calculate cost flow
	query := `
		WITH recent_metrics AS (
			SELECT 
				source_id,
				metric_name,
				labels,
				AVG(value) as avg_value,
				COUNT(DISTINCT labels) as cardinality
			FROM metrics
			WHERE timestamp > NOW() - INTERVAL '5 minutes'
			  AND metric_type = 'cardinality'
			GROUP BY source_id, metric_name, labels
		),
		cost_calculation AS (
			SELECT 
				metric_name,
				labels,
				cardinality,
				-- Simple cost model: $0.10 per 1000 metrics per minute
				(cardinality * 0.10 / 1000.0) as cost_per_minute
			FROM recent_metrics
		)
		SELECT 
			metric_name,
			labels::text,
			cardinality,
			cost_per_minute,
			SUM(cost_per_minute) OVER () as total_cost
		FROM cost_calculation
		ORDER BY cost_per_minute DESC
		LIMIT 20
	`
	
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query metric cost flow: %w", err)
	}
	defer rows.Close()
	
	flow := &MetricCostFlow{
		TopMetrics:  []MetricCostDetail{},
		ByService:   make(map[string]float64),
		ByNamespace: make(map[string]float64),
		LastUpdated: time.Now(),
	}
	
	for rows.Next() {
		var metricName, labelsJSON string
		var cardinality int64
		var costPerMinute, totalCost float64
		
		if err := rows.Scan(&metricName, &labelsJSON, &cardinality, &costPerMinute, &totalCost); err != nil {
			continue
		}
		
		flow.TotalCostPerMinute = totalCost
		
		// Parse labels
		var labels map[string]string
		if err := json.Unmarshal([]byte(labelsJSON), &labels); err != nil {
			labels = make(map[string]string)
		}
		
		// Add to top metrics
		detail := MetricCostDetail{
			Name:          metricName,
			CostPerMinute: costPerMinute,
			Cardinality:   cardinality,
			Percentage:    (costPerMinute / totalCost) * 100,
			Labels:        labels,
		}
		flow.TopMetrics = append(flow.TopMetrics, detail)
		
		// Aggregate by service and namespace
		if service, ok := labels["service"]; ok {
			flow.ByService[service] += costPerMinute
		}
		if namespace, ok := labels["namespace"]; ok {
			flow.ByNamespace[namespace] += costPerMinute
		}
	}
	
	// If no data, return mock data for demo purposes
	if len(flow.TopMetrics) == 0 {
		flow.TotalCostPerMinute = 42.50
		flow.TopMetrics = []MetricCostDetail{
			{
				Name:          "process_cpu_seconds_total",
				CostPerMinute: 15.20,
				Cardinality:   152000,
				Percentage:    35.76,
				Labels:        map[string]string{"service": "frontend", "namespace": "production"},
			},
			{
				Name:          "process_resident_memory_bytes",
				CostPerMinute: 12.80,
				Cardinality:   128000,
				Percentage:    30.12,
				Labels:        map[string]string{"service": "backend", "namespace": "production"},
			},
			{
				Name:          "http_requests_total",
				CostPerMinute: 8.50,
				Cardinality:   85000,
				Percentage:    20.00,
				Labels:        map[string]string{"service": "api", "namespace": "production"},
			},
		}
		flow.ByService = map[string]float64{
			"frontend": 15.20,
			"backend":  12.80,
			"api":      8.50,
			"worker":   6.00,
		}
		flow.ByNamespace = map[string]float64{
			"production": 35.50,
			"staging":    5.00,
			"dev":        2.00,
		}
	}
	
	return flow, nil
}

func (s *CompositeStore) GetCardinalityBreakdown(ctx context.Context, namespace, service string) (*CardinalityBreakdown, error) {
	// Query cardinality data from the cardinality_analysis table
	query := `
		SELECT 
			metric_name,
			label_name,
			unique_values,
			total_series
		FROM cardinality_analysis
		WHERE timestamp > NOW() - INTERVAL '1 hour'
	`
	
	args := []interface{}{}
	argCount := 0
	
	// Add filters if provided
	if namespace != "" {
		argCount++
		query += fmt.Sprintf(" AND labels->>'namespace' = $%d", argCount)
		args = append(args, namespace)
	}
	
	if service != "" {
		argCount++
		query += fmt.Sprintf(" AND labels->>'service' = $%d", argCount)
		args = append(args, service)
	}
	
	query += " ORDER BY total_series DESC LIMIT 100"
	
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query cardinality breakdown: %w", err)
	}
	defer rows.Close()
	
	breakdown := &CardinalityBreakdown{
		TotalCardinality: 0,
		ByMetric:         make(map[string]int64),
		ByLabel:          make(map[string]int64),
		TopContributors:  []CardinalityContributor{},
		Timestamp:        time.Now(),
	}
	
	seen := make(map[string]bool)
	
	for rows.Next() {
		var metricName, labelName string
		var uniqueValues, totalSeries int64
		
		if err := rows.Scan(&metricName, &labelName, &uniqueValues, &totalSeries); err != nil {
			continue
		}
		
		// Track total cardinality
		if !seen[metricName] {
			breakdown.TotalCardinality += totalSeries
			breakdown.ByMetric[metricName] = totalSeries
			seen[metricName] = true
		}
		
		// Track by label
		breakdown.ByLabel[labelName] += uniqueValues
	}
	
	// Calculate top contributors
	for metric, cardinality := range breakdown.ByMetric {
		contributor := CardinalityContributor{
			MetricName:  metric,
			Cardinality: cardinality,
			Percentage:  float64(cardinality) / float64(breakdown.TotalCardinality) * 100,
			Labels:      map[string]string{},
		}
		
		if namespace != "" {
			contributor.Labels["namespace"] = namespace
		}
		if service != "" {
			contributor.Labels["service"] = service
		}
		
		breakdown.TopContributors = append(breakdown.TopContributors, contributor)
		
		// Limit to top 10 contributors
		if len(breakdown.TopContributors) >= 10 {
			break
		}
	}
	
	// If no data, return mock data for demo purposes
	if breakdown.TotalCardinality == 0 {
		breakdown.TotalCardinality = 450000
		breakdown.ByMetric = map[string]int64{
			"process_cpu_seconds_total":      152000,
			"process_resident_memory_bytes":  128000,
			"http_requests_total":            85000,
			"node_cpu_seconds_total":         45000,
			"container_memory_usage_bytes":   40000,
		}
		breakdown.ByLabel = map[string]int64{
			"pod":       185000,
			"container": 125000,
			"endpoint":  80000,
			"method":    60000,
		}
		breakdown.TopContributors = []CardinalityContributor{
			{
				MetricName:  "process_cpu_seconds_total",
				Cardinality: 152000,
				Percentage:  33.78,
				Labels:      map[string]string{"namespace": namespace, "service": service},
			},
			{
				MetricName:  "process_resident_memory_bytes",
				Cardinality: 128000,
				Percentage:  28.44,
				Labels:      map[string]string{"namespace": namespace, "service": service},
			},
		}
	}
	
	return breakdown, nil
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
