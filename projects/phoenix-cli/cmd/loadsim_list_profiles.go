package cmd

import (
	"context"
	"fmt"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/config"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
	"github.com/spf13/cobra"
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
		Aliases     []string
		Description string
		MaxDuration string
		Impact      struct {
			CPU     string
			Memory  string
			Network string
		}
		UseCases []string
	}{
		{
			Name:        "high-cardinality",
			Aliases:     []string{"high-card"},
			Description: "Simulates high cardinality metrics explosion",
			MaxDuration: "30m",
			Impact: struct {
				CPU     string
				Memory  string
				Network string
			}{
				CPU:     "low",
				Memory:  "high",
				Network: "medium",
			},
			UseCases: []string{
				"Testing cardinality reduction processors",
				"Validating memory limits and safeguards",
				"Demonstrating the need for Phoenix optimization",
			},
		},
		{
			Name:        "realistic",
			Aliases:     []string{"normal"},
			Description: "Simulates normal production workload",
			MaxDuration: "1h",
			Impact: struct {
				CPU     string
				Memory  string
				Network string
			}{
				CPU:     "medium",
				Memory:  "low",
				Network: "low",
			},
			UseCases: []string{
				"Baseline performance testing",
				"Comparing optimized vs non-optimized pipelines",
				"General system validation",
			},
		},
		{
			Name:        "spike",
			Aliases:     []string{},
			Description: "Simulates traffic spikes and recovery",
			MaxDuration: "10m",
			Impact: struct {
				CPU     string
				Memory  string
				Network string
			}{
				CPU:     "variable",
				Memory:  "low",
				Network: "variable",
			},
			UseCases: []string{
				"Testing auto-scaling triggers",
				"Validating buffer and queue handling",
				"Checking recovery behavior",
			},
		},
		{
			Name:        "steady",
			Aliases:     []string{"process-churn"},
			Description: "Maintains constant load for stability testing",
			MaxDuration: "24h",
			Impact: struct {
				CPU     string
				Memory  string
				Network string
			}{
				CPU:     "low",
				Memory:  "minimal",
				Network: "low",
			},
			UseCases: []string{
				"Baseline measurements",
				"Long-term stability testing",
				"Resource leak detection",
			},
		},
	}

	// Table header
	headers := []string{"Profile", "Aliases", "Description", "Max Duration", "Resource Impact"}
	var data [][]string

	for _, profile := range profiles {
		aliasStr := "-"
		if len(profile.Aliases) > 0 {
			aliasStr = profile.Aliases[0]
			for i := 1; i < len(profile.Aliases); i++ {
				aliasStr += ", " + profile.Aliases[i]
			}
		}

		impactStr := fmt.Sprintf("CPU: %s, Mem: %s, Net: %s",
			profile.Impact.CPU, profile.Impact.Memory, profile.Impact.Network)

		data = append(data, []string{
			profile.Name,
			aliasStr,
			profile.Description,
			profile.MaxDuration,
			impactStr,
		})
	}

	output.Table(headers, data)
	fmt.Println()

	// Show detailed information
	fmt.Println("Profile Details:")
	fmt.Println()

	for _, profile := range profiles {
		fmt.Printf("%s (%s)\n", output.Bold(profile.Name), profile.Description)
		if len(profile.Aliases) > 0 {
			fmt.Printf("  Aliases: %v\n", profile.Aliases)
		}
		fmt.Printf("  Max Duration: %s\n", profile.MaxDuration)
		fmt.Printf("  Resource Impact: CPU=%s, Memory=%s, Network=%s\n",
			profile.Impact.CPU, profile.Impact.Memory, profile.Impact.Network)
		fmt.Println("  Use Cases:")
		for _, useCase := range profile.UseCases {
			fmt.Printf("    - %s\n", useCase)
		}
		fmt.Println()
	}

	fmt.Println("Usage Examples:")
	fmt.Println("  phoenix-cli loadsim start --profile high-cardinality --duration 5m")
	fmt.Println("  phoenix-cli loadsim start --profile realistic --duration 30m")
	fmt.Println("  phoenix-cli loadsim start --profile spike --duration 2m")
	fmt.Println("  phoenix-cli loadsim start --profile steady --duration 1h")
	fmt.Println()
	fmt.Println("Note: Profiles can be referenced by name or alias")

	return nil
}

func init() {
	loadsimCmd.AddCommand(loadsimListProfilesCmd)
}
