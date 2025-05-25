package completion

import (
	"strings"

	"github.com/phoenix-vnext/platform/services/phoenix-cli/internal/client"
	"github.com/phoenix-vnext/platform/services/phoenix-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RegisterCompletions registers shell completion functions for all commands
func RegisterCompletions(rootCmd *cobra.Command) {
	// Register custom completion functions
	_ = rootCmd.RegisterFlagCompletionFunc("output", OutputFormatCompletion)
}

// ExperimentIDCompletion provides completion for experiment IDs
func ExperimentIDCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	apiClient, err := getAPIClient(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	req := client.ListExperimentsRequest{}
	experiments, err := apiClient.ListExperiments(req)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, exp := range experiments {
		if strings.HasPrefix(exp.ID, toComplete) {
			// Format: ID:NAME (STATUS)
			suggestion := exp.ID + ":" + exp.Name + " (" + exp.Status + ")"
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// DeploymentIDCompletion provides completion for deployment IDs
func DeploymentIDCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	apiClient, err := getAPIClient(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	req := client.ListPipelineDeploymentsRequest{}
	deployments, err := apiClient.ListPipelineDeployments(req)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, dep := range deployments {
		if strings.HasPrefix(dep.ID, toComplete) {
			// Format: ID:NAME (STATUS)
			suggestion := dep.ID + ":" + dep.DeploymentName + " (" + dep.Status + ")"
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// ExperimentNameCompletion provides completion for experiment names
func ExperimentNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	apiClient, err := getAPIClient(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	req := client.ListExperimentsRequest{}
	experiments, err := apiClient.ListExperiments(req)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, exp := range experiments {
		if strings.HasPrefix(exp.Name, toComplete) {
			suggestions = append(suggestions, exp.Name)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// PipelineTemplateCompletion provides completion for pipeline template names
func PipelineTemplateCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	apiClient, err := getAPIClient(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	pipelines, err := apiClient.ListPipelines()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, pipeline := range pipelines {
		if strings.HasPrefix(pipeline.ID, toComplete) {
			suggestions = append(suggestions, pipeline.ID)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// OutputFormatCompletion provides completion for output format flag
func OutputFormatCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	formats := []string{"table", "json", "yaml"}
	var suggestions []string
	
	for _, format := range formats {
		if strings.HasPrefix(format, toComplete) {
			suggestions = append(suggestions, format)
		}
	}
	
	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// ExperimentStatusCompletion provides completion for experiment status
func ExperimentStatusCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	statuses := []string{"pending", "running", "completed", "failed", "cancelled"}
	var suggestions []string
	
	for _, status := range statuses {
		if strings.HasPrefix(status, toComplete) {
			suggestions = append(suggestions, status)
		}
	}
	
	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// NamespaceCompletion provides completion for Kubernetes namespaces
func NamespaceCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// For now, return common namespaces
	// In a real implementation, this would query the Kubernetes API
	namespaces := []string{"default", "phoenix-system", "monitoring", "production", "staging"}
	var suggestions []string
	
	for _, ns := range namespaces {
		if strings.HasPrefix(ns, toComplete) {
			suggestions = append(suggestions, ns)
		}
	}
	
	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// getAPIClient creates an API client from command context
func getAPIClient(cmd *cobra.Command) (*client.APIClient, error) {
	// Get endpoint from viper config
	endpoint := viper.GetString("api.endpoint")
	if endpoint == "" {
		endpoint = "http://localhost:8080"
	}

	// Get token from config
	cfg := config.New()
	token := cfg.GetToken()

	return client.NewAPIClient(endpoint, token), nil
}

// getNamespace gets namespace from flags or config
func getNamespace(cmd *cobra.Command) string {
	namespace, _ := cmd.Flags().GetString("namespace")
	if namespace == "" {
		namespace = viper.GetString("namespace")
	}
	if namespace == "" {
		namespace = "default"
	}
	return namespace
}