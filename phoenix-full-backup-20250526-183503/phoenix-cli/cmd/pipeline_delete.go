package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/config"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
)

var (
	deleteForce bool
)

// pipelineDeleteCmd represents the pipeline delete command
var pipelineDeleteCmd = &cobra.Command{
	Use:   "delete <deployment-id>",
	Short: "Remove deployed pipeline",
	Long: `Remove a deployed pipeline and clean up all associated resources.

This command deletes a pipeline deployment, stopping all collectors
and removing the configuration from target nodes.

Examples:
  # Delete a pipeline deployment
  phoenix pipeline delete deploy-abc123

  # Force delete without confirmation
  phoenix pipeline delete deploy-abc123 --force`,
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

		// Get deployment info first
		deployment, err := apiClient.GetPipelineDeployment(deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get deployment: %w", err)
		}

		// Display deployment information
		output.Warning(fmt.Sprintf("About to delete pipeline deployment: %s", deployment.Name))
		fmt.Printf("Pipeline: %s\n", deployment.Pipeline)
		fmt.Printf("Namespace: %s\n", deployment.Namespace)
		fmt.Printf("Status: %s\n", deployment.Status)
		
		if len(deployment.TargetNodes) > 0 {
			fmt.Printf("\nThis will affect %d target nodes:\n", len(deployment.TargetNodes))
			for node := range deployment.TargetNodes {
				fmt.Printf("  - %s\n", node)
			}
		}

		// Confirm deletion unless forced
		if !deleteForce {
			fmt.Println()
			confirmed, err := output.Confirm("Are you sure you want to delete this deployment?")
			if err != nil {
				return err
			}
			if !confirmed {
				output.Info("Deletion cancelled")
				return nil
			}
		}

		// Delete the deployment
		output.Info("Deleting pipeline deployment...")
		
		if err := apiClient.DeletePipelineDeployment(deploymentID); err != nil {
			return fmt.Errorf("failed to delete deployment: %w", err)
		}

		output.Success("Pipeline deployment deleted successfully")
		
		// Show summary
		fmt.Println("\nDeleted resources:")
		fmt.Printf("  - Deployment: %s\n", deployment.Name)
		fmt.Printf("  - Pipeline: %s\n", deployment.Pipeline)
		if deployment.Instances != nil {
			fmt.Printf("  - Collectors: %d instances\n", deployment.Instances.Desired)
		}

		return nil
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineDeleteCmd)

	pipelineDeleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, 
		"Force deletion without confirmation")
}