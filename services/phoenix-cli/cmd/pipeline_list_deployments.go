package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/phoenix/platform/services/phoenix-cli/internal/client"
	"github.com/phoenix/platform/services/phoenix-cli/internal/config"
	"github.com/phoenix/platform/services/phoenix-cli/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	listDeployNamespace string
	listDeployStatus    string
)

// listDeploymentsCmd represents the pipeline list-deployments command
var listDeploymentsCmd = &cobra.Command{
	Use:   "list-deployments",
	Short: "List pipeline deployments",
	Long: `List all pipeline deployments.

Examples:
  # List all deployments
  phoenix pipeline list-deployments

  # List deployments in a specific namespace
  phoenix pipeline list-deployments --namespace phoenix-prod

  # List only active deployments
  phoenix pipeline list-deployments --status active

  # Output as JSON
  phoenix pipeline list-deployments -o json`,
	RunE: runListDeployments,
}

func init() {
	pipelineCmd.AddCommand(listDeploymentsCmd)

	listDeploymentsCmd.Flags().StringVar(&listDeployNamespace, "namespace", "", "Filter by namespace")
	listDeploymentsCmd.Flags().StringVar(&listDeployStatus, "status", "", "Filter by status (pending, active, failed)")
}

func runListDeployments(cmd *cobra.Command, args []string) error {
	// Get config and check authentication
	cfg := config.New()
	token := cfg.GetToken()
	if token == "" {
		return fmt.Errorf("not authenticated. Please run: phoenix auth login")
	}

	// Create API client
	apiClient := client.NewAPIClient(cfg.GetAPIEndpoint(), token)

	// Prepare request
	req := client.ListPipelineDeploymentsRequest{
		Namespace: listDeployNamespace,
		Status:    listDeployStatus,
	}

	// List deployments
	deployments, err := apiClient.ListPipelineDeployments(req)
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	// Display results
	switch outputFormat {
	case "json":
		data, _ := json.MarshalIndent(deployments, "", "  ")
		fmt.Println(string(data))
	case "yaml":
		data, _ := yaml.Marshal(deployments)
		fmt.Print(string(data))
	default:
		// Table format
		if len(deployments) == 0 {
			fmt.Println("No deployments found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tPIPELINE\tNAMESPACE\tSTATUS\tINSTANCES\tCREATED")
		
		for _, d := range deployments {
			instances := "N/A"
			if d.Instances != nil {
				instances = fmt.Sprintf("%d/%d", d.Instances.Ready, d.Instances.Desired)
			}
			
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				d.ID[:8], // Short ID
				truncate(d.DeploymentName, 20),
				truncate(d.PipelineName, 20),
				d.Namespace,
				output.ColorizeStatus(d.Status),
				instances,
				d.CreatedAt.Format("2006-01-02 15:04"),
			)
		}
		w.Flush()

		fmt.Printf("\nTotal: %d deployments\n", len(deployments))
	}

	return nil
}