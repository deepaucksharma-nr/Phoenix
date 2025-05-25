package cmd

import (
	"github.com/spf13/cobra"
)

// pipelineCmd represents the pipeline command group
var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Manage pipeline configurations",
	Long: `Manage Phoenix pipeline configurations and deployments.

This includes:
  - Listing available pipeline templates
  - Deploying pipelines directly (without experiments)
  - Managing pipeline deployments
  - Validating pipeline configurations`,
}

func init() {
	rootCmd.AddCommand(pipelineCmd)
}