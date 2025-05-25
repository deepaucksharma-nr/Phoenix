package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

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

	store := &PostgresStore{db: db}
	
	// Create tables if they don't exist
	if err := store.createTables(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return store, nil
}

// CreateExperiment creates a new experiment
func (s *PostgresStore) CreateExperiment(ctx context.Context, exp *models.Experiment) error {
	// Serialize target_nodes to JSON
	targetNodesJSON, err := json.Marshal(exp.TargetNodes)
	if err != nil {
		return fmt.Errorf("failed to marshal target_nodes: %w", err)
	}

	query := `
		INSERT INTO experiments (
			id, name, description, baseline_pipeline, candidate_pipeline, 
			status, target_nodes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = s.db.ExecContext(ctx, query,
		exp.ID, exp.Name, exp.Description, exp.BaselinePipeline, exp.CandidatePipeline,
		exp.Status, string(targetNodesJSON), exp.CreatedAt, exp.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create experiment: %w", err)
	}

	return nil
}

// GetExperiment retrieves an experiment by ID
func (s *PostgresStore) GetExperiment(ctx context.Context, id string) (*models.Experiment, error) {
	query := `
		SELECT id, name, description, baseline_pipeline, candidate_pipeline,
		       status, target_nodes, created_at, updated_at, started_at, completed_at
		FROM experiments WHERE id = $1
	`

	row := s.db.QueryRowContext(ctx, query, id)

	var exp models.Experiment
	var targetNodesJSON string
	var startedAt, completedAt sql.NullTime

	err := row.Scan(
		&exp.ID, &exp.Name, &exp.Description, &exp.BaselinePipeline, &exp.CandidatePipeline,
		&exp.Status, &targetNodesJSON, &exp.CreatedAt, &exp.UpdatedAt, &startedAt, &completedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("experiment not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get experiment: %w", err)
	}

	// Deserialize target_nodes from JSON
	if err := json.Unmarshal([]byte(targetNodesJSON), &exp.TargetNodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal target_nodes: %w", err)
	}

	// Handle nullable timestamps
	if startedAt.Valid {
		exp.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		exp.CompletedAt = &completedAt.Time
	}

	return &exp, nil
}

// ListExperiments lists experiments with pagination
func (s *PostgresStore) ListExperiments(ctx context.Context, limit, offset int) ([]*models.Experiment, error) {
	query := `
		SELECT id, name, description, baseline_pipeline, candidate_pipeline,
		       status, target_nodes, created_at, updated_at, started_at, completed_at
		FROM experiments 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list experiments: %w", err)
	}
	defer rows.Close()

	var experiments []*models.Experiment
	for rows.Next() {
		var exp models.Experiment
		var targetNodesJSON string
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&exp.ID, &exp.Name, &exp.Description, &exp.BaselinePipeline, &exp.CandidatePipeline,
			&exp.Status, &targetNodesJSON, &exp.CreatedAt, &exp.UpdatedAt, &startedAt, &completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan experiment row: %w", err)
		}

		// Deserialize target_nodes from JSON
		if err := json.Unmarshal([]byte(targetNodesJSON), &exp.TargetNodes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal target_nodes: %w", err)
		}

		// Handle nullable timestamps
		if startedAt.Valid {
			exp.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			exp.CompletedAt = &completedAt.Time
		}

		experiments = append(experiments, &exp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate experiment rows: %w", err)
	}

	return experiments, nil
}

// UpdateExperiment updates an experiment
func (s *PostgresStore) UpdateExperiment(ctx context.Context, exp *models.Experiment) error {
	// Serialize target_nodes to JSON
	targetNodesJSON, err := json.Marshal(exp.TargetNodes)
	if err != nil {
		return fmt.Errorf("failed to marshal target_nodes: %w", err)
	}

	query := `
		UPDATE experiments SET 
			name = $2, description = $3, baseline_pipeline = $4, candidate_pipeline = $5,
			status = $6, target_nodes = $7, updated_at = $8, started_at = $9, completed_at = $10
		WHERE id = $1
	`

	exp.UpdatedAt = time.Now()

	_, err = s.db.ExecContext(ctx, query,
		exp.ID, exp.Name, exp.Description, exp.BaselinePipeline, exp.CandidatePipeline,
		exp.Status, string(targetNodesJSON), exp.UpdatedAt, exp.StartedAt, exp.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update experiment: %w", err)
	}

	return nil
}

// createTables creates the database tables if they don't exist
func (s *PostgresStore) createTables(ctx context.Context) error {
	schema := `
		CREATE TABLE IF NOT EXISTS experiments (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			baseline_pipeline VARCHAR(255) NOT NULL,
			candidate_pipeline VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			target_nodes JSONB,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			started_at TIMESTAMP WITH TIME ZONE,
			completed_at TIMESTAMP WITH TIME ZONE
		);

		CREATE INDEX IF NOT EXISTS idx_experiments_status ON experiments(status);
		CREATE INDEX IF NOT EXISTS idx_experiments_created_at ON experiments(created_at DESC);
	`

	_, err := s.db.ExecContext(ctx, schema)
	return err
}

// Close closes the database connection
func (s *PostgresStore) Close() error {
	return s.db.Close()
}