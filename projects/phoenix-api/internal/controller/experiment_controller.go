package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/store"
	"github.com/phoenix/platform/projects/phoenix-api/internal/tasks"
	"github.com/rs/zerolog/log"
)

type ExperimentController struct {
	store     store.Store
	taskQueue *tasks.Queue
}

func NewExperimentController(store store.Store, taskQueue *tasks.Queue) *ExperimentController {
	return &ExperimentController{
		store:     store,
		taskQueue: taskQueue,
	}
}

// StartExperiment initiates an experiment by creating tasks for agents
func (c *ExperimentController) StartExperiment(ctx context.Context, exp *models.Experiment) error {
	log.Info().Str("experiment_id", exp.ID).Msg("Starting experiment")

	// Update experiment phase to deploying
	if err := c.store.UpdateExperimentPhase(ctx, exp.ID, "deploying"); err != nil {
		return fmt.Errorf("failed to update experiment phase: %w", err)
	}

	// Create tasks for each target host
	for _, host := range exp.Config.TargetHosts {
		// Baseline collector task
		baselineTask := &models.Task{
			HostID:       host,
			ExperimentID: exp.ID,
			Type:         "collector",
			Action:       "start",
			Priority:     1,
			Config: map[string]interface{}{
				"id":        fmt.Sprintf("%s-baseline", exp.ID),
				"variant":   "baseline",
				"configUrl": exp.Config.BaselineTemplate.URL,
				"vars":      exp.Config.BaselineTemplate.Variables,
			},
		}

		if err := c.taskQueue.Enqueue(ctx, baselineTask); err != nil {
			return fmt.Errorf("failed to enqueue baseline task for host %s: %w", host, err)
		}

		// Candidate collector task
		candidateTask := &models.Task{
			HostID:       host,
			ExperimentID: exp.ID,
			Type:         "collector",
			Action:       "start",
			Priority:     1,
			Config: map[string]interface{}{
				"id":        fmt.Sprintf("%s-candidate", exp.ID),
				"variant":   "candidate",
				"configUrl": exp.Config.CandidateTemplate.URL,
				"vars":      exp.Config.CandidateTemplate.Variables,
			},
		}

		if err := c.taskQueue.Enqueue(ctx, candidateTask); err != nil {
			return fmt.Errorf("failed to enqueue candidate task for host %s: %w", host, err)
		}

		// Load simulation task if configured
		if exp.Config.LoadProfile != "" {
			loadTask := &models.Task{
				HostID:       host,
				ExperimentID: exp.ID,
				Type:         "loadsim",
				Action:       "start",
				Priority:     0, // Lower priority, run after collectors
				Config: map[string]interface{}{
					"profile":  exp.Config.LoadProfile,
					"duration": exp.Config.Duration.String(),
				},
			}

			if err := c.taskQueue.Enqueue(ctx, loadTask); err != nil {
				return fmt.Errorf("failed to enqueue load simulation task for host %s: %w", host, err)
			}
		}
	}

	// Create experiment event
	event := &models.ExperimentEvent{
		ExperimentID: exp.ID,
		EventType:    "experiment_started",
		Phase:        "deploying",
		Message:      fmt.Sprintf("Experiment started with %d hosts", len(exp.Config.TargetHosts)),
	}

	if err := c.store.CreateExperimentEvent(ctx, event); err != nil {
		log.Error().Err(err).Msg("Failed to create experiment event")
	}

	return nil
}

// StopExperiment stops all tasks related to an experiment
func (c *ExperimentController) StopExperiment(ctx context.Context, experimentID string) error {
	log.Info().Str("experiment_id", experimentID).Msg("Stopping experiment")

	// Get experiment
	exp, err := c.store.GetExperiment(ctx, experimentID)
	if err != nil {
		return fmt.Errorf("failed to get experiment: %w", err)
	}

	// Create stop tasks for each host
	for _, host := range exp.Config.TargetHosts {
		// Stop baseline collector
		stopBaselineTask := &models.Task{
			HostID:       host,
			ExperimentID: experimentID,
			Type:         "collector",
			Action:       "stop",
			Priority:     2, // High priority
			Config: map[string]interface{}{
				"id": fmt.Sprintf("%s-baseline", experimentID),
			},
		}

		if err := c.taskQueue.Enqueue(ctx, stopBaselineTask); err != nil {
			log.Error().Err(err).Str("host", host).Msg("Failed to enqueue stop baseline task")
		}

		// Stop candidate collector
		stopCandidateTask := &models.Task{
			HostID:       host,
			ExperimentID: experimentID,
			Type:         "collector",
			Action:       "stop",
			Priority:     2,
			Config: map[string]interface{}{
				"id": fmt.Sprintf("%s-candidate", experimentID),
			},
		}

		if err := c.taskQueue.Enqueue(ctx, stopCandidateTask); err != nil {
			log.Error().Err(err).Str("host", host).Msg("Failed to enqueue stop candidate task")
		}

		// Stop load simulation if running
		stopLoadTask := &models.Task{
			HostID:       host,
			ExperimentID: experimentID,
			Type:         "loadsim",
			Action:       "stop",
			Priority:     2,
			Config:       map[string]interface{}{},
		}

		if err := c.taskQueue.Enqueue(ctx, stopLoadTask); err != nil {
			log.Error().Err(err).Str("host", host).Msg("Failed to enqueue stop load task")
		}
	}

	// Update experiment phase to stopping
	if err := c.store.UpdateExperimentPhase(ctx, experimentID, "stopping"); err != nil {
		return fmt.Errorf("failed to update experiment phase: %w", err)
	}

	return nil
}

