package cmd

import (
	"fmt"
	"strings"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/config"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	deploymentName   string
	deployPipeline   string
	deployTarget     string
	deploySelector   map[string]string
	deployParams     map[string]string
	deployCPURequest string
	deployCPULimit   string
	deployMemRequest string
	deployMemLimit   string
)

// deployPipelineCmd represents the pipeline deploy command
var pipelineDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a pipeline directly",
	Long: `Deploy a pipeline configuration directly without running an experiment.

This is useful for:
  - Rolling out proven optimizations to production
  - Deploying standard configurations
  - Emergency rollbacks to baseline

Examples:
  # Deploy a pipeline to specific agents
  phoenix pipeline deploy \
    --name production-optimization \
    --pipeline process-topk-v1 \
    --target "production" \
    --selector "environment=production,tier=frontend" \
    --param top_k=20

  # Deploy with resource limits
  phoenix pipeline deploy \
    --name staging-test \
    --pipeline process-priority-filter-v1 \
    --target "staging" \
    --selector "app=api" \
    --param critical_processes=nginx,envoy \
    --cpu-limit 500m \
    --memory-limit 512Mi`,
	RunE: runPipelineDeploy,
}

func init() {
	pipelineCmd.AddCommand(pipelineDeployCmd)

	// Required flags
	pipelineDeployCmd.Flags().StringVarP(&deploymentName, "name", "n", "", "Deployment name (required)")
	pipelineDeployCmd.Flags().StringVar(&deployPipeline, "pipeline", "", "Pipeline template to deploy (required)")
	pipelineDeployCmd.Flags().StringVar(&deployTarget, "target", "default", "Target environment or agent group")
	pipelineDeployCmd.Flags().StringToStringVar(&deploySelector, "selector", nil, "Target node selector labels (required)")

	pipelineDeployCmd.MarkFlagRequired("name")
	pipelineDeployCmd.MarkFlagRequired("pipeline")
	pipelineDeployCmd.MarkFlagRequired("selector")

	// Optional flags
	pipelineDeployCmd.Flags().StringToStringVar(&deployParams, "param", nil, "Pipeline parameters (can be specified multiple times)")
	pipelineDeployCmd.Flags().StringVar(&deployCPURequest, "cpu-request", "100m", "CPU request")
	pipelineDeployCmd.Flags().StringVar(&deployCPULimit, "cpu-limit", "500m", "CPU limit")
	pipelineDeployCmd.Flags().StringVar(&deployMemRequest, "memory-request", "128Mi", "Memory request")
	pipelineDeployCmd.Flags().StringVar(&deployMemLimit, "memory-limit", "512Mi", "Memory limit")
}

func runPipelineDeploy(cmd *cobra.Command, args []string) error {
	// Get config and check authentication
	cfg := config.New()
	token := cfg.GetToken()
	if token == "" {
		return fmt.Errorf("not authenticated. Please run: phoenix auth login")
	}

	// Create API client
	apiClient := client.NewAPIClient(cfg.GetAPIEndpoint(), token)

	// Convert string params to interface{} map
	parameters := make(map[string]interface{})
	for k, v := range deployParams {
		// Handle comma-separated lists as arrays
		if strings.Contains(v, ",") {
			parameters[k] = strings.Split(v, ",")
		} else {
			parameters[k] = v
		}
	}

	// Prepare deployment request
	req := client.CreatePipelineDeploymentRequest{
		DeploymentName: deploymentName,
		PipelineName:   deployPipeline,
		TargetEnv:      deployTarget,
		TargetNodes:    deploySelector,
		Parameters:     parameters,
		Resources: &client.ResourceRequirements{
			Requests: client.ResourceList{
				CPU:    deployCPURequest,
				Memory: deployMemRequest,
			},
			Limits: client.ResourceList{
				CPU:    deployCPULimit,
				Memory: deployMemLimit,
			},
		},
	}

	// Display deployment details
	fmt.Println("Deployment Configuration:")
	fmt.Printf("  Name:      %s\n", deploymentName)
	fmt.Printf("  Pipeline:  %s\n", deployPipeline)
	fmt.Printf("  Target:    %s\n", deployTarget)
	fmt.Printf("  Selectors: %s\n", formatSelector(deploySelector))

	if len(parameters) > 0 {
		fmt.Println("  Parameters:")
		for k, v := range parameters {
			fmt.Printf("    %s: %v\n", k, v)
		}
	}

	// Create deployment
	fmt.Printf("\nDeploying pipeline...\n")
	deployment, err := apiClient.CreatePipelineDeployment(req)
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	output.PrintSuccess("Pipeline deployment created successfully!")

	status, err := apiClient.GetPipelineDeploymentStatus(deployment.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve deployment status: %w", err)
	}

	// Display deployment info
	fmt.Printf("\nDeployment Details:\n")
	fmt.Printf("  ID:     %s\n", status.DeploymentID)
	fmt.Printf("  Status: %s\n", status.Status)
	fmt.Printf("  Phase:  %s\n", status.Phase)

	fmt.Printf("\nTo check deployment status:\n")
	fmt.Printf("  phoenix pipeline status %s\n", status.DeploymentID)

	fmt.Printf("\nTo list all deployments:\n")
	fmt.Printf("  phoenix pipeline list-deployments --target %s\n", deployTarget)

	return nil
}

func formatSelector(selector map[string]string) string {
	if len(selector) == 0 {
		return "none"
	}

	parts := []string{}
	for k, v := range selector {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ", ")
}
