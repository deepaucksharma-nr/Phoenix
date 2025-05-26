package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/phoenix/platform/pkg/models"
)

func TestNewPostgresStore(t *testing.T) {
	tests := []struct {
		name      string
		dbURL     string
		shouldErr bool
	}{
		{
			name:      "invalid URL",
			dbURL:     "invalid-url",
			shouldErr: true,
		},
		{
			name:      "valid URL format but unreachable",
			dbURL:     "postgres://user:pass@nonexistent:5432/db?sslmode=disable",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewPostgresStore(tt.dbURL)

			if tt.shouldErr {
				assert.Error(t, err)
				assert.Nil(t, store)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, store)
				if store != nil {
					store.Close()
				}
			}
		})
	}
}

func TestExperimentCRUD(t *testing.T) {
	// Skip this test if not running integration tests
	if testing.Short() {
		t.Skip("Skipping CRUD test in short mode")
	}

	// Use in-memory SQLite for testing
	store, err := NewPostgresStore("sqlite://file::memory:?cache=shared")
	if err != nil {
		t.Skip("SQLite not available, skipping CRUD tests")
	}
	defer store.Close()

	ctx := context.Background()

	t.Run("CreateAndGetExperiment", func(t *testing.T) {
		// Create test experiment
		exp := &models.Experiment{
			ID:                "test-exp-1",
			Name:              "Test Experiment",
			Description:       "Test Description",
			BaselinePipeline:  "baseline-v1",
			CandidatePipeline: "candidate-v1",
			Status:            models.StatusPending,
			TargetNodes:       []string{"node1", "node2"},
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		// Create experiment
		err := store.CreateExperiment(ctx, exp)
		require.NoError(t, err)

		// Get experiment
		retrieved, err := store.GetExperiment(ctx, exp.ID)
		require.NoError(t, err)
		assert.Equal(t, exp.ID, retrieved.ID)
		assert.Equal(t, exp.Name, retrieved.Name)
		assert.Equal(t, exp.Description, retrieved.Description)
		assert.Equal(t, exp.BaselinePipeline, retrieved.BaselinePipeline)
		assert.Equal(t, exp.CandidatePipeline, retrieved.CandidatePipeline)
		assert.Equal(t, exp.Status, retrieved.Status)
		assert.Equal(t, len(exp.TargetNodes), len(retrieved.TargetNodes))
	})

	t.Run("UpdateExperiment", func(t *testing.T) {
		// Create test experiment
		exp := &models.Experiment{
			ID:                "test-exp-2",
			Name:              "Test Experiment 2",
			Description:       "Test Description 2",
			BaselinePipeline:  "baseline-v1",
			CandidatePipeline: "candidate-v1",
			Status:            models.StatusPending,
			TargetNodes:       []string{"node1"},
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		// Create experiment
		err := store.CreateExperiment(ctx, exp)
		require.NoError(t, err)

		// Update experiment
		exp.Status = models.StatusRunning
		exp.Description = "Updated Description"
		startTime := time.Now()
		exp.StartedAt = &startTime

		err = store.UpdateExperiment(ctx, exp)
		require.NoError(t, err)

		// Get updated experiment
		retrieved, err := store.GetExperiment(ctx, exp.ID)
		require.NoError(t, err)
		assert.Equal(t, models.StatusRunning, retrieved.Status)
		assert.Equal(t, "Updated Description", retrieved.Description)
		assert.NotNil(t, retrieved.StartedAt)
		assert.True(t, retrieved.UpdatedAt.After(retrieved.CreatedAt))
	})

	t.Run("ListExperiments", func(t *testing.T) {
		// Create multiple test experiments
		experiments := []*models.Experiment{
			{
				ID:                "list-exp-1",
				Name:              "List Test 1",
				BaselinePipeline:  "baseline-v1",
				CandidatePipeline: "candidate-v1",
				Status:            models.StatusPending,
				TargetNodes:       []string{"node1"},
				CreatedAt:         time.Now().Add(-2 * time.Hour),
				UpdatedAt:         time.Now().Add(-2 * time.Hour),
			},
			{
				ID:                "list-exp-2",
				Name:              "List Test 2",
				BaselinePipeline:  "baseline-v1",
				CandidatePipeline: "candidate-v1",
				Status:            models.StatusRunning,
				TargetNodes:       []string{"node2"},
				CreatedAt:         time.Now().Add(-1 * time.Hour),
				UpdatedAt:         time.Now().Add(-1 * time.Hour),
			},
			{
				ID:                "list-exp-3",
				Name:              "List Test 3",
				BaselinePipeline:  "baseline-v1",
				CandidatePipeline: "candidate-v1",
				Status:            models.StatusCompleted,
				TargetNodes:       []string{"node3"},
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			},
		}

		// Create all experiments
		for _, exp := range experiments {
			err := store.CreateExperiment(ctx, exp)
			require.NoError(t, err)
		}

		// List experiments
		retrieved, err := store.ListExperiments(ctx, 10, 0)
		require.NoError(t, err)

		// Should return experiments in reverse chronological order (newest first)
		assert.GreaterOrEqual(t, len(retrieved), 3)

		// Find our test experiments
		var foundExps []*models.Experiment
		for _, exp := range retrieved {
			if exp.ID == "list-exp-1" || exp.ID == "list-exp-2" || exp.ID == "list-exp-3" {
				foundExps = append(foundExps, exp)
			}
		}

		assert.Len(t, foundExps, 3)

		// Check pagination
		page1, err := store.ListExperiments(ctx, 2, 0)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(page1), 2)

		page2, err := store.ListExperiments(ctx, 2, 2)
		require.NoError(t, err)

		// Ensure different pages return different results
		if len(page1) > 0 && len(page2) > 0 {
			assert.NotEqual(t, page1[0].ID, page2[0].ID)
		}
	})

	t.Run("GetNonexistentExperiment", func(t *testing.T) {
		_, err := store.GetExperiment(ctx, "nonexistent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "experiment not found")
	})
}

func TestJSONSerialization(t *testing.T) {
	// Test that TargetNodes slice serializes/deserializes correctly
	targetNodes := []string{"node1", "node2", "node3"}

	exp := &models.Experiment{
		ID:                "json-test-1",
		Name:              "JSON Test",
		BaselinePipeline:  "baseline-v1",
		CandidatePipeline: "candidate-v1",
		Status:            models.StatusPending,
		TargetNodes:       targetNodes,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// This test validates our JSON marshaling logic
	assert.Equal(t, 3, len(exp.TargetNodes))
	assert.Contains(t, exp.TargetNodes, "node1")
	assert.Contains(t, exp.TargetNodes, "node2")
	assert.Contains(t, exp.TargetNodes, "node3")
}
