package cmd

import (
	"fmt"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/config"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
	"github.com/spf13/cobra"
)

var experimentRollbackCmd = &cobra.Command{
	Use:   "rollback [experiment-id]",
	Short: "Rollback an experiment to baseline state",
	Long: `Rollback an experiment by stopping the candidate pipeline and 
keeping only the baseline pipeline running.

This command can be used when an experiment shows degraded performance
or unexpected behavior, allowing you to quickly revert to the baseline
configuration.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		experimentID := args[0]
		
		// Get flags
		instant, _ := cmd.Flags().GetBool("instant")
		reason, _ := cmd.Flags().GetString("reason")
		
		// Get config and check authentication
		cfg := config.New()
		token := cfg.GetToken()
		if token == "" {
			return fmt.Errorf("not authenticated. Please run: phoenix auth login")
		}
		
		// Create API client
		apiClient := client.NewAPIClient(cfg.GetAPIEndpoint(), token)
		
		// Rollback experiment
		result, err := apiClient.RollbackExperiment(experimentID, instant, reason)
		if err != nil {
			return fmt.Errorf("failed to rollback experiment: %w", err)
		}
		
		// Display result
		output.PrintSuccess("Experiment rollback initiated")
		
		if result["status"] == "success" {
			fmt.Printf("Experiment ID: %s\n", experimentID)
			if msg, ok := result["message"].(string); ok {
				fmt.Printf("Message: %s\n", msg)
			}
			if hostsAffected, ok := result["hosts_affected"].(float64); ok {
				fmt.Printf("Hosts affected: %d\n", int(hostsAffected))
			}
		}
		
		return nil
	},
}

func init() {
	experimentCmd.AddCommand(experimentRollbackCmd)
	
	// Add flags
	experimentRollbackCmd.Flags().Bool("instant", false, "Perform instant rollback (skip graceful shutdown)")
	experimentRollbackCmd.Flags().String("reason", "", "Reason for rollback")
}