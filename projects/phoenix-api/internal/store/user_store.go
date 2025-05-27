package store

import (
	"context"
	"errors"
	"github.com/phoenix/platform/pkg/database"
	"time"

	"github.com/google/uuid"
	internalModels "github.com/phoenix/platform/projects/phoenix-api/internal/models"
)

var (
	// ErrNotFound is returned when a requested resource is not found
	ErrNotFound = errors.New("not found")
)

// CreateUser creates a new user in the database
func (s *CompositeStore) CreateUser(ctx context.Context, user *internalModels.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, username, email, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := s.pipelineStore.db.DB().ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		true, // is_active defaults to true
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// GetUser retrieves a user by ID
func (s *CompositeStore) GetUser(ctx context.Context, userID string) (*internalModels.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, is_active, last_login, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &internalModels.User{}
	err := s.pipelineStore.db.DB().QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == database.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *CompositeStore) GetUserByUsername(ctx context.Context, username string) (*internalModels.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, is_active, last_login, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	user := &internalModels.User{}
	err := s.pipelineStore.db.DB().QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == database.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUserLastLogin updates the last login timestamp for a user
func (s *CompositeStore) UpdateUserLastLogin(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET last_login = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := s.pipelineStore.db.DB().ExecContext(ctx, query, time.Now(), time.Now(), userID)
	return err
}

// ListUsers retrieves a paginated list of users
func (s *CompositeStore) ListUsers(ctx context.Context, offset, limit int) ([]*internalModels.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, is_active, last_login, created_at, updated_at
		FROM users
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.pipelineStore.db.DB().QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*internalModels.User
	for rows.Next() {
		user := &internalModels.User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.IsActive,
			&user.LastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUser updates user details
func (s *CompositeStore) UpdateUser(ctx context.Context, user *internalModels.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET username = $1,
			email = $2,
			password_hash = $3,
			role = $4,
			is_active = $5,
			updated_at = $6
		WHERE id = $7
	`

	result, err := s.pipelineStore.db.DB().ExecContext(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.IsActive,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteUser soft deletes a user by setting is_active to false
func (s *CompositeStore) DeleteUser(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET is_active = false,
			updated_at = $1
		WHERE id = $2
	`

	result, err := s.pipelineStore.db.DB().ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
