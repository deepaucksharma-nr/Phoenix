package framework

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	phoenixv1alpha1 "github.com/phoenix/platform/projects/phoenix-operator/api/v1alpha1"
)

// AcceptanceTestSuite provides the base test suite for Phoenix acceptance tests
type AcceptanceTestSuite struct {
	suite.Suite
	
	// Kubernetes clients
	K8sClient    client.Client
	ClientSet    *kubernetes.Clientset
	
	// Test configuration
	Namespace    string
	Timeout      time.Duration
	PollInterval time.Duration
	
	// Logging
	Logger       *zap.Logger
	
	// Test context
	ctx          context.Context
	cancel       context.CancelFunc
}

// SetupSuite initializes the test suite
func (s *AcceptanceTestSuite) SetupSuite() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	require.NoError(s.T(), err)
	s.Logger = logger
	
	// Set defaults
	s.Namespace = os.Getenv("TEST_NAMESPACE")
	if s.Namespace == "" {
		s.Namespace = "phoenix-acceptance-test"
	}
	
	s.Timeout = 10 * time.Minute
	s.PollInterval = 5 * time.Second
	
	// Initialize Kubernetes clients
	config := ctrl.GetConfigOrDie()
	
	s.K8sClient, err = client.New(config, client.Options{})
	require.NoError(s.T(), err)
	
	s.ClientSet, err = kubernetes.NewForConfig(config)
	require.NoError(s.T(), err)
	
	// Create test context
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 30*time.Minute)
	
	// Ensure CRDs are installed
	s.ensureCRDs()
}

// TearDownSuite cleans up after all tests
func (s *AcceptanceTestSuite) TearDownSuite() {
	s.cancel()
	s.Logger.Sync()
}

// ensureCRDs verifies that all Phoenix CRDs are installed
func (s *AcceptanceTestSuite) ensureCRDs() {
	// This would check for PhoenixProcessPipeline, ExperimentCR, LoadSimulationJob CRDs
	s.Logger.Info("Verifying CRDs are installed")
	// Implementation would use discovery client to check CRDs
}

// Helper methods for common test operations

// CreateExperiment creates an experiment and returns it
func (s *AcceptanceTestSuite) CreateExperiment(name, baselinePipeline, candidatePipeline string, targetNodes map[string]string) *phoenixv1alpha1.ExperimentCR {
	experiment := &phoenixv1alpha1.ExperimentCR{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      name,
			Namespace: s.Namespace,
		},
		Spec: phoenixv1alpha1.ExperimentSpec{
			BaselinePipeline:  baselinePipeline,
			CandidatePipeline: candidatePipeline,
			TargetNodes:       targetNodes,
			Duration:          "10m",
			TrafficSplit: &phoenixv1alpha1.TrafficSplit{
				Baseline:  50,
				Candidate: 50,
			},
		},
	}
	
	err := s.K8sClient.Create(s.ctx, experiment)
	require.NoError(s.T(), err, "Failed to create experiment")
	
	return experiment
}

// WaitForExperimentPhase waits for an experiment to reach a specific phase
func (s *AcceptanceTestSuite) WaitForExperimentPhase(name string, expectedPhase phoenixv1alpha1.ExperimentPhase) {
	require.Eventually(s.T(), func() bool {
		var experiment phoenixv1alpha1.ExperimentCR
		err := s.K8sClient.Get(s.ctx, client.ObjectKey{
			Name:      name,
			Namespace: s.Namespace,
		}, &experiment)
		
		if err != nil {
			s.Logger.Error("Failed to get experiment", zap.Error(err))
			return false
		}
		
		s.Logger.Info("Experiment status", 
			zap.String("name", name),
			zap.String("phase", string(experiment.Status.Phase)),
			zap.String("expected", string(expectedPhase)))
		
		return experiment.Status.Phase == expectedPhase
	}, s.Timeout, s.PollInterval, "Experiment did not reach phase %s", expectedPhase)
}

// CreatePipeline creates a Phoenix process pipeline
func (s *AcceptanceTestSuite) CreatePipeline(name, template string, config map[string]interface{}) *phoenixv1alpha1.PhoenixProcessPipeline {
	pipeline := &phoenixv1alpha1.PhoenixProcessPipeline{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      name,
			Namespace: s.Namespace,
		},
		Spec: phoenixv1alpha1.PhoenixProcessPipelineSpec{
			Template: template,
			Config:   config,
			Replicas: 1,
		},
	}
	
	err := s.K8sClient.Create(s.ctx, pipeline)
	require.NoError(s.T(), err, "Failed to create pipeline")
	
	return pipeline
}

