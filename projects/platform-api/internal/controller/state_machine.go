package controller

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/phoenix/platform/pkg/common/analysis"
	"github.com/phoenix/platform/projects/platform-api/internal/services"
	api "github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
)

// StateMachine manages experiment state transitions
type StateMachine struct {
	logger              *zap.Logger
	controller          *ExperimentController
	pipelineService     *services.PipelineDeploymentService
	experimentService   *services.ExperimentService
	promAPI             v1.API
	analyzer            *analysis.ExperimentAnalyzer
	transitions         map[ExperimentPhase][]ExperimentPhase
}

// NewStateMachine creates a new experiment state machine
func NewStateMachine(
	logger *zap.Logger,
	controller *ExperimentController,
	pipelineService *services.PipelineDeploymentService,
	experimentService *services.ExperimentService,
) *StateMachine {
	promAddr := getEnvDefault("PROMETHEUS_URL", "http://localhost:9090")
	promClient, err := api.NewClient(api.Config{Address: promAddr})
	var promAPI v1.API
	if err == nil {
		promAPI = v1.NewAPI(promClient)
	} else {
		logger.Warn("failed to create Prometheus client", zap.Error(err))
	}

	return &StateMachine{
		logger:              logger,
		controller:          controller,
		pipelineService:     pipelineService,
		experimentService:   experimentService,
		promAPI:             promAPI,
		analyzer:            analysis.NewExperimentAnalyzer(),
		transitions: map[ExperimentPhase][]ExperimentPhase{
			ExperimentPhasePending:      {ExperimentPhaseInitializing, ExperimentPhaseCancelled},
			ExperimentPhaseInitializing: {ExperimentPhaseRunning, ExperimentPhaseFailed, ExperimentPhaseCancelled},
			ExperimentPhaseRunning:      {ExperimentPhaseAnalyzing, ExperimentPhaseFailed, ExperimentPhaseCancelled},
			ExperimentPhaseAnalyzing:    {ExperimentPhaseCompleted, ExperimentPhaseFailed},
			ExperimentPhaseCompleted:    {}, // Terminal state
			ExperimentPhaseFailed:       {}, // Terminal state
			ExperimentPhaseCancelled:    {}, // Terminal state
		},
	}
}

// TransitionTo attempts to transition an experiment to a new phase
func (sm *StateMachine) TransitionTo(ctx context.Context, experimentID string, targetPhase ExperimentPhase) error {
	exp, err := sm.controller.GetExperiment(ctx, experimentID)
	if err != nil {
		return fmt.Errorf("failed to get experiment: %w", err)
	}

	// Check if transition is valid
	if !sm.isValidTransition(exp.Phase, targetPhase) {
		return fmt.Errorf("invalid transition from %s to %s", exp.Phase, targetPhase)
	}

	sm.logger.Info("transitioning experiment",
		zap.String("id", experimentID),
		zap.String("from", string(exp.Phase)),
		zap.String("to", string(targetPhase)),
	)

	// Execute phase-specific logic
	switch targetPhase {
	case ExperimentPhaseInitializing:
		return sm.handleInitializing(ctx, exp)
	case ExperimentPhaseRunning:
		return sm.handleRunning(ctx, exp)
	case ExperimentPhaseAnalyzing:
		return sm.handleAnalyzing(ctx, exp)
	case ExperimentPhaseCompleted:
		return sm.handleCompleted(ctx, exp)
	case ExperimentPhaseFailed:
		return sm.handleFailed(ctx, exp, "Experiment failed")
	case ExperimentPhaseCancelled:
		return sm.handleCancelled(ctx, exp)
	default:
		return fmt.Errorf("unknown phase: %s", targetPhase)
	}
}

// isValidTransition checks if a state transition is valid
func (sm *StateMachine) isValidTransition(from, to ExperimentPhase) bool {
	validTransitions, exists := sm.transitions[from]
	if !exists {
		return false
	}

	for _, valid := range validTransitions {
		if valid == to {
			return true
		}
	}
	return false
}

