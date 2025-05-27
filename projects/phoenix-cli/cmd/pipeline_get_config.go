package cmd

import (
	"fmt"
	"os"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/config"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	getConfigOutput string
)

// pipelineGetConfigCmd represents the pipeline get-config command
var pipelineGetConfigCmd = &cobra.Command{
	Use:   "get-config <deployment-id>",
	Short: "Retrieve running configuration from a deployment",
	Long: `Retrieve the active OTel collector configuration from a running deployment.

This command fetches the actual configuration being used by the deployed
collectors, which may include runtime modifications or merged configurations.

Examples:
  # Get configuration from a deployment
  phoenix pipeline get-config deploy-abc123

  # Save configuration to a file
  phoenix pipeline get-config deploy-abc123 --output my-config.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deploymentID := args[0]

		// Get API client configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.Token == "" {
			return fmt.Errorf("not authenticated. Please run 'phoenix auth login' first")
		}

		// Create API client
		apiClient := client.NewAPIClient(cfg.APIEndpoint, cfg.Token)

		// Get deployment to verify it exists
		deployment, err := apiClient.GetPipelineDeployment(deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get deployment: %w", err)
		}

		output.Success(fmt.Sprintf("Retrieving configuration for deployment: %s", deployment.Name))
		fmt.Printf("Pipeline: %s\n", deployment.Pipeline)
		fmt.Printf("Namespace: %s\n\n", deployment.Namespace)

		// Get the active configuration
		configYAML, err := apiClient.GetPipelineConfig(deploymentID)
		if err != nil {
			return fmt.Errorf("failed to retrieve configuration: %w", err)
		}

		// If output file specified, write to file
		if getConfigOutput != "" {
			if err := writeToFile(getConfigOutput, configYAML); err != nil {
				return fmt.Errorf("failed to write configuration to file: %w", err)
			}
			output.Success(fmt.Sprintf("Configuration saved to: %s", getConfigOutput))
		} else {
			// Otherwise, print to stdout
			fmt.Println("=== Active Configuration ===")
			fmt.Println(configYAML)
		}

		return nil
	},
}

// writeToFile writes content to a file
func writeToFile(filename string, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

func init() {
	pipelineCmd.AddCommand(pipelineGetConfigCmd)

	pipelineGetConfigCmd.Flags().StringVarP(&getConfigOutput, "output", "o", "",
		"Output file to save the configuration")
}
