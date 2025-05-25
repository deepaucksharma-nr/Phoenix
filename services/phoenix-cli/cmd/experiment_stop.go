package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/phoenix-vnext/platform/services/phoenix-cli/internal/client"
	"github.com/phoenix-vnext/platform/services/phoenix-cli/internal/config"
	"github.com/phoenix-vnext/platform/services/phoenix-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	stopForce  bool
	stopReason string
)

// stopExperimentCmd represents the experiment stop command
var stopExperimentCmd = &cobra.Command{
	Use:   "stop [ID]",
	Short: "Stop a running experiment",
	Long: `Stop a running experiment.

This will stop both baseline and candidate pipelines and mark the experiment as cancelled.

Examples:
  # Stop an experiment
  phoenix experiment stop exp-123

  # Stop with a reason
  phoenix experiment stop exp-123 --reason "High error rate detected"

  # Force stop without confirmation
  phoenix experiment stop exp-123 --force`,
	Args: cobra.ExactArgs(1),
	RunE: runExperimentStop,
}

func init() {
	experimentCmd.AddCommand(stopExperimentCmd)

	stopExperimentCmd.Flags().BoolVarP(&stopForce, "force", "f", false, "Force stop without confirmation")
	stopExperimentCmd.Flags().StringVarP(&stopReason, "reason", "r", "", "Reason for stopping the experiment")
}

func runExperimentStop(cmd *cobra.Command, args []string) error {
	experimentID := args[0]

	// Get config and check authentication
	cfg := config.New()
	token := cfg.GetToken()
	if token == "" {
		return fmt.Errorf("not authenticated. Please run: phoenix auth login")
	}

	// Create API client
	apiClient := client.NewAPIClient(cfg.GetAPIEndpoint(), token)

	// Get current experiment status
	experiment, err := apiClient.GetExperiment(experimentID)
	if err != nil {
		return fmt.Errorf("failed to get experiment: %w", err)
	}

	// Check if experiment can be stopped
	if experiment.Status != "running" && experiment.Status != "initializing" {
		return fmt.Errorf("experiment is %s, can only stop running or initializing experiments", experiment.Status)
	}

	// Confirm unless force flag is set
	if !stopForce {
		fmt.Printf("Are you sure you want to stop experiment '%s'?\n", experiment.Name)
		fmt.Printf("This will terminate both baseline and candidate pipelines.\n")
		fmt.Print("\nType 'yes' to confirm: ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "yes" {
			fmt.Println("Operation cancelled.")
			return nil
		}
	}

	// Stop the experiment
	fmt.Printf("Stopping experiment '%s'...\n", experiment.Name)
	err = apiClient.StopExperiment(experimentID)
	if err != nil {
		return fmt.Errorf("failed to stop experiment: %w", err)
	}

	output.PrintSuccess("Experiment stopped successfully!")

	// If there was a duration, show how long it ran
	if experiment.StartedAt != nil {
		duration := formatDuration(time.Since(*experiment.StartedAt))
		fmt.Printf("\nExperiment ran for: %s\n", duration)
	}

	// Show next steps
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  • Review partial results: phoenix experiment metrics %s\n", experimentID)
	fmt.Printf("  • Create a new experiment with adjusted parameters\n")

	return nil
}