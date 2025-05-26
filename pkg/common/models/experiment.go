package models

import (
	"time"
)

// Experiment represents a Phoenix experiment
type Experiment struct {
	ID                string            `json:"id" db:"id"`
	Name              string            `json:"name" db:"name"`
	Description       string            `json:"description" db:"description"`
	BaselinePipeline  string            `json:"baseline_pipeline" db:"baseline_pipeline"`
	CandidatePipeline string            `json:"candidate_pipeline" db:"candidate_pipeline"`
	Status            string            `json:"status" db:"status"`
	TargetNodes       map[string]string `json:"target_nodes" db:"target_nodes"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" db:"updated_at"`
	StartedAt         *time.Time        `json:"started_at" db:"started_at"`
	CompletedAt       *time.Time        `json:"completed_at" db:"completed_at"`
}

// ExperimentStatus represents possible experiment states
const (
	ExperimentStatusPending   = "pending"
	ExperimentStatusRunning   = "running"
	ExperimentStatusCompleted = "completed"
	ExperimentStatusFailed    = "failed"
	ExperimentStatusStopped   = "stopped"
)