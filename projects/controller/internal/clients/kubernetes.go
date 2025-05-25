package clients

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// TODO: Move pipeline API types to packages/contracts/k8s/
	// pipelinev1alpha1 "github.com/phoenix/platform/packages/contracts/k8s/pipeline/v1alpha1"
)

// PipelineDeployment represents a pipeline deployment request
type PipelineDeployment struct {
	ExperimentID     string                 `json:"experiment_id"`
	PipelineName     string                 `json:"pipeline_name"`
	PipelineType     string                 `json:"pipeline_type"` // "baseline" or "candidate"
	TargetNodes      []string               `json:"target_nodes"`
	ConfigID         string                 `json:"config_id"`
	Variables        map[string]interface{} `json:"variables"`
	Namespace        string                 `json:"namespace"`
}

// PipelineStatus represents the status of a deployed pipeline
type PipelineStatus struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Phase     string `json:"phase"`
	Ready     bool   `json:"ready"`
	Message   string `json:"message"`
	PodCount  int32  `json:"pod_count"`
}

// KubernetesClient handles Kubernetes operations for experiments
type KubernetesClient struct {
	logger    *zap.Logger
	clientset *kubernetes.Clientset
	client    client.Client
	scheme    *runtime.Scheme
}

// NewKubernetesClient creates a new Kubernetes client
func NewKubernetesClient(logger *zap.Logger) (*KubernetesClient, error) {
	// Try in-cluster config first, then fall back to kubeconfig
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubernetes config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	// Create runtime scheme and add our CRDs
	scheme := runtime.NewScheme()
	if err := pipelinev1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add pipeline CRDs to scheme: %w", err)
	}

	// Create controller-runtime client
	runtimeClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create controller-runtime client: %w", err)
	}

	return &KubernetesClient{
		logger:    logger,
		clientset: clientset,
		client:    runtimeClient,
		scheme:    scheme,
	}, nil
}

// DeployPipeline deploys a pipeline using PhoenixProcessPipeline CRD
func (k *KubernetesClient) DeployPipeline(ctx context.Context, deployment *PipelineDeployment) error {
	k.logger.Info("deploying pipeline",
		zap.String("experiment_id", deployment.ExperimentID),
		zap.String("pipeline_name", deployment.PipelineName),
		zap.String("pipeline_type", deployment.PipelineType),
		zap.String("namespace", deployment.Namespace),
	)

	// Create PhoenixProcessPipeline resource
	pipelineResource := &pipelinev1alpha1.PhoenixProcessPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", deployment.ExperimentID, deployment.PipelineType),
			Namespace: deployment.Namespace,
			Labels: map[string]string{
				"phoenix.newrelic.com/experiment-id": deployment.ExperimentID,
				"phoenix.newrelic.com/pipeline-type": deployment.PipelineType,
				"phoenix.newrelic.com/managed-by":    "experiment-controller",
			},
			Annotations: map[string]string{
				"phoenix.newrelic.com/created-at": time.Now().Format(time.RFC3339),
			},
		},
		Spec: pipelinev1alpha1.PhoenixProcessPipelineSpec{
			ExperimentID: deployment.ExperimentID,
			Variant:      deployment.PipelineType,
			ConfigMap:    fmt.Sprintf("%s-%s-config", deployment.ExperimentID, deployment.PipelineType),
		},
	}

	// Add node selector if target nodes are specified
	if len(deployment.TargetNodes) > 0 {
		// For simplicity, use the first target node
		// In real implementation, you might create multiple CRDs or use node affinity
		pipelineResource.Spec.NodeSelector = map[string]string{
			"kubernetes.io/hostname": deployment.TargetNodes[0],
		}
	}

	// Note: Variables would be stored in the ConfigMap, not directly in the CRD spec
	// The ConfigMap would be created separately with the actual OTel configuration

	// Create the resource
	if err := k.client.Create(ctx, pipelineResource); err != nil {
		return fmt.Errorf("failed to create PhoenixProcessPipeline: %w", err)
	}

	k.logger.Info("pipeline deployed successfully",
		zap.String("name", pipelineResource.Name),
		zap.String("namespace", pipelineResource.Namespace),
	)

	return nil
}