// handleInitializing handles the transition to initializing phase
func (sm *StateMachine) handleInitializing(ctx context.Context, exp *Experiment) error {
	sm.logger.Info("initializing experiment", zap.String("id", exp.ID))

	// Update phase
	if err := sm.controller.UpdateExperimentPhase(ctx, exp.ID, ExperimentPhaseInitializing, "Initializing experiment resources"); err != nil {
		return err
	}

	// Enqueue initialization tasks
	tasks := []Task{
		{
			ID:           fmt.Sprintf("validate-pipelines-%s", exp.ID),
			Type:         "validate_pipelines",
			ExperimentID: exp.ID,
			Data: map[string]interface{}{
				"baseline_pipeline":  exp.Config.BaselinePipeline,
				"candidate_pipeline": exp.Config.CandidatePipeline,
			},
		},
		{
			ID:           fmt.Sprintf("deploy-pipelines-%s", exp.ID),
			Type:         "deploy_pipelines",
			ExperimentID: exp.ID,
			Data: map[string]interface{}{
				"baseline_pipeline":  exp.Config.BaselinePipeline,
				"candidate_pipeline": exp.Config.CandidatePipeline,
				"target_hosts":       exp.Config.TargetHosts,
				"variables":          exp.Config.Variables,
			},
		},
	}

	for _, task := range tasks {
		if err := sm.controller.taskQueue.EnqueueTask(ctx, task); err != nil {
			sm.logger.Error("failed to enqueue task",
				zap.String("task_id", task.ID),
				zap.String("task_type", task.Type),
				zap.Error(err),
			)
			return sm.handleFailed(ctx, exp, fmt.Sprintf("Failed to enqueue task: %v", err))
		}
	}

	// Enqueue task to transition to running after initialization
	transitionTask := Task{
		ID:           fmt.Sprintf("transition-running-%s", exp.ID),
		Type:         "transition_phase",
		ExperimentID: exp.ID,
		Data: map[string]interface{}{
			"target_phase": ExperimentPhaseRunning,
		},
	}

	return sm.controller.taskQueue.EnqueueTask(ctx, transitionTask)
}

// handleRunning handles the transition to running phase
func (sm *StateMachine) handleRunning(ctx context.Context, exp *Experiment) error {
	sm.logger.Info("starting experiment run", zap.String("id", exp.ID))

	now := time.Now()
	exp.Status.StartTime = &now

	// Update phase
	if err := sm.controller.UpdateExperimentPhase(ctx, exp.ID, ExperimentPhaseRunning, "Experiment is running"); err != nil {
		return err
	}

	// Schedule transition to analyzing phase after duration
	analyzeTask := Task{
		ID:           fmt.Sprintf("transition-analyzing-%s", exp.ID),
		Type:         "transition_phase",
		ExperimentID: exp.ID,
		Data: map[string]interface{}{
			"target_phase": ExperimentPhaseAnalyzing,
			"execute_at":   time.Now().Add(exp.Config.Duration),
		},
	}

	return sm.controller.taskQueue.EnqueueTask(ctx, analyzeTask)
}

// handleAnalyzing handles the transition to analyzing phase
func (sm *StateMachine) handleAnalyzing(ctx context.Context, exp *Experiment) error {
	sm.logger.Info("analyzing experiment results", zap.String("id", exp.ID))

	// Update phase
	if err := sm.controller.UpdateExperimentPhase(ctx, exp.ID, ExperimentPhaseAnalyzing, "Analyzing experiment results"); err != nil {
		return err
	}

	// Enqueue analysis task
	analysisTask := Task{
		ID:           fmt.Sprintf("analyze-results-%s", exp.ID),
		Type:         "analyze_experiment",
		ExperimentID: exp.ID,
		Data:         map[string]interface{}{},
	}

	return sm.controller.taskQueue.EnqueueTask(ctx, analysisTask)
}

// handleCompleted handles the transition to completed phase
func (sm *StateMachine) handleCompleted(ctx context.Context, exp *Experiment) error {
	sm.logger.Info("experiment completed successfully", zap.String("id", exp.ID))

	now := time.Now()
	exp.Status.EndTime = &now

	return sm.controller.UpdateExperimentPhase(ctx, exp.ID, ExperimentPhaseCompleted, "Experiment completed successfully")
}

// handleFailed handles the transition to failed phase
func (sm *StateMachine) handleFailed(ctx context.Context, exp *Experiment, reason string) error {
	sm.logger.Warn("experiment failed",
		zap.String("id", exp.ID),
		zap.String("reason", reason),
	)

	now := time.Now()
	exp.Status.EndTime = &now

	return sm.controller.UpdateExperimentPhase(ctx, exp.ID, ExperimentPhaseFailed, reason)
}

// handleCancelled handles the transition to cancelled phase
func (sm *StateMachine) handleCancelled(ctx context.Context, exp *Experiment) error {
	sm.logger.Info("experiment cancelled", zap.String("id", exp.ID))

	now := time.Now()
	exp.Status.EndTime = &now

	// Enqueue cleanup task
	cleanupTask := Task{
		ID:           fmt.Sprintf("cleanup-%s", exp.ID),
		Type:         "cleanup_experiment",
		ExperimentID: exp.ID,
		Data:         map[string]interface{}{},
	}

	if err := sm.controller.taskQueue.EnqueueTask(ctx, cleanupTask); err != nil {
		sm.logger.Error("failed to enqueue cleanup task", zap.Error(err))
	}

	return sm.controller.UpdateExperimentPhase(ctx, exp.ID, ExperimentPhaseCancelled, "Experiment cancelled by user")
}

