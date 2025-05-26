package completion

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// GetExperimentIDs returns a list of experiment IDs for shell completion
func GetExperimentIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// In a real implementation, this would fetch from the API
	// For now, return some example IDs
	experimentIDs := []string{
		"exp-12345678", "exp-87654321", "exp-11223344", "exp-99887766",
	}
	
	// Filter based on what the user has typed
	var matches []string
	for _, id := range experimentIDs {
		if strings.HasPrefix(id, toComplete) {
			matches = append(matches, id)
		}
	}
	
	return matches, cobra.ShellCompDirectiveNoFileComp
}

// GetPipelineNames returns a list of pipeline names for shell completion
func GetPipelineNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// In a real implementation, this would fetch from the API
	pipelineNames := []string{
		"process-baseline-v1", "process-adaptive-v1", "process-intelligent-v1",
		"process-aggregated-v1", "process-minimal-v1", "process-topk-v1",
	}
	
	var matches []string
	for _, name := range pipelineNames {
		if strings.HasPrefix(name, toComplete) {
			matches = append(matches, name)
		}
	}
	
	return matches, cobra.ShellCompDirectiveNoFileComp
}

// GetLoadSimProfiles returns a list of load simulation profiles for shell completion
func GetLoadSimProfiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	profiles := []string{
		"realistic", "high-cardinality", "process-churn", "custom",
	}
	
	var matches []string
	for _, profile := range profiles {
		if strings.HasPrefix(profile, toComplete) {
			matches = append(matches, profile)
		}
	}
	
	return matches, cobra.ShellCompDirectiveNoFileComp
}

// GetNamespaces returns a list of Kubernetes namespaces for shell completion
func GetNamespaces(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// In a real implementation, this would use kubectl or the Kubernetes API
	namespaces := []string{
		"default", "phoenix-dev", "phoenix-staging", "phoenix-prod",
		"kube-system", "monitoring",
	}
	
	var matches []string
	for _, ns := range namespaces {
		if strings.HasPrefix(ns, toComplete) {
			matches = append(matches, ns)
		}
	}
	
	return matches, cobra.ShellCompDirectiveNoFileComp
}

// GetOutputFormats returns a list of supported output formats
func GetOutputFormats(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	formats := []string{"table", "json", "yaml"}
	
	var matches []string
	for _, format := range formats {
		if strings.HasPrefix(format, toComplete) {
			matches = append(matches, format)
		}
	}
	
	return matches, cobra.ShellCompDirectiveNoFileComp
}

// ValidateExperimentID validates that an experiment ID has the correct format
func ValidateExperimentID(experimentID string) error {
	if experimentID == "" {
		return fmt.Errorf("experiment ID cannot be empty")
	}
	
	if !strings.HasPrefix(experimentID, "exp-") {
		return fmt.Errorf("experiment ID must start with 'exp-'")
	}
	
	if len(experimentID) != 12 { // exp- + 8 chars
		return fmt.Errorf("experiment ID must be 12 characters long (exp- + 8 character ID)")
	}
	
	return nil
}

// RegisterCompletions registers completion functions for commands
func RegisterCompletions(rootCmd *cobra.Command) {
	// This function would register shell completion for various commands
	// For now, we'll leave it as a placeholder
}