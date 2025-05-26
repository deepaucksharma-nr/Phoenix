package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/config"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
)

var (
	rollbackVersion string
	rollbackForce   bool
)

// pipelineRollbackCmd represents the pipeline rollback command
var pipelineRollbackCmd = &cobra.Command{
	Use:   "rollback <deployment-id>",
	Short: "Revert to previous pipeline version",
	Long: `Revert a pipeline deployment to a previous version.

This command rolls back a deployment to a previously deployed configuration.
By default, it rolls back to the previous version, but you can specify
a specific version to roll back to.

Examples:
  # Rollback to previous version
  phoenix pipeline rollback deploy-abc123

  # Rollback to specific version
  phoenix pipeline rollback deploy-abc123 --version v2

  # Force rollback without confirmation
  phoenix pipeline rollback deploy-abc123 --force`,
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

		// Get current deployment info
		deployment, err := apiClient.GetPipelineDeployment(deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get deployment: %w", err)
		}

		output.Info(fmt.Sprintf("Current deployment: %s", deployment.Name))
		fmt.Printf("Pipeline: %s\n", deployment.Pipeline)
		fmt.Printf("Status: %s\n", deployment.Status)
		fmt.Printf("Namespace: %s\n\n", deployment.Namespace)

		// Confirm rollback unless forced
		if !rollbackForce {
			targetVersion := rollbackVersion
			if targetVersion == "" {
				targetVersion = "previous"
			}
			
			confirmed, err := output.Confirm(fmt.Sprintf("Are you sure you want to rollback to %s version?", targetVersion))
			if err != nil {
				return err
			}
			if !confirmed {
				output.Info("Rollback cancelled")
				return nil
			}
		}

		// Perform rollback
		rollbackReq := client.RollbackPipelineRequest{
			Version: rollbackVersion,
		}

		output.Info("Initiating rollback...")
		
		newDeployment, err := apiClient.RollbackPipeline(deploymentID, rollbackReq)
		if err != nil {
			return fmt.Errorf("failed to rollback pipeline: %w", err)
		}

		output.Success("Pipeline rollback initiated successfully")
		
		// Show new status
		data := [][]string{
			{"Deployment ID", newDeployment.ID},
			{"Pipeline", newDeployment.Pipeline},
			{"Status", newDeployment.Status},
			{"Phase", newDeployment.Phase},
		}

		if rollbackVersion != "" {
			data = append(data, []string{"Rolled Back To", rollbackVersion})
		}

		output.Table([]string{"Field", "Value"}, data)

		fmt.Println("\nMonitor rollback progress with:")
		fmt.Printf("  phoenix pipeline status %s --watch\n", deploymentID)

		return nil
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineRollbackCmd)

	pipelineRollbackCmd.Flags().StringVarP(&rollbackVersion, "version", "v", "", 
		"Specific version to rollback to (default: previous)")
	pipelineRollbackCmd.Flags().BoolVarP(&rollbackForce, "force", "f", false, 
		"Force rollback without confirmation")
}