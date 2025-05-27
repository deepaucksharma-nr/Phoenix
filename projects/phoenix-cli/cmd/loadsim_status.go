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

var loadSimWatch bool

// loadsimStatusCmd represents the loadsim status command
var loadsimStatusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Show status of load simulations",
	Long: `Show the status of one or all load simulations.

If no name is provided, lists all load simulations in the system.
Use the --watch flag to continuously monitor status updates.

Examples:
  # Show all load simulations
  phoenix loadsim status

  # Show specific load simulation
  phoenix loadsim status loadsim-12345678

  # Watch load simulation status
  phoenix loadsim status loadsim-12345678 --watch`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
		loadSimClient := client.NewLoadSimulationClient(apiClient)

		ctx := context.Background()

		if len(args) == 0 {
			// List all load simulations
			return listLoadSimulations(ctx, loadSimClient)
		}

		// Show specific load simulation
		name := args[0]

		// If the user provided an experiment ID, convert it to loadsim name
		if len(name) == 12 && name[:4] == "exp-" {
			name = fmt.Sprintf("loadsim-%s", name[4:])
		}

		if loadSimWatch {
			return watchLoadSimulation(ctx, loadSimClient, name)
		}

		return showLoadSimulation(ctx, loadSimClient, name)
	},
}

func listLoadSimulations(ctx context.Context, loadSimClient *client.LoadSimulationClient) error {
	list, err := loadSimClient.List(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list load simulations: %w", err)
	}

	if len(list) == 0 {
		output.Info("No load simulations found")
		return nil
	}

	headers := []string{"Name", "Experiment", "Profile", "Duration", "Status", "Started"}
	var data [][]string

	for _, sim := range list {
		started := "Not started"
		if sim.StartTime != nil {
			started = sim.StartTime.Format("2006-01-02 15:04:05")
		}

		data = append(data, []string{
			sim.Name,
			sim.ExperimentID,
			sim.Profile,
			sim.Duration,
			string(sim.Status),
			started,
		})
	}

	output.Table(headers, data)
	return nil
}

func showLoadSimulation(ctx context.Context, loadSimClient *client.LoadSimulationClient, name string) error {
	sim, err := loadSimClient.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get load simulation: %w", err)
	}

	output.Success(fmt.Sprintf("Load Simulation: %s", sim.Name))

	data := [][]string{
		{"Experiment ID", sim.ExperimentID},
		{"Profile", sim.Profile},
		{"Duration", sim.Duration},
		{"Process Count", fmt.Sprintf("%d", sim.ProcessCount)},
		{"Status", string(sim.Status)},
	}

	if sim.Message != "" {
		data = append(data, []string{"Message", sim.Message})
	}

	if sim.StartTime != nil {
		data = append(data, []string{"Start Time", sim.StartTime.Format(time.RFC3339)})
		if sim.EndTime == nil {
			elapsed := time.Since(*sim.StartTime).Round(time.Second)
			data = append(data, []string{"Elapsed", elapsed.String()})
		}
	}

	if sim.EndTime != nil {
		data = append(data, []string{"End Time", sim.EndTime.Format(time.RFC3339)})
		if sim.StartTime != nil {
			duration := sim.EndTime.Sub(*sim.StartTime).Round(time.Second)
			data = append(data, []string{"Total Duration", duration.String()})
		}
	}

	output.Table([]string{"Field", "Value"}, data)
	return nil
}

func watchLoadSimulation(ctx context.Context, loadSimClient *client.LoadSimulationClient, name string) error {
	output.Info(fmt.Sprintf("Watching load simulation %s (press Ctrl+C to stop)...\n", name))

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Show initial status
	if err := showLoadSimulation(ctx, loadSimClient, name); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			fmt.Print("\033[H\033[2J") // Clear screen
			fmt.Println()
			if err := showLoadSimulation(ctx, loadSimClient, name); err != nil {
				return err
			}
		}
	}
}

func init() {
	loadsimCmd.AddCommand(loadsimStatusCmd)

	loadsimStatusCmd.Flags().BoolVarP(&loadSimWatch, "watch", "w", false,
		"Watch status updates continuously")
}
