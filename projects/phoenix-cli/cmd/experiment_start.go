package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/config"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
	"github.com/spf13/cobra"
)

// startExperimentCmd represents the experiment start command
var startExperimentCmd = &cobra.Command{
	Use:   "start [ID]",
	Short: "Start an experiment",
	Long: `Start a pending experiment.

The experiment must be in 'pending' status to be started.

Examples:
  # Start an experiment
  phoenix experiment start exp-123

  # Start and follow progress
  phoenix experiment start exp-123 --follow`,
	Args: cobra.ExactArgs(1),
	RunE: runExperimentStart,
}

var startAndFollow bool

func init() {
	experimentCmd.AddCommand(startExperimentCmd)

	startExperimentCmd.Flags().BoolVarP(&startAndFollow, "follow", "f", false, "Follow experiment progress after starting")
}

func runExperimentStart(cmd *cobra.Command, args []string) error {
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

	// Check if experiment can be started
	if experiment.Phase != "pending" {
		return fmt.Errorf("experiment is %s, can only start pending experiments", experiment.Phase)
	}

	// Start the experiment
	fmt.Printf("Starting experiment '%s'...\n", experiment.Name)
	err = apiClient.StartExperiment(experimentID)
	if err != nil {
		return fmt.Errorf("failed to start experiment: %w", err)
	}

	// Wait for pipelines to be deployed via agents
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Poll for pipeline deployment status
	fmt.Println("Waiting for pipeline deployments...")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for pipeline deployments")
		case <-ticker.C:
			// Check deployment status via API
			exp, err := apiClient.GetExperiment(experimentID)
			if err == nil && exp.Phase == "running" {
				fmt.Println("Pipelines deployed successfully")
				goto deployed
			}
		}
	}
deployed:

	output.PrintSuccess("Experiment started successfully!")

	// Show deployment information
	fmt.Printf("\nDeployment Information:\n")
	fmt.Printf("  Baseline Pipeline:  %s\n", experiment.BaselinePipeline)
	fmt.Printf("  Candidate Pipeline: %s\n", experiment.CandidatePipeline)
	fmt.Printf("  Target Nodes:       %s\n", formatTargetNodes(experiment.TargetNodes))

	if startAndFollow {
		fmt.Printf("\nFollowing experiment progress...\n")
		return followExperimentStatus(apiClient, experimentID)
	}

	fmt.Printf("\nTo monitor progress, run:\n")
	fmt.Printf("  phoenix experiment status %s --follow\n", experimentID)

	return nil
}

func formatTargetNodes(nodes map[string]string) string {
	if len(nodes) == 0 {
		return "none"
	}

	result := ""
	i := 0
	for k, v := range nodes {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s=%s", k, v)
		i++
	}
	return result
}
