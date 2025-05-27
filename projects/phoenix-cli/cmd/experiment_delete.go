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
	expDeleteForce bool
	expDeleteAll   bool
)

// deleteExperimentCmd represents the experiment delete command
var deleteExperimentCmd = &cobra.Command{
	Use:   "delete [ID...]",
	Short: "Delete one or more experiments",
	Long: `Delete one or more experiments.

This will permanently remove experiment data including metrics and results.
Running experiments cannot be deleted and must be stopped first.

Examples:
  # Delete a single experiment
  phoenix experiment delete exp-123

  # Delete multiple experiments
  phoenix experiment delete exp-123 exp-456 exp-789

  # Delete without confirmation
  phoenix experiment delete exp-123 --force

  # Delete all completed experiments (interactive)
  phoenix experiment delete --all --status completed`,
	Args: func(cmd *cobra.Command, args []string) error {
		if expDeleteAll && len(args) > 0 {
			return fmt.Errorf("cannot specify experiment IDs when using --all flag")
		}
		if !expDeleteAll && len(args) == 0 {
			return fmt.Errorf("requires at least one experiment ID or --all flag")
		}
		return nil
	},
	RunE: runExperimentDelete,
}

func init() {
	experimentCmd.AddCommand(deleteExperimentCmd)

	deleteExperimentCmd.Flags().BoolVarP(&expDeleteForce, "force", "f", false, "Force delete without confirmation")
	deleteExperimentCmd.Flags().BoolVar(&expDeleteAll, "all", false, "Delete all experiments matching filter criteria")
	deleteExperimentCmd.Flags().String("status", "", "Filter by status when using --all (completed, failed, cancelled)")
}

func runExperimentDelete(cmd *cobra.Command, args []string) error {
	// Get config and check authentication
	cfg := config.New()
	token := cfg.GetToken()
	if token == "" {
		return fmt.Errorf("not authenticated. Please run: phoenix auth login")
	}

	// Create API client
	apiClient := client.NewAPIClient(cfg.GetAPIEndpoint(), token)

	var experimentsToDelete []client.Experiment

	if expDeleteAll {
		// Get status filter
		statusFilter, _ := cmd.Flags().GetString("status")
		
		// List experiments with filter
		experiments, err := apiClient.ListExperiments(client.ListExperimentsRequest{})
		if err != nil {
			return fmt.Errorf("failed to list experiments: %w", err)
		}

		// Filter by status if specified
		for _, exp := range experiments.Experiments {
			if statusFilter == "" || exp.Phase == statusFilter {
				// Skip running experiments
				if exp.Phase == "running" || exp.Phase == "initializing" {
					continue
				}
				experimentsToDelete = append(experimentsToDelete, exp)
			}
		}

		if len(experimentsToDelete) == 0 {
			fmt.Println("No experiments found matching the criteria.")
			return nil
		}
	} else {
		// Get specific experiments
		for _, id := range args {
			experiment, err := apiClient.GetExperiment(id)
			if err != nil {
				output.PrintWarning(fmt.Sprintf("Failed to get experiment %s: %v", id, err))
				continue
			}

			// Check if experiment can be deleted
			if experiment.Phase == "running" || experiment.Phase == "initializing" {
				output.PrintWarning(fmt.Sprintf("Cannot delete %s experiment '%s' (%s). Stop it first.", 
					experiment.Phase, experiment.Name, experiment.ID))
				continue
			}

			experimentsToDelete = append(experimentsToDelete, *experiment)
		}
	}

	if len(experimentsToDelete) == 0 {
		fmt.Println("No experiments to delete.")
		return nil
	}

	// Show what will be deleted
	fmt.Printf("The following %d experiment(s) will be deleted:\n\n", len(experimentsToDelete))
	for _, exp := range experimentsToDelete {
		fmt.Printf("  • %s - %s (%s)\n", exp.ID, exp.Name, exp.Phase)
	}

	// Confirm unless force flag is set
	if !expDeleteForce {
		fmt.Printf("\nThis action cannot be undone. All experiment data will be permanently removed.\n")
		fmt.Print("\nType 'yes' to confirm: ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "yes" {
			fmt.Println("Operation cancelled.")
			return nil
		}
	}

	// Delete experiments
	successCount := 0
	failCount := 0

	for _, exp := range experimentsToDelete {
		fmt.Printf("Deleting experiment '%s'...", exp.Name)
		err := apiClient.DeleteExperiment(exp.ID)
		if err != nil {
			fmt.Printf(" ✗\n")
			output.PrintError(fmt.Errorf("Failed to delete %s: %v", exp.ID, err))
			failCount++
		} else {
			fmt.Printf(" ✓\n")
			successCount++
		}
	}

	// Summary
	fmt.Println()
	if successCount > 0 {
		output.PrintSuccess(fmt.Sprintf("Successfully deleted %d experiment(s)", successCount))
	}
	if failCount > 0 {
		output.PrintWarning(fmt.Sprintf("Failed to delete %d experiment(s)", failCount))
	}

	return nil
}