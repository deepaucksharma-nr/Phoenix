package client

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	phoenixv1alpha1 "github.com/phoenix-vnext/platform/projects/loadsim-operator/api/v1alpha1"
	phoenixclientset "github.com/phoenix-vnext/platform/projects/loadsim-operator/pkg/generated/clientset/versioned"
)

// Interface defines the Kubernetes client interface
type Interface interface {
	PhoenixV1alpha1() phoenixclientset.PhoenixV1alpha1Interface
	Kubernetes() kubernetes.Interface
}

// kubernetesClient implements the Interface
type kubernetesClient struct {
	phoenixClient phoenixclientset.Interface
	k8sClient     kubernetes.Interface
}

// PhoenixV1alpha1 returns the Phoenix v1alpha1 client
func (c *kubernetesClient) PhoenixV1alpha1() phoenixclientset.PhoenixV1alpha1Interface {
	return c.phoenixClient.PhoenixV1alpha1()
}

// Kubernetes returns the standard Kubernetes client
func (c *kubernetesClient) Kubernetes() kubernetes.Interface {
	return c.k8sClient
}

// GetKubernetesClient creates a new Kubernetes client
func GetKubernetesClient() (Interface, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	// Create Phoenix client
	phoenixClient, err := phoenixclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Phoenix client: %w", err)
	}

	// Create standard Kubernetes client
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &kubernetesClient{
		phoenixClient: phoenixClient,
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