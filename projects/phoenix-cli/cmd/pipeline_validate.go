package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// pipelineValidateCmd represents the pipeline validate command
var pipelineValidateCmd = &cobra.Command{
	Use:   "validate <config-file>",
	Short: "Validate a local OTel collector configuration",
	Long: `Validate a local OTel collector configuration file.

This command checks the YAML syntax and validates the configuration
against OTel collector requirements. It ensures all required sections
are present and properly formatted.

Examples:
  # Validate a local pipeline configuration
  phoenix pipeline validate my-pipeline.yaml

  # Validate a pipeline from the catalog
  phoenix pipeline validate configs/pipelines/catalog/process/process-topk-v1.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile := args[0]

		// Check if file exists
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return fmt.Errorf("configuration file not found: %s", configFile)
		}

		// Read the configuration file
		content, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read configuration file: %w", err)
		}

		// Validate YAML syntax
		var config map[string]interface{}
		if err := yaml.Unmarshal(content, &config); err != nil {
			return fmt.Errorf("invalid YAML syntax: %w", err)
		}

		output.Info(fmt.Sprintf("Validating pipeline configuration: %s", configFile))

		// Perform validation checks
		var errors []string
		var warnings []string

		// Check required top-level sections
		requiredSections := []string{"receivers", "processors", "exporters", "service"}
		for _, section := range requiredSections {
			if _, ok := config[section]; !ok {
				errors = append(errors, fmt.Sprintf("Missing required section: %s", section))
			}
		}

		// Validate service pipelines
		if service, ok := config["service"].(map[string]interface{}); ok {
			if pipelines, ok := service["pipelines"].(map[string]interface{}); ok {
				// Check if metrics pipeline exists (required for process metrics)
				if _, ok := pipelines["metrics"]; !ok {
					errors = append(errors, "Missing 'metrics' pipeline in service.pipelines")
				} else {
					// Validate metrics pipeline structure
					if metricsPipeline, ok := pipelines["metrics"].(map[string]interface{}); ok {
						validatePipeline("metrics", metricsPipeline, &errors, &warnings)
					}
				}
			} else {
				errors = append(errors, "Missing 'pipelines' in service section")
			}
		}

		// Validate receivers
		if receivers, ok := config["receivers"].(map[string]interface{}); ok {
			if len(receivers) == 0 {
				errors = append(errors, "No receivers defined")
			}
			// Check for hostmetrics receiver (required for process metrics)
			if _, ok := receivers["hostmetrics"]; !ok {
				warnings = append(warnings, "Missing 'hostmetrics' receiver - required for process metrics collection")
			}
		}

		// Validate processors
		if processors, ok := config["processors"].(map[string]interface{}); ok {
			// Check for memory_limiter (recommended)
			if _, ok := processors["memory_limiter"]; !ok {
				warnings = append(warnings, "Missing 'memory_limiter' processor - recommended for production")
			}
			// Check for Phoenix-specific processors
			phoenixProcessors := []string{"phoenix/topk", "phoenix/adaptive_filter", "phoenix/sampling"}
			hasPhoenixProcessor := false
			for _, proc := range phoenixProcessors {
				if _, ok := processors[proc]; ok {
					hasPhoenixProcessor = true
					break
				}
			}
			if !hasPhoenixProcessor {
				warnings = append(warnings, "No Phoenix-specific processors found - pipeline may not reduce cardinality")
			}
		}

		// Try to validate with otelcol if available
		if otelcolPath, err := exec.LookPath("otelcol"); err == nil {
			output.Info("Running otelcol validation...")
			validateCmd := exec.Command(otelcolPath, "validate", "--config", configFile)
			var stderr bytes.Buffer
			validateCmd.Stderr = &stderr

			if err := validateCmd.Run(); err != nil {
				errors = append(errors, fmt.Sprintf("OTel collector validation failed: %s", stderr.String()))
			} else {
				output.Success("OTel collector validation passed")
			}
		} else {
			warnings = append(warnings, "otelcol not found in PATH - skipping collector-specific validation")
		}

		// Display results
		if len(errors) > 0 {
			output.Error("Validation failed with errors:")
			for _, err := range errors {
				fmt.Printf("  ❌ %s\n", err)
			}
		}

		if len(warnings) > 0 {
			fmt.Println()
			output.Warning("Validation warnings:")
			for _, warn := range warnings {
				fmt.Printf("  ⚠️  %s\n", warn)
			}
		}

		if len(errors) == 0 {
			fmt.Println()
			output.Success("Pipeline configuration is valid!")

			// Show summary
			fmt.Println("\nConfiguration Summary:")
			if receivers, ok := config["receivers"].(map[string]interface{}); ok {
				fmt.Printf("  Receivers: %s\n", strings.Join(getKeys(receivers), ", "))
			}
			if processors, ok := config["processors"].(map[string]interface{}); ok {
				fmt.Printf("  Processors: %s\n", strings.Join(getKeys(processors), ", "))
			}
			if exporters, ok := config["exporters"].(map[string]interface{}); ok {
				fmt.Printf("  Exporters: %s\n", strings.Join(getKeys(exporters), ", "))
			}
		}

		if len(errors) > 0 {
			return fmt.Errorf("validation failed with %d error(s)", len(errors))
		}

		return nil
	},
}

func validatePipeline(name string, pipeline map[string]interface{}, errors *[]string, warnings *[]string) {
	// Check for required pipeline components
	components := []string{"receivers", "processors", "exporters"}
	for _, component := range components {
		if val, ok := pipeline[component]; ok {
			// Check if it's a non-empty array
			switch v := val.(type) {
			case []interface{}:
				if len(v) == 0 {
					*warnings = append(*warnings, fmt.Sprintf("Pipeline '%s' has empty %s list", name, component))
				}
			case nil:
				*errors = append(*errors, fmt.Sprintf("Pipeline '%s' has null %s", name, component))
			}
		} else {
			*errors = append(*errors, fmt.Sprintf("Pipeline '%s' missing %s", name, component))
		}
	}
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func init() {
	pipelineCmd.AddCommand(pipelineValidateCmd)
}
