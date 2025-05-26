package controller

import (
	"context"
	"fmt"
	"time"
	
	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/store"
	"github.com/rs/zerolog/log"
)

// ExperimentStateMachine manages experiment state transitions
type ExperimentStateMachine struct {
	store     store.Store
	controller *ExperimentController
}

// NewExperimentStateMachine creates a new state machine
func NewExperimentStateMachine(store store.Store, controller *ExperimentController) *ExperimentStateMachine {
	return &ExperimentStateMachine{
		store:     store,
		controller: controller,
	}
}

// ValidateTransition checks if a state transition is valid
func (sm *ExperimentStateMachine) ValidateTransition(currentPhase, newPhase string) error {
	validTransitions := map[string][]string{
		"created":     {"deploying"},
		"deploying":   {"initializing", "failed"},
		"initializing": {"running", "failed"},
		"running":     {"analyzing", "stopping", "failed"},
		"analyzing":   {"completed", "failed"},
		"stopping":    {"stopped", "failed"},
		"completed":   {}, // Terminal state
		"stopped":     {}, // Terminal state
		"failed":      {}, // Terminal state
	}
	
	allowed, exists := validTransitions[currentPhase]
	if !exists {
		return fmt.Errorf("unknown phase: %s", currentPhase)
	}
	
	for _, phase := range allowed {
		if phase == newPhase {
			return nil
		}
	}
	
	return fmt.Errorf("invalid transition from %s to %s", currentPhase, newPhase)
}

// TransitionExperiment moves an experiment to a new phase
func (sm *ExperimentStateMachine) TransitionExperiment(ctx context.Context, experimentID, newPhase string) error {
	// Get current experiment state
	exp, err := sm.store.GetExperiment(ctx, experimentID)
	if err != nil {
		return fmt.Errorf("failed to get experiment: %w", err)
	}
	
	// Validate transition
	if err := sm.ValidateTransition(exp.Phase, newPhase); err != nil {
		return err
	}
	
	// Update phase
	if err := sm.store.UpdateExperimentPhase(ctx, experimentID, newPhase); err != nil {
		return fmt.Errorf("failed to update phase: %w", err)
	}
	
	// Create event
	event := &models.ExperimentEvent{
		ExperimentID: experimentID,
		EventType:    "phase_changed",
		Phase:        newPhase,
		Message:      fmt.Sprintf("Experiment transitioned from %s to %s", exp.Phase, newPhase),
		Metadata: map[string]interface{}{
			"previous_phase": exp.Phase,
			"new_phase":      newPhase,
		},
	}
	
	if err := sm.store.CreateExperimentEvent(ctx, event); err != nil {
		log.Error().Err(err).Msg("Failed to create phase change event")
	}
	
	// Trigger phase-specific actions
	return sm.handlePhaseActions(ctx, exp, newPhase)
}

// handlePhaseActions executes actions based on the new phase
func (sm *ExperimentStateMachine) handlePhaseActions(ctx context.Context, exp *models.Experiment, newPhase string) error {
	switch newPhase {
	case "deploying":
		// Already handled by StartExperiment
		return nil
		
	case "initializing":
		// Check if all collectors are deployed
		go sm.monitorDeployment(exp.ID)
		return nil
		
	case "running":
		// Start metrics collection
		go sm.startMetricsCollection(exp.ID)
		
		// Schedule automatic stop if duration is set
		if exp.Config.Duration > 0 {
			go sm.scheduleStop(exp.ID, exp.Config.Duration)
		}
		return nil
		
	case "analyzing":
		// Trigger KPI calculation
		go sm.triggerAnalysis(exp.ID)
		return nil
		
	case "stopped", "failed":
		// Clean up resources
		go sm.cleanupExperiment(exp.ID)
		return nil
		
	default:
		return nil
	}
}

