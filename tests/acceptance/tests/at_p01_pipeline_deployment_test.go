package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/phoenix-vnext/platform/tests/acceptance/framework"
)

// AT_P01_PipelineDeploymentTestSuite tests pipeline deployment functionality
type AT_P01_PipelineDeploymentTestSuite struct {
	framework.AcceptanceTestSuite
}

func TestAT_P01_PipelineDeployment(t *testing.T) {
	suite.Run(t, new(AT_P01_PipelineDeploymentTestSuite))
}

func (s *AT_P01_PipelineDeploymentTestSuite) TestPipelineDeploymentWithinSLA() {
	testID := "AT-P01"
	s.Logger.Info("Starting acceptance test", zap.String("testID", testID))
	
	startTime := time.Now()
	results := make(map[string]interface{})
	
	// Test case: Deploy a Phoenix process pipeline and verify it's ready within 10 minutes
	pipelineName := "test-pipeline-deployment"
	
	// Create pipeline
	s.Logger.Info("Creating pipeline", zap.String("name", pipelineName))
	deployStartTime := time.Now()
	
	pipeline := s.CreatePipeline(pipelineName, "process-baseline-v1", map[string]interface{}{
		"sampling_rate":     0.1,
		"aggregation_interval": "30s",
		"retention_days":    7,
	})
	
	// Wait for pipeline to be ready
	s.WaitForPipelineReady(pipelineName)
	
	deploymentTime := time.Since(deployStartTime)
	s.Logger.Info("Pipeline deployed", 
		zap.String("name", pipelineName),
		zap.Duration("deploymentTime", deploymentTime))
	
	// Validate deployment time is within SLA (10 minutes)
	s.Require().LessOrEqual(deploymentTime, 10*time.Minute, 
		"Pipeline deployment exceeded 10 minute SLA")
	
	// Verify pipeline is processing metrics
	s.Logger.Info("Verifying pipeline is processing metrics")
	time.Sleep(30 * time.Second) // Allow time for metrics to flow
	
	// Check pipeline metrics
	metrics := s.CollectMetrics(1 * time.Minute)
	s.ValidateMetrics(metrics, map[string]func(float64) bool{
		"data_points_processed": func(v float64) bool { return v > 0 },
		"error_rate":           func(v float64) bool { return v < 0.01 },
	})
	
	// Record results
	results["deployment_time_seconds"] = deploymentTime.Seconds()
	results["pipeline_name"] = pipelineName
	results["metrics_collected"] = metrics
	
	// Generate KPI results
	kpis := map[string]framework.KPIResult{
		"deployment_time": {
			Name:        "Pipeline Deployment Time",
			Value:       deploymentTime.Seconds(),
			Target:      600, // 10 minutes in seconds
			Unit:        "seconds",
			Passed:      deploymentTime <= 10*time.Minute,
			Description: "Time to deploy and ready a Phoenix process pipeline",
		},
	}
	
	// Cleanup
	s.Logger.Info("Cleaning up pipeline")
	s.CleanupPipeline(pipelineName)
	
	// Generate report
	report := framework.TestReport{
		TestName:    "Pipeline Deployment Test",
		TestID:      testID,
		Timestamp:   startTime,
		Duration:    time.Since(startTime),
		Passed:      !s.T().Failed(),
		Results:     results,
		Environment: s.getEnvironmentInfo(),
		KPIs:        kpis,
	}
	
	err := report.Save(s.T().Name())
	s.Require().NoError(err, "Failed to save test report")
}

func (s *AT_P01_PipelineDeploymentTestSuite) TestMultiplePipelineDeployments() {
	testID := "AT-P01-Multiple"
	s.Logger.Info("Starting multiple pipeline deployment test", zap.String("testID", testID))
	
	// Test deploying multiple pipelines concurrently
	pipelineCount := 3
	pipelines := make([]string, pipelineCount)
	
	startTime := time.Now()
	
	// Create pipelines concurrently
	for i := 0; i < pipelineCount; i++ {
		pipelines[i] = fmt.Sprintf("test-pipeline-%d", i)
		go func(name string, template string) {
			s.CreatePipeline(name, template, map[string]interface{}{
				"sampling_rate": 0.1,
			})
		}(pipelines[i], "process-baseline-v1")
	}
	
	// Wait for all pipelines to be ready
	for _, pipelineName := range pipelines {
		s.WaitForPipelineReady(pipelineName)
	}
	
	totalDeploymentTime := time.Since(startTime)
	
	// All pipelines should be ready within 10 minutes
	s.Require().LessOrEqual(totalDeploymentTime, 10*time.Minute,
		"Multiple pipeline deployment exceeded 10 minute SLA")
	
	// Cleanup
	for _, pipelineName := range pipelines {
		s.CleanupPipeline(pipelineName)
	}
	
	s.Logger.Info("Multiple pipeline deployment completed",
		zap.Int("count", pipelineCount),
		zap.Duration("totalTime", totalDeploymentTime))
}

// Helper method to cleanup pipeline
func (s *AT_P01_PipelineDeploymentTestSuite) CleanupPipeline(name string) {
	pipeline := &phoenixv1alpha1.PhoenixProcessPipeline{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      name,
			Namespace: s.Namespace,
		},
	}
	
	err := s.K8sClient.Delete(s.ctx, pipeline)
	if err != nil && !client.IgnoreNotFound(err) != nil {
		s.Logger.Error("Failed to delete pipeline", zap.Error(err))
	}
}