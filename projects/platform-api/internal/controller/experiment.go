package controller

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ExperimentPhase represents the current phase of an experiment
type ExperimentPhase string

const (
	// ExperimentPhasePending indicates the experiment is pending
	ExperimentPhasePending ExperimentPhase = "Pending"
	// ExperimentPhaseInitializing indicates the experiment is being initialized
	ExperimentPhaseInitializing ExperimentPhase = "Initializing"
	// ExperimentPhaseRunning indicates the experiment is running
	ExperimentPhaseRunning ExperimentPhase = "Running"
	// ExperimentPhaseAnalyzing indicates the experiment results are being analyzed
	ExperimentPhaseAnalyzing ExperimentPhase = "Analyzing"
	// ExperimentPhaseCompleted indicates the experiment has completed
	ExperimentPhaseCompleted ExperimentPhase = "Completed"
	// ExperimentPhaseFailed indicates the experiment has failed
	ExperimentPhaseFailed ExperimentPhase = "Failed"
	// ExperimentPhaseCancelled indicates the experiment was cancelled
	ExperimentPhaseCancelled ExperimentPhase = "Cancelled"
)

// Experiment represents an A/B testing experiment
type Experiment struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Phase       ExperimentPhase        `json:"phase"`
	Config      ExperimentConfig       `json:"config"`
	Status      ExperimentStatus       `json:"status"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ExperimentConfig contains the configuration for an experiment
type ExperimentConfig struct {
	BaselinePipeline  string            `json:"baseline_pipeline"`
	CandidatePipeline string            `json:"candidate_pipeline"`
	TargetHosts       []string          `json:"target_hosts"`
	Duration          time.Duration     `json:"duration"`
	SuccessCriteria   SuccessCriteria   `json:"success_criteria"`
	Variables         map[string]string `json:"variables"`
}

// SuccessCriteria defines what constitutes a successful experiment
type SuccessCriteria struct {
	MinCardinalityReduction float64 `json:"min_cardinality_reduction"`
	MaxCPUOverhead          float64 `json:"max_cpu_overhead"`
	MaxMemoryOverhead       float64 `json:"max_memory_overhead"`
	CriticalProcessCoverage float64 `json:"critical_process_coverage"`
}

// ExperimentStatus contains the current status of an experiment
type ExperimentStatus struct {
	Phase          ExperimentPhase        `json:"phase"`
	Message        string                 `json:"message"`
	StartTime      *time.Time             `json:"start_time,omitempty"`
	EndTime        *time.Time             `json:"end_time,omitempty"`
	Results        *ExperimentResults     `json:"results,omitempty"`
	Conditions     []ExperimentCondition  `json:"conditions"`
	AnalysisReport string                 `json:"analysis_report,omitempty"`
}

// ExperimentResults contains the results of a completed experiment
type ExperimentResults struct {
	BaselineMetrics   MetricsSnapshot `json:"baseline_metrics"`
	CandidateMetrics  MetricsSnapshot `json:"candidate_metrics"`
	CardinalityReduction float64      `json:"cardinality_reduction"`
	CPUOverhead          float64      `json:"cpu_overhead"`
	MemoryOverhead       float64      `json:"memory_overhead"`
	ProcessCoverage      float64      `json:"process_coverage"`
	Recommendation       string       `json:"recommendation"`
	StatisticalAnalysis  interface{} `json:"statistical_analysis,omitempty"`
}

// MetricsSnapshot represents metrics at a point in time
type MetricsSnapshot struct {
	Timestamp         time.Time `json:"timestamp"`
	TimeSeriesCount   int64     `json:"time_series_count"`
	SamplesPerSecond  float64   `json:"samples_per_second"`
	CPUUsage          float64   `json:"cpu_usage"`
	MemoryUsage       float64   `json:"memory_usage"`
	ProcessCount      int64     `json:"process_count"`
	CriticalProcesses []string  `json:"critical_processes"`
}

// ExperimentCondition represents a condition or event in the experiment lifecycle
type ExperimentCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastTransitionTime time.Time `json:"last_transition_time"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
}

