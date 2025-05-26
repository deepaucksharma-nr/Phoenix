package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/phoenix-vnext/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix-vnext/platform/projects/phoenix-cli/internal/config"
	"github.com/phoenix-vnext/platform/projects/phoenix-cli/internal/output"
)

// loadsimListProfilesCmd represents the loadsim list-profiles command
var loadsimListProfilesCmd = &cobra.Command{
	Use:   "list-profiles",
	Short: "List available load simulation profiles",
	Long: `List all available load simulation profiles with their descriptions.

Each profile simulates different process patterns to test various aspects
of metrics collection and cardinality reduction.`,
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

		// Get profiles from API
		ctx := context.Background()
		profiles, err := loadSimClient.GetProfiles(ctx)
		if err != nil {
			// Fall back to showing static profiles if API fails
			return showStaticProfiles()
		}

		output.Success("Available Load Simulation Profiles")
		fmt.Println()

		// Display profiles from API
		headers := []string{"Profile", "Description", "Process Count", "Churn Rate", "CPU Pattern", "Mem Pattern"}
		var data [][]string

		for _, profile := range profiles {
			data = append(data, []string{
				profile.Name,
				profile.Description,
				fmt.Sprintf("%d", profile.ProcessCount),
				fmt.Sprintf("%.1f%%", profile.ChurnRate*100),
				profile.CPUPattern,
				profile.MemPattern,
			})
		}

		output.Table(headers, data)
		fmt.Println()

		fmt.Println("Usage Examples:")
		fmt.Println("  phoenix loadsim start exp-12345678 --profile realistic --duration 1h")
		fmt.Println("  phoenix loadsim start exp-12345678 --profile high-cardinality --process-count 500")
		fmt.Println("  phoenix loadsim start exp-12345678 --profile process-churn --duration 30m")

		return nil
	},
}

func showStaticProfiles() error {
	output.Success("Available Load Simulation Profiles")
	fmt.Println()

	profiles := []struct {
		Name        string
		Description string
		Patterns    []string
	}{
		{
			Name:        "realistic",
			Description: "Simulates a typical production environment",
			Patterns: []string{
				"70% long-running processes (web servers, databases)",
				"30% short-lived processes (cron jobs, batch tasks)",
				"Steady CPU/memory usage with occasional spikes",
				"Process names follow common patterns (webapp-N, job-N)",
			},
		},
		{
			Name:        "high-cardinality",
			Description: "Tests cardinality reduction effectiveness",
			Patterns: []string{
				"Many unique process names",
				"Combinations of service/environment/region",
				"Random CPU and memory patterns",
				"Designed to create high metrics volume",
			},
		},
		{
			Name:        "process-churn",
			Description: "Rapid process creation and destruction",
			Patterns: []string{
				"80% churn rate every 2 seconds",
				"Short-lived processes (5 second lifetime)",
				"Spiky CPU and memory usage",
				"Tests metric collection under high churn",
			},
		},
		{
			Name:        "custom",
			Description: "User-defined patterns (requires configuration)",
			Patterns: []string{
				"Configurable process patterns",
				"Custom CPU/memory profiles",
				"Adjustable churn rates",
				"Define via environment variables or config file",
			},
		},
	}

	for _, profile := range profiles {
		fmt.Printf("Profile: %s\n", output.Bold(profile.Name))
		fmt.Printf("Description: %s\n", profile.Description)
		fmt.Println("Patterns:")
		for _, pattern := range profile.Patterns {
			fmt.Printf("  - %s\n", pattern)
		}
		fmt.Println()
	}

	fmt.Println("Usage Examples:")
	fmt.Println("  phoenix loadsim start exp-12345678 --profile realistic --duration 1h")
	fmt.Println("  phoenix loadsim start exp-12345678 --profile high-cardinality --process-count 500")
	fmt.Println("  phoenix loadsim start exp-12345678 --profile process-churn --duration 30m")

	return nil
}

func init() {
	loadsimCmd.AddCommand(loadsimListProfilesCmd)
}