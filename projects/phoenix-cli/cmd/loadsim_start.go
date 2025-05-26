package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/phoenix-vnext/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix-vnext/platform/projects/phoenix-cli/internal/config"
	"github.com/phoenix-vnext/platform/projects/phoenix-cli/internal/output"
)

var (
	loadSimProfile      string
	loadSimDuration     string
	loadSimProcessCount int32
	loadSimNodeSelector map[string]string
)

// loadsimStartCmd represents the loadsim start command
var loadsimStartCmd = &cobra.Command{
	Use:   "start <experiment-id>",
	Short: "Start a load simulation for an experiment",
	Long: `Start a load simulation for an experiment using a predefined profile.

Available profiles:
  - realistic: Mix of long-running and short-lived processes
  - high-cardinality: Many unique process names to test cardinality reduction
  - process-churn: Rapid process creation/destruction
  - custom: User-defined patterns (requires additional configuration)

Examples:
  # Start a realistic load simulation for 1 hour
  phoenix loadsim start exp-12345678 --profile realistic --duration 1h

  # Start high-cardinality simulation with 500 processes
  phoenix loadsim start exp-12345678 --profile high-cardinality --process-count 500

  # Start process churn simulation on specific nodes
  phoenix loadsim start exp-12345678 --profile process-churn --node-selector workload=test`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		experimentID := args[0]

		// Validate experiment ID format
		if !isValidExperimentID(experimentID) {
			return fmt.Errorf("invalid experiment ID format: %s (expected: exp-XXXXXXXX)", experimentID)
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

		// Create load simulation request
		req := client.CreateLoadSimulationRequest{
			ExperimentID: experimentID,
			Profile:      loadSimProfile,
			Duration:     loadSimDuration,
			ProcessCount: loadSimProcessCount,
			NodeSelector: loadSimNodeSelector,
		}

		// Create the load simulation
		ctx := context.Background()
		loadSim, err := loadSimClient.Start(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create load simulation: %w", err)
		}

		// Output the result
		output.Success("Load simulation started successfully")
		
		data := [][]string{
			{"Name", loadSim.Name},
			{"Experiment ID", loadSim.ExperimentID},
			{"Profile", loadSim.Profile},
			{"Duration", loadSim.Duration},
			{"Process Count", fmt.Sprintf("%d", loadSim.ProcessCount)},
			{"Status", string(loadSim.Status)},
		}
		
		output.Table([]string{"Field", "Value"}, data)
		
		fmt.Fprintf(os.Stdout, "\nMonitor status with: phoenix loadsim status %s\n", loadSim.Name)

		return nil
	},
}

func init() {
	loadsimCmd.AddCommand(loadsimStartCmd)

	loadsimStartCmd.Flags().StringVarP(&loadSimProfile, "profile", "p", "realistic", 
		"Load simulation profile (realistic, high-cardinality, process-churn, custom)")
	loadsimStartCmd.Flags().StringVarP(&loadSimDuration, "duration", "d", "30m", 
		"Duration of the load simulation (e.g., 30m, 1h)")
	loadsimStartCmd.Flags().Int32Var(&loadSimProcessCount, "process-count", 100, 
		"Number of processes to simulate")
	loadsimStartCmd.Flags().StringToStringVar(&loadSimNodeSelector, "node-selector", nil, 
		"Node selector for load simulation pods (key=value pairs)")
}

// isValidExperimentID checks if the experiment ID follows the expected format
func isValidExperimentID(id string) bool {
	if len(id) != 12 {
		return false
	}
	if id[:4] != "exp-" {
		return false
	}
	// Check that the rest contains only lowercase letters and numbers
	for _, c := range id[4:] {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}