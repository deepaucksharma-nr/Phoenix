package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	phoenixv1alpha1 "github.com/phoenix-vnext/platform/projects/phoenix-operator/api/v1alpha1"
	"github.com/phoenix-vnext/platform/tests/acceptance/framework"
)

// AT_P02_ExperimentCreationTestSuite tests experiment creation and management
type AT_P02_ExperimentCreationTestSuite struct {
	framework.AcceptanceTestSuite
}

func TestAT_P02_ExperimentCreation(t *testing.T) {
	suite.Run(t, new(AT_P02_ExperimentCreationTestSuite))
}

func (s *AT_P02_ExperimentCreationTestSuite) TestBasicExperimentCreation() {
	testID := "AT-P02"
	s.Logger.Info("Starting experiment creation test", zap.String("testID", testID))
	
	startTime := time.Now()
	results := make(map[string]interface{})
	
	// Setup: Create baseline and candidate pipelines
	baselineName := "baseline-pipeline"
	candidateName := "candidate-pipeline"
	
	s.Logger.Info("Creating baseline and candidate pipelines")
	
	baseline := s.CreatePipeline(baselineName, "process-baseline-v1", map[string]interface{}{
		"sampling_rate":        0.1,
		"aggregation_interval": "30s",
	})
	
	candidate := s.CreatePipeline(candidateName, "process-adaptive-v1", map[string]interface{}{
		"sampling_rate":        0.05,
		"aggregation_interval": "60s",
		"adaptive_threshold":   0.8,
	})
	
	// Wait for pipelines to be ready
	s.WaitForPipelineReady(baselineName)
	s.WaitForPipelineReady(candidateName)
	
	// Create experiment
	experimentName := "test-experiment-basic"
	targetNodes := map[string]string{
		"environment": "test",
		"region":      "us-west-1",
	}
	
	s.Logger.Info("Creating experiment", zap.String("name", experimentName))
	experimentCreateTime := time.Now()
	
	experiment := s.CreateExperiment(experimentName, baselineName, candidateName, targetNodes)
	
	// Wait for experiment to reach Running phase
	s.WaitForExperimentPhase(experimentName, "Running")
	
	experimentSetupTime := time.Since(experimentCreateTime)
	
	// Validate experiment setup time (should be under 5 minutes)
	s.Require().LessOrEqual(experimentSetupTime, 5*time.Minute,
		"Experiment setup exceeded 5 minute expected time")
	
	// Verify experiment configuration
	var exp phoenixv1alpha1.ExperimentCR
	err := s.K8sClient.Get(s.ctx, client.ObjectKey{
		Name:      experimentName,
		Namespace: s.Namespace,
	}, &exp)
	s.Require().NoError(err)
	
	s.Equal(baselineName, exp.Spec.BaselinePipeline)
	s.Equal(candidateName, exp.Spec.CandidatePipeline)
	s.Equal(targetNodes, exp.Spec.TargetNodes)
	
	// Record results
	results["experiment_name"] = experimentName
	results["setup_time_seconds"] = experimentSetupTime.Seconds()
	results["baseline_pipeline"] = baselineName
	results["candidate_pipeline"] = candidateName
	
	// Generate KPI results
	kpis := map[string]framework.KPIResult{
		"experiment_setup_time": {
			Name:        "Experiment Setup Time",
			Value:       experimentSetupTime.Seconds(),
			Target:      300, // 5 minutes in seconds
			Unit:        "seconds",
			Passed:      experimentSetupTime <= 5*time.Minute,
			Description: "Time to create and start an A/B experiment",
		},
	}
	
	// Cleanup
	s.Logger.Info("Cleaning up experiment and pipelines")
	s.CleanupExperiment(experimentName)
	s.CleanupPipeline(baselineName)
	s.CleanupPipeline(candidateName)
	
	// Generate report
	report := framework.TestReport{
		TestName:    "Basic Experiment Creation Test",
		TestID:      testID,
		Timestamp:   startTime,
		Duration:    time.Since(startTime),
		Passed:      !s.T().Failed(),
		Results:     results,
		Environment: s.getEnvironmentInfo(),
		KPIs:        kpis,
	}
	
	err = report.Save(s.T().Name())
	s.Require().NoError(err, "Failed to save test report")
}

