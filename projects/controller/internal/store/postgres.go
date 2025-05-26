package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/phoenix-vnext/platform/packages/go-common/store"
	"github.com/phoenix-vnext/platform/projects/controller/internal/controller"
	"go.uber.org/zap"
)

// PostgresStore implements the ExperimentStore interface using PostgreSQL
type PostgresStore struct {
	store  *store.PostgresStore
	logger *zap.Logger
}

// NewPostgresStore creates a new PostgreSQL-backed experiment store
func NewPostgresStore(connectionString string, logger *zap.Logger) (*PostgresStore, error) {
	baseStore, err := store.NewPostgresStore(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres store: %w", err)
	}

	// Set connection pool settings
	db := baseStore.DB()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	s := &PostgresStore{
		store:  baseStore,
		logger: logger,
	}

	// Run migrations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.migrate(ctx); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return s, nil
}

// Close closes the database connection
func (s *PostgresStore) Close() error {
	return s.store.Close()
}

// CreateExperiment creates a new experiment in the database
func (s *PostgresStore) CreateExperiment(ctx context.Context, exp *controller.Experiment) error {
	query := `
		INSERT INTO experiments (
			id, name, description, phase, 
			config, status, metadata,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	configJSON, err := json.Marshal(exp.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	statusJSON, err := json.Marshal(exp.Status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	metadataJSON, err := json.Marshal(exp.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = s.store.DB().ExecContext(ctx, query,
		exp.ID,
		exp.Name,
		exp.Description,
		exp.Phase,
		configJSON,
		statusJSON,
		metadataJSON,
		exp.CreatedAt,
		exp.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("experiment with ID %s already exists", exp.ID)
		}
		return fmt.Errorf("failed to insert experiment: %w", err)
	}

	s.logger.Info("created experiment in database",
		zap.String("id", exp.ID),
		zap.String("name", exp.Name),
	)

	return nil
}

// GetExperiment retrieves an experiment by ID
func (s *PostgresStore) GetExperiment(ctx context.Context, id string) (*controller.Experiment, error) {
	query := `
		SELECT 
			id, name, description, phase,
			config, status, metadata,
			created_at, updated_at
		FROM experiments
		WHERE id = $1
	`

	var exp controller.Experiment
	var configJSON, statusJSON, metadataJSON []byte

	err := s.store.DB().QueryRowContext(ctx, query, id).Scan(
		&exp.ID,
		&exp.Name,
		&exp.Description,
		&exp.Phase,
		&configJSON,
		&statusJSON,
		&metadataJSON,
		&exp.CreatedAt,
		&exp.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("experiment not found: %s", id)
		}
		return nil, fmt.Errorf("failed to query experiment: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(configJSON, &exp.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := json.Unmarshal(statusJSON, &exp.Status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &exp.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &exp, nil
}

// UpdateExperiment updates an existing experiment
func (s *PostgresStore) UpdateExperiment(ctx context.Context, exp *controller.Experiment) error {
	query := `
		UPDATE experiments SET
			name = $2,
			description = $3,
			phase = $4,
			config = $5,
			status = $6,
			metadata = $7,
			updated_at = $8
		WHERE id = $1
	`

	configJSON, err := json.Marshal(exp.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	statusJSON, err := json.Marshal(exp.Status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	metadataJSON, err := json.Marshal(exp.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	result, err := s.store.DB().ExecContext(ctx, query,
		exp.ID,
		exp.Name,
		exp.Description,
		exp.Phase,
		configJSON,
		statusJSON,
		metadataJSON,
		exp.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update experiment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("experiment not found: %s", exp.ID)
	}

	s.logger.Info("updated experiment in database",
		zap.String("id", exp.ID),
		zap.String("phase", string(exp.Phase)),
	)

	return nil
}

// ListExperiments retrieves experiments based on the provided filter
func (s *PostgresStore) ListExperiments(ctx context.Context, filter controller.ExperimentFilter) ([]*controller.Experiment, error) {
	query := `
		SELECT 
			id, name, description, phase,
			config, status, metadata,
			created_at, updated_at
		FROM experiments
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 0

	// Add phase filter if specified
	if filter.Phase != nil {
		argCount++
		query += fmt.Sprintf(" AND phase = $%d", argCount)
		args = append(args, *filter.Phase)
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

	// Add pagination
	if filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := s.store.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query experiments: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Ignore close errors
		}
	}()

	experiments := []*controller.Experiment{}
	for rows.Next() {
		var exp controller.Experiment
		var configJSON, statusJSON, metadataJSON []byte

		err := rows.Scan(
			&exp.ID,
			&exp.Name,
			&exp.Description,
			&exp.Phase,
			&configJSON,
			&statusJSON,
			&metadataJSON,
			&exp.CreatedAt,
			&exp.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan experiment: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(configJSON, &exp.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		if err := json.Unmarshal(statusJSON, &exp.Status); err != nil {
			return nil, fmt.Errorf("failed to unmarshal status: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &exp.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		experiments = append(experiments, &exp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating experiments: %w", err)
	}

	return experiments, nil
}

// DB returns the underlying database connection from the base store
func (s *PostgresStore) DB() *sql.DB {
    return s.store.DB()
}

// Close method is already defined above at line 52

// migrate runs database migrations
func (s *PostgresStore) migrate(ctx context.Context) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS experiments (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			phase VARCHAR(50) NOT NULL,
			config JSONB NOT NULL,
			status JSONB NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_experiments_phase ON experiments(phase)`,
		`CREATE INDEX IF NOT EXISTS idx_experiments_created_at ON experiments(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_experiments_updated_at ON experiments(updated_at DESC)`,
	}

	for i, migration := range migrations {
		s.logger.Debug("running migration", zap.Int("index", i))
		if _, err := s.store.DB().ExecContext(ctx, migration); err != nil {
			return fmt.Errorf("failed to run migration %d: %w", i, err)
		}
	}

	s.logger.Info("database migrations completed successfully")
	return nil
}