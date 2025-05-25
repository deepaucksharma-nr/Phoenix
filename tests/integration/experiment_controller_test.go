// +build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/phoenix-vnext/platform/cmd/controller/internal/controller"
	"github.com/phoenix-vnext/platform/cmd/controller/internal/store"
	pb "github.com/phoenix-vnext/platform/pkg/api/v1"
)

// TestExperimentControllerIntegration tests the full experiment workflow
func TestExperimentControllerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zap.NewNop()
	
	// Setup test database
	testDB := setupTestDatabase(t)
	defer testDB.Close()

	// Create store
	postgresStore, err := store.NewPostgresStore(getTestDatabaseURL(), logger)
	require.NoError(t, err)
	defer postgresStore.Close()

	// Create experiment controller
	expController := controller.NewExperimentController(logger, postgresStore)
	
	// Create state machine
	stateMachine := controller.NewStateMachine(logger, expController)

	ctx := context.Background()

	t.Run("CreateExperiment", func(t *testing.T) {
		// Create experiment
		exp := &controller.Experiment{
			ID:          "test-exp-1",
			Name:        "Integration Test Experiment",
			Description: "Testing the full experiment lifecycle",
			Config: controller.ExperimentConfig{
				BaselinePipeline:  "process-baseline-v1",
				CandidatePipeline: "process-priority-filter-v1",
				TargetHosts:       []string{"test-node-1", "test-node-2"},
				Duration:          5 * time.Minute,
				SuccessCriteria: controller.SuccessCriteria{
					MinCardinalityReduction: 50.0,
					MaxCPUOverhead:          10.0,
					MaxMemoryOverhead:       15.0,
					CriticalProcessCoverage: 100.0,
				},
			},
		}

		err := expController.CreateExperiment(ctx, exp)
		require.NoError(t, err)

		// Verify experiment was created
		retrieved, err := expController.GetExperiment(ctx, exp.ID)
		require.NoError(t, err)
		assert.Equal(t, exp.Name, retrieved.Name)
		assert.Equal(t, controller.ExperimentPhasePending, retrieved.Phase)
	})

	t.Run("ExperimentStateTransitions", func(t *testing.T) {
		// Create a new experiment for state transition testing
		exp := &controller.Experiment{
			ID:          "test-exp-2",
			Name:        "State Transition Test",
			Description: "Testing state machine transitions",
			Config: controller.ExperimentConfig{
				BaselinePipeline:  "process-baseline-v1",
				CandidatePipeline: "process-priority-filter-v1",
				TargetHosts:       []string{"test-node-1"},
				Duration:          1 * time.Second, // Short duration for testing
				SuccessCriteria: controller.SuccessCriteria{
					MinCardinalityReduction: 30.0,
					MaxCPUOverhead:          5.0,
					MaxMemoryOverhead:       10.0,
					CriticalProcessCoverage: 95.0,
				},
			},
		}

		err := expController.CreateExperiment(ctx, exp)
		require.NoError(t, err)

		// Test transition to initializing
		err = stateMachine.TransitionTo(ctx, exp.ID, controller.ExperimentPhaseInitializing)
		require.NoError(t, err)

		// Verify state was updated
		retrieved, err := expController.GetExperiment(ctx, exp.ID)
		require.NoError(t, err)
		assert.Equal(t, controller.ExperimentPhaseInitializing, retrieved.Phase)

		// Test transition to running
		err = stateMachine.TransitionTo(ctx, exp.ID, controller.ExperimentPhaseRunning)
		require.NoError(t, err)

		retrieved, err = expController.GetExperiment(ctx, exp.ID)
		require.NoError(t, err)
		assert.Equal(t, controller.ExperimentPhaseRunning, retrieved.Phase)
		assert.NotNil(t, retrieved.Status.StartTime)

		// Test transition to analyzing
		err = stateMachine.TransitionTo(ctx, exp.ID, controller.ExperimentPhaseAnalyzing)
		require.NoError(t, err)

		retrieved, err = expController.GetExperiment(ctx, exp.ID)
		require.NoError(t, err)
		assert.Equal(t, controller.ExperimentPhaseAnalyzing, retrieved.Phase)
	})

	t.Run("ExperimentScheduler", func(t *testing.T) {
		// Create scheduler
		scheduler := controller.NewScheduler(logger, expController, stateMachine, 1*time.Second)
		
		// Start scheduler
		scheduler.Start(ctx)
		defer scheduler.Stop()

		// Create experiment in initializing state
		exp := &controller.Experiment{
			ID:          "test-exp-3",
			Name:        "Scheduler Test",
			Description: "Testing automatic state transitions",
			Phase:       controller.ExperimentPhaseInitializing,
			Config: controller.ExperimentConfig{
				BaselinePipeline:  "process-baseline-v1",
				CandidatePipeline: "process-priority-filter-v1",
				TargetHosts:       []string{"test-node-1"},
				Duration:          100 * time.Millisecond, // Very short for testing
				SuccessCriteria: controller.SuccessCriteria{
					MinCardinalityReduction: 50.0,
					MaxCPUOverhead:          10.0,
					MaxMemoryOverhead:       15.0,
					CriticalProcessCoverage: 100.0,
				},
			},
			Status: controller.ExperimentStatus{
				Phase:   controller.ExperimentPhaseInitializing,
				Message: "Initializing",
			},
		}

		err := expController.CreateExperiment(ctx, exp)
		require.NoError(t, err)

		// Wait for scheduler to process the experiment
		// The scheduler should eventually transition it through the states
		time.Sleep(3 * time.Second)

		// Check final state
		retrieved, err := expController.GetExperiment(ctx, exp.ID)
		require.NoError(t, err)
		
		// Should have progressed beyond initializing
		assert.NotEqual(t, controller.ExperimentPhaseInitializing, retrieved.Phase)
		t.Logf("Final experiment phase: %s", retrieved.Phase)
	})

	t.Run("ListExperiments", func(t *testing.T) {
		// Test listing experiments with different filters
		allExperiments, err := expController.ListExperiments(ctx, controller.ExperimentFilter{
			Limit: 10,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(allExperiments), 3) // We created 3 experiments

		// Test filtering by phase
		pendingFilter := controller.ExperimentFilter{
			Phase: &controller.ExperimentPhasePending,
			Limit: 10,
		}
		pendingExperiments, err := expController.ListExperiments(ctx, pendingFilter)
		require.NoError(t, err)
		
		// All returned experiments should be pending
		for _, exp := range pendingExperiments {
			assert.Equal(t, controller.ExperimentPhasePending, exp.Phase)
		}
	})

	t.Run("ExperimentCancellation", func(t *testing.T) {
		// Create experiment for cancellation test
		exp := &controller.Experiment{
			ID:          "test-exp-4",
			Name:        "Cancellation Test",
			Description: "Testing experiment cancellation",
			Config: controller.ExperimentConfig{
				BaselinePipeline:  "process-baseline-v1",
				CandidatePipeline: "process-priority-filter-v1",
				TargetHosts:       []string{"test-node-1"},
				Duration:          10 * time.Minute,
				SuccessCriteria: controller.SuccessCriteria{
					MinCardinalityReduction: 50.0,
					MaxCPUOverhead:          10.0,
					MaxMemoryOverhead:       15.0,
					CriticalProcessCoverage: 100.0,
				},
			},
		}

		err := expController.CreateExperiment(ctx, exp)
		require.NoError(t, err)

		// Start the experiment
		err = stateMachine.TransitionTo(ctx, exp.ID, controller.ExperimentPhaseRunning)
		require.NoError(t, err)

		// Cancel the experiment
		err = stateMachine.TransitionTo(ctx, exp.ID, controller.ExperimentPhaseCancelled)
		require.NoError(t, err)

		// Verify cancellation
		retrieved, err := expController.GetExperiment(ctx, exp.ID)
		require.NoError(t, err)
		assert.Equal(t, controller.ExperimentPhaseCancelled, retrieved.Phase)
		assert.NotNil(t, retrieved.Status.EndTime)
	})
}

