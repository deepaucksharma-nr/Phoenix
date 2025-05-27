package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var pipelineListVersionsCmd = &cobra.Command{
	Use:   "list-versions [deployment-id]",
	Short: "List all versions of a pipeline deployment",
	Long: `List all versions of a pipeline deployment with their metadata.
	
This command shows the version history of a deployment, including:
- Version number
- Deployment timestamp
- Deployed by
- Notes/description
- Parameters used`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deploymentID := args[0]

		// Get API client
		client, err := getAPIClient()
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		// List deployment versions
		versions, err := client.ListPipelineDeploymentVersions(deploymentID)
		if err != nil {
			return fmt.Errorf("failed to list deployment versions: %w", err)
		}

		// Display versions
		if outputFormat == "json" {
			data, _ := json.MarshalIndent(versions, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(versions) == 0 {
			fmt.Println("No versions found for deployment:", deploymentID)
			return nil
		}

		// Table output
		fmt.Printf("Deployment ID: %s\n", deploymentID)
		fmt.Println("\nVersions:")
		fmt.Printf("%-8s %-20s %-15s %-50s\n", "VERSION", "DEPLOYED AT", "DEPLOYED BY", "NOTES")
		fmt.Println(strings.Repeat("-", 95))

		for _, v := range versions {
			version := v.(map[string]interface{})
			
			versionNum := fmt.Sprintf("%v", version["version"])
			deployedAt := ""
			if ts, ok := version["deployed_at"].(string); ok {
				if t, err := time.Parse(time.RFC3339, ts); err == nil {
					deployedAt = t.Format("2006-01-02 15:04:05")
				} else {
					deployedAt = ts
				}
			}
			deployedBy := fmt.Sprintf("%v", version["deployed_by"])
			notes := fmt.Sprintf("%v", version["notes"])
			
			// Truncate notes if too long
			if len(notes) > 47 {
				notes = notes[:47] + "..."
			}

			fmt.Printf("%-8s %-20s %-15s %-50s\n",
				versionNum,
				deployedAt,
				deployedBy,
				notes,
			)
		}

		return nil
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineListVersionsCmd)
}