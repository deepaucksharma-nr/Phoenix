package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	vmConfigOutput     string
	vmExporterEndpoint string
)

// pipelineVMConfigCmd generates a static collector config for VM hosts
var pipelineVMConfigCmd = &cobra.Command{
	Use:   "vm-config <pipeline-name>",
	Short: "Generate static collector config for a VM",
	Long: `Generate a static OpenTelemetry Collector configuration for running on a virtual machine.

The command reads a pipeline template from the local catalog and writes a single YAML
configuration that can be used directly with the otelcol binary outside Kubernetes.`,
	Args: cobra.ExactArgs(1),
	RunE: runPipelineVMConfig,
}

func init() {
	pipelineCmd.AddCommand(pipelineVMConfigCmd)
	pipelineVMConfigCmd.Flags().StringVarP(&vmConfigOutput, "output", "o", "collector.yaml", "Output config file")
	pipelineVMConfigCmd.Flags().StringVar(&vmExporterEndpoint, "exporter-endpoint", "", "Override OTLP exporter endpoint")
}

func runPipelineVMConfig(cmd *cobra.Command, args []string) error {
	pipelineName := args[0]

	possiblePaths := []string{
		filepath.Join(".", "configs", "pipelines", "catalog", "process", pipelineName+".yaml"),
		filepath.Join("..", "configs", "pipelines", "catalog", "process", pipelineName+".yaml"),
		filepath.Join("..", "..", "configs", "pipelines", "catalog", "process", pipelineName+".yaml"),
	}

	var content []byte
	var err error
	for _, path := range possiblePaths {
		content, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("pipeline '%s' not found", pipelineName)
	}

	if vmExporterEndpoint != "" {
		var config map[string]interface{}
		if err := yaml.Unmarshal(content, &config); err != nil {
			return fmt.Errorf("invalid yaml: %w", err)
		}
		if exporters, ok := config["exporters"].(map[string]interface{}); ok {
			if otlp, ok := exporters["otlp"].(map[string]interface{}); ok {
				otlp["endpoint"] = vmExporterEndpoint
				exporters["otlp"] = otlp
			}
			config["exporters"] = exporters
		}
		if modified, err := yaml.Marshal(config); err == nil {
			content = modified
		} else {
			return fmt.Errorf("failed to marshal updated config: %w", err)
		}
	}

	if err := os.WriteFile(vmConfigOutput, content, 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	output.Success(fmt.Sprintf("Collector configuration written to: %s", vmConfigOutput))
	return nil
}