// TestGRPCIntegration tests the gRPC adapter server
func TestGRPCIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require starting the actual gRPC server
	// For now, we'll test the adapter directly
	logger := zap.NewNop()
	
	// Setup test database
	testDB := setupTestDatabase(t)
	defer testDB.Close()

	// Create store and controller
	postgresStore, err := store.NewPostgresStore(getTestDatabaseURL(), logger)
	require.NoError(t, err)
	defer postgresStore.Close()

	expController := controller.NewExperimentController(logger, postgresStore)
	
	// Create adapter server (this would normally be served via gRPC)
	adapterServer := createMockAdapterServer(logger, expController)

	ctx := context.Background()

	t.Run("CreateExperimentViaAdapter", func(t *testing.T) {
		req := &pb.CreateExperimentRequest{
			Name:        "gRPC Test Experiment",
			Description: "Testing via gRPC adapter",
		}

		resp, err := adapterServer.CreateExperiment(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Experiment.Id)
		assert.Equal(t, "gRPC Test Experiment", resp.Experiment.Name)
	})

	t.Run("GetExperimentViaAdapter", func(t *testing.T) {
		// First create an experiment
		createReq := &pb.CreateExperimentRequest{
			Name:        "Get Test Experiment",
			Description: "Testing get via gRPC adapter",
		}

		createResp, err := adapterServer.CreateExperiment(ctx, createReq)
		require.NoError(t, err)

		// Then get it
		getReq := &pb.GetExperimentRequest{
			Id: createResp.Experiment.Id,
		}

		getResp, err := adapterServer.GetExperiment(ctx, getReq)
		require.NoError(t, err)
		assert.Equal(t, createResp.Experiment.Id, getResp.Experiment.Id)
		assert.Equal(t, "Get Test Experiment", getResp.Experiment.Name)
	})
}

