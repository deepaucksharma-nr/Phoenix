package cmd

import (
	"github.com/spf13/cobra"
)

// loadsimCmd represents the loadsim command
var loadsimCmd = &cobra.Command{
	Use:   "loadsim",
	Short: "Manage load simulations for experiments",
	Long: `Manage load simulations to generate process metrics for Phoenix experiments.

Load simulations create realistic process patterns to test different pipeline
configurations and measure their effectiveness at reducing metrics cardinality.`,
}

func init() {
	rootCmd.AddCommand(loadsimCmd)
}