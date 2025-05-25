package controller

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Scheduler manages the periodic processing of experiments
type Scheduler struct {
	logger       *zap.Logger
	controller   *ExperimentController
	stateMachine *StateMachine
	interval     time.Duration
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// NewScheduler creates a new experiment scheduler
func NewScheduler(logger *zap.Logger, controller *ExperimentController, stateMachine *StateMachine, interval time.Duration) *Scheduler {
	return &Scheduler{
		logger:       logger,
		controller:   controller,
		stateMachine: stateMachine,
		interval:     interval,
		stopCh:       make(chan struct{}),
	}
}

// Start begins the scheduler loop
func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Info("starting experiment scheduler", zap.Duration("interval", s.interval))

	s.wg.Add(1)
	go s.run(ctx)
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	s.logger.Info("stopping experiment scheduler")
	close(s.stopCh)
	s.wg.Wait()
}

// run is the main scheduler loop
func (s *Scheduler) run(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run immediately on start
	s.reconcileExperiments(ctx)

	for {
		select {
		case <-ticker.C:
			s.reconcileExperiments(ctx)
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// reconcileExperiments processes all experiments and updates their state
func (s *Scheduler) reconcileExperiments(ctx context.Context) {
	s.logger.Debug("reconciling experiments")

	// Get all non-terminal experiments
	experiments, err := s.controller.store.ListExperiments(ctx, ExperimentFilter{})
	if err != nil {
		s.logger.Error("failed to list experiments", zap.Error(err))
		return
	}

	for _, exp := range experiments {
		// Skip terminal states
		if s.isTerminalPhase(exp.Phase) {
			continue
		}

		// Process experiment based on current phase
		s.processExperiment(ctx, exp)
	}
}

// processExperiment handles state transitions for a single experiment
func (s *Scheduler) processExperiment(ctx context.Context, exp *Experiment) {
	s.logger.Debug("processing experiment",
		zap.String("id", exp.ID),
		zap.String("phase", string(exp.Phase)),
	)

	switch exp.Phase {
	case ExperimentPhasePending:
		// Transition to initializing
		if err := s.stateMachine.TransitionTo(ctx, exp.ID, ExperimentPhaseInitializing); err != nil {
			s.logger.Error("failed to transition to initializing",
				zap.String("experiment_id", exp.ID),
				zap.Error(err),
			)
		}

	case ExperimentPhaseInitializing:
		// Check if initialization is complete
		if s.isInitializationComplete(exp) {
			if err := s.stateMachine.TransitionTo(ctx, exp.ID, ExperimentPhaseRunning); err != nil {
				s.logger.Error("failed to transition to running",
					zap.String("experiment_id", exp.ID),
					zap.Error(err),
				)
			}
		} else if s.hasTimedOut(exp, 10*time.Minute) {
			s.logger.Warn("experiment initialization timed out", zap.String("id", exp.ID))
			if err := s.stateMachine.handleFailed(ctx, exp, "Initialization timed out"); err != nil {
				s.logger.Error("failed to handle experiment failure", zap.Error(err))
			}
		}

	case ExperimentPhaseRunning:
		// Check if experiment duration has elapsed
		if exp.Status.StartTime != nil {
			elapsed := time.Since(*exp.Status.StartTime)
			if elapsed >= exp.Config.Duration {
				if err := s.stateMachine.TransitionTo(ctx, exp.ID, ExperimentPhaseAnalyzing); err != nil {
					s.logger.Error("failed to transition to analyzing",
						zap.String("experiment_id", exp.ID),
						zap.Error(err),
					)
				}
			}
		}

	case ExperimentPhaseAnalyzing:
		// Check if analysis is complete
		if s.isAnalysisComplete(exp) {
			if s.stateMachine.meetsSuccessCriteria(exp, exp.Status.Results) {
				if err := s.stateMachine.TransitionTo(ctx, exp.ID, ExperimentPhaseCompleted); err != nil {
					s.logger.Error("failed to transition to completed",
						zap.String("experiment_id", exp.ID),
						zap.Error(err),
					)
				}
			} else {
				if err := s.stateMachine.handleFailed(ctx, exp, "Did not meet success criteria"); err != nil {
					s.logger.Error("failed to handle experiment failure", zap.Error(err))
				}
			}
		} else if s.hasTimedOut(exp, 30*time.Minute) {
			s.logger.Warn("experiment analysis timed out", zap.String("id", exp.ID))
			if err := s.stateMachine.handleFailed(ctx, exp, "Analysis timed out"); err != nil {
				s.logger.Error("failed to handle experiment failure", zap.Error(err))
			}
		}
	}
}

// isTerminalPhase checks if a phase is terminal
func (s *Scheduler) isTerminalPhase(phase ExperimentPhase) bool {
	return phase == ExperimentPhaseCompleted ||
		phase == ExperimentPhaseFailed ||
		phase == ExperimentPhaseCancelled
}

// isInitializationComplete checks if experiment initialization is complete
func (s *Scheduler) isInitializationComplete(exp *Experiment) bool {
	// TODO: Check actual initialization status
	// For now, check if we've been in this phase for at least 5 seconds
	for _, condition := range exp.Status.Conditions {
		if condition.Type == string(ExperimentPhaseInitializing) {
			return time.Since(condition.LastTransitionTime) > 5*time.Second
		}
	}
	return false
}

// isAnalysisComplete checks if experiment analysis is complete
func (s *Scheduler) isAnalysisComplete(exp *Experiment) bool {
	// Check if results are populated
	return exp.Status.Results != nil
}

// hasTimedOut checks if an experiment has been in its current phase too long
func (s *Scheduler) hasTimedOut(exp *Experiment, timeout time.Duration) bool {
	// Find the most recent phase transition
	var lastTransition time.Time
	for _, condition := range exp.Status.Conditions {
		if condition.Type == string(exp.Phase) && condition.LastTransitionTime.After(lastTransition) {
			lastTransition = condition.LastTransitionTime
		}
	}

	if lastTransition.IsZero() {
		// Use update time if no transition found
		lastTransition = exp.UpdatedAt
	}

	return time.Since(lastTransition) > timeout
}

// ReconcileExperiment manually triggers reconciliation for a specific experiment
func (s *Scheduler) ReconcileExperiment(ctx context.Context, experimentID string) error {
	exp, err := s.controller.GetExperiment(ctx, experimentID)
	if err != nil {
		return err
	}

	s.processExperiment(ctx, exp)
	return nil
}