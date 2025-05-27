package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/store"
	"github.com/rs/zerolog/log"
)

type Queue struct {
	store store.Store
}

func NewQueue(store store.Store) *Queue {
	return &Queue{
		store: store,
	}
}

// Enqueue adds a new task to the queue
func (q *Queue) Enqueue(ctx context.Context, task *models.Task) error {
	task.Status = "pending"
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	
	if err := q.store.CreateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	
	log.Info().
		Str("task_id", task.ID).
		Str("host_id", task.HostID).
		Str("type", task.Type).
		Str("action", task.Action).
		Msg("Task enqueued")
	
	return nil
}

// GetPendingTasks retrieves pending tasks for a specific host with long polling
func (q *Queue) GetPendingTasks(ctx context.Context, hostID string) ([]*models.Task, error) {
	// Try to get tasks immediately
	tasks, err := q.store.GetPendingTasksForHost(ctx, hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending tasks: %w", err)
	}
	
	// If we have tasks, return them immediately
	if len(tasks) > 0 {
		return tasks, nil
	}
	
	// Otherwise, implement long polling
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			// Context cancelled (timeout or client disconnect)
			return []*models.Task{}, nil
			
		case <-ticker.C:
			// Check for new tasks
			tasks, err := q.store.GetPendingTasksForHost(ctx, hostID)
			if err != nil {
				return nil, fmt.Errorf("failed to get pending tasks: %w", err)
			}
			
			if len(tasks) > 0 {
				return tasks, nil
			}
		}
	}
}

// GetTask retrieves a specific task by ID
func (q *Queue) GetTask(ctx context.Context, taskID string) (*models.Task, error) {
	task, err := q.store.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

// UpdateTaskStatus updates the status of a task
func (q *Queue) UpdateTaskStatus(ctx context.Context, taskID string, status string) error {
	task, err := q.store.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	
	task.Status = status
	task.UpdatedAt = time.Now()
	
	switch status {
	case "assigned":
		task.AssignedAt = &task.UpdatedAt
	case "running":
		task.StartedAt = &task.UpdatedAt
	case "completed", "failed":
		task.CompletedAt = &task.UpdatedAt
	}
	
	if err := q.store.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	
	log.Info().
		Str("task_id", taskID).
		Str("status", status).
		Msg("Task status updated")
	
	return nil
}

// UpdateTaskStatusWithResult updates task status with result data
func (q *Queue) UpdateTaskStatusWithResult(ctx context.Context, taskID string, status string, result map[string]interface{}, errorMessage string) error {
	task, err := q.store.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	
	task.Status = status
	task.Result = result
	task.ErrorMessage = errorMessage
	task.UpdatedAt = time.Now()
	
	switch status {
	case "running":
		task.StartedAt = &task.UpdatedAt
	case "completed", "failed":
		task.CompletedAt = &task.UpdatedAt
	}
	
	if err := q.store.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	
	// If task failed and has retries left, create a new task
	if status == "failed" && task.RetryCount < 3 {
		retryTask := &models.Task{
			HostID:       task.HostID,
			ExperimentID: task.ExperimentID,
			Type:         task.Type,
			Action:       task.Action,
			Config:       task.Config,
			Priority:     task.Priority - 1, // Slightly lower priority
			RetryCount:   task.RetryCount + 1,
		}
		
		if err := q.Enqueue(ctx, retryTask); err != nil {
			log.Error().Err(err).Str("task_id", taskID).Msg("Failed to enqueue retry task")
		} else {
			log.Info().
				Str("original_task_id", taskID).
				Int("retry_count", retryTask.RetryCount).
				Msg("Retry task enqueued")
		}
	}
	
	log.Info().
		Str("task_id", taskID).
		Str("status", status).
		Bool("has_error", errorMessage != "").
		Msg("Task completed with result")
	
	return nil
}

// GetTasksForExperiment retrieves all tasks for a specific experiment
func (q *Queue) GetTasksForExperiment(ctx context.Context, experimentID string) ([]*models.Task, error) {
	tasks, err := q.store.GetTasksByExperiment(ctx, experimentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks for experiment: %w", err)
	}
	return tasks, nil
}

// CancelTasksForExperiment cancels all pending tasks for an experiment
func (q *Queue) CancelTasksForExperiment(ctx context.Context, experimentID string) error {
	tasks, err := q.GetTasksForExperiment(ctx, experimentID)
	if err != nil {
		return err
	}
	
	for _, task := range tasks {
		if task.Status == "pending" || task.Status == "assigned" {
			if err := q.UpdateTaskStatus(ctx, task.ID, "cancelled"); err != nil {
				log.Error().Err(err).Str("task_id", task.ID).Msg("Failed to cancel task")
			}
		}
	}
	
	return nil
}

// GetTaskStats returns statistics about tasks in the queue
func (q *Queue) GetTaskStats(ctx context.Context) (map[string]interface{}, error) {
	stats, err := q.store.GetTaskStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task stats: %w", err)
	}
	return stats, nil
}

// Run starts the background worker for task queue maintenance
func (q *Queue) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	log.Info().Msg("Task queue background worker started")
	
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Task queue background worker stopping")
			return
			
		case <-ticker.C:
			// Process stale tasks
			if err := q.processStaleTask(ctx); err != nil {
				log.Error().Err(err).Msg("Failed to process stale tasks")
			}
			
			// Clean up old completed tasks
			if err := q.cleanupOldTasks(ctx); err != nil {
				log.Error().Err(err).Msg("Failed to cleanup old tasks")
			}
		}
	}
}

// processStaleTask marks assigned tasks that haven't been updated as failed
func (q *Queue) processStaleTask(ctx context.Context) error {
	// Get tasks that have been assigned for more than 5 minutes without update
	staleTasks, err := q.store.GetStaleTasks(ctx, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to get stale tasks: %w", err)
	}
	
	for _, task := range staleTasks {
		log.Warn().
			Str("task_id", task.ID).
			Str("host_id", task.HostID).
			Str("status", task.Status).
			Msg("Marking stale task as failed")
			
		if err := q.UpdateTaskStatusWithResult(ctx, task.ID, "failed", nil, "Task timed out"); err != nil {
			log.Error().Err(err).Str("task_id", task.ID).Msg("Failed to mark task as failed")
		}
	}
	
	return nil
}

// cleanupOldTasks removes completed tasks older than 24 hours
func (q *Queue) cleanupOldTasks(ctx context.Context) error {
	cutoff := time.Now().Add(-24 * time.Hour)
	err := q.store.DeleteOldTasks(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("failed to delete old tasks: %w", err)
	}
	
	return nil
}