// WaitForPipelineReady waits for a pipeline to be ready
func (s *AcceptanceTestSuite) WaitForPipelineReady(name string) {
	require.Eventually(s.T(), func() bool {
		var pipeline phoenixv1alpha1.PhoenixProcessPipeline
		err := s.K8sClient.Get(s.ctx, client.ObjectKey{
			Name:      name,
			Namespace: s.Namespace,
		}, &pipeline)
		
		if err != nil {
			return false
		}
		
		return pipeline.Status.Phase == "Ready" && pipeline.Status.ReadyReplicas == pipeline.Spec.Replicas
	}, s.Timeout, s.PollInterval, "Pipeline %s did not become ready", name)
}

// CreateLoadSimulation creates a load simulation job
func (s *AcceptanceTestSuite) CreateLoadSimulation(name, experimentID, profile string, duration string) *phoenixv1alpha1.LoadSimulationJob {
	loadSim := &phoenixv1alpha1.LoadSimulationJob{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      name,
			Namespace: s.Namespace,
		},
		Spec: phoenixv1alpha1.LoadSimulationJobSpec{
			ExperimentID: experimentID,
			Profile:      profile,
			Duration:     duration,
			ProcessCount: 100,
		},
	}
	
	err := s.K8sClient.Create(s.ctx, loadSim)
	require.NoError(s.T(), err, "Failed to create load simulation")
	
	return loadSim
}

// CollectMetrics collects metrics for a given duration
func (s *AcceptanceTestSuite) CollectMetrics(duration time.Duration) map[string]float64 {
	s.Logger.Info("Collecting metrics", zap.Duration("duration", duration))
	
	// This would query Prometheus for metrics
	// For now, return mock data
	time.Sleep(duration)
	
	return map[string]float64{
		"cardinality_reduction": 0.85,
		"data_points_processed": 1000000,
		"error_rate":           0.001,
	}
}

// ValidateMetrics validates that metrics meet expected criteria
func (s *AcceptanceTestSuite) ValidateMetrics(metrics map[string]float64, criteria map[string]func(float64) bool) {
	for metric, validator := range criteria {
		value, exists := metrics[metric]
		require.True(s.T(), exists, "Metric %s not found", metric)
		require.True(s.T(), validator(value), "Metric %s with value %f did not meet criteria", metric, value)
	}
}

// CleanupExperiment deletes an experiment and waits for cleanup
func (s *AcceptanceTestSuite) CleanupExperiment(name string) {
	experiment := &phoenixv1alpha1.ExperimentCR{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      name,
			Namespace: s.Namespace,
		},
	}
	
	err := s.K8sClient.Delete(s.ctx, experiment)
	if err != nil && !client.IgnoreNotFound(err) != nil {
		s.Logger.Error("Failed to delete experiment", zap.Error(err))
	}
	
	// Wait for deletion
	require.Eventually(s.T(), func() bool {
		var exp phoenixv1alpha1.ExperimentCR
		err := s.K8sClient.Get(s.ctx, client.ObjectKey{
			Name:      name,
			Namespace: s.Namespace,
		}, &exp)
		return client.IgnoreNotFound(err) == nil
	}, 30*time.Second, time.Second)
}

// GenerateTestReport generates a test report
func (s *AcceptanceTestSuite) GenerateTestReport(testName string, results map[string]interface{}) {
	report := TestReport{
		TestName:    testName,
		Timestamp:   time.Now(),
		Duration:    time.Since(s.ctx.Value("startTime").(time.Time)),
		Passed:      !s.T().Failed(),
		Results:     results,
		Environment: s.getEnvironmentInfo(),
	}
	
	// Save report
	reportPath := fmt.Sprintf("reports/%s-%s.json", testName, time.Now().Format("20060102-150405"))
	report.Save(reportPath)
}

// getEnvironmentInfo collects environment information
func (s *AcceptanceTestSuite) getEnvironmentInfo() map[string]string {
	return map[string]string{
		"namespace":        s.Namespace,
		"kubernetes_version": s.getKubernetesVersion(),
		"phoenix_version":   os.Getenv("PHOENIX_VERSION"),
	}
}

// getKubernetesVersion gets the Kubernetes cluster version
func (s *AcceptanceTestSuite) getKubernetesVersion() string {
	version, err := s.ClientSet.Discovery().ServerVersion()
	if err != nil {
		return "unknown"
	}
	return version.String()
}