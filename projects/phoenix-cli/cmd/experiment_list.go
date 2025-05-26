package cmd

import (
	"fmt"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/config"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	listStatus   string
	listPageSize int
	listAll      bool
)

// listExperimentCmd represents the list experiments command
var listExperimentCmd = &cobra.Command{
	Use:   "list",
	Short: "List experiments",
	Long: `List all experiments with optional filtering.

Examples:
  # List all experiments
  phoenix experiment list

  # List only running experiments
  phoenix experiment list --status running

  # List experiments in JSON format
  phoenix experiment list -o json

  # List all experiments (no pagination)
  phoenix experiment list --all`,
	RunE: runListExperiments,
}

func init() {
	experimentCmd.AddCommand(listExperimentCmd)

	listExperimentCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status (pending, running, completed, failed)")
	listExperimentCmd.Flags().IntVar(&listPageSize, "page-size", 20, "Number of experiments per page")
	listExperimentCmd.Flags().BoolVar(&listAll, "all", false, "List all experiments without pagination")
}

func runListExperiments(cmd *cobra.Command, args []string) error {
	// Get config and check authentication
	cfg := config.New()
	token := cfg.GetToken()
	if token == "" {
		return fmt.Errorf("not authenticated. Please run: phoenix auth login")
	}

	// Create API client
	apiClient := client.NewAPIClient(cfg.GetAPIEndpoint(), token)

	// Prepare request
	req := client.ListExperimentsRequest{
		Status:   listStatus,
		PageSize: listPageSize,
	}

	if listAll {
		req.PageSize = 1000 // Large number to get all
	}

	// List experiments
	experiments, err := apiClient.ListExperiments(req)
	if err != nil {
		return fmt.Errorf("failed to list experiments: %w", err)
	}

	// Display results
	if len(experiments) == 0 {
		fmt.Println("No experiments found.")
		return nil
	}

	output.PrintExperimentList(experiments, outputFormat)

	return nil
}