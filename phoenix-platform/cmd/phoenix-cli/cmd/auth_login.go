package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/phoenix/platform/cmd/phoenix-cli/internal/auth"
	"github.com/phoenix/platform/cmd/phoenix-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	username string
	password string
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Phoenix Platform",
	Long: `Log in to Phoenix Platform using your credentials.

If username and password are not provided via flags, you will be prompted to enter them.
The authentication token will be stored securely in your local configuration.`,
	RunE: runLogin,
}

func init() {
	authCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVarP(&username, "username", "u", "", "Username for authentication")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Password for authentication (use with caution)")
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Get username if not provided
	if username == "" {
		fmt.Print("Username: ")
		fmt.Scanln(&username)
		if username == "" {
			return fmt.Errorf("username is required")
		}
	}

	// Get password if not provided
	if password == "" {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println() // New line after password input
		password = string(passwordBytes)
		if password == "" {
			return fmt.Errorf("password is required")
		}
	}

	// Create auth client
	authClient := auth.NewClient(apiEndpoint)

	// Attempt login
	fmt.Println("Logging in to Phoenix Platform...")
	token, err := authClient.Login(username, password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Store token in config
	cfg := config.New()
	if err := cfg.SetToken(token); err != nil {
		return fmt.Errorf("failed to store authentication token: %w", err)
	}

	// Also store the API endpoint if it was provided
	if cmd.Flags().Changed("api-endpoint") {
		if err := cfg.SetAPIEndpoint(apiEndpoint); err != nil {
			return fmt.Errorf("failed to store API endpoint: %w", err)
		}
	}

	fmt.Println("✓ Successfully logged in!")
	fmt.Printf("Authentication token stored in: %s\n", cfg.GetConfigPath())

	return nil
}