// ExperimentController manages the lifecycle of experiments
type ExperimentController struct {
	logger    *zap.Logger
	store     ExperimentStore
	taskQueue TaskQueue
}

// ExperimentStore defines the interface for experiment persistence
type ExperimentStore interface {
	CreateExperiment(ctx context.Context, exp *Experiment) error
	GetExperiment(ctx context.Context, id string) (*Experiment, error)
	UpdateExperiment(ctx context.Context, exp *Experiment) error
	ListExperiments(ctx context.Context, filter ExperimentFilter) ([]*Experiment, error)
}

// TaskQueue defines the interface for task queue operations
type TaskQueue interface {
	EnqueueTask(ctx context.Context, task Task) error
}

// Task represents a task in the queue
type Task struct {
	ID           string
	Type         string
	ExperimentID string
	Data         map[string]interface{}
}

// ExperimentFilter defines filters for listing experiments
type ExperimentFilter struct {
	Phase  *ExperimentPhase
	Limit  int
	Offset int
}

// NewExperimentController creates a new experiment controller
func NewExperimentController(logger *zap.Logger, store ExperimentStore, taskQueue TaskQueue) *ExperimentController {
	return &ExperimentController{
		logger:    logger,
		store:     store,
		taskQueue: taskQueue,
	}
}

// CreateExperiment creates a new experiment
func (c *ExperimentController) CreateExperiment(ctx context.Context, exp *Experiment) error {
	c.logger.Info("creating experiment",
		zap.String("id", exp.ID),
		zap.String("name", exp.Name),
	)

	// Set initial state
	exp.Phase = ExperimentPhasePending
	exp.CreatedAt = time.Now()
	exp.UpdatedAt = time.Now()
	exp.Status.Phase = ExperimentPhasePending
	exp.Status.Message = "Experiment created"

	// Validate experiment configuration
	if err := c.validateExperiment(exp); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Store experiment
	if err := c.store.CreateExperiment(ctx, exp); err != nil {
		return fmt.Errorf("failed to store experiment: %w", err)
	}

	// Enqueue task to start experiment processing
	task := Task{
		ID:           fmt.Sprintf("experiment-%s-init", exp.ID),
		Type:         "experiment_initialize",
		ExperimentID: exp.ID,
		Data: map[string]interface{}{
			"phase": ExperimentPhaseInitializing,
		},
	}

	if err := c.taskQueue.EnqueueTask(ctx, task); err != nil {
		c.logger.Error("failed to enqueue initialization task", zap.Error(err))
		// Don't fail experiment creation if task enqueue fails
		// The experiment can be manually started later
	}

	return nil
}

// GetExperiment retrieves an experiment by ID
func (c *ExperimentController) GetExperiment(ctx context.Context, id string) (*Experiment, error) {
	return c.store.GetExperiment(ctx, id)
}

// UpdateExperimentPhase updates the phase of an experiment
func (c *ExperimentController) UpdateExperimentPhase(ctx context.Context, id string, phase ExperimentPhase, message string) error {
	exp, err := c.store.GetExperiment(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get experiment: %w", err)
	}

	exp.Phase = phase
	exp.Status.Phase = phase
	exp.Status.Message = message
	exp.UpdatedAt = time.Now()

	// Add condition
	condition := ExperimentCondition{
		Type:               string(phase),
		Status:             "True",
		LastTransitionTime: time.Now(),
		Message:            message,
	}
	exp.Status.Conditions = append(exp.Status.Conditions, condition)

	return c.store.UpdateExperiment(ctx, exp)
}

// ProcessExperimentTask processes a task for an experiment
func (c *ExperimentController) ProcessExperimentTask(ctx context.Context, task Task) error {
	c.logger.Info("processing experiment task",
		zap.String("task_id", task.ID),
		zap.String("task_type", task.Type),
		zap.String("experiment_id", task.ExperimentID),
	)

	switch task.Type {
	case "experiment_initialize":
		return c.handleInitializeTask(ctx, task)
	case "experiment_run":
		return c.handleRunTask(ctx, task)
	case "experiment_analyze":
		return c.handleAnalyzeTask(ctx, task)
	default:
		return fmt.Errorf("unknown task type: %s", task.Type)
	}
}

