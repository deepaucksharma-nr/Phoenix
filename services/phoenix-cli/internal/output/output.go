package output

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// PrintJSON prints data as formatted JSON
func PrintJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// PrintYAML prints data as YAML
func PrintYAML(data interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(data)
}

// PrintTable prints data in table format
func PrintTable(headers []string, rows [][]string) {
	w := os.Stdout
	fmt.Fprintln(w)
	
	// Print headers
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, h)
	}
	fmt.Fprintln(w)
	
	// Print separator
	for i := range headers {
		if i > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, "---")
	}
	fmt.Fprintln(w)
	
	// Print rows
	for _, row := range rows {
		for i, col := range row {
			if i > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, col)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w)
}

// PrintOverlapWarning prints a warning about overlapping experiments
func PrintOverlapWarning(overlaps []interface{}) {
	if len(overlaps) > 0 {
		fmt.Println("\n⚠️  Warning: The following experiments have overlapping target nodes:")
		for _, overlap := range overlaps {
			fmt.Printf("  - %v\n", overlap)
		}
		fmt.Println()
	}
}

// PrintExperiment prints experiment details
func PrintExperiment(exp interface{}) {
	fmt.Printf("Experiment created successfully!\n")
	fmt.Printf("ID: %v\n", exp)
}

// PrintExperimentList prints a list of experiments
func PrintExperimentList(experiments []interface{}) {
	if len(experiments) == 0 {
		fmt.Println("No experiments found")
		return
	}
	
	fmt.Printf("Found %d experiments:\n", len(experiments))
	for _, exp := range experiments {
		fmt.Printf("  - %v\n", exp)
	}
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	fmt.Printf("⚠️  Warning: %s\n", message)
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// PrintError prints an error message
func PrintError(message string) {
	fmt.Printf("❌ Error: %s\n", message)
}

// ColorizeStatus adds color to status strings
func ColorizeStatus(status string) string {
	// For CLI without color support, just return the status
	// In a real implementation, you could use a color library
	return status
}

// FormatBytes formats bytes to human readable format
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