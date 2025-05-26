package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	phoenixv1alpha1 "github.com/phoenix/platform/projects/loadsim-operator/api/v1alpha1"
)

var (
	testEnv *envtest.Environment
	k8sClient client.Client
	ctx context.Context
	cancel context.CancelFunc
)

func TestMain(m *testing.M) {
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// Setup test environment
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{"../../config/crd/bases"},
	}

	cfg, err := testEnv.Start()
	if err != nil {
		panic(fmt.Sprintf("Failed to start test environment: %v", err))
	}

	scheme := runtime.NewScheme()
	err = clientgoscheme.AddToScheme(scheme)
	if err != nil {
		panic(fmt.Sprintf("Failed to add client-go scheme: %v", err))
	}

	err = phoenixv1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(fmt.Sprintf("Failed to add Phoenix scheme: %v", err))
	}

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		panic(fmt.Sprintf("Failed to create k8s client: %v", err))
	}

	// Run tests
	code := m.Run()

	// Cleanup
	err = testEnv.Stop()
	if err != nil {
		panic(fmt.Sprintf("Failed to stop test environment: %v", err))
	}

	// Exit with the test result code
	panic(code)
}

func TestLoadSimulationJobCreation(t *testing.T) {
	// Create a test LoadSimulationJob
	loadSimJob := &phoenixv1alpha1.LoadSimulationJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-loadsim",
			Namespace: "default",
		},
		Spec: phoenixv1alpha1.LoadSimulationJobSpec{
			ExperimentID: "exp-12345678",
			Profile:      "realistic",
			Duration:     "30m",
			ProcessCount: 100,
		},
	}

	// Create the resource
	err := k8sClient.Create(ctx, loadSimJob)
	require.NoError(t, err, "Failed to create LoadSimulationJob")

	// Verify the resource was created
	created := &phoenixv1alpha1.LoadSimulationJob{}
	err = k8sClient.Get(ctx, client.ObjectKey{
		Name:      "test-loadsim",
		Namespace: "default",
	}, created)
	require.NoError(t, err, "Failed to get created LoadSimulationJob")

	// Assert the spec matches
	assert.Equal(t, "exp-12345678", created.Spec.ExperimentID)
	assert.Equal(t, "realistic", created.Spec.Profile)
	assert.Equal(t, "30m", created.Spec.Duration)
	assert.Equal(t, int32(100), created.Spec.ProcessCount)

	// Cleanup
	err = k8sClient.Delete(ctx, created)
	require.NoError(t, err, "Failed to delete LoadSimulationJob")
}

func TestLoadSimulationJobValidation(t *testing.T) {
	tests := []struct {
		name        string
		spec        phoenixv1alpha1.LoadSimulationJobSpec
		shouldError bool
	}{
		{
			name: "valid realistic profile",
			spec: phoenixv1alpha1.LoadSimulationJobSpec{
				ExperimentID: "exp-abcd1234",
				Profile:      "realistic",
				Duration:     "1h",
				ProcessCount: 200,
			},
			shouldError: false,
		},
		{
			name: "valid high-cardinality profile",
			spec: phoenixv1alpha1.LoadSimulationJobSpec{
				ExperimentID: "exp-efgh5678",
				Profile:      "high-cardinality",
				Duration:     "45m",
				ProcessCount: 500,
			},
			shouldError: false,
		},
		{
			name: "valid process-churn profile",
			spec: phoenixv1alpha1.LoadSimulationJobSpec{
				ExperimentID: "exp-ijkl9012",
				Profile:      "process-churn",
				Duration:     "15m",
				ProcessCount: 50,
			},
			shouldError: false,
		},
		{
			name: "invalid experiment ID format",
			spec: phoenixv1alpha1.LoadSimulationJobSpec{
				ExperimentID: "invalid-id",
				Profile:      "realistic",
				Duration:     "30m",
				ProcessCount: 100,
			},
			shouldError: true,
		},
		{
			name: "invalid profile",
			spec: phoenixv1alpha1.LoadSimulationJobSpec{
				ExperimentID: "exp-12345678",
				Profile:      "invalid-profile",
				Duration:     "30m",
				ProcessCount: 100,
			},
			shouldError: true,
		},
		{
			name: "invalid duration format",
			spec: phoenixv1alpha1.LoadSimulationJobSpec{
				ExperimentID: "exp-12345678",
				Profile:      "realistic",
				Duration:     "invalid-duration",
				ProcessCount: 100,
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loadSimJob := &phoenixv1alpha1.LoadSimulationJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("test-validation-%d", time.Now().UnixNano()),
					Namespace: "default",
				},
				Spec: tt.spec,
			}

			err := k8sClient.Create(ctx, loadSimJob)
			if tt.shouldError {
				assert.Error(t, err, "Expected validation error but got none")
			} else {
				assert.NoError(t, err, "Unexpected validation error")
				if err == nil {
					// Cleanup on success
					_ = k8sClient.Delete(ctx, loadSimJob)
				}
			}
		})
	}
}

