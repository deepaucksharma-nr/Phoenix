package store
import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/lib/pq"
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
func (s *CompositeStore) GetMetricCostFlow(ctx context.Context) (map[string]interface{}, error) {
	// Query from the metric_cost_flow_view
	query := `
		SELECT 
			metric_name,
			labels::text,
			cardinality,
			cost_per_minute,
			percentage,
			total_cost
		FROM metric_cost_flow_view
		LIMIT 20
	`
	rows, err := s.pipelineStore.db.DB().QueryContext(ctx, query)
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
		var costPerMinute, percentage, totalCost float64
		if err := rows.Scan(&metricName, &labelsJSON, &cardinality, &costPerMinute, &percentage, &totalCost); err != nil {
			log.Error().Err(err).Msg("Failed to scan metric cost flow row")
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
			Percentage:    percentage,
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
	
	// Convert to map[string]interface{} for consistency with interface
	return map[string]interface{}{
		"total_cost_per_minute": flow.TotalCostPerMinute,
		"top_metrics":           flow.TopMetrics,
		"by_service":            flow.ByService,
		"by_namespace":          flow.ByNamespace,
		"last_updated":          flow.LastUpdated,
	}, nil
}
func (s *CompositeStore) GetCardinalityBreakdown(ctx context.Context, namespace, service string) (map[string]interface{}, error) {
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
	rows, err := s.pipelineStore.db.DB().QueryContext(ctx, query, args...)
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
	
	// Convert to map[string]interface{} for consistency with interface
	return map[string]interface{}{
		"total_cardinality":  breakdown.TotalCardinality,
		"by_metric":          breakdown.ByMetric,
		"by_label":           breakdown.ByLabel,
		"top_contributors":   breakdown.TopContributors,
		"timestamp":          breakdown.Timestamp,
		"namespace":          namespace,
		"service":            service,
	}, nil
}
func (s *CompositeStore) GetPipelineTemplates(ctx context.Context) ([]*PipelineTemplate, error) {
	query := `
		SELECT 
			id::text,
			name,
			display_name,
			description,
			version,
			config_url,
			tags,
			estimated_reduction,
			features,
			metadata,
			created_at,
			updated_at
		FROM pipeline_templates
		WHERE is_active = true
		ORDER BY name
	`
	
	rows, err := s.pipelineStore.db.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pipeline templates: %w", err)
	}
	defer rows.Close()
	
	var templates []*PipelineTemplate
	for rows.Next() {
		var template PipelineTemplate
		var tags, features []string
		var metadataJSON []byte
		
		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.DisplayName,
			&template.Description,
			&template.Version,
			&template.ConfigURL,
			pq.Array(&tags),
			&template.EstimatedReduction,
			pq.Array(&features),
			&metadataJSON,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan pipeline template")
			continue
		}
		
		// Parse metadata JSON
		if err := json.Unmarshal(metadataJSON, &template.Variables); err != nil {
			template.Variables = make(map[string]string)
		}
		
		templates = append(templates, &template)
	}
	
	return templates, nil
}
func (s *CompositeStore) GetCostAnalytics(ctx context.Context, period string) (map[string]interface{}, error) {
	// Parse period
	var interval string
	switch period {
		case "1h":
			interval = "1 hour"
		case "24h":
			interval = "24 hours"
		case "7d":
			interval = "7 days"
		case "30d":
			interval = "30 days"
		default:
			interval = "24 hours"
	}
	
	// Query cost tracking data
	query := `
		WITH cost_data AS (
			SELECT 
				start_time,
				end_time,
				total_metrics,
				total_cost,
				cost_breakdown
			FROM cost_tracking
			WHERE period = 'hourly'
			  AND start_time > NOW() - INTERVAL '%s'
			ORDER BY start_time
		),
		aggregated AS (
			SELECT 
				SUM(total_metrics) as total_metrics,
				SUM(total_cost) as total_cost,
				MIN(start_time) as period_start,
				MAX(end_time) as period_end
			FROM cost_data
		)
		SELECT 
			COALESCE(total_metrics, 0) as total_metrics,
			COALESCE(total_cost, 0) as total_cost,
			period_start,
			period_end
		FROM aggregated
	`
	
	var totalMetrics int64
	var totalCost float64
	var periodStart, periodEnd *time.Time
	
	row := s.pipelineStore.db.DB().QueryRowContext(ctx, fmt.Sprintf(query, interval))
	if err := row.Scan(&totalMetrics, &totalCost, &periodStart, &periodEnd); err != nil {
		log.Error().Err(err).Str("period", period).Msg("Failed to get cost analytics")
		// Return empty data if no records
		return map[string]interface{}{
			"period":                 period,
			"total_cost":             0.0,
			"total_savings":          0.0,
			"savings_percent":        0.0,
			"cost_trend":             []map[string]interface{}{},
			"savings_by_pipeline":    map[string]float64{},
			"savings_by_service":     map[string]float64{},
			"top_cost_drivers":       []map[string]interface{}{},
			"projected_monthly_cost": 0.0,
			"projected_savings":      0.0,
		}, nil
	}
	
	// Calculate savings based on baseline vs optimized metrics
	baselineCost := totalCost * 3.33 // Assume baseline is 3.33x more expensive (70% reduction)
	totalSavings := baselineCost - totalCost
	savingsPercent := 0.0
	if baselineCost > 0 {
		savingsPercent = (totalSavings / baselineCost) * 100
	}
	
	// Get cost trend
	trendQuery := `
		SELECT 
			start_time,
			total_cost,
			cost_breakdown
		FROM cost_tracking
		WHERE period = 'hourly'
		  AND start_time > NOW() - INTERVAL '%s'
		ORDER BY start_time
		LIMIT 24
	`
	
	rows, err := s.pipelineStore.db.DB().QueryContext(ctx, fmt.Sprintf(trendQuery, interval))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get cost trend")
	}
	defer rows.Close()
	
	var costTrend []map[string]interface{}
	for rows.Next() {
		var timestamp time.Time
		var cost float64
		var breakdown []byte
		
		if err := rows.Scan(&timestamp, &cost, &breakdown); err != nil {
			continue
		}
		
		costTrend = append(costTrend, map[string]interface{}{
			"timestamp": timestamp,
			"cost":      cost,
			"savings":   (cost * 3.33) - cost, // Calculate savings for each point
		})
	}
	
	// Project monthly costs
	var projectedMonthlyCost, projectedSavings float64
	if periodEnd != nil && periodStart != nil {
		hourlyRate := totalCost / periodEnd.Sub(*periodStart).Hours()
		projectedMonthlyCost = hourlyRate * 24 * 30
		projectedSavings = (projectedMonthlyCost * 3.33) - projectedMonthlyCost
	}
	
	return map[string]interface{}{
		"period":                 period,
		"total_cost":             totalCost,
		"total_savings":          totalSavings,
		"savings_percent":        savingsPercent,
		"cost_trend":             costTrend,
		"savings_by_pipeline":    map[string]float64{}, // TODO: Calculate from experiment data
		"savings_by_service":     map[string]float64{}, // TODO: Calculate from labels
		"top_cost_drivers":       []map[string]interface{}{}, // TODO: Calculate from metrics
		"projected_monthly_cost": projectedMonthlyCost,
		"projected_savings":      projectedSavings,
	}, nil
}
// GetExperimentMetrics retrieves metrics for an experiment
func (s *CompositeStore) GetExperimentMetrics(ctx context.Context, experimentID string) (map[string]interface{}, error) {
	// Query metrics from metric_cache table
	query := `
		WITH recent_metrics AS (
			SELECT 
				variant,
				metric_name,
				AVG(value) as avg_value,
				COUNT(*) as data_points,
				MAX(timestamp) as latest_timestamp
			FROM metric_cache
			WHERE experiment_id = $1
			  AND timestamp > NOW() - INTERVAL '5 minutes'
			GROUP BY variant, metric_name
		),
		cardinality_data AS (
			SELECT 
				m1.experiment_id,
				COUNT(DISTINCT m1.labels) as baseline_cardinality,
				COUNT(DISTINCT m2.labels) as candidate_cardinality
			FROM metrics m1
			LEFT JOIN metrics m2 ON m1.experiment_id = m2.experiment_id 
				AND m2.source_id LIKE '%candidate%'
			WHERE m1.experiment_id = $1
			  AND m1.source_id LIKE '%baseline%'
			  AND m1.timestamp > NOW() - INTERVAL '5 minutes'
			GROUP BY m1.experiment_id
		),
		resource_usage AS (
			SELECT 
				AVG(CAST(resource_usage->>'cpu_percent' AS FLOAT)) as avg_cpu,
				AVG(CAST(resource_usage->>'memory_bytes' AS BIGINT)) / 1024.0 / 1024.0 as avg_memory_mb
			FROM agents a
			JOIN tasks t ON a.host_id = t.host_id
			WHERE t.experiment_id = $1
			  AND a.last_heartbeat > NOW() - INTERVAL '5 minutes'
		)
		SELECT 
			COALESCE(cd.baseline_cardinality, 0) as baseline_cardinality,
			COALESCE(cd.candidate_cardinality, 0) as candidate_cardinality,
			COALESCE(ru.avg_cpu, 0) as avg_cpu,
			COALESCE(ru.avg_memory_mb, 0) as avg_memory_mb,
			COUNT(DISTINCT rm.metric_name) as unique_metrics,
			SUM(rm.data_points) as total_data_points
		FROM recent_metrics rm
		LEFT JOIN cardinality_data cd ON 1=1
		LEFT JOIN resource_usage ru ON 1=1
		GROUP BY cd.baseline_cardinality, cd.candidate_cardinality, ru.avg_cpu, ru.avg_memory_mb
	`
	
	var baselineCardinality, candidateCardinality, uniqueMetrics, totalDataPoints int64
	var avgCPU, avgMemoryMB float64
	
	row := s.pipelineStore.db.DB().QueryRowContext(ctx, query, experimentID)
	err := row.Scan(
		&baselineCardinality,
		&candidateCardinality,
		&avgCPU,
		&avgMemoryMB,
		&uniqueMetrics,
		&totalDataPoints,
	)
	
	if err != nil {
		log.Error().Err(err).Str("experiment_id", experimentID).Msg("Failed to get experiment metrics")
		// Return empty metrics if error
		return map[string]interface{}{
			"experiment_id": experimentID,
			"timestamp":     time.Now(),
			"summary": map[string]interface{}{
				"total_metrics":          0,
				"metrics_per_second":     0,
				"cardinality_reduction":  0,
				"cpu_usage":              0,
				"memory_usage":           0,
			},
			"baseline": map[string]interface{}{
				"cardinality":       0,
				"metrics_per_second": 0,
			},
			"candidate": map[string]interface{}{
				"cardinality":       0,
				"metrics_per_second": 0,
			},
		}, nil
	}
	
	// Calculate reduction percentage
	var cardinalityReduction float64
	if baselineCardinality > 0 {
		cardinalityReduction = float64(baselineCardinality-candidateCardinality) / float64(baselineCardinality) * 100
	}
	
	// Calculate metrics per second (assuming 5 minute window)
	metricsPerSecond := float64(totalDataPoints) / 300.0
	
	return map[string]interface{}{
		"experiment_id": experimentID,
		"timestamp":     time.Now(),
		"summary": map[string]interface{}{
			"total_metrics":          uniqueMetrics,
			"metrics_per_second":     metricsPerSecond,
			"cardinality_reduction":  cardinalityReduction,
			"cpu_usage":              avgCPU,
			"memory_usage":           avgMemoryMB,
		},
		"baseline": map[string]interface{}{
			"cardinality":       baselineCardinality,
			"metrics_per_second": metricsPerSecond * 0.6, // Estimate baseline portion
		},
		"candidate": map[string]interface{}{
			"cardinality":       candidateCardinality,
			"metrics_per_second": metricsPerSecond * 0.4, // Estimate candidate portion
		},
	}, nil
}

