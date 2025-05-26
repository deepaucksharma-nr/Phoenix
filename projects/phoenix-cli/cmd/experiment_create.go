package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/config"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	expName           string
	expDescription    string
	baselinePipeline  string
	candidatePipeline string
	targetSelector    map[string]string
	duration          time.Duration
	criticalProcesses []string
	topK              int
	checkOverlap      bool
	force             bool
)

// createExperimentCmd represents the create experiment command
var createExperimentCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new experiment",
	Long: `Create a new A/B experiment to test pipeline optimizations.

Examples:
  # Create a simple experiment
  phoenix experiment create --name "reduce-cardinality" \
    --baseline process-baseline-v1 \
    --candidate process-topk-v1 \
    --target-selector "app=webserver"

  # Create experiment with critical processes
  phoenix experiment create --name "priority-filter-test" \
    --baseline process-baseline-v1 \
    --candidate process-priority-filter-v1 \
    --target-selector "environment=production" \
    --critical-processes "nginx,postgres,redis"

  # Check for overlaps before creating
  phoenix experiment create --name "test-optimization" \
    --baseline process-baseline-v1 \
    --candidate process-adaptive-v1 \
    --target-selector "tier=frontend" \
    --check-overlap`,
	RunE: runCreateExperiment,
}

func init() {
	experimentCmd.AddCommand(createExperimentCmd)

	// Required flags
	createExperimentCmd.Flags().StringVarP(&expName, "name", "n", "", "Experiment name (required)")
	createExperimentCmd.Flags().StringVar(&baselinePipeline, "baseline", "", "Baseline pipeline template (required)")
	createExperimentCmd.Flags().StringVar(&candidatePipeline, "candidate", "", "Candidate pipeline template (required)")
	createExperimentCmd.Flags().StringToStringVar(&targetSelector, "target-selector", nil, "Target node selector labels (required)")
	
	createExperimentCmd.MarkFlagRequired("name")
	createExperimentCmd.MarkFlagRequired("baseline")
	createExperimentCmd.MarkFlagRequired("candidate")
	createExperimentCmd.MarkFlagRequired("target-selector")

	// Optional flags
	createExperimentCmd.Flags().StringVarP(&expDescription, "description", "d", "", "Experiment description")
	createExperimentCmd.Flags().DurationVar(&duration, "duration", 1*time.Hour, "Experiment duration")
	createExperimentCmd.Flags().StringSliceVar(&criticalProcesses, "critical-processes", nil, "List of critical processes to monitor")
	createExperimentCmd.Flags().IntVar(&topK, "top-k", 10, "Number of top processes to keep (for topk pipeline)")
	createExperimentCmd.Flags().BoolVar(&checkOverlap, "check-overlap", false, "Check for overlapping experiments")
	createExperimentCmd.Flags().BoolVarP(&force, "force", "f", false, "Force creation even with warnings")
}

func runCreateExperiment(cmd *cobra.Command, args []string) error {
	// Get config and check authentication
	cfg := config.New()
	token := cfg.GetToken()
	if token == "" {
		return fmt.Errorf("not authenticated. Please run: phoenix auth login")
	}

	// Create API client
	apiClient := client.NewAPIClient(cfg.GetAPIEndpoint(), token)

	// Prepare experiment request
	req := client.CreateExperimentRequest{
		Name:              expName,
		Description:       expDescription,
		BaselinePipeline:  baselinePipeline,
		CandidatePipeline: candidatePipeline,
		TargetNodes:       targetSelector,
		Duration:          duration,
		Parameters:        make(map[string]interface{}),
	}

	// Add pipeline-specific parameters
	if len(criticalProcesses) > 0 {
		req.Parameters["critical_processes"] = criticalProcesses
	}
	if strings.Contains(candidatePipeline, "topk") && cmd.Flags().Changed("top-k") {
		req.Parameters["top_k"] = topK
	}

	// Check for overlaps if requested
	if checkOverlap {
		fmt.Println("Checking for experiment overlaps...")
		overlap, err := apiClient.CheckExperimentOverlap(req)
		if err != nil {
			return fmt.Errorf("failed to check overlap: %w", err)
		}

		if overlap.HasOverlap {
			output.PrintOverlapWarning(overlap)
			
			if overlap.Severity == "blocking" {
				return fmt.Errorf("cannot create experiment due to blocking overlap")
			}

			if !force {
				fmt.Print("\nDo you want to proceed anyway? [y/N]: ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" {
					fmt.Println("Experiment creation cancelled.")
					return nil
				}
			}
		} else {
			fmt.Println("✓ No overlapping experiments found")
		}
	}

	// Create experiment
	fmt.Printf("Creating experiment '%s'...\n", expName)
	experiment, err := apiClient.CreateExperiment(req)
	if err != nil {
		return fmt.Errorf("failed to create experiment: %w", err)
	}

	// Display result
	output.PrintExperiment(experiment, outputFormat)
	
	fmt.Println("\n✓ Experiment created successfully!")
	fmt.Printf("\nTo start the experiment, run:\n  phoenix experiment start %s\n", experiment.ID)
	fmt.Printf("\nTo monitor status, run:\n  phoenix experiment status %s --follow\n", experiment.ID)

	return nil
}