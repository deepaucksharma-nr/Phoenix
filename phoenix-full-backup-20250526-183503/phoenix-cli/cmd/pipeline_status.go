package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/config"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
)

var (
	pipelineStatusNamespace string
	pipelineStatusWatch     bool
)

// pipelineStatusCmd represents the pipeline status command
var pipelineStatusCmd = &cobra.Command{
	Use:   "status <deployment-id>",
	Short: "Show deployment status with metrics",
	Long: `Show the status of a deployed pipeline including health metrics.

This command displays the current status of a pipeline deployment,
including instance health, throughput metrics, and cardinality reduction.

Examples:
  # Show status of a specific deployment
  phoenix pipeline status deploy-abc123

  # Show status with namespace filter
  phoenix pipeline status deploy-abc123 --namespace production

  # Watch status updates
  phoenix pipeline status deploy-abc123 --watch`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deploymentID := args[0]

		// Get API client configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.Token == "" {
			return fmt.Errorf("not authenticated. Please run 'phoenix auth login' first")
		}

		// Create API client
		apiClient := client.NewAPIClient(cfg.APIEndpoint, cfg.Token)
		
		// Get deployment status
		deployment, err := apiClient.GetPipelineDeployment(deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get deployment status: %w", err)
		}

		// Display status
		output.Success(fmt.Sprintf("Pipeline Deployment: %s", deployment.Name))
		
		// Basic information
		data := [][]string{
			{"Deployment ID", deployment.ID},
			{"Pipeline", deployment.Pipeline},
			{"Namespace", deployment.Namespace},
			{"Status", deployment.Status},
			{"Phase", deployment.Phase},
		}

		// Instance information
		if deployment.Instances != nil {
			data = append(data, []string{"Instances", fmt.Sprintf("%d/%d ready", 
				deployment.Instances.Ready, deployment.Instances.Desired)})
		}

		// Metrics if available
		if deployment.Metrics != nil {
			data = append(data, []string{"", ""}) // Empty row for spacing
			data = append(data, []string{"=== Metrics ===", ""})
			data = append(data, []string{"Cardinality", fmt.Sprintf("%d series", deployment.Metrics.Cardinality)})
			data = append(data, []string{"Throughput", deployment.Metrics.Throughput})
			data = append(data, []string{"Error Rate", fmt.Sprintf("%.2f%%", deployment.Metrics.ErrorRate)})
			data = append(data, []string{"CPU Usage", fmt.Sprintf("%.1f%%", deployment.Metrics.CPUUsage)})
			data = append(data, []string{"Memory Usage", fmt.Sprintf("%.1f%%", deployment.Metrics.MemoryUsage)})
		}

		// Target nodes
		if len(deployment.TargetNodes) > 0 {
			data = append(data, []string{"", ""}) // Empty row for spacing
			data = append(data, []string{"=== Target Nodes ===", ""})
			for host, status := range deployment.TargetNodes {
				data = append(data, []string{host, status})
			}
		}

		output.Table([]string{"Field", "Value"}, data)

		// Show health indicators
		fmt.Println("\nHealth Status:")
		if deployment.Status == "active" && deployment.Phase == "running" {
			output.Success("✓ Pipeline is healthy and processing metrics")
		} else if deployment.Status == "failed" {
			output.Error("✗ Pipeline has failed - check logs for details")
		} else if deployment.Status == "updating" {
			output.Warning("⟳ Pipeline is updating...")
		} else {
			output.Info("◎ Pipeline is in transitional state")
		}

		// Calculate cardinality reduction if we have metrics
		if deployment.Metrics != nil && deployment.Metrics.Cardinality > 0 {
			// Assume baseline of 10K series for 100 processes
			baseline := int64(10000)
			reduction := float64(baseline-deployment.Metrics.Cardinality) / float64(baseline) * 100
			if reduction > 0 {
				fmt.Printf("\nCardinality Reduction: %.1f%% (from ~%d to %d series)\n", 
					reduction, baseline, deployment.Metrics.Cardinality)
			}
		}

		return nil
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineStatusCmd)

	pipelineStatusCmd.Flags().StringVarP(&pipelineStatusNamespace, "namespace", "n", "", 
		"Filter by namespace")
	pipelineStatusCmd.Flags().BoolVarP(&pipelineStatusWatch, "watch", "w", false, 
		"Watch status updates")
}