func TestLoadSimulationJobWithCustomProfile(t *testing.T) {
	// Create a LoadSimulationJob with custom profile
	customProfile := &phoenixv1alpha1.CustomProfile{
		ChurnRate: 0.5,
		Patterns: []phoenixv1alpha1.ProcessPattern{
			{
				NameTemplate: "custom-proc-{{.Index}}",
				CPUPattern:   "steady",
				MemPattern:   "growing",
				Lifetime:     "60s",
				Count:        25,
			},
			{
				NameTemplate: "worker-{{.Index}}",
				CPUPattern:   "spiky",
				MemPattern:   "steady",
				Lifetime:     "120s",
				Count:        10,
			},
		},
	}

	loadSimJob := &phoenixv1alpha1.LoadSimulationJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-custom-profile",
			Namespace: "default",
		},
		Spec: phoenixv1alpha1.LoadSimulationJobSpec{
			ExperimentID:  "exp-custom01",
			Profile:       "custom",
			Duration:      "10m",
			ProcessCount:  35,
			CustomProfile: customProfile,
		},
	}

	// Create the resource
	err := k8sClient.Create(ctx, loadSimJob)
	require.NoError(t, err, "Failed to create LoadSimulationJob with custom profile")

	// Verify the custom profile was set correctly
	created := &phoenixv1alpha1.LoadSimulationJob{}
	err = k8sClient.Get(ctx, client.ObjectKey{
		Name:      "test-custom-profile",
		Namespace: "default",
	}, created)
	require.NoError(t, err, "Failed to get created LoadSimulationJob")

	assert.NotNil(t, created.Spec.CustomProfile)
	assert.Equal(t, 0.5, created.Spec.CustomProfile.ChurnRate)
	assert.Len(t, created.Spec.CustomProfile.Patterns, 2)

	// Verify first pattern
	pattern1 := created.Spec.CustomProfile.Patterns[0]
	assert.Equal(t, "custom-proc-{{.Index}}", pattern1.NameTemplate)
	assert.Equal(t, "steady", pattern1.CPUPattern)
	assert.Equal(t, "growing", pattern1.MemPattern)
	assert.Equal(t, "60s", pattern1.Lifetime)
	assert.Equal(t, int32(25), pattern1.Count)

	// Cleanup
	err = k8sClient.Delete(ctx, created)
	require.NoError(t, err, "Failed to delete LoadSimulationJob")
}

func TestLoadSimulationJobWithNodeSelector(t *testing.T) {
	// Create a LoadSimulationJob with node selector
	nodeSelector := map[string]string{
		"workload": "test",
		"zone":     "us-west-1a",
	}

	loadSimJob := &phoenixv1alpha1.LoadSimulationJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-node-selector",
			Namespace: "default",
		},
		Spec: phoenixv1alpha1.LoadSimulationJobSpec{
			ExperimentID: "exp-nodes123",
			Profile:      "realistic",
			Duration:     "20m",
			ProcessCount: 150,
			NodeSelector: nodeSelector,
		},
	}

	// Create the resource
	err := k8sClient.Create(ctx, loadSimJob)
	require.NoError(t, err, "Failed to create LoadSimulationJob with node selector")

	// Verify the node selector was set correctly
	created := &phoenixv1alpha1.LoadSimulationJob{}
	err = k8sClient.Get(ctx, client.ObjectKey{
		Name:      "test-node-selector",
		Namespace: "default",
	}, created)
	require.NoError(t, err, "Failed to get created LoadSimulationJob")

	assert.Equal(t, nodeSelector, created.Spec.NodeSelector)
	assert.Equal(t, "test", created.Spec.NodeSelector["workload"])
	assert.Equal(t, "us-west-1a", created.Spec.NodeSelector["zone"])

	// Cleanup
	err = k8sClient.Delete(ctx, created)
	require.NoError(t, err, "Failed to delete LoadSimulationJob")
}