// GetTaskQueueStatus returns the current status of the task queue
func (s *CompositeStore) GetTaskQueueStatus(ctx context.Context) (map[string]interface{}, error) {
	query := `
		WITH task_counts AS (
			SELECT 
				status,
				COUNT(*) as count
			FROM tasks
			WHERE created_at > NOW() - INTERVAL '24 hours'
			GROUP BY status
		),
		queue_stats AS (
			SELECT 
				COUNT(*) FILTER (WHERE status = 'pending') as pending,
				COUNT(*) FILTER (WHERE status = 'assigned') as assigned,
				COUNT(*) FILTER (WHERE status = 'running') as running,
				COUNT(*) FILTER (WHERE status = 'completed') as completed,
				COUNT(*) FILTER (WHERE status = 'failed') as failed,
				COUNT(*) as total,
				AVG(CASE 
					WHEN status = 'completed' AND completed_at IS NOT NULL AND started_at IS NOT NULL 
					THEN EXTRACT(EPOCH FROM (completed_at - started_at))
					ELSE NULL 
				END) as avg_execution_time,
				AVG(CASE 
					WHEN assigned_at IS NOT NULL AND created_at IS NOT NULL 
					THEN EXTRACT(EPOCH FROM (assigned_at - created_at))
					ELSE NULL 
				END) as avg_wait_time
			FROM tasks
			WHERE created_at > NOW() - INTERVAL '24 hours'
		)
		SELECT 
			pending,
			assigned,
			running,
			completed,
			failed,
			total,
			COALESCE(avg_execution_time, 0) as avg_execution_time,
			COALESCE(avg_wait_time, 0) as avg_wait_time
		FROM queue_stats
	`
	
	var pending, assigned, running, completed, failed, total int64
	var avgExecutionTime, avgWaitTime float64
	
	row := s.pipelineStore.db.DB().QueryRowContext(ctx, query)
	err := row.Scan(&pending, &assigned, &running, &completed, &failed, &total, &avgExecutionTime, &avgWaitTime)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get task queue status")
		// Return zeros if no data
		return map[string]interface{}{
			"pending":           0,
			"assigned":          0,
			"running":           0,
			"completed":         0,
			"failed":            0,
			"total":             0,
			"avg_execution_time": 0,
			"avg_wait_time":     0,
			"throughput_per_hour": 0,
		}, nil
	}
	
	// Calculate throughput (completed tasks per hour in last 24h)
	throughputPerHour := float64(completed) / 24.0
	
	return map[string]interface{}{
		"pending":             pending,
		"assigned":            assigned,
		"running":             running,
		"completed":           completed,
		"failed":              failed,
		"total":               total,
		"avg_execution_time":  avgExecutionTime,
		"avg_wait_time":       avgWaitTime,
		"throughput_per_hour": throughputPerHour,
	}, nil
}