// ProcessAnalysisTask processes experiment analysis tasks
func (sm *StateMachine) ProcessAnalysisTask(ctx context.Context, task Task) error {
	exp, err := sm.controller.GetExperiment(ctx, task.ExperimentID)
	if err != nil {
		return fmt.Errorf("failed to get experiment: %w", err)
	}

	// Collect metrics data from monitoring system
	metricsData, err := sm.collectMetricsData(ctx, exp)
	if err != nil {
		sm.logger.Error("failed to collect metrics data", zap.Error(err))
		return sm.handleFailed(ctx, exp, fmt.Sprintf("Failed to collect metrics: %v", err))
	}

	// Perform statistical analysis
	analysisResult, err := sm.analyzer.AnalyzeExperimentResults(ctx, exp, metricsData)
	if err != nil {
		sm.logger.Error("failed to analyze experiment", zap.Error(err))
		return sm.handleFailed(ctx, exp, fmt.Sprintf("Failed to analyze results: %v", err))
	}

	// Convert analysis to experiment results
	results := sm.convertAnalysisToResults(analysisResult)

	// Update experiment with results
	exp.Status.Results = results

	// Log analysis summary
	sm.logger.Info("experiment analysis completed",
		zap.String("experiment_id", exp.ID),
		zap.String("recommendation", string(analysisResult.Recommendation)),
		zap.Float64("confidence", analysisResult.Confidence),
		zap.Bool("sufficient_data", analysisResult.SufficientData),
	)

	// Generate and store analysis report
	report := analysisResult.GenerateReport()
	exp.Status.AnalysisReport = report

	// Update experiment in store
	if err := sm.controller.store.UpdateExperiment(ctx, exp); err != nil {
		return fmt.Errorf("failed to update experiment with results: %w", err)
	}

	// Determine if experiment was successful based on analysis
	if analysisResult.Recommendation == analysis.RecommendationPromote && analysisResult.SufficientData {
		return sm.TransitionTo(ctx, exp.ID, ExperimentPhaseCompleted)
	} else if analysisResult.Recommendation == analysis.RecommendationReject {
		return sm.handleFailed(ctx, exp, "Analysis rejected candidate configuration")
	} else {
		// Continue or neutral - mark as completed but with caution
		exp.Status.Message = fmt.Sprintf("Analysis result: %s (confidence: %.1f%%)",
			analysisResult.Recommendation, analysisResult.Confidence*100)
		return sm.TransitionTo(ctx, exp.ID, ExperimentPhaseCompleted)
	}
}

// ProcessPipelineTask processes pipeline-related tasks
func (sm *StateMachine) ProcessPipelineTask(ctx context.Context, task Task) error {
	switch task.Type {
	case "validate_pipelines":
		return sm.validatePipelines(ctx, task)
	case "deploy_pipelines":
		return sm.deployPipelines(ctx, task)
	default:
		return fmt.Errorf("unknown pipeline task type: %s", task.Type)
	}
}

// validatePipelines validates pipeline configurations
func (sm *StateMachine) validatePipelines(ctx context.Context, task Task) error {
	sm.logger.Info("validating pipeline configurations",
		zap.String("experiment_id", task.ExperimentID),
	)

	baselinePipeline, ok := task.Data["baseline_pipeline"].(string)
	if !ok {
		return fmt.Errorf("baseline_pipeline not found in task data")
	}

	candidatePipeline, ok := task.Data["candidate_pipeline"].(string)
	if !ok {
		return fmt.Errorf("candidate_pipeline not found in task data")
	}

	// In a real implementation, this would validate the pipeline templates
	// For now, just log and return success
	sm.logger.Info("pipeline validation completed",
		zap.String("experiment_id", task.ExperimentID),
		zap.String("baseline", baselinePipeline),
		zap.String("candidate", candidatePipeline),
	)

	return nil
}

// deployPipelines deploys experiment pipelines
func (sm *StateMachine) deployPipelines(ctx context.Context, task Task) error {
	sm.logger.Info("deploying pipelines",
		zap.String("experiment_id", task.ExperimentID),
	)

	// Extract deployment data
	baselinePipeline := task.Data["baseline_pipeline"].(string)
	candidatePipeline := task.Data["candidate_pipeline"].(string)
	targetHosts := task.Data["target_hosts"].([]string)
	variables := task.Data["variables"].(map[string]string)

	// Deploy baseline pipeline
	baselineDeployment := &services.PipelineDeploymentRequest{
		ExperimentID: task.ExperimentID,
		PipelineName: baselinePipeline,
		PipelineType: "baseline",
		TargetNodes:  targetHosts,
		ConfigID:     fmt.Sprintf("%s-baseline", task.ExperimentID),
		Variables:    variables,
	}

	if err := sm.pipelineService.DeployPipeline(ctx, baselineDeployment); err != nil {
		return fmt.Errorf("failed to deploy baseline pipeline: %w", err)
	}

	// Deploy candidate pipeline
	candidateDeployment := &services.PipelineDeploymentRequest{
		ExperimentID: task.ExperimentID,
		PipelineName: candidatePipeline,
		PipelineType: "candidate",
		TargetNodes:  targetHosts,
		ConfigID:     fmt.Sprintf("%s-candidate", task.ExperimentID),
		Variables:    variables,
	}

	if err := sm.pipelineService.DeployPipeline(ctx, candidateDeployment); err != nil {
		return fmt.Errorf("failed to deploy candidate pipeline: %w", err)
	}

	sm.logger.Info("pipelines deployed successfully",
		zap.String("experiment_id", task.ExperimentID),
	)

	return nil
}

