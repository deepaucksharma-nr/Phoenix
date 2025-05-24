package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/phoenix/platform/pkg/models"
)

// Store interface defines data access methods
type Store interface {
	CreateExperiment(ctx context.Context, exp *models.Experiment) error
	GetExperiment(ctx context.Context, id string) (*models.Experiment, error)
	ListExperiments(ctx context.Context, limit, offset int) ([]*models.Experiment, error)
	UpdateExperiment(ctx context.Context, exp *models.Experiment) error
	Close() error
}

// ExperimentStore is an alias for Store (for compatibility)
type ExperimentStore = Store

// PostgresStore implements Store using PostgreSQL
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a new PostgreSQL store
func NewPostgresStore(dbURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStore{db: db}, nil
}

// CreateExperiment creates a new experiment
func (s *PostgresStore) CreateExperiment(ctx context.Context, exp *models.Experiment) error {
	// TODO: Implement
	return nil
}

// GetExperiment retrieves an experiment by ID
func (s *PostgresStore) GetExperiment(ctx context.Context, id string) (*models.Experiment, error) {
	// TODO: Implement
	return &models.Experiment{ID: id}, nil
}

// ListExperiments lists experiments with pagination
func (s *PostgresStore) ListExperiments(ctx context.Context, limit, offset int) ([]*models.Experiment, error) {
	// TODO: Implement
	return []*models.Experiment{}, nil
}

// UpdateExperiment updates an experiment
func (s *PostgresStore) UpdateExperiment(ctx context.Context, exp *models.Experiment) error {
	// TODO: Implement
	return nil
}

// Close closes the database connection
func (s *PostgresStore) Close() error {
	return s.db.Close()
}