package models

import (
	"time"
)

// Experiment represents an A/B test between different processing pipelines
type Experiment struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	BaselinePipeline   string     `json:"baseline_pipeline"`
	CandidatePipeline  string     `json:"candidate_pipeline"`
	Status             string     `json:"status"`
	TargetNodes        []string   `json:"target_nodes"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	StartedAt          *time.Time `json:"started_at,omitempty"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
}

// ExperimentStatus constants
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCanceled  = "canceled"
)