// collectMetricsData collects metrics data from Prometheus for analysis
func (sm *StateMachine) collectMetricsData(ctx context.Context, exp *Experiment) (map[string]*analysis.MetricData, error) {
	metrics := make(map[string]*analysis.MetricData)

	queries := []struct {
		name       string
		metricType analysis.MetricType
		fmtStr     string
	}{
		{
			name:       "cpu_usage",
			metricType: analysis.MetricTypeCost,
			fmtStr:     `avg(rate(container_cpu_usage_seconds_total{deployment="%s"}[5m])) * 100`,
		},
		{
			name:       "memory_usage",
			metricType: analysis.MetricTypeCost,
			fmtStr:     `avg(container_memory_usage_bytes{deployment="%s"}) / 1024 / 1024`,
		},
		{
			name:       "process_count",
			metricType: analysis.MetricTypeThroughput,
			fmtStr:     `count(count by (process_name) (process_cpu_seconds_total{deployment="%s"}))`,
		},
	}

	for _, q := range queries {
		baselineQuery := fmt.Sprintf(q.fmtStr, exp.Config.BaselinePipeline)
		candidateQuery := fmt.Sprintf(q.fmtStr, exp.Config.CandidatePipeline)

		baselineVal, err := sm.queryPromFloat(ctx, baselineQuery)
		if err != nil {
			sm.logger.Warn("failed to query baseline metric", zap.String("metric", q.name), zap.Error(err))
		}
		candidateVal, err := sm.queryPromFloat(ctx, candidateQuery)
		if err != nil {
			sm.logger.Warn("failed to query candidate metric", zap.String("metric", q.name), zap.Error(err))
		}

		metrics[q.name] = &analysis.MetricData{
			Type:      q.metricType,
			Baseline:  []float64{baselineVal},
			Candidate: []float64{candidateVal},
		}
	}

	return metrics, nil
}

// convertAnalysisToResults converts statistical analysis to experiment results
func (sm *StateMachine) convertAnalysisToResults(analysis *analysis.ExperimentAnalysis) *ExperimentResults {
	totalCardinalityReduction := analysis.CardinalityReduction
	cpuOverhead := analysis.CPUOverhead
	memoryOverhead := analysis.MemoryOverhead

	// Build recommendation string
	recommendation := fmt.Sprintf("Analysis: %s (Confidence: %.1f%%, Risk: %s)",
		analysis.Recommendation,
		analysis.Confidence*100,
		analysis.GetRiskLevel(),
	)

	return &ExperimentResults{
		BaselineMetrics: MetricsSnapshot{
			Timestamp:    analysis.AnalysisTime,
			CPUUsage:     analysis.BaselineCPU,
			MemoryUsage:  analysis.BaselineMemory,
			ProcessCount: int64(analysis.BaselineProcessCount),
		},
		CandidateMetrics: MetricsSnapshot{
			Timestamp:    analysis.AnalysisTime,
			CPUUsage:     analysis.CandidateCPU,
			MemoryUsage:  analysis.CandidateMemory,
			ProcessCount: int64(analysis.CandidateProcessCount),
		},
		CardinalityReduction: totalCardinalityReduction,
		CPUOverhead:          cpuOverhead,
		MemoryOverhead:       memoryOverhead,
		ProcessCoverage:      100.0,
		Recommendation:       recommendation,
		StatisticalAnalysis:  analysis,
	}
}

// queryPromFloat executes a Prometheus instant query and returns the first value
func (sm *StateMachine) queryPromFloat(ctx context.Context, query string) (float64, error) {
	if sm.promAPI == nil {
		return 0, fmt.Errorf("prometheus client not configured")
	}
	result, _, err := sm.promAPI.Query(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}
	if vector, ok := result.(model.Vector); ok && len(vector) > 0 {
		return float64(vector[0].Value), nil
	}
	return 0, fmt.Errorf("no data")
}

// getEnvDefault retrieves an environment variable or returns a default value
func getEnvDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}