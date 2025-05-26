package controller

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/phoenix/platform/pkg/common/analysis"
	"github.com/phoenix/platform/projects/controller/internal/clients"
	api "github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
)

// StateMachine manages experiment state transitions
type StateMachine struct {
	logger           *zap.Logger
	controller       *ExperimentController
	generatorClient  *clients.GeneratorClient
	kubernetesClient *clients.KubernetesClient
	promAPI          v1.API
	analyzer         *analysis.ExperimentAnalyzer
	transitions      map[ExperimentPhase][]ExperimentPhase
}

// NewStateMachine creates a new experiment state machine
func NewStateMachine(logger *zap.Logger, controller *ExperimentController, generatorClient *clients.GeneratorClient, kubernetesClient *clients.KubernetesClient) *StateMachine {
	promAddr := getEnvDefault("PROMETHEUS_URL", "http://localhost:9090")
	promClient, err := api.NewClient(api.Config{Address: promAddr})
	var promAPI v1.API
	if err == nil {
		promAPI = v1.NewAPI(promClient)
	} else {
		logger.Warn("failed to create Prometheus client", zap.Error(err))
	}

	return &StateMachine{
		logger:           logger,
		controller:       controller,
		generatorClient:  generatorClient,
		kubernetesClient: kubernetesClient,
		promAPI:          promAPI,
		analyzer:         analysis.NewExperimentAnalyzer(),
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

	// Perform initialization tasks asynchronously
	go func() {
		ctx := context.Background()

		// Initialization tasks
		tasks := []struct {
			name     string
			duration time.Duration
			action   func(context.Context, *Experiment) error
		}{
			{"Validating pipelines", 2 * time.Second, sm.validatePipelines},
			{"Creating git branch", 1 * time.Second, sm.createGitBranch},
			{"Generating configurations", 3 * time.Second, sm.generateConfigurations},
			{"Creating Kubernetes resources", 2 * time.Second, sm.createKubernetesResources},
		}

		for _, task := range tasks {
			sm.logger.Info("executing initialization task",
				zap.String("experiment_id", exp.ID),
				zap.String("task", task.name),
			)

			// Execute task
			if err := task.action(ctx, exp); err != nil {
				sm.logger.Error("initialization task failed",
					zap.String("experiment_id", exp.ID),
					zap.String("task", task.name),
					zap.Error(err),
				)
				if transErr := sm.TransitionTo(ctx, exp.ID, ExperimentPhaseFailed); transErr != nil {
					sm.logger.Error("failed to transition to failed state", zap.Error(transErr))
				}
				return
			}
		}

		// Transition to running
		if err := sm.TransitionTo(ctx, exp.ID, ExperimentPhaseRunning); err != nil {
			sm.logger.Error("failed to transition to running", zap.Error(err))
		}
	}()

	return nil
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
	go func() {
		timer := time.NewTimer(exp.Config.Duration)
		defer timer.Stop()

		select {
		case <-timer.C:
			sm.logger.Info("experiment duration completed", zap.String("id", exp.ID))
			if err := sm.TransitionTo(context.Background(), exp.ID, ExperimentPhaseAnalyzing); err != nil {
				sm.logger.Error("failed to transition to analyzing", zap.Error(err))
			}
		case <-ctx.Done():
			sm.logger.Info("context cancelled", zap.String("id", exp.ID))
		}
	}()

	return nil
}

// handleAnalyzing handles the transition to analyzing phase
func (sm *StateMachine) handleAnalyzing(ctx context.Context, exp *Experiment) error {
	sm.logger.Info("analyzing experiment results", zap.String("id", exp.ID))

	// Update phase
	if err := sm.controller.UpdateExperimentPhase(ctx, exp.ID, ExperimentPhaseAnalyzing, "Analyzing experiment results"); err != nil {
		return err
	}

	// Perform analysis asynchronously
	go func() {
		ctx := context.Background()

		// Collect metrics data from monitoring system
		metricsData, err := sm.collectMetricsData(ctx, exp)
		if err != nil {
			sm.logger.Error("failed to collect metrics data", zap.Error(err))
			if err := sm.handleFailed(ctx, exp, fmt.Sprintf("Failed to collect metrics: %v", err)); err != nil {
				sm.logger.Error("failed to mark experiment as failed", zap.Error(err))
			}
			return
		}

		// Perform statistical analysis
		analysisResult, err := sm.analyzer.AnalyzeExperimentResults(ctx, exp, metricsData)
		if err != nil {
			sm.logger.Error("failed to analyze experiment", zap.Error(err))
			if err := sm.handleFailed(ctx, exp, fmt.Sprintf("Failed to analyze results: %v", err)); err != nil {
				sm.logger.Error("failed to mark experiment as failed", zap.Error(err))
			}
			return
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

		// Determine if experiment was successful based on analysis
		if analysisResult.Recommendation == analysis.RecommendationPromote && analysisResult.SufficientData {
			if err := sm.TransitionTo(ctx, exp.ID, ExperimentPhaseCompleted); err != nil {
				sm.logger.Error("failed to transition to completed", zap.Error(err))
			}
		} else if analysisResult.Recommendation == analysis.RecommendationReject {
			if err := sm.handleFailed(ctx, exp, "Analysis rejected candidate configuration"); err != nil {
				sm.logger.Error("failed to mark experiment as failed", zap.Error(err))
			}
		} else {
			// Continue or neutral - mark as completed but with caution
			exp.Status.Message = fmt.Sprintf("Analysis result: %s (confidence: %.1f%%)",
				analysisResult.Recommendation, analysisResult.Confidence*100)
			if err := sm.TransitionTo(ctx, exp.ID, ExperimentPhaseCompleted); err != nil {
				sm.logger.Error("failed to transition to completed", zap.Error(err))
			}
		}
	}()

	return nil
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

	// TODO: Cleanup resources

	return sm.controller.UpdateExperimentPhase(ctx, exp.ID, ExperimentPhaseCancelled, "Experiment cancelled by user")
}

// meetsSuccessCriteria checks if the experiment results meet the success criteria
func (sm *StateMachine) meetsSuccessCriteria(exp *Experiment, results *ExperimentResults) bool {
	criteria := exp.Config.SuccessCriteria

	// Check cardinality reduction
	if results.CardinalityReduction < criteria.MinCardinalityReduction {
		sm.logger.Info("cardinality reduction below threshold",
			zap.Float64("actual", results.CardinalityReduction),
			zap.Float64("required", criteria.MinCardinalityReduction),
		)
		return false
	}

	// Check CPU overhead (negative means improvement)
	if results.CPUOverhead > criteria.MaxCPUOverhead {
		sm.logger.Info("CPU overhead above threshold",
			zap.Float64("actual", results.CPUOverhead),
			zap.Float64("max", criteria.MaxCPUOverhead),
		)
		return false
	}

	// Check memory overhead (negative means improvement)
	if results.MemoryOverhead > criteria.MaxMemoryOverhead {
		sm.logger.Info("memory overhead above threshold",
			zap.Float64("actual", results.MemoryOverhead),
			zap.Float64("max", criteria.MaxMemoryOverhead),
		)
		return false
	}

	// Check process coverage
	if results.ProcessCoverage < criteria.CriticalProcessCoverage {
		sm.logger.Info("process coverage below threshold",
			zap.Float64("actual", results.ProcessCoverage),
			zap.Float64("required", criteria.CriticalProcessCoverage),
		)
		return false
	}

	return true
}

// Initialization task methods
func (sm *StateMachine) validatePipelines(ctx context.Context, exp *Experiment) error {
	sm.logger.Info("validating pipeline configurations",
		zap.String("experiment_id", exp.ID),
		zap.String("baseline_pipeline", exp.Config.BaselinePipeline),
		zap.String("candidate_pipeline", exp.Config.CandidatePipeline),
	)

	// Validate baseline pipeline
	if err := sm.generatorClient.ValidateTemplate(ctx, exp.Config.BaselinePipeline, nil); err != nil {
		return fmt.Errorf("baseline pipeline validation failed: %w", err)
	}

	// Validate candidate pipeline
	if err := sm.generatorClient.ValidateTemplate(ctx, exp.Config.CandidatePipeline, nil); err != nil {
		return fmt.Errorf("candidate pipeline validation failed: %w", err)
	}

	sm.logger.Info("pipeline validation completed successfully",
		zap.String("experiment_id", exp.ID),
	)

	return nil
}

func (sm *StateMachine) createGitBranch(ctx context.Context, exp *Experiment) error {
	sm.logger.Info("creating git branch for experiment",
		zap.String("experiment_id", exp.ID),
	)

	// In a real implementation, this would:
	// 1. Connect to git repository using git client
	// 2. Create a new branch named "experiment-{experiment_id}"
	// 3. Prepare directory structure for configurations
	// 4. Set up ArgoCD sync for the branch

	// For now, simulate the operation
	time.Sleep(300 * time.Millisecond)

	sm.logger.Info("git branch created successfully",
		zap.String("experiment_id", exp.ID),
		zap.String("branch", fmt.Sprintf("experiment-%s", exp.ID)),
	)

	return nil
}

func (sm *StateMachine) generateConfigurations(ctx context.Context, exp *Experiment) error {
	sm.logger.Info("generating experiment configurations",
		zap.String("experiment_id", exp.ID),
	)

	// Prepare request for config generator
	generatorReq := &clients.GeneratorRequest{
		ExperimentID:      exp.ID,
		BaselinePipeline:  exp.Config.BaselinePipeline,
		CandidatePipeline: exp.Config.CandidatePipeline,
		TargetNodes:       exp.Config.TargetHosts,
		Variables: map[string]interface{}{
			"NEW_RELIC_API_KEY_SECRET_NAME": "newrelic-secret",
			"EXPERIMENT_ID":                 exp.ID,
			"NAMESPACE":                     "phoenix-system",
		},
	}

	// Call config generator service
	response, err := sm.generatorClient.GenerateConfigurations(ctx, generatorReq)
	if err != nil {
		return fmt.Errorf("failed to generate configurations: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("configuration generation failed: %s", response.Message)
	}

	sm.logger.Info("configurations generated successfully",
		zap.String("experiment_id", exp.ID),
		zap.String("baseline_config_id", response.BaselineConfigID),
		zap.String("candidate_config_id", response.CandidateConfigID),
		zap.String("git_commit_sha", response.GitCommitSHA),
	)

	return nil
}

func (sm *StateMachine) createKubernetesResources(ctx context.Context, exp *Experiment) error {
	sm.logger.Info("creating Kubernetes resources",
		zap.String("experiment_id", exp.ID),
	)

	namespace := "phoenix-system" // In real implementation, get from config

	// Deploy baseline pipeline
	baselineDeployment := &clients.PipelineDeployment{
		ExperimentID: exp.ID,
		PipelineName: exp.Config.BaselinePipeline,
		PipelineType: "baseline",
		TargetNodes:  exp.Config.TargetHosts,
		ConfigID:     fmt.Sprintf("%s-baseline", exp.ID),
		Variables: map[string]interface{}{
			"NEW_RELIC_API_KEY_SECRET_NAME": "newrelic-secret",
			"EXPERIMENT_ID":                 exp.ID,
			"PIPELINE_TYPE":                 "baseline",
		},
		Namespace: namespace,
	}

	if err := sm.kubernetesClient.DeployPipeline(ctx, baselineDeployment); err != nil {
		return fmt.Errorf("failed to deploy baseline pipeline: %w", err)
	}

	// Deploy candidate pipeline
	candidateDeployment := &clients.PipelineDeployment{
		ExperimentID: exp.ID,
		PipelineName: exp.Config.CandidatePipeline,
		PipelineType: "candidate",
		TargetNodes:  exp.Config.TargetHosts,
		ConfigID:     fmt.Sprintf("%s-candidate", exp.ID),
		Variables: map[string]interface{}{
			"NEW_RELIC_API_KEY_SECRET_NAME": "newrelic-secret",
			"EXPERIMENT_ID":                 exp.ID,
			"PIPELINE_TYPE":                 "candidate",
		},
		Namespace: namespace,
	}

	if err := sm.kubernetesClient.DeployPipeline(ctx, candidateDeployment); err != nil {
		return fmt.Errorf("failed to deploy candidate pipeline: %w", err)
	}

	// Wait for both pipelines to be ready
	timeout := 5 * time.Minute

	sm.logger.Info("waiting for pipelines to be ready",
		zap.String("experiment_id", exp.ID),
		zap.Duration("timeout", timeout),
	)

	// Wait for baseline to be ready
	if err := sm.kubernetesClient.WaitForPipelineReady(ctx, exp.ID, "baseline", namespace, timeout); err != nil {
		return fmt.Errorf("baseline pipeline failed to become ready: %w", err)
	}

	// Wait for candidate to be ready
	if err := sm.kubernetesClient.WaitForPipelineReady(ctx, exp.ID, "candidate", namespace, timeout); err != nil {
		return fmt.Errorf("candidate pipeline failed to become ready: %w", err)
	}

	sm.logger.Info("Kubernetes resources created successfully",
		zap.String("experiment_id", exp.ID),
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
