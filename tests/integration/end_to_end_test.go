// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/phoenix-vnext/platform/cmd/controller/internal/controller"
	"github.com/phoenix-vnext/platform/cmd/controller/internal/store"
	"github.com/phoenix-vnext/platform/pkg/generator"
)

// TestEndToEndWorkflow tests the complete experiment workflow
// from creation through config generation to completion
func TestEndToEndWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end test in short mode")
	}

	logger := zap.NewNop()
	ctx := context.Background()

	// Setup experiment controller
	testDB := setupTestDatabase(t)
	defer testDB.Close()
	defer CleanupTestData(t)

	postgresStore, err := store.NewPostgresStore(getTestDatabaseURL(), logger)
	require.NoError(t, err)
	defer postgresStore.Close()

	expController := controller.NewExperimentController(logger, postgresStore)
	stateMachine := controller.NewStateMachine(logger, expController)

	// Setup config generator
	generatorService := generator.NewService(logger, "https://github.com/phoenix/configs", "")
	generatorServer := httptest.NewServer(createGenerateConfigHandler(logger, generatorService))
	defer generatorServer.Close()

	t.Run("CompleteExperimentLifecycle", func(t *testing.T) {
		// Step 1: Create experiment
		exp := &controller.Experiment{
			ID:          "e2e-test-1",
			Name:        "End-to-End Test Experiment",
			Description: "Testing complete workflow",
			Config: controller.ExperimentConfig{
				BaselinePipeline:  "process-baseline-v1",
				CandidatePipeline: "process-priority-filter-v1",
				TargetHosts:       []string{"e2e-node-1", "e2e-node-2"},
				Duration:          1 * time.Second, // Short for testing
				SuccessCriteria: controller.SuccessCriteria{
					MinCardinalityReduction: 50.0,
					MaxCPUOverhead:          10.0,
					MaxMemoryOverhead:       15.0,
					CriticalProcessCoverage: 100.0,
				},
				Variables: map[string]interface{}{
					"NEW_RELIC_API_KEY":       "test-key-123",
					"NEW_RELIC_OTLP_ENDPOINT": "https://otlp.nr-data.net:4317",
				},
			},
		}

		err := expController.CreateExperiment(ctx, exp)
		require.NoError(t, err)

		// Verify experiment was created
		created, err := expController.GetExperiment(ctx, exp.ID)
		require.NoError(t, err)
		assert.Equal(t, controller.ExperimentPhasePending, created.Phase)

		// Step 2: Transition to initializing and generate configs
		err = stateMachine.TransitionTo(ctx, exp.ID, controller.ExperimentPhaseInitializing)
		require.NoError(t, err)

		// Simulate config generation request
		genReq := &generator.GenerateRequest{
			ExperimentID:      exp.ID,
			BaselinePipeline:  exp.Config.BaselinePipeline,
			CandidatePipeline: exp.Config.CandidatePipeline,
			TargetHosts:       exp.Config.TargetHosts,
			Variables: map[string]string{
				"NEW_RELIC_API_KEY":       "test-key-123",
				"NEW_RELIC_OTLP_ENDPOINT": "https://otlp.nr-data.net:4317",
			},
			Duration: exp.Config.Duration,
		}

		// Call config generator via HTTP
		configResp := callConfigGenerator(t, generatorServer.URL, genReq)
		
		// Verify config generation
		assert.Equal(t, exp.ID, configResp.ExperimentID)
		assert.NotEmpty(t, configResp.BaselineConfig)
		assert.NotEmpty(t, configResp.CandidateConfig)
		assert.NotEmpty(t, configResp.KubernetesManifests)

		// Step 3: Transition to running
		err = stateMachine.TransitionTo(ctx, exp.ID, controller.ExperimentPhaseRunning)
		require.NoError(t, err)

		running, err := expController.GetExperiment(ctx, exp.ID)
		require.NoError(t, err)
		assert.Equal(t, controller.ExperimentPhaseRunning, running.Phase)
		assert.NotNil(t, running.Status.StartTime)

		// Step 4: Wait for experiment duration to complete
		time.Sleep(2 * time.Second) // Wait longer than experiment duration

		// Step 5: Transition to analyzing
		err = stateMachine.TransitionTo(ctx, exp.ID, controller.ExperimentPhaseAnalyzing)
		require.NoError(t, err)

		analyzing, err := expController.GetExperiment(ctx, exp.ID)
		require.NoError(t, err)
		assert.Equal(t, controller.ExperimentPhaseAnalyzing, analyzing.Phase)

		// Step 6: Complete analysis and transition to completed
		err = stateMachine.TransitionTo(ctx, exp.ID, controller.ExperimentPhaseCompleted)
		require.NoError(t, err)

		completed, err := expController.GetExperiment(ctx, exp.ID)
		require.NoError(t, err)
		assert.Equal(t, controller.ExperimentPhaseCompleted, completed.Phase)
		assert.NotNil(t, completed.Status.EndTime)

		t.Logf("Experiment %s completed successfully", exp.ID)
	})

	t.Run("ExperimentWithScheduler", func(t *testing.T) {
		// Create scheduler with short interval for testing
		scheduler := controller.NewScheduler(logger, expController, stateMachine, 500*time.Millisecond)
		scheduler.Start(ctx)
		defer scheduler.Stop()

		// Create experiment
		exp := &controller.Experiment{
			ID:          "e2e-scheduler-test",
			Name:        "Scheduler E2E Test",
			Description: "Testing with scheduler automation",
			Phase:       controller.ExperimentPhaseInitializing, // Start in initializing
			Config: controller.ExperimentConfig{
				BaselinePipeline:  "process-baseline-v1",
				CandidatePipeline: "process-priority-filter-v1",
				TargetHosts:       []string{"scheduler-node-1"},
				Duration:          100 * time.Millisecond, // Very short
				SuccessCriteria: controller.SuccessCriteria{
					MinCardinalityReduction: 30.0, // Lower threshold for success
					MaxCPUOverhead:          20.0,
					MaxMemoryOverhead:       25.0,
					CriticalProcessCoverage: 90.0,
				},
			},
			Status: controller.ExperimentStatus{
				Phase:   controller.ExperimentPhaseInitializing,
				Message: "Initializing via scheduler",
			},
		}

		err := expController.CreateExperiment(ctx, exp)
		require.NoError(t, err)

		// Wait for scheduler to process the experiment
		success := WaitForCondition(func() bool {
			current, err := expController.GetExperiment(ctx, exp.ID)
			if err != nil {
				return false
			}
			// Check if experiment has progressed beyond initializing
			return current.Phase != controller.ExperimentPhaseInitializing
		}, 10*time.Second, 200*time.Millisecond)

		assert.True(t, success, "Scheduler should have processed the experiment")

		// Check final state
		final, err := expController.GetExperiment(ctx, exp.ID)
		require.NoError(t, err)
		
		t.Logf("Final experiment state: %s", final.Phase)
		
		// Should have moved beyond initializing
		assert.NotEqual(t, controller.ExperimentPhaseInitializing, final.Phase)
		
		// Should be in a terminal state or advanced state
		terminalStates := []controller.ExperimentPhase{
			controller.ExperimentPhaseCompleted,
			controller.ExperimentPhaseFailed,
			controller.ExperimentPhaseCancelled,
			controller.ExperimentPhaseRunning,
			controller.ExperimentPhaseAnalyzing,
		}
		
		assert.Contains(t, terminalStates, final.Phase)
	})

	t.Run("ParallelExperiments", func(t *testing.T) {
		// Test running multiple experiments in parallel
		numExperiments := 3
		experimentIDs := make([]string, numExperiments)

		// Create multiple experiments
		for i := 0; i < numExperiments; i++ {
			expID := fmt.Sprintf("parallel-exp-%d", i)
			experimentIDs[i] = expID

			exp := &controller.Experiment{
				ID:          expID,
				Name:        fmt.Sprintf("Parallel Experiment %d", i),
				Description: "Testing parallel execution",
				Config: controller.ExperimentConfig{
					BaselinePipeline:  "process-baseline-v1",
					CandidatePipeline: "process-priority-filter-v1",
					TargetHosts:       []string{fmt.Sprintf("parallel-node-%d", i)},
					Duration:          100 * time.Millisecond,
					SuccessCriteria: controller.SuccessCriteria{
						MinCardinalityReduction: 40.0,
						MaxCPUOverhead:          15.0,
						MaxMemoryOverhead:       20.0,
						CriticalProcessCoverage: 95.0,
					},
				},
			}

			err := expController.CreateExperiment(ctx, exp)
			require.NoError(t, err)
		}

		// Generate configs for all experiments in parallel
		for _, expID := range experimentIDs {
			exp, err := expController.GetExperiment(ctx, expID)
			require.NoError(t, err)

			genReq := &generator.GenerateRequest{
				ExperimentID:      expID,
				BaselinePipeline:  exp.Config.BaselinePipeline,
				CandidatePipeline: exp.Config.CandidatePipeline,
				TargetHosts:       exp.Config.TargetHosts,
				Duration:          exp.Config.Duration,
			}

			// Call config generator
			configResp := callConfigGenerator(t, generatorServer.URL, genReq)
			assert.Equal(t, expID, configResp.ExperimentID)
		}

		// Verify all experiments were created successfully
		filter := controller.ExperimentFilter{Limit: 10}
		experiments, err := expController.ListExperiments(ctx, filter)
		require.NoError(t, err)

		// Count parallel experiments
		parallelCount := 0
		for _, exp := range experiments {
			if len(exp.ID) > 12 && exp.ID[:12] == "parallel-exp" {
				parallelCount++
			}
		}

		assert.Equal(t, numExperiments, parallelCount, "All parallel experiments should be created")
	})
}