// handleInitializeTask handles experiment initialization tasks
func (c *ExperimentController) handleInitializeTask(ctx context.Context, task Task) error {
	// Update phase to initializing
	if err := c.UpdateExperimentPhase(ctx, task.ExperimentID, ExperimentPhaseInitializing, "Starting experiment initialization"); err != nil {
		return err
	}

	// Enqueue next task to run the experiment
	runTask := Task{
		ID:           fmt.Sprintf("experiment-%s-run", task.ExperimentID),
		Type:         "experiment_run",
		ExperimentID: task.ExperimentID,
		Data: map[string]interface{}{
			"phase": ExperimentPhaseRunning,
		},
	}

	return c.taskQueue.EnqueueTask(ctx, runTask)
}

// handleRunTask handles experiment run tasks
func (c *ExperimentController) handleRunTask(ctx context.Context, task Task) error {
	exp, err := c.GetExperiment(ctx, task.ExperimentID)
	if err != nil {
		return err
	}

	// Update phase to running
	now := time.Now()
	exp.Status.StartTime = &now
	if err := c.UpdateExperimentPhase(ctx, task.ExperimentID, ExperimentPhaseRunning, "Experiment is running"); err != nil {
		return err
	}

	// Schedule analyze task after duration
	analyzeTask := Task{
		ID:           fmt.Sprintf("experiment-%s-analyze", task.ExperimentID),
		Type:         "experiment_analyze",
		ExperimentID: task.ExperimentID,
		Data: map[string]interface{}{
			"phase":      ExperimentPhaseAnalyzing,
			"execute_at": time.Now().Add(exp.Config.Duration),
		},
	}

	return c.taskQueue.EnqueueTask(ctx, analyzeTask)
}

// handleAnalyzeTask handles experiment analysis tasks
func (c *ExperimentController) handleAnalyzeTask(ctx context.Context, task Task) error {
	// Update phase to analyzing
	if err := c.UpdateExperimentPhase(ctx, task.ExperimentID, ExperimentPhaseAnalyzing, "Analyzing experiment results"); err != nil {
		return err
	}

	// In a real implementation, this would:
	// 1. Collect metrics from monitoring system
	// 2. Perform statistical analysis
	// 3. Generate recommendations
	// 4. Update experiment with results

	// For now, just mark as completed
	return c.UpdateExperimentPhase(ctx, task.ExperimentID, ExperimentPhaseCompleted, "Experiment completed successfully")
}

// validateExperiment validates the experiment configuration
func (c *ExperimentController) validateExperiment(exp *Experiment) error {
	if exp.Name == "" {
		return fmt.Errorf("experiment name is required")
	}

	if exp.Config.BaselinePipeline == "" {
		return fmt.Errorf("baseline pipeline is required")
	}

	if exp.Config.CandidatePipeline == "" {
		return fmt.Errorf("candidate pipeline is required")
	}

	if len(exp.Config.TargetHosts) == 0 {
		return fmt.Errorf("at least one target host is required")
	}

	if exp.Config.Duration <= 0 {
		return fmt.Errorf("duration must be positive")
	}

	// Validate success criteria
	if exp.Config.SuccessCriteria.MinCardinalityReduction < 0 || exp.Config.SuccessCriteria.MinCardinalityReduction > 100 {
		return fmt.Errorf("min cardinality reduction must be between 0 and 100")
	}

	if exp.Config.SuccessCriteria.CriticalProcessCoverage < 0 || exp.Config.SuccessCriteria.CriticalProcessCoverage > 100 {
		return fmt.Errorf("critical process coverage must be between 0 and 100")
	}

	return nil
}

// ListExperiments retrieves experiments based on the provided filter
func (c *ExperimentController) ListExperiments(ctx context.Context, filter ExperimentFilter) ([]*Experiment, error) {
	return c.store.ListExperiments(ctx, filter)
}