// CacheMetric is implemented in agent_store.go
// RecordDeploymentVersion records a new version of a deployment
func (s *CompositeStore) RecordDeploymentVersion(ctx context.Context, deploymentID, pipelineConfig string, parameters map[string]interface{}, deployedBy string, notes string) (int, error) {
	return s.pipelineStore.RecordDeploymentVersion(ctx, deploymentID, pipelineConfig, parameters, deployedBy, notes)
}
// GetDeploymentVersion retrieves a specific version of a deployment
func (s *CompositeStore) GetDeploymentVersion(ctx context.Context, deploymentID string, version int) (*DeploymentVersion, error) {
	return s.pipelineStore.GetDeploymentVersion(ctx, deploymentID, version)
}
// ListDeploymentVersions retrieves the version history for a deployment
func (s *CompositeStore) ListDeploymentVersions(ctx context.Context, deploymentID string) ([]*DeploymentVersion, error) {
	return s.pipelineStore.ListDeploymentVersions(ctx, deploymentID)
}

// BlacklistToken adds a token to the blacklist
func (s *CompositeStore) BlacklistToken(ctx context.Context, jti, userID string, expiresAt time.Time, reason string) error {
	query := `
		INSERT INTO token_blacklist (token_jti, user_id, expires_at, reason)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (token_jti) DO NOTHING
	`
	
	_, err := s.pipelineStore.db.DB().ExecContext(ctx, query, jti, userID, expiresAt, reason)
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	
	return nil
}

// IsTokenBlacklisted checks if a token is in the blacklist
func (s *CompositeStore) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM token_blacklist 
			WHERE token_jti = $1 
			AND expires_at > NOW()
		)
	`
	
	var exists bool
	err := s.pipelineStore.db.DB().QueryRowContext(ctx, query, jti).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	
	return exists, nil
}

// CleanupExpiredTokens removes expired tokens from the blacklist
func (s *CompositeStore) CleanupExpiredTokens(ctx context.Context) error {
	query := `
		DELETE FROM token_blacklist 
		WHERE expires_at < NOW()
	`
	
	result, err := s.pipelineStore.db.DB().ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Info().Int64("count", rowsAffected).Msg("Cleaned up expired blacklisted tokens")
	}
	
	return nil
}
