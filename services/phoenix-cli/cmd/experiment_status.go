package cmd

import (
	"fmt"
	"time"

	"github.com/phoenix-vnext/platform/services/phoenix-cli/internal/client"
	"github.com/phoenix-vnext/platform/services/phoenix-cli/internal/config"
	"github.com/phoenix-vnext/platform/services/phoenix-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	statusFollow   bool
	statusInterval time.Duration
)

// statusExperimentCmd represents the experiment status command
var statusExperimentCmd = &cobra.Command{
	Use:   "status [ID]",
	Short: "Get experiment status",
	Long: `Get the current status of an experiment.

Examples:
  # Get status of an experiment
  phoenix experiment status exp-123

  # Follow experiment progress
  phoenix experiment status exp-123 --follow

  # Follow with custom interval
  phoenix experiment status exp-123 --follow --interval 10s`,
	Args: cobra.ExactArgs(1),
	RunE: runExperimentStatus,
}

func init() {
	experimentCmd.AddCommand(statusExperimentCmd)

	statusExperimentCmd.Flags().BoolVarP(&statusFollow, "follow", "f", false, "Follow experiment progress")
	statusExperimentCmd.Flags().DurationVar(&statusInterval, "interval", 5*time.Second, "Update interval when following")
}

func runExperimentStatus(cmd *cobra.Command, args []string) error {
	experimentID := args[0]

	// Get config and check authentication
	cfg := config.New()
	token := cfg.GetToken()
	if token == "" {
		return fmt.Errorf("not authenticated. Please run: phoenix auth login")
	}

	// Create API client
	apiClient := client.NewAPIClient(cfg.GetAPIEndpoint(), token)

	if statusFollow {
		// Follow mode - continuously update status
		return followExperimentStatus(apiClient, experimentID)
	}

	// Single status check
	experiment, err := apiClient.GetExperiment(experimentID)
	if err != nil {
		return fmt.Errorf("failed to get experiment status: %w", err)
	}

	// Display experiment details
	output.PrintExperiment(experiment)

	// Show deployment status if available
	if experiment.Status == "running" || experiment.Status == "initializing" {
		fmt.Println("\nDeployment Status:")
		fmt.Printf("  Baseline:  🟢 Running\n")
		fmt.Printf("  Candidate: 🟢 Running\n")
	}

	return nil
}

func followExperimentStatus(apiClient *client.APIClient, experimentID string) error {
	fmt.Printf("Following experiment %s (press Ctrl+C to stop)...\n\n", experimentID)

	ticker := time.NewTicker(statusInterval)
	defer ticker.Stop()

	// Initial status
	if err := displayExperimentStatus(apiClient, experimentID); err != nil {
		return err
	}

	// Follow updates
	for range ticker.C {
		// Clear screen (simple approach - just add newlines)
		fmt.Printf("\n")
		
		if err := displayExperimentStatus(apiClient, experimentID); err != nil {
			return err
		}

		// Check if experiment is complete
		experiment, _ := apiClient.GetExperiment(experimentID)
		if experiment != nil && (experiment.Status == "completed" || experiment.Status == "failed" || experiment.Status == "cancelled") {
			fmt.Printf("\nExperiment %s. Stopping follow mode.\n", experiment.Status)
			break
		}
	}

	return nil
}

func displayExperimentStatus(apiClient *client.APIClient, experimentID string) error {
	experiment, err := apiClient.GetExperiment(experimentID)
	if err != nil {
		return fmt.Errorf("failed to get experiment: %w", err)
	}

	// Display summary
	fmt.Printf("Experiment: %s (%s)\n", experiment.Name, experiment.ID[:8])
	fmt.Printf("Status:     %s\n", output.ColorizeStatus(experiment.Status))
	
	if experiment.StartedAt != nil {
		duration := time.Since(*experiment.StartedAt)
		fmt.Printf("Duration:   %s\n", formatDuration(duration))
	}

	// Display metrics if available
	if experiment.Status == "running" && experiment.Results != nil {
		fmt.Printf("\nCurrent Metrics:\n")
		fmt.Printf("  Baseline Cardinality:   %d\n", experiment.Results.BaselineMetrics.Cardinality)
		fmt.Printf("  Candidate Cardinality:  %d\n", experiment.Results.CandidateMetrics.Cardinality)
		fmt.Printf("  Reduction:              %.1f%%\n", experiment.Results.CardinalityReduction)
		
		if experiment.Results.BaselineMetrics.ErrorRate > 0 || experiment.Results.CandidateMetrics.ErrorRate > 0 {
			fmt.Printf("\n  Error Rates:\n")
			fmt.Printf("    Baseline:  %.2f%%\n", experiment.Results.BaselineMetrics.ErrorRate*100)
			fmt.Printf("    Candidate: %.2f%%\n", experiment.Results.CandidateMetrics.ErrorRate*100)
		}
	}

	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}