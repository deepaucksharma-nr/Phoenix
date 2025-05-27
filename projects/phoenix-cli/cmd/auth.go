package cmd

import (
	"github.com/spf13/cobra"
)

// authCmd represents the auth command group
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long: `Manage authentication with the Phoenix Platform.

This includes logging in, logging out, and checking authentication status.`,
}

func init() {
	rootCmd.AddCommand(authCmd)
}
