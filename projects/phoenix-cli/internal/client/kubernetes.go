package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	loadSimJobGVR = schema.GroupVersionResource{
		Group:    "phoenix.io",
		Version:  "v1alpha1",
		Resource: "loadsimulationjobs",
	}
)

// LoadSimulationJobInterface provides operations for LoadSimulationJob resources
type LoadSimulationJobInterface interface {
	Create(ctx context.Context, job *LoadSimulationJob, opts metav1.CreateOptions) (*LoadSimulationJob, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*LoadSimulationJob, error)
	List(ctx context.Context, opts metav1.ListOptions) (*LoadSimulationJobList, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
}

// PhoenixV1alpha1Interface provides access to Phoenix v1alpha1 resources
type PhoenixV1alpha1Interface interface {
	LoadSimulationJobs(namespace string) LoadSimulationJobInterface
}

// Interface defines the Kubernetes client interface
type Interface interface {
	PhoenixV1alpha1() PhoenixV1alpha1Interface
	Kubernetes() kubernetes.Interface
	WaitForPipelineReady(ctx context.Context, experimentID, pipelineType, namespace string, timeout time.Duration) error
}

// kubernetesClient implements the Interface
type kubernetesClient struct {
	dynamicClient dynamic.Interface
	k8sClient     kubernetes.Interface
}

// loadSimJobClient implements LoadSimulationJobInterface
type loadSimJobClient struct {
	client    dynamic.NamespaceableResourceInterface
	namespace string
}

// phoenixV1alpha1Client implements PhoenixV1alpha1Interface
type phoenixV1alpha1Client struct {
	dynamicClient dynamic.Interface
}

// PhoenixV1alpha1 returns the Phoenix v1alpha1 client
func (c *kubernetesClient) PhoenixV1alpha1() PhoenixV1alpha1Interface {
	return &phoenixV1alpha1Client{dynamicClient: c.dynamicClient}
}

// Kubernetes returns the standard Kubernetes client
func (c *kubernetesClient) Kubernetes() kubernetes.Interface {
	return c.k8sClient
}

// LoadSimulationJobs returns a LoadSimulationJobInterface for the given namespace
func (c *phoenixV1alpha1Client) LoadSimulationJobs(namespace string) LoadSimulationJobInterface {
	return &loadSimJobClient{
		client:    c.dynamicClient.Resource(loadSimJobGVR),
		namespace: namespace,
	}
}

// Create creates a new LoadSimulationJob
func (c *loadSimJobClient) Create(ctx context.Context, job *LoadSimulationJob, opts metav1.CreateOptions) (*LoadSimulationJob, error) {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(job)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to unstructured: %w", err)
	}

	result, err := c.client.Namespace(c.namespace).Create(ctx, &unstructured.Unstructured{Object: unstructuredObj}, opts)
	if err != nil {
		return nil, err
	}

	var resultJob LoadSimulationJob
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(result.Object, &resultJob); err != nil {
		return nil, fmt.Errorf("failed to convert from unstructured: %w", err)
	}

	return &resultJob, nil
}

// Get retrieves a LoadSimulationJob by name
func (c *loadSimJobClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*LoadSimulationJob, error) {
	result, err := c.client.Namespace(c.namespace).Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}

	var job LoadSimulationJob
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(result.Object, &job); err != nil {
		return nil, fmt.Errorf("failed to convert from unstructured: %w", err)
	}

	return &job, nil
}

// List retrieves a list of LoadSimulationJobs
func (c *loadSimJobClient) List(ctx context.Context, opts metav1.ListOptions) (*LoadSimulationJobList, error) {
	result, err := c.client.Namespace(c.namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var jobList LoadSimulationJobList
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(result.Object, &jobList); err != nil {
		return nil, fmt.Errorf("failed to convert from unstructured: %w", err)
	}

	return &jobList, nil
}

// Delete deletes a LoadSimulationJob by name
func (c *loadSimJobClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Namespace(c.namespace).Delete(ctx, name, opts)
}

// GetKubernetesClient creates a new Kubernetes client
func GetKubernetesClient() (Interface, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	// Create dynamic client for CRDs
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Create standard Kubernetes client
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &kubernetesClient{
		dynamicClient: dynamicClient,
		k8sClient:     k8sClient,
	}, nil
}

// getKubeConfig gets the Kubernetes configuration
func getKubeConfig() (*rest.Config, error) {
	// Try in-cluster config first
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}

	// Fall back to kubeconfig file
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Check if KUBECONFIG env var is set
	if kubeconfigEnv := os.Getenv("KUBECONFIG"); kubeconfigEnv != "" {
		kubeconfig = kubeconfigEnv
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}

	return config, nil
}

// WaitForPipelineReady waits until a pipeline appears ready
func (c *kubernetesClient) WaitForPipelineReady(ctx context.Context, experimentID, pipelineType, namespace string, timeout time.Duration) error {
	// Placeholder implementation until CRDs are available
	select {
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for pipeline to be ready")
	case <-time.After(2 * time.Second):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
