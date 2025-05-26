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
	
	// Update experiment phase
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
	
	// Update experiment phase
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
	if exp.Phase != "completed" {
		return fmt.Errorf("experiment must be in completed phase to promote")
	}
	
	// Update production pipeline template
	template := &models.PipelineTemplate{
		Name:        "production",
		Description: fmt.Sprintf("Promoted from experiment %s", experimentID),
		ConfigURL:   exp.Config.CandidateTemplate.URL,
		Variables:   exp.Config.CandidateTemplate.Variables,
		Metadata: map[string]interface{}{
			"promoted_from": experimentID,
			"promoted_at":   time.Now(),
		},
	}
	
	if err := c.store.UpsertPipelineTemplate(ctx, template); err != nil {
		return fmt.Errorf("failed to update production template: %w", err)
	}
	
	// Update experiment phase
	if err := c.store.UpdateExperimentPhase(ctx, experimentID, "promoted"); err != nil {
		return fmt.Errorf("failed to update experiment phase: %w", err)
	}
	
	// Create promotion event
	event := &models.ExperimentEvent{
		ExperimentID: experimentID,
		EventType:    "experiment_promoted",
		Phase:        "promoted",
		Message:      "Candidate configuration promoted to production",
	}
	
	if err := c.store.CreateExperimentEvent(ctx, event); err != nil {
		log.Error().Err(err).Msg("Failed to create promotion event")
	}
	
	return nil
}

// CheckExperimentStatus monitors active pipelines and updates experiment phase
func (c *ExperimentController) CheckExperimentStatus(ctx context.Context, experimentID string) error {
	// Get active pipelines for this experiment
	pipelines, err := c.store.GetActivePipelines(ctx, experimentID)
	if err != nil {
		return fmt.Errorf("failed to get active pipelines: %w", err)
	}
	
	// Count running pipelines by variant
	baselineRunning := 0
	candidateRunning := 0
	
	for _, pipeline := range pipelines {
		if pipeline.Status == "running" {
			if pipeline.Variant == "baseline" {
				baselineRunning++
			} else if pipeline.Variant == "candidate" {
				candidateRunning++
			}
		}
	}
	
	// Get experiment
	exp, err := c.store.GetExperiment(ctx, experimentID)
	if err != nil {
		return fmt.Errorf("failed to get experiment: %w", err)
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