package cmd

import (
	"fmt"
	"strings"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/config"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	promoteVariant string
	promoteForce   bool
)

// promoteExperimentCmd represents the experiment promote command
var promoteExperimentCmd = &cobra.Command{
	Use:   "promote [ID]",
	Short: "Promote an experiment variant",
	Long: `Promote the winning variant of a completed experiment.

This will apply the chosen pipeline configuration to the target nodes permanently.

Examples:
  # Promote the candidate variant
  phoenix experiment promote exp-123 --variant candidate

  # Promote the baseline (rollback)
  phoenix experiment promote exp-123 --variant baseline

  # Force promotion without confirmation
  phoenix experiment promote exp-123 --variant candidate --force`,
	Args: cobra.ExactArgs(1),
	RunE: runExperimentPromote,
}

func init() {
	experimentCmd.AddCommand(promoteExperimentCmd)

	promoteExperimentCmd.Flags().StringVarP(&promoteVariant, "variant", "v", "", "Variant to promote (baseline or candidate)")
	promoteExperimentCmd.Flags().BoolVarP(&promoteForce, "force", "f", false, "Force promotion without confirmation")

	promoteExperimentCmd.MarkFlagRequired("variant")
}

func runExperimentPromote(cmd *cobra.Command, args []string) error {
	experimentID := args[0]

	// Validate variant
	if promoteVariant != "baseline" && promoteVariant != "candidate" {
		return fmt.Errorf("variant must be 'baseline' or 'candidate', got: %s", promoteVariant)
	}

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

	// Check if experiment can be promoted
	if experiment.Phase != "completed" && experiment.Phase != "running" {
		if !promoteForce {
			return fmt.Errorf("experiment is %s, typically only completed experiments should be promoted", experiment.Phase)
		}
		output.PrintWarning(fmt.Sprintf("Promoting %s experiment - results may be incomplete", experiment.Phase))
	}

	// Show promotion details
	fmt.Printf("Promotion Details:\n")
	fmt.Printf("  Experiment:    %s (%s)\n", experiment.Name, experiment.ID[:8])
	fmt.Printf("  Variant:       %s\n", promoteVariant)

	pipelineName := experiment.BaselinePipeline
	if promoteVariant == "candidate" {
		pipelineName = experiment.CandidatePipeline
	}
	fmt.Printf("  Pipeline:      %s\n", pipelineName)
	fmt.Printf("  Target Nodes:  %s\n", formatTargetNodes(experiment.TargetNodes))

	// Show results if available
	if experiment.Results != nil {
		fmt.Printf("\nExperiment Results:\n")
		fmt.Printf("  Cardinality Reduction: %.1f%%\n", experiment.Results.CardinalityReduction)
		fmt.Printf("  Cost Reduction:        %.1f%%\n", experiment.Results.CostReduction)

		if experiment.Results.Recommendation != "" {
			fmt.Printf("  Recommendation:        %s\n", experiment.Results.Recommendation)
		}
	}

	// Confirm unless force flag is set
	if !promoteForce {
		fmt.Printf("\nThis will permanently apply the %s configuration to all target nodes.\n", promoteVariant)
		fmt.Print("Are you sure? Type 'yes' to confirm: ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "yes" {
			fmt.Println("Promotion cancelled.")
			return nil
		}
	}

	// Promote the variant
	fmt.Printf("\nPromoting %s variant...\n", promoteVariant)
	err = apiClient.PromoteExperiment(experimentID, promoteVariant)
	if err != nil {
		return fmt.Errorf("failed to promote experiment: %w", err)
	}

	output.PrintSuccess("Variant promoted successfully!")

	// Show next steps
	fmt.Printf("\nThe %s pipeline configuration has been applied to:\n", promoteVariant)
	for k, v := range experiment.TargetNodes {
		fmt.Printf("  • %s=%s\n", k, v)
	}

	if promoteVariant == "candidate" && experiment.Results != nil {
		fmt.Printf("\nExpected benefits:\n")
		fmt.Printf("  • %.1f%% reduction in metrics cardinality\n", experiment.Results.CardinalityReduction)
		fmt.Printf("  • %.1f%% reduction in observability costs\n", experiment.Results.CostReduction)
	}

	fmt.Printf("\nTo deploy this configuration more broadly:\n")
	fmt.Printf("  phoenix pipeline deploy --name %s --selector <broader-selector>\n", pipelineName)

	return nil
}