// GetPipelineStatus retrieves the status of a deployed pipeline
func (k *KubernetesClient) GetPipelineStatus(ctx context.Context, experimentID, pipelineType, namespace string) (*PipelineStatus, error) {
	pipelineName := fmt.Sprintf("%s-%s", experimentID, pipelineType)

	k.logger.Debug("getting pipeline status",
		zap.String("name", pipelineName),
		zap.String("namespace", namespace),
	)

	// Get the PhoenixProcessPipeline resource
	pipeline := &pipelinev1alpha1.PhoenixProcessPipeline{}
	namespacedName := types.NamespacedName{
		Name:      pipelineName,
		Namespace: namespace,
	}

	if err := k.client.Get(ctx, namespacedName, pipeline); err != nil {
		return nil, fmt.Errorf("failed to get PhoenixProcessPipeline: %w", err)
	}

	// Determine if pipeline is ready
	ready := pipeline.Status.Phase == "Running" && pipeline.Status.ReadyNodes > 0
	
	// Get message from conditions if available
	message := "No status available"
	if len(pipeline.Status.Conditions) > 0 {
		lastCondition := pipeline.Status.Conditions[len(pipeline.Status.Conditions)-1]
		message = lastCondition.Message
	}

	status := &PipelineStatus{
		Name:      pipeline.Name,
		Namespace: pipeline.Namespace,
		Phase:     pipeline.Status.Phase,
		Ready:     ready,
		Message:   message,
		PodCount:  pipeline.Status.ReadyNodes,
	}

	return status, nil
}

// DeletePipeline removes a deployed pipeline
func (k *KubernetesClient) DeletePipeline(ctx context.Context, experimentID, pipelineType, namespace string) error {
	pipelineName := fmt.Sprintf("%s-%s", experimentID, pipelineType)

	k.logger.Info("deleting pipeline",
		zap.String("name", pipelineName),
		zap.String("namespace", namespace),
	)

	pipeline := &pipelinev1alpha1.PhoenixProcessPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pipelineName,
			Namespace: namespace,
		},
	}

	if err := k.client.Delete(ctx, pipeline); err != nil {
		return fmt.Errorf("failed to delete PhoenixProcessPipeline: %w", err)
	}

	k.logger.Info("pipeline deleted successfully",
		zap.String("name", pipelineName),
	)

	return nil
}

// ListExperimentPipelines lists all pipelines for a specific experiment
func (k *KubernetesClient) ListExperimentPipelines(ctx context.Context, experimentID, namespace string) ([]*PipelineStatus, error) {
	k.logger.Debug("listing experiment pipelines",
		zap.String("experiment_id", experimentID),
		zap.String("namespace", namespace),
	)

	pipelineList := &pipelinev1alpha1.PhoenixProcessPipelineList{}
	listOptions := &client.ListOptions{
		Namespace: namespace,
	}

	// Add label selector for experiment ID
	labelSelector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			"phoenix.newrelic.com/experiment-id": experimentID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create label selector: %w", err)
	}
	listOptions.LabelSelector = labelSelector

	if err := k.client.List(ctx, pipelineList, listOptions); err != nil {
		return nil, fmt.Errorf("failed to list PhoenixProcessPipelines: %w", err)
	}

	var statuses []*PipelineStatus
	for _, pipeline := range pipelineList.Items {
		// Determine if pipeline is ready
		ready := pipeline.Status.Phase == "Running" && pipeline.Status.ReadyNodes > 0
		
		// Get message from conditions if available
		message := "No status available"
		if len(pipeline.Status.Conditions) > 0 {
			lastCondition := pipeline.Status.Conditions[len(pipeline.Status.Conditions)-1]
			message = lastCondition.Message
		}

		status := &PipelineStatus{
			Name:      pipeline.Name,
			Namespace: pipeline.Namespace,
			Phase:     pipeline.Status.Phase,
			Ready:     ready,
			Message:   message,
			PodCount:  pipeline.Status.ReadyNodes,
		}
		statuses = append(statuses, status)
	}

	k.logger.Info("listed experiment pipelines",
		zap.String("experiment_id", experimentID),
		zap.Int("count", len(statuses)),
	)

	return statuses, nil
}

// WaitForPipelineReady waits for a pipeline to become ready
func (k *KubernetesClient) WaitForPipelineReady(ctx context.Context, experimentID, pipelineType, namespace string, timeout time.Duration) error {
	k.logger.Info("waiting for pipeline to be ready",
		zap.String("experiment_id", experimentID),
		zap.String("pipeline_type", pipelineType),
		zap.Duration("timeout", timeout),
	)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for pipeline to be ready")
		case <-ticker.C:
			status, err := k.GetPipelineStatus(ctx, experimentID, pipelineType, namespace)
			if err != nil {
				k.logger.Warn("failed to get pipeline status", zap.Error(err))
				continue
			}

			if status.Ready {
				k.logger.Info("pipeline is ready",
					zap.String("name", status.Name),
					zap.String("phase", status.Phase),
				)
				return nil
			}

			k.logger.Debug("pipeline not ready yet",
				zap.String("name", status.Name),
				zap.String("phase", status.Phase),
				zap.String("message", status.Message),
			)
		}
	}
}