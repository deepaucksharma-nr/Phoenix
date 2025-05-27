package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/config"
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
		// Get config
		cfg := config.New()
		apiURL := cfg.GetAPIEndpoint()

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

		// Fleet status placeholder
		fmt.Printf("\nFleet Status:\n")
		fmt.Printf("  Feature under development\n")

		return nil
	},
}

var uiCostCmd = &cobra.Command{
	Use:   "cost",
	Short: "Show real-time cost flow",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Cost flow feature is under development")
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
		fmt.Println("\nWizard feature is under development")
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
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}

func checkEndpoint(url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func checkWebSocket(baseURL string) bool {
	// Simple check - just verify the endpoint exists
	wsURL := baseURL + "/api/v1/ws"
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(wsURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	// WebSocket upgrade would return 400, but that's ok - it means endpoint exists
	return resp.StatusCode == 400 || resp.StatusCode == 200
}

func formatStatus(ok bool) string {
	if ok {
		return "✅ Online"
	}
	return "❌ Offline"
}

// Helper function to parse comma-separated strings
func parseCommaSeparated(input string) []string {
	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
