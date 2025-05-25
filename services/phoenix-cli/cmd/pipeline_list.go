package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/phoenix/platform/services/phoenix-cli/internal/client"
	"github.com/phoenix/platform/services/phoenix-cli/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	pipelineType string
)

// listPipelineCmd represents the pipeline list command
var listPipelineCmd = &cobra.Command{
	Use:   "list",
	Short: "List available pipeline templates",
	Long: `List all available pipeline templates in the Phoenix platform.

Examples:
  # List all pipelines
  phoenix pipeline list

  # List only optimization pipelines
  phoenix pipeline list --type optimization

  # List as JSON
  phoenix pipeline list -o json`,
	RunE: runPipelineList,
}

func init() {
	pipelineCmd.AddCommand(listPipelineCmd)

	listPipelineCmd.Flags().StringVar(&pipelineType, "type", "", "Filter by pipeline type (baseline, optimization)")
}

func runPipelineList(cmd *cobra.Command, args []string) error {
	// Get config and check authentication
	cfg := config.New()
	token := cfg.GetToken()
	if token == "" {
		return fmt.Errorf("not authenticated. Please run: phoenix auth login")
	}

	// Create API client
	apiClient := client.NewAPIClient(cfg.GetAPIEndpoint(), token)

	// List pipelines
	pipelines, err := apiClient.ListPipelines()
	if err != nil {
		return fmt.Errorf("failed to list pipelines: %w", err)
	}

	// Filter by type if specified
	if pipelineType != "" {
		filtered := []client.Pipeline{}
		for _, p := range pipelines {
			if p.Type == pipelineType {
				filtered = append(filtered, p)
			}
		}
		pipelines = filtered
	}

	// Display results
	switch outputFormat {
	case "json":
		data, _ := json.MarshalIndent(pipelines, "", "  ")
		fmt.Println(string(data))
	case "yaml":
		data, _ := yaml.Marshal(pipelines)
		fmt.Print(string(data))
	default:
		// Table format
		if len(pipelines) == 0 {
			fmt.Println("No pipelines found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tDESCRIPTION\tPARAMETERS")
		
		for _, p := range pipelines {
			params := "none"
			if len(p.Parameters) > 0 {
				paramList := []string{}
				for k := range p.Parameters {
					paramList = append(paramList, k)
				}
				params = fmt.Sprintf("%v", paramList)
			}
			
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				p.Name,
				p.Type,
				truncate(p.Description, 40),
				params,
			)
		}
		w.Flush()

		fmt.Printf("\nTotal: %d pipelines\n", len(pipelines))
	}

	return nil
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}