// monitorDeployment monitors pipeline deployment progress
func (sm *ExperimentStateMachine) monitorDeployment(experimentID string) {
	ctx := context.Background()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	timeout := time.After(10 * time.Minute)
	
	for {
		select {
		case <-ticker.C:
			// Check task status
			tasks, err := sm.store.GetTasksByExperiment(ctx, experimentID)
			if err != nil {
				log.Error().Err(err).Str("experiment_id", experimentID).Msg("Failed to get tasks")
				continue
			}
			
			// Count collector tasks
			var totalCollectors, runningCollectors int
			for _, task := range tasks {
				if task.Type == "collector" && task.Action == "start" {
					totalCollectors++
					if task.Status == "completed" {
						runningCollectors++
					} else if task.Status == "failed" {
						// Deployment failed
						sm.TransitionExperiment(ctx, experimentID, "failed")
						return
					}
				}
			}
			
			// Check if all collectors are running
			if totalCollectors > 0 && runningCollectors == totalCollectors {
				log.Info().
					Str("experiment_id", experimentID).
					Int("collectors", runningCollectors).
					Msg("All collectors deployed successfully")
				
				// Transition to running
				if err := sm.TransitionExperiment(ctx, experimentID, "initializing"); err != nil {
					log.Error().Err(err).Msg("Failed to transition to initializing")
				}
				
				// Wait a bit for collectors to stabilize
				time.Sleep(10 * time.Second)
				
				if err := sm.TransitionExperiment(ctx, experimentID, "running"); err != nil {
					log.Error().Err(err).Msg("Failed to transition to running")
				}
				return
			}
			
		case <-timeout:
			// Deployment timeout
			log.Error().
				Str("experiment_id", experimentID).
				Msg("Deployment timeout")
			sm.TransitionExperiment(ctx, experimentID, "failed")
			return
		}
	}
}

// startMetricsCollection begins collecting metrics for the experiment
func (sm *ExperimentStateMachine) startMetricsCollection(experimentID string) {
	log.Info().Str("experiment_id", experimentID).Msg("Starting metrics collection")
	
	// Create event
	ctx := context.Background()
	event := &models.ExperimentEvent{
		ExperimentID: experimentID,
		EventType:    "metrics_collection_started",
		Phase:        "running",
		Message:      "Started collecting metrics from pipelines",
	}
	
	if err := sm.store.CreateExperimentEvent(ctx, event); err != nil {
		log.Error().Err(err).Msg("Failed to create metrics collection event")
	}
}

// scheduleStop schedules automatic experiment stop
func (sm *ExperimentStateMachine) scheduleStop(experimentID string, duration time.Duration) {
	log.Info().
		Str("experiment_id", experimentID).
		Dur("duration", duration).
		Msg("Scheduling automatic stop")
	
	time.Sleep(duration)
	
	ctx := context.Background()
	
	// Check if still running
	exp, err := sm.store.GetExperiment(ctx, experimentID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get experiment for scheduled stop")
		return
	}
	
	if exp.Phase == "running" {
		log.Info().Str("experiment_id", experimentID).Msg("Automatically stopping experiment")
		
		// Stop the experiment
		if err := sm.controller.StopExperiment(ctx, experimentID); err != nil {
			log.Error().Err(err).Msg("Failed to stop experiment")
			return
		}
		
		// Transition to analyzing
		time.Sleep(5 * time.Second) // Wait for stop tasks to be processed
		if err := sm.TransitionExperiment(ctx, experimentID, "analyzing"); err != nil {
			log.Error().Err(err).Msg("Failed to transition to analyzing")
		}
	}
}

// triggerAnalysis starts the analysis phase
func (sm *ExperimentStateMachine) triggerAnalysis(experimentID string) {
	log.Info().Str("experiment_id", experimentID).Msg("Triggering experiment analysis")
	
	ctx := context.Background()
	
	// Wait for metrics to be flushed
	time.Sleep(30 * time.Second)
	
	// TODO: Call KPI calculator here
	// For now, just transition to completed
	time.Sleep(10 * time.Second)
	
	if err := sm.TransitionExperiment(ctx, experimentID, "completed"); err != nil {
		log.Error().Err(err).Msg("Failed to transition to completed")
	}
}

// cleanupExperiment cleans up experiment resources
func (sm *ExperimentStateMachine) cleanupExperiment(experimentID string) {
	log.Info().Str("experiment_id", experimentID).Msg("Cleaning up experiment resources")
	
	ctx := context.Background()
	
	// Stop any running tasks
	tasks, err := sm.store.GetTasksByExperiment(ctx, experimentID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tasks for cleanup")
		return
	}
	
	for _, task := range tasks {
		if task.Status == "running" || task.Status == "pending" {
			// Mark as cancelled
			task.Status = "cancelled"
			if err := sm.store.UpdateTask(ctx, task); err != nil {
				log.Error().Err(err).Str("task_id", task.ID).Msg("Failed to cancel task")
			}
		}
	}
	
	// Create cleanup event
	event := &models.ExperimentEvent{
		ExperimentID: experimentID,
		EventType:    "cleanup_completed",
		Phase:        "stopped",
		Message:      "Experiment resources cleaned up",
	}
	
	if err := sm.store.CreateExperimentEvent(ctx, event); err != nil {
		log.Error().Err(err).Msg("Failed to create cleanup event")
	}
}