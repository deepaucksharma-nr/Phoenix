package cmd

import (
	"fmt"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/config"
	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from Phoenix Platform",
	Long:  `Log out from Phoenix Platform by removing the stored authentication token.`,
	RunE:  runLogout,
}

func init() {
	authCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	cfg := config.New()
	
	// Check if we're logged in
	if cfg.GetToken() == "" {
		fmt.Println("You are not logged in.")
		return nil
	}

	// Clear the token
	if err := cfg.ClearToken(); err != nil {
		return fmt.Errorf("failed to clear authentication token: %w", err)
	}

	fmt.Println("✓ Successfully logged out!")
	return nil
}