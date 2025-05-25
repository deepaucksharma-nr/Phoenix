package cmd

import (
	"fmt"

	"github.com/phoenix/platform/services/phoenix-cli/internal/auth"
	"github.com/phoenix/platform/services/phoenix-cli/internal/config"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long:  `Check if you are currently authenticated with the Phoenix Platform.`,
	RunE:  runAuthStatus,
}

func init() {
	authCmd.AddCommand(statusCmd)
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	cfg := config.New()
	token := cfg.GetToken()

	if token == "" {
		fmt.Println("Not logged in.")
		fmt.Println("\nTo log in, run: phoenix auth login")
		return nil
	}

	// Verify token with API
	authClient := auth.NewClient(cfg.GetAPIEndpoint())
	userInfo, err := authClient.VerifyToken(token)
	if err != nil {
		fmt.Println("Authentication token is invalid or expired.")
		fmt.Println("\nPlease log in again: phoenix auth login")
		return nil
	}

	fmt.Println("✓ Authenticated")
	fmt.Printf("  User:     %s\n", userInfo.Username)
	fmt.Printf("  Email:    %s\n", userInfo.Email)
	fmt.Printf("  Roles:    %v\n", userInfo.Roles)
	fmt.Printf("  Endpoint: %s\n", cfg.GetAPIEndpoint())

	return nil
}