// PromoteExperiment promotes the candidate configuration to production
func (c *ExperimentController) PromoteExperiment(ctx context.Context, experimentID string) error {
	log.Info().Str("experiment_id", experimentID).Msg("Promoting experiment")

	// Get experiment
	exp, err := c.store.GetExperiment(ctx, experimentID)
	if err != nil {
		return fmt.Errorf("failed to get experiment: %w", err)
	}

	// Verify experiment is in completed phase
	if exp.Phase != models.PhaseCompleted {
		return fmt.Errorf("experiment must be in completed phase to promote")
	}

	// For MVP, we'll simply record the promotion in the experiment metadata
	// In the future, this could update an actual pipeline template in a registry
	
	// Update experiment metadata to track promotion
	if exp.Metadata == nil {
		exp.Metadata = make(map[string]interface{})
	}
	exp.Metadata["promoted"] = true
	exp.Metadata["promoted_at"] = time.Now()
	exp.Metadata["promoted_config"] = map[string]interface{}{
		"template_url": exp.Config.CandidateTemplate.URL,
		"variables":    exp.Config.CandidateTemplate.Variables,
	}
	
	// Update experiment status
	exp.Status.KPIs["promotion_status"] = 1.0 // 1.0 indicates promoted
	
	// Save the updated experiment
	if err := c.store.UpdateExperiment(ctx, exp); err != nil {
		return fmt.Errorf("failed to update experiment with promotion status: %w", err)
	}

	// Create an event for the promotion
	event := &models.ExperimentEvent{
		ExperimentID: experimentID,
		EventType:    "promoted",
		Phase:        "promoted",
		Message:      "Experiment promoted to production",
		Metadata: map[string]interface{}{
			"promoted_template": exp.Config.CandidateTemplate.URL,
		},
	}
	
	if err := c.store.CreateExperimentEvent(ctx, event); err != nil {
		log.Error().Err(err).Msg("Failed to create promotion event")
	}

	// Update experiment phase to 'promoted'
	if err := c.store.UpdateExperimentPhase(ctx, experimentID, "promoted"); err != nil {
		return fmt.Errorf("failed to update experiment phase: %w", err)
	}

	log.Info().
		Str("experiment_id", experimentID).
		Str("template", exp.Config.CandidateTemplate.URL).
		Msg("Experiment promoted successfully")

	return nil
}

// CheckExperimentStatus monitors active pipelines and updates experiment phase
func (c *ExperimentController) CheckExperimentStatus(ctx context.Context, experimentID string) error {
	// TODO: Implement GetActivePipelines in store interface
	// For now, we'll assume pipelines are running if the experiment is in running phase
	exp, err := c.store.GetExperiment(ctx, experimentID)
	if err != nil {
		return fmt.Errorf("failed to get experiment: %w", err)
	}

	// For now, we'll assume all pipelines are running if experiment is in appropriate phase
	// In a real implementation, we would check the actual pipeline status
	baselineRunning := 0
	candidateRunning := 0
	
	// TODO: When GetActivePipelines is implemented, count actual running pipelines
	// For now, assume they're running based on experiment phase
	if exp.Phase == "running" || exp.Phase == "deploying" {
		baselineRunning = len(exp.Config.TargetHosts)
		candidateRunning = len(exp.Config.TargetHosts)
	}

	expectedHosts := len(exp.Config.TargetHosts)

	// Update phase based on pipeline status
	switch exp.Phase {
	case "deploying":
		if baselineRunning == expectedHosts && candidateRunning == expectedHosts {
			// All pipelines are running
			if err := c.store.UpdateExperimentPhase(ctx, experimentID, "running"); err != nil {
				return fmt.Errorf("failed to update phase to running: %w", err)
			}

			// Start monitoring phase after configured warmup
			go func() {
				time.Sleep(exp.Config.WarmupDuration)
				if err := c.store.UpdateExperimentPhase(context.Background(), experimentID, "monitoring"); err != nil {
					log.Error().Err(err).Msg("Failed to update phase to monitoring")
				}
			}()
		}

	case "stopping":
		if baselineRunning == 0 && candidateRunning == 0 {
			// All pipelines have stopped
			if err := c.store.UpdateExperimentPhase(ctx, experimentID, "stopped"); err != nil {
				return fmt.Errorf("failed to update phase to stopped: %w", err)
			}
		}
	}

	return nil
}