func (s *AT_P02_ExperimentCreationTestSuite) TestExperimentWithLoadSimulation() {
	testID := "AT-P02-Load"
	s.Logger.Info("Starting experiment with load simulation test", zap.String("testID", testID))
	
	startTime := time.Now()
	
	// Create pipelines
	baselineName := "baseline-load-test"
	candidateName := "candidate-load-test"
	
	baseline := s.CreatePipeline(baselineName, "process-baseline-v1", map[string]interface{}{
		"sampling_rate": 0.1,
	})
	
	candidate := s.CreatePipeline(candidateName, "process-topk-v1", map[string]interface{}{
		"sampling_rate": 0.05,
		"top_k":         100,
	})
	
	s.WaitForPipelineReady(baselineName)
	s.WaitForPipelineReady(candidateName)
	
	// Create experiment
	experimentName := "test-experiment-load"
	experiment := s.CreateExperiment(experimentName, baselineName, candidateName, map[string]string{
		"workload": "high-cardinality",
	})
	
	s.WaitForExperimentPhase(experimentName, "Running")
	
	// Create load simulation
	loadSimName := "test-load-realistic"
	loadSim := s.CreateLoadSimulation(loadSimName, experiment.Name, "realistic", "10m")
	
	// Wait for load simulation to start
	s.WaitForLoadSimulationPhase(loadSimName, "Running")
	
	// Let experiment run for a few minutes
	s.Logger.Info("Letting experiment run with load simulation for 3 minutes")
	time.Sleep(3 * time.Minute)
	
	// Collect metrics during experiment
	metrics := s.CollectMetrics(1 * time.Minute)
	
	// Validate metrics show activity
	s.ValidateMetrics(metrics, map[string]func(float64) bool{
		"data_points_processed": func(v float64) bool { return v > 1000 },
		"cardinality_reduction": func(v float64) bool { return v > 0.5 },
		"error_rate":           func(v float64) bool { return v < 0.01 },
	})
	
	// Cleanup
	s.CleanupLoadSimulation(loadSimName)
	s.CleanupExperiment(experimentName)
	s.CleanupPipeline(baselineName)
	s.CleanupPipeline(candidateName)
	
	totalTime := time.Since(startTime)
	
	s.Logger.Info("Experiment with load simulation completed",
		zap.Duration("totalTime", totalTime),
		zap.Any("metrics", metrics))
}

func (s *AT_P02_ExperimentCreationTestSuite) TestConcurrentExperiments() {
	testID := "AT-P02-Concurrent"
	s.Logger.Info("Starting concurrent experiments test", zap.String("testID", testID))
	
	// Test creating multiple experiments concurrently
	experimentCount := 3
	experiments := make([]string, experimentCount)
	
	// Create pipelines for each experiment
	for i := 0; i < experimentCount; i++ {
		baselineName := fmt.Sprintf("baseline-%d", i)
		candidateName := fmt.Sprintf("candidate-%d", i)
		
		s.CreatePipeline(baselineName, "process-baseline-v1", map[string]interface{}{
			"sampling_rate": 0.1,
		})
		
		s.CreatePipeline(candidateName, "process-adaptive-v1", map[string]interface{}{
			"sampling_rate": 0.05,
		})
		
		s.WaitForPipelineReady(baselineName)
		s.WaitForPipelineReady(candidateName)
		
		// Create experiment
		experimentName := fmt.Sprintf("concurrent-exp-%d", i)
		experiments[i] = experimentName
		
		s.CreateExperiment(experimentName, baselineName, candidateName, map[string]string{
			"test_id": fmt.Sprintf("%d", i),
		})
	}
	
	// Wait for all experiments to be running
	for _, expName := range experiments {
		s.WaitForExperimentPhase(expName, "Running")
	}
	
	s.Logger.Info("All concurrent experiments are running")
	
	// Let them run briefly
	time.Sleep(1 * time.Minute)
	
	// Cleanup all experiments and pipelines
	for i, expName := range experiments {
		s.CleanupExperiment(expName)
		s.CleanupPipeline(fmt.Sprintf("baseline-%d", i))
		s.CleanupPipeline(fmt.Sprintf("candidate-%d", i))
	}
	
	s.Logger.Info("Concurrent experiments test completed",
		zap.Int("experimentCount", experimentCount))
}

// Helper methods

func (s *AT_P02_ExperimentCreationTestSuite) WaitForLoadSimulationPhase(name string, expectedPhase string) {
	s.Require().Eventually(func() bool {
		var loadSim phoenixv1alpha1.LoadSimulationJob
		err := s.K8sClient.Get(s.ctx, client.ObjectKey{
			Name:      name,
			Namespace: s.Namespace,
		}, &loadSim)
		
		if err != nil {
			return false
		}
		
		return string(loadSim.Status.Phase) == expectedPhase
	}, s.Timeout, s.PollInterval, "LoadSimulation did not reach phase %s", expectedPhase)
}

func (s *AT_P02_ExperimentCreationTestSuite) CleanupLoadSimulation(name string) {
	loadSim := &phoenixv1alpha1.LoadSimulationJob{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      name,
			Namespace: s.Namespace,
		},
	}
	
	err := s.K8sClient.Delete(s.ctx, loadSim)
	if err != nil && client.IgnoreNotFound(err) != nil {
		s.Logger.Error("Failed to delete load simulation", zap.Error(err))
	}
}

func (s *AT_P02_ExperimentCreationTestSuite) CleanupPipeline(name string) {
	pipeline := &phoenixv1alpha1.PhoenixProcessPipeline{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      name,
			Namespace: s.Namespace,
		},
	}
	
	err := s.K8sClient.Delete(s.ctx, pipeline)
	if err != nil && client.IgnoreNotFound(err) != nil {
		s.Logger.Error("Failed to delete pipeline", zap.Error(err))
	}
}

func (s *AT_P02_ExperimentCreationTestSuite) getEnvironmentInfo() map[string]string {
	return map[string]string{
		"namespace":          s.Namespace,
		"kubernetes_version": s.getKubernetesVersion(),
		"phoenix_version":    "v0.1.0",
	}
}

func (s *AT_P02_ExperimentCreationTestSuite) getKubernetesVersion() string {
	version, err := s.ClientSet.Discovery().ServerVersion()
	if err != nil {
		return "unknown"
	}
	return version.String()
}