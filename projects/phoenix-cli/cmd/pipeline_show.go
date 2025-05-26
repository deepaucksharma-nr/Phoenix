package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
)

// pipelineShowCmd represents the pipeline show command
var pipelineShowCmd = &cobra.Command{
	Use:   "show <pipeline-name>",
	Short: "Display the YAML configuration of a catalog pipeline",
	Long: `Display the YAML configuration of a pipeline from the catalog.

This command reads and displays the OTel collector configuration
for the specified pipeline template.

Examples:
  # Show the process-baseline-v1 pipeline
  phoenix pipeline show process-baseline-v1

  # Show the process-topk-v1 pipeline
  phoenix pipeline show process-topk-v1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pipelineName := args[0]
		
		// Try multiple possible locations
		possiblePaths := []string{
			filepath.Join(".", "configs", "pipelines", "catalog", "process", pipelineName+".yaml"),
			filepath.Join("..", "configs", "pipelines", "catalog", "process", pipelineName+".yaml"),
			filepath.Join("..", "..", "configs", "pipelines", "catalog", "process", pipelineName+".yaml"),
			filepath.Join("/Users/deepaksharma/Desktop/src/Phoenix", "configs", "pipelines", "catalog", "process", pipelineName+".yaml"),
		}

		var content []byte
		var err error
		var foundPath string

		// Try to read from local files first
		for _, path := range possiblePaths {
			content, err = os.ReadFile(path)
			if err == nil {
				foundPath = path
				break
			}
		}

		// If not found locally, try to fetch from API
		if foundPath == "" {
			// TODO: Implement API endpoint to fetch pipeline configs
			// For now, we'll show an error
			return fmt.Errorf("pipeline '%s' not found in catalog. Available pipelines:\n"+
				"  - process-baseline-v1\n"+
				"  - process-sampling-v1\n"+
				"  - process-topk-v1\n"+
				"  - process-adaptive-filter-v1\n"+
				"  - process-anomaly-v1", pipelineName)
		}

		// Display the pipeline configuration
		output.Success(fmt.Sprintf("Pipeline: %s", pipelineName))
		fmt.Printf("Path: %s\n\n", foundPath)
		
		// Print the YAML content
		fmt.Println(string(content))

		return nil
	},
}

// fetchPipelineFromAPI attempts to fetch pipeline config from API
func fetchPipelineFromAPI(pipelineName string) ([]byte, error) {
	// TODO: Implement when API endpoint is available
	url := fmt.Sprintf("http://localhost:8080/api/v1/pipelines/%s/config", pipelineName)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch pipeline: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func init() {
	pipelineCmd.AddCommand(pipelineShowCmd)
}