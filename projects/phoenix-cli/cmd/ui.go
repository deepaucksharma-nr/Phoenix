package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "UI-related commands",
	Long:  `Commands for interacting with the Phoenix Dashboard and UI features`,
}

var uiOpenCmd = &cobra.Command{
	Use:   "open",
	Short: "Open Phoenix Dashboard in browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		dashboardURL := "http://localhost:3000"
		if url := os.Getenv("PHOENIX_DASHBOARD_URL"); url != "" {
			dashboardURL = url
		}

		fmt.Printf("Opening Phoenix Dashboard at %s...\n", dashboardURL)
		return openBrowser(dashboardURL)
	},
}

var uiStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show UI component status",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check API
		apiStatus := checkEndpoint(apiURL + "/health")
		fmt.Printf("Phoenix API:       %s\n", formatStatus(apiStatus))

		// Check WebSocket
		wsStatus := checkWebSocket(apiURL)
		fmt.Printf("WebSocket Hub:     %s\n", formatStatus(wsStatus))

		// Check Dashboard
		dashboardURL := "http://localhost:3000"
		dashboardStatus := checkEndpoint(dashboardURL)
		fmt.Printf("Phoenix Dashboard: %s\n", formatStatus(dashboardStatus))

		// Check fleet
		fleet, err := apiClient.GetFleetStatus()
		if err == nil {
			fmt.Printf("\nFleet Status:\n")
			fmt.Printf("  Total Agents:   %d\n", fleet.TotalAgents)
			fmt.Printf("  Healthy:        %d\n", fleet.HealthyAgents)
			fmt.Printf("  Offline:        %d\n", fleet.OfflineAgents)
			fmt.Printf("  Total Savings:  ₹%.2f/hour\n", fleet.TotalSavings)
		}

		return nil
	},
}

var uiCostCmd = &cobra.Command{
	Use:   "cost",
	Short: "Show real-time cost flow",
	RunE: func(cmd *cobra.Command, args []string) error {
		costFlow, err := apiClient.GetMetricCostFlow()
		if err != nil {
			return fmt.Errorf("failed to get cost flow: %w", err)
		}

		fmt.Printf("Total Cost Rate: ₹%.2f/minute\n\n", costFlow.TotalCostRate)
		fmt.Println("Top Cost Drivers:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("%-40s %12s %10s %8s\n", "Metric", "Cost/min", "Cardinality", "Percent")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		for i, metric := range costFlow.TopMetrics {
			if i >= 10 {
				break
			}
			name := metric.MetricName
			if len(name) > 40 {
				name = name[:37] + "..."
			}
			fmt.Printf("%-40s ₹%10.2f %10d %7.1f%%\n",
				name,
				metric.CostPerMinute,
				metric.Cardinality,
				metric.Percentage,
			)
		}

		return nil
	},
}

var uiWizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "Launch interactive experiment wizard",
	Long:  `Launch an interactive wizard to create experiments without writing YAML`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Phoenix Experiment Wizard")
		fmt.Println("========================")

		// Interactive prompts
		var name, description string
		var hosts []string
		var pipelineType string
		var duration int

		fmt.Print("\nExperiment name: ")
		fmt.Scanln(&name)

		fmt.Print("Description (optional): ")
		fmt.Scanln(&description)

		// Show available host groups
		fleet, _ := apiClient.GetFleetStatus()
		if fleet != nil {
			fmt.Println("\nAvailable host groups:")
			groups := make(map[string]int)
			for _, agent := range fleet.Agents {
				groups[agent.Group]++
			}
			for group, count := range groups {
				fmt.Printf("  - %s (%d hosts)\n", group, count)
			}
		}

		fmt.Print("\nSelect hosts (comma-separated groups or 'all'): ")
		var hostInput string
		fmt.Scanln(&hostInput)
		if hostInput == "all" {
			hosts = []string{"all"}
		} else {
			// Parse comma-separated groups
			hosts = parseCommaSeparated(hostInput)
		}

		// Show pipeline templates
		templates, _ := apiClient.GetPipelineTemplates()
		if templates != nil {
			fmt.Println("\nAvailable pipeline templates:")
			for _, tmpl := range templates {
				fmt.Printf("  - %s: %s (-%d%% cost)\n",
					tmpl.ID,
					tmpl.Name,
					tmpl.EstimatedSavings,
				)
			}
		}

		fmt.Print("\nSelect pipeline template: ")
		fmt.Scanln(&pipelineType)

		fmt.Print("Duration (hours) [24]: ")
		fmt.Scanln(&duration)
		if duration == 0 {
			duration = 24
		}

		// Create experiment
		wizardData := map[string]interface{}{
			"name":           name,
			"description":    description,
			"host_selector":  hosts,
			"pipeline_type":  pipelineType,
			"duration_hours": duration,
		}

		experiment, err := apiClient.CreateExperimentWizard(wizardData)
		if err != nil {
			return fmt.Errorf("failed to create experiment: %w", err)
		}

		fmt.Printf("\n✅ Experiment created successfully!\n")
		fmt.Printf("ID: %s\n", experiment.ID)
		fmt.Printf("Estimated savings: %d%%\n", experiment.EstimatedSavings)
		fmt.Printf("\nView in dashboard: http://localhost:3000/experiments/%s\n", experiment.ID)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)
	uiCmd.AddCommand(uiOpenCmd)
	uiCmd.AddCommand(uiStatusCmd)
	uiCmd.AddCommand(uiCostCmd)
	uiCmd.AddCommand(uiWizardCmd)
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		return fmt.Errorf("unsupported platform")
	}

	return exec.Command(cmd, args...).Start()
}

func checkEndpoint(url string) bool {
	resp, err := httpClient.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func checkWebSocket(baseURL string) bool {
	// Simple check - just verify the endpoint exists
	wsURL := baseURL + "/api/v1/ws"
	resp, err := httpClient.Get(wsURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	// WebSocket endpoint typically returns 400 for non-WebSocket requests
	return resp.StatusCode == 400 || resp.StatusCode == 426
}

func formatStatus(ok bool) string {
	if ok {
		return "✅ Running"
	}
	return "❌ Not available"
}

func parseCommaSeparated(input string) []string {
	var result []string
	for _, item := range strings.Split(input, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, "group:"+trimmed)
		}
	}
	return result
}