package cmd

import (
	"github.com/spf13/cobra"
)

// experimentCmd represents the experiment command group
var experimentCmd = &cobra.Command{
	Use:     "experiment",
	Aliases: []string{"exp"},
	Short:   "Manage experiments",
	Long: `Manage Phoenix Platform experiments.

This includes creating, listing, monitoring, and controlling experiments.`,
}

func init() {
	rootCmd.AddCommand(experimentCmd)
}