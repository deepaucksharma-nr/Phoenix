package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/config"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
)

// loadsimStopCmd represents the loadsim stop command
var loadsimStopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a running load simulation",
	Long: `Stop a running load simulation by name.

This will gracefully terminate all simulated processes and clean up resources.

Examples:
  # Stop a load simulation
  phoenix loadsim stop loadsim-12345678

  # Stop using experiment ID
  phoenix loadsim stop exp-12345678`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// If the user provided an experiment ID, convert it to loadsim name
		if len(name) == 12 && name[:4] == "exp-" {
			name = fmt.Sprintf("loadsim-%s", name[4:])
		}

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

		// Get the current status before stopping
		ctx := context.Background()
		loadSim, err := loadSimClient.Get(ctx, name)
		if err != nil {
			return fmt.Errorf("failed to get load simulation: %w", err)
		}

		// Stop the load simulation
		err = loadSimClient.Stop(ctx, name)
		if err != nil {
			return fmt.Errorf("failed to stop load simulation: %w", err)
		}

		output.Success("Load simulation stop initiated")
		
		data := [][]string{
			{"Name", loadSim.Name},
			{"Experiment ID", loadSim.ExperimentID},
			{"Profile", loadSim.Profile},
			{"Previous Status", string(loadSim.Status)},
		}
		
		output.Table([]string{"Field", "Value"}, data)

		return nil
	},
}

func init() {
	loadsimCmd.AddCommand(loadsimStopCmd)
}