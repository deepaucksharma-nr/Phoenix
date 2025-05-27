package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/client"
	"gopkg.in/yaml.v3"
)

// PrintExperiment prints a single experiment in the specified format
func PrintExperiment(exp *client.Experiment, format string) {
	switch format {
	case "json":
		data, _ := json.MarshalIndent(exp, "", "  ")
		fmt.Println(string(data))
	case "yaml":
		data, _ := yaml.Marshal(exp)
		fmt.Print(string(data))
	default:
		// Table format
		fmt.Printf("ID:          %s\n", exp.ID)
		fmt.Printf("Name:        %s\n", exp.Name)
		fmt.Printf("Status:      %s\n", ColorizeStatus(exp.Status))
		fmt.Printf("Description: %s\n", exp.Description)
		fmt.Printf("Baseline:    %s\n", exp.BaselinePipeline)
		fmt.Printf("Candidate:   %s\n", exp.CandidatePipeline)
		fmt.Printf("Target:      %s\n", formatTargetNodes(exp.TargetNodes))
		fmt.Printf("Created:     %s\n", exp.CreatedAt.Format(time.RFC3339))

		if exp.StartedAt != nil {
			fmt.Printf("Started:     %s\n", exp.StartedAt.Format(time.RFC3339))
		}
		if exp.CompletedAt != nil {
			fmt.Printf("Completed:   %s\n", exp.CompletedAt.Format(time.RFC3339))
		}

		if exp.Results != nil {
			fmt.Printf("\nResults:\n")
			fmt.Printf("  Cardinality Reduction: %.1f%%\n", exp.Results.CardinalityReduction)
			fmt.Printf("  Cost Reduction:        %.1f%%\n", exp.Results.CostReduction)
			fmt.Printf("  Recommendation:        %s\n", exp.Results.Recommendation)
		}
	}
}

// PrintExperimentList prints a list of experiments in the specified format
func PrintExperimentList(experiments []client.Experiment, format string) {
	switch format {
	case "json":
		data, _ := json.MarshalIndent(experiments, "", "  ")
		fmt.Println(string(data))
	case "yaml":
		data, _ := yaml.Marshal(experiments)
		fmt.Print(string(data))
	default:
		// Table format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tBASELINE\tCANDIDATE\tCREATED")

		for _, exp := range experiments {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				exp.ID[:8], // Short ID
				truncate(exp.Name, 30),
				ColorizeStatus(exp.Status),
				truncate(exp.BaselinePipeline, 20),
				truncate(exp.CandidatePipeline, 20),
				exp.CreatedAt.Format("2006-01-02 15:04"),
			)
		}
		w.Flush()
	}
}

// PrintOverlapWarning prints an overlap warning
func PrintOverlapWarning(overlap *client.OverlapResult) {
	severity := strings.ToUpper(overlap.Severity)

	fmt.Printf("\n⚠️  %s: %s\n", severity, overlap.Message)

	if len(overlap.ConflictingExpIDs) > 0 {
		fmt.Printf("\nConflicting experiments:\n")
		for _, id := range overlap.ConflictingExpIDs {
			fmt.Printf("  - %s\n", id)
		}
	}

	if len(overlap.AffectedNodes) > 0 {
		fmt.Printf("\nAffected nodes (%d):\n", len(overlap.AffectedNodes))
		// Show first 5 nodes
		for i, node := range overlap.AffectedNodes {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(overlap.AffectedNodes)-5)
				break
			}
			fmt.Printf("  - %s\n", node)
		}
	}

	if len(overlap.Suggestions) > 0 {
		fmt.Printf("\nSuggestions:\n")
		for _, suggestion := range overlap.Suggestions {
			fmt.Printf("  • %s\n", suggestion)
		}
	}
}

// PrintError prints an error message in a consistent format
func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// Success prints a success message (alias for PrintSuccess)
func Success(message string) {
	PrintSuccess(message)
}

// PrintInfo prints an info message
func PrintInfo(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// Info prints an info message (alias for PrintInfo)
func Info(message string) {
	PrintInfo(message)
}

// Table prints data in a table format
func Table(headers []string, data [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Print headers
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Print data rows
	for _, row := range data {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	w.Flush()
}

// Bold returns text formatted in bold (for terminals that support it)
func Bold(text string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", text)
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// ColorizeStatus adds visual indicators to status strings
func ColorizeStatus(status string) string {
	// In a real implementation, you might use a color library
	// For now, just return the status with a prefix
	switch status {
	case "running":
		return "🟢 " + status
	case "completed":
		return "✅ " + status
	case "failed":
		return "❌ " + status
	case "pending":
		return "⏳ " + status
	default:
		return status
	}
}

func formatTargetNodes(nodes map[string]string) string {
	if len(nodes) == 0 {
		return "none"
	}

	var parts []string
	for k, v := range nodes {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ", ")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Warning prints a warning message (alias for PrintWarning)
func Warning(message string) {
	PrintWarning(message)
}

// Error prints an error message (alias for PrintError)
func Error(message string) {
	fmt.Fprintf(os.Stderr, "❌ %s\n", message)
}

// Confirm prompts the user for confirmation and returns true if they confirm
func Confirm(prompt string) (bool, error) {
	fmt.Printf("%s [y/N]: ", prompt)

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil && err.Error() != "unexpected newline" {
		return false, err
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

// PrintJSON prints data as JSON to the specified writer
func PrintJSON(writer interface{}, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if w, ok := writer.(interface{ Write([]byte) (int, error) }); ok {
		_, err = w.Write(jsonData)
		return err
	}

	fmt.Print(string(jsonData))
	return nil
}

// PrintYAML prints data as YAML to the specified writer
func PrintYAML(writer interface{}, data interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	if w, ok := writer.(interface{ Write([]byte) (int, error) }); ok {
		_, err = w.Write(yamlData)
		return err
	}

	fmt.Print(string(yamlData))
	return nil
}

// FormatBytes formats byte counts in human readable form
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// TruncateString truncates a string to the specified length
func TruncateString(s string, maxLen int) string {
	return truncate(s, maxLen)
}