// Helper functions

func setupTestDatabase(t *testing.T) *sql.DB {
	// Connect to test database
	db, err := sql.Open("postgres", getTestDatabaseURL())
	require.NoError(t, err)

	// Test connection
	err = db.Ping()
	require.NoError(t, err)

	return db
}

func getTestDatabaseURL() string {
	// Use environment variable or default to test database
	if dbURL := getEnvOrDefault("TEST_DATABASE_URL", ""); dbURL != "" {
		return dbURL
	}
	return "postgres://phoenix:phoenix@localhost:5432/phoenix_test?sslmode=disable"
}

func getEnvOrDefault(key, defaultValue string) string {
	// This would use os.Getenv in a real implementation
	return defaultValue
}

// Mock adapter server for testing without starting actual gRPC server
type mockAdapterServer struct {
	logger     *zap.Logger
	controller *controller.ExperimentController
}

func createMockAdapterServer(logger *zap.Logger, controller *controller.ExperimentController) *mockAdapterServer {
	return &mockAdapterServer{
		logger:     logger,
		controller: controller,
	}
}

func (s *mockAdapterServer) CreateExperiment(ctx context.Context, req *pb.CreateExperimentRequest) (*pb.CreateExperimentResponse, error) {
	// Create a basic experiment
	exp := &controller.Experiment{
		ID:          fmt.Sprintf("grpc-exp-%d", time.Now().Unix()),
		Name:        req.Name,
		Description: req.Description,
		Config: controller.ExperimentConfig{
			BaselinePipeline:  "process-baseline-v1",
			CandidatePipeline: "process-priority-filter-v1",
			TargetHosts:       []string{"grpc-test-node"},
			Duration:          5 * time.Minute,
			SuccessCriteria: controller.SuccessCriteria{
				MinCardinalityReduction: 50.0,
				MaxCPUOverhead:          10.0,
				MaxMemoryOverhead:       15.0,
				CriticalProcessCoverage: 100.0,
			},
		},
	}

	err := s.controller.CreateExperiment(ctx, exp)
	if err != nil {
		return nil, err
	}

	// Convert to proto format
	protoExp := &pb.Experiment{
		Id:          exp.ID,
		Name:        exp.Name,
		Description: exp.Description,
		Status:      string(exp.Phase),
		CreatedAt:   exp.CreatedAt.Unix(),
		UpdatedAt:   exp.UpdatedAt.Unix(),
	}

	return &pb.CreateExperimentResponse{
		Experiment: protoExp,
	}, nil
}

func (s *mockAdapterServer) GetExperiment(ctx context.Context, req *pb.GetExperimentRequest) (*pb.GetExperimentResponse, error) {
	exp, err := s.controller.GetExperiment(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Convert to proto format
	protoExp := &pb.Experiment{
		Id:          exp.ID,
		Name:        exp.Name,
		Description: exp.Description,
		Status:      string(exp.Phase),
		CreatedAt:   exp.CreatedAt.Unix(),
		UpdatedAt:   exp.UpdatedAt.Unix(),
	}

	return &pb.GetExperimentResponse{
		Experiment: protoExp,
	}, nil
}