// TestErrorRecovery tests error handling and recovery scenarios
func TestErrorRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error recovery test in short mode")
	}

	logger := zap.NewNop()
	ctx := context.Background()

	// Setup
	testDB := setupTestDatabase(t)
	defer testDB.Close()
	defer CleanupTestData(t)

	postgresStore, err := store.NewPostgresStore(getTestDatabaseURL(), logger)
	require.NoError(t, err)
	defer postgresStore.Close()

	expController := controller.NewExperimentController(logger, postgresStore)
	stateMachine := controller.NewStateMachine(logger, expController)

	t.Run("InvalidTransition", func(t *testing.T) {
		// Create experiment
		exp := &controller.Experiment{
			ID:          "error-test-1",
			Name:        "Error Recovery Test",
			Description: "Testing error handling",
			Config: controller.ExperimentConfig{
				BaselinePipeline:  "process-baseline-v1",
				CandidatePipeline: "process-priority-filter-v1",
				TargetHosts:       []string{"error-node-1"},
				Duration:          1 * time.Minute,
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

		// Try invalid transition (pending -> completed)
		err = stateMachine.TransitionTo(ctx, exp.ID, controller.ExperimentPhaseCompleted)
		assert.Error(t, err, "Should not allow invalid transition")

		// Verify experiment is still in pending state
		current, err := expController.GetExperiment(ctx, exp.ID)
		require.NoError(t, err)
		assert.Equal(t, controller.ExperimentPhasePending, current.Phase)
	})

	t.Run("ExperimentCancellation", func(t *testing.T) {
		// Create and start experiment
		exp := &controller.Experiment{
			ID:          "cancel-test-1",
			Name:        "Cancellation Test",
			Description: "Testing experiment cancellation",
			Config: controller.ExperimentConfig{
				BaselinePipeline:  "process-baseline-v1",
				CandidatePipeline: "process-priority-filter-v1",
				TargetHosts:       []string{"cancel-node-1"},
				Duration:          10 * time.Minute, // Long duration
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

		// Start experiment
		err = stateMachine.TransitionTo(ctx, exp.ID, controller.ExperimentPhaseRunning)
		require.NoError(t, err)

		// Cancel experiment
		err = stateMachine.TransitionTo(ctx, exp.ID, controller.ExperimentPhaseCancelled)
		require.NoError(t, err)

		// Verify cancellation
		cancelled, err := expController.GetExperiment(ctx, exp.ID)
		require.NoError(t, err)
		assert.Equal(t, controller.ExperimentPhaseCancelled, cancelled.Phase)
		assert.NotNil(t, cancelled.Status.EndTime)
	})
}

// Helper functions

func callConfigGenerator(t *testing.T, serverURL string, req *generator.GenerateRequest) *generator.GenerateResponse {
	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	resp, err := http.Post(serverURL+"/api/v1/generate", "application/json", bytes.NewReader(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var genResp generator.GenerateResponse
	err = json.NewDecoder(resp.Body).Decode(&genResp)
	require.NoError(t, err)

	return &genResp
}