package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/phoenix-vnext/platform/cmd/phoenix-cli/internal/client"
	"github.com/phoenix-vnext/platform/cmd/phoenix-cli/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	metricsTimeRange string
	metricsRaw       bool
)

// metricsExperimentCmd represents the experiment metrics command
var metricsExperimentCmd = &cobra.Command{
	Use:   "metrics [ID]",
	Short: "View experiment metrics",
	Long: `View detailed metrics for an experiment.

This shows time-series data for both baseline and candidate pipelines,
including cardinality, resource usage, and error rates.

Examples:
  # View current metrics
  phoenix experiment metrics exp-123

  # View metrics for last hour
  phoenix experiment metrics exp-123 --range 1h

  # Export raw metrics data as JSON
  phoenix experiment metrics exp-123 --raw -o json`,
	Args: cobra.ExactArgs(1),
	RunE: runExperimentMetrics,
}

func init() {
	experimentCmd.AddCommand(metricsExperimentCmd)

	metricsExperimentCmd.Flags().StringVar(&metricsTimeRange, "range", "30m", "Time range for metrics (e.g., 1h, 30m, 24h)")
	metricsExperimentCmd.Flags().BoolVar(&metricsRaw, "raw", false, "Show raw metrics data")
}

func runExperimentMetrics(cmd *cobra.Command, args []string) error {
	experimentID := args[0]

	// Get config and check authentication
	cfg := config.New()
	token := cfg.GetToken()
	if token == "" {
		return fmt.Errorf("not authenticated. Please run: phoenix auth login")
	}

	// Create API client
	apiClient := client.NewAPIClient(cfg.GetAPIEndpoint(), token)

	// Get experiment details
	experiment, err := apiClient.GetExperiment(experimentID)
	if err != nil {
		return fmt.Errorf("failed to get experiment: %w", err)
	}

	// Get metrics
	metrics, err := apiClient.GetExperimentMetrics(experimentID)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	if metricsRaw || outputFormat == "json" || outputFormat == "yaml" {
		// Output raw metrics data
		printMetricsRaw(metrics)
		return nil
	}

	// Display formatted metrics
	fmt.Printf("Experiment: %s (%s)\n", experiment.Name, experiment.ID[:8])
	fmt.Printf("Status:     %s\n", experiment.Status)
	fmt.Printf("Time Range: %s\n\n", metricsTimeRange)

	// Display summary
	if experiment.Results != nil {
		fmt.Println("Summary Metrics:")
		fmt.Println("================")
		fmt.Printf("Cardinality Reduction: %.1f%%\n", experiment.Results.CardinalityReduction)
		fmt.Printf("Cost Reduction:        %.1f%%\n", experiment.Results.CostReduction)
		fmt.Printf("\n")
	}

	// Display detailed metrics
	fmt.Println("Baseline Pipeline:")
	fmt.Println("==================")
	displayMetricsSummary(metrics.Baseline)

	fmt.Println("\nCandidate Pipeline:")
	fmt.Println("===================")
	displayMetricsSummary(metrics.Candidate)

	// Display comparison
	if len(metrics.Baseline.Cardinality) > 0 && len(metrics.Candidate.Cardinality) > 0 {
		fmt.Println("\nComparison:")
		fmt.Println("============")
		
		baselineLatest := metrics.Baseline.Cardinality[len(metrics.Baseline.Cardinality)-1].Value
		candidateLatest := metrics.Candidate.Cardinality[len(metrics.Candidate.Cardinality)-1].Value
		reduction := (baselineLatest - candidateLatest) / baselineLatest * 100
		
		fmt.Printf("Current Cardinality Reduction: %.1f%%\n", reduction)
		
		// Resource usage comparison
		if len(metrics.Baseline.CPUUsage) > 0 && len(metrics.Candidate.CPUUsage) > 0 {
			baselineCPU := metrics.Baseline.CPUUsage[len(metrics.Baseline.CPUUsage)-1].Value
			candidateCPU := metrics.Candidate.CPUUsage[len(metrics.Candidate.CPUUsage)-1].Value
			cpuDiff := candidateCPU - baselineCPU
			
			if cpuDiff > 0 {
				fmt.Printf("CPU Overhead: +%.1f%%\n", cpuDiff)
			} else {
				fmt.Printf("CPU Savings: %.1f%%\n", -cpuDiff)
			}
		}
	}

	// Show recommendation
	if experiment.Results != nil && experiment.Results.Recommendation != "" {
		fmt.Printf("\nRecommendation: %s\n", experiment.Results.Recommendation)
	}

	return nil
}

func displayMetricsSummary(data client.TimeSeriesData) {
	if len(data.Cardinality) > 0 {
		latest := data.Cardinality[len(data.Cardinality)-1]
		fmt.Printf("  Cardinality:    %.0f metrics\n", latest.Value)
	}
	
	if len(data.CPUUsage) > 0 {
		latest := data.CPUUsage[len(data.CPUUsage)-1]
		fmt.Printf("  CPU Usage:      %.1f%%\n", latest.Value)
	}
	
	if len(data.MemoryUsage) > 0 {
		latest := data.MemoryUsage[len(data.MemoryUsage)-1]
		fmt.Printf("  Memory Usage:   %.1f MB\n", latest.Value)
	}
	
	if len(data.NetworkTraffic) > 0 {
		latest := data.NetworkTraffic[len(data.NetworkTraffic)-1]
		fmt.Printf("  Network:        %.1f KB/s\n", latest.Value)
	}
}

func printMetricsRaw(metrics *client.ExperimentMetrics) {
	switch outputFormat {
	case "json":
		data, _ := json.MarshalIndent(metrics, "", "  ")
		fmt.Println(string(data))
	case "yaml":
		data, _ := yaml.Marshal(metrics)
		fmt.Print(string(data))
	default:
		// Table format for raw data
		fmt.Println("Baseline Metrics:")
		printTimeSeriesTable(metrics.Baseline)
		fmt.Println("\nCandidate Metrics:")
		printTimeSeriesTable(metrics.Candidate)
	}
}

func printTimeSeriesTable(data client.TimeSeriesData) {
	if len(data.Cardinality) == 0 {
		fmt.Println("  No data available")
		return
	}
	
	fmt.Println("  Time                  Cardinality  CPU%   Memory(MB)  Network(KB/s)")
	fmt.Println("  ==================== ============ ====== =========== =============")
	
	// Show last 10 data points
	start := 0
	if len(data.Cardinality) > 10 {
		start = len(data.Cardinality) - 10
	}
	
	for i := start; i < len(data.Cardinality); i++ {
		point := data.Cardinality[i]
		cpu := "N/A"
		memory := "N/A"
		network := "N/A"
		
		if i < len(data.CPUUsage) {
			cpu = fmt.Sprintf("%.1f", data.CPUUsage[i].Value)
		}
		if i < len(data.MemoryUsage) {
			memory = fmt.Sprintf("%.1f", data.MemoryUsage[i].Value)
		}
		if i < len(data.NetworkTraffic) {
			network = fmt.Sprintf("%.1f", data.NetworkTraffic[i].Value)
		}
		
		fmt.Printf("  %s %12.0f %6s %11s %13s\n",
			point.Time.Format("2006-01-02 15:04:05"),
			point.Value,
			cpu,
			memory,
			network,
		)
	}
}