package completion

import (
	"context"
	"strings"
	"time"

	"github.com/phoenix/platform/services/phoenix-cli/internal/client"
	"github.com/spf13/cobra"
)

// CompletionCache caches completion results to improve performance
type CompletionCache struct {
	experiments map[string]time.Time
	deployments map[string]time.Time
	templates   map[string]time.Time
	ttl         time.Duration
}

var cache = &CompletionCache{
	experiments: make(map[string]time.Time),
	deployments: make(map[string]time.Time),
	templates:   make(map[string]time.Time),
	ttl:         5 * time.Minute,
}

// ExperimentIDCompletion provides completion for experiment IDs
func ExperimentIDCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	apiClient, err := getAPIClient(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Get namespace from flags or config
	namespace := getNamespace(cmd)

	experiments, err := apiClient.ListExperiments(namespace, "")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, exp := range experiments.Experiments {
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

	namespace := getNamespace(cmd)

	// This would require implementing ListDeployments in the API client
	// For now, return empty suggestions
	var suggestions []string

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// ExperimentNameCompletion provides completion for experiment names
func ExperimentNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	apiClient, err := getAPIClient(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	namespace := getNamespace(cmd)

	experiments, err := apiClient.ListExperiments(namespace, "")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, exp := range experiments.Experiments {
		if strings.HasPrefix(exp.Name, toComplete) {
			suggestions = append(suggestions, exp.Name)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// PipelineTemplateCompletion provides completion for pipeline templates
func PipelineTemplateCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Static list of known pipeline templates
	templates := []string{
		"process-baseline-v1",
		"process-intelligent-v1",
		"process-topk-v1",
		"process-aggregated-v1",
		"process-priority-filter-v1",
		"process-adaptive-v1",
		"process-minimal-v1",
	}

	var suggestions []string
	for _, template := range templates {
		if strings.HasPrefix(template, toComplete) {
			suggestions = append(suggestions, template)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// NamespaceCompletion provides completion for Kubernetes namespaces
func NamespaceCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Common namespace suggestions
	namespaces := []string{
		"default",
		"production",
		"staging",
		"development",
		"testing",
		"phoenix-system",
	}

	var suggestions []string
	for _, ns := range namespaces {
		if strings.HasPrefix(ns, toComplete) {
			suggestions = append(suggestions, ns)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// StatusCompletion provides completion for experiment/deployment status
func StatusCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	statuses := []string{
		"pending",
		"running",
		"completed",
		"failed",
		"stopped",
		"promoting",
		"promoted",
	}

	var suggestions []string
	for _, status := range statuses {
		if strings.HasPrefix(status, toComplete) {
			suggestions = append(suggestions, status)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// OutputFormatCompletion provides completion for output formats
func OutputFormatCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	formats := []string{
		"table",
		"json",
		"yaml",
	}

	var suggestions []string
	for _, format := range formats {
		if strings.HasPrefix(format, toComplete) {
			suggestions = append(suggestions, format)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// TrafficSplitCompletion provides completion for traffic split ratios
func TrafficSplitCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	splits := []string{
		"50/50",
		"80/20",
		"90/10",
		"70/30",
		"60/40",
		"95/5",
		"10/90",
		"20/80",
		"30/70",
	}

	var suggestions []string
	for _, split := range splits {
		if strings.HasPrefix(split, toComplete) {
			suggestions = append(suggestions, split)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// DurationCompletion provides completion for duration values
func DurationCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	durations := []string{
		"15m",
		"30m",
		"1h",
		"2h",
		"4h",
		"8h",
		"12h",
		"24h",
		"48h",
		"7d",
	}

	var suggestions []string
	for _, duration := range durations {
		if strings.HasPrefix(duration, toComplete) {
			suggestions = append(suggestions, duration)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// SelectorCompletion provides completion for Kubernetes selectors
func SelectorCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	selectors := []string{
		"app=webserver",
		"app=database",
		"app=api",
		"app=worker",
		"role=frontend",
		"role=backend",
		"env=production",
		"env=staging",
		"env=development",
		"tier=web",
		"tier=data",
		"component=cache",
		"component=queue",
	}

	var suggestions []string
	for _, selector := range selectors {
		if strings.HasPrefix(selector, toComplete) {
			suggestions = append(suggestions, selector)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// ConfigKeyCompletion provides completion for configuration keys
func ConfigKeyCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	keys := []string{
		"api_url",
		"auth_token",
		"default_namespace",
		"output_format",
		"debug",
		"timeout",
		"username",
	}

	var suggestions []string
	for _, key := range keys {
		if strings.HasPrefix(key, toComplete) {
			suggestions = append(suggestions, key)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// PluginNameCompletion provides completion for plugin names
func PluginNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// This would require loading the plugin manager
	// For now, return common plugin names
	plugins := []string{
		"metrics-exporter",
		"cost-analyzer",
		"alert-manager",
		"config-validator",
	}

	var suggestions []string
	for _, plugin := range plugins {
		if strings.HasPrefix(plugin, toComplete) {
			suggestions = append(suggestions, plugin)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// MetricNameCompletion provides completion for metric names
func MetricNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	metrics := []string{
		"cost_reduction",
		"data_loss",
		"progress",
		"cardinality",
		"ingestion_rate",
		"error_rate",
		"latency",
		"throughput",
		"memory_usage",
		"cpu_usage",
	}

	var suggestions []string
	for _, metric := range metrics {
		if strings.HasPrefix(metric, toComplete) {
			suggestions = append(suggestions, metric)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// Helper functions

func getAPIClient(cmd *cobra.Command) (*client.APIClient, error) {
	// This would need to be implemented to get the API client
	// For now, return a placeholder
	return nil, nil
}

func getNamespace(cmd *cobra.Command) string {
	// Try to get namespace from flags first
	if ns, err := cmd.Flags().GetString("namespace"); err == nil && ns != "" {
		return ns
	}

	// Fallback to config or default
	return "default"
}

// RegisterCompletions registers completion functions for all commands
func RegisterCompletions(rootCmd *cobra.Command) {
	// Walk through all commands and register appropriate completions
	registerCommandCompletions(rootCmd)
}

func registerCommandCompletions(cmd *cobra.Command) {
	// Register completion for common flags
	if cmd.Flags().Lookup("namespace") != nil {
		cmd.RegisterFlagCompletionFunc("namespace", NamespaceCompletion)
	}
	if cmd.Flags().Lookup("output") != nil {
		cmd.RegisterFlagCompletionFunc("output", OutputFormatCompletion)
	}
	if cmd.Flags().Lookup("status") != nil {
		cmd.RegisterFlagCompletionFunc("status", StatusCompletion)
	}
	if cmd.Flags().Lookup("pipeline-a") != nil {
		cmd.RegisterFlagCompletionFunc("pipeline-a", PipelineTemplateCompletion)
	}
	if cmd.Flags().Lookup("pipeline-b") != nil {
		cmd.RegisterFlagCompletionFunc("pipeline-b", PipelineTemplateCompletion)
	}
	if cmd.Flags().Lookup("template") != nil {
		cmd.RegisterFlagCompletionFunc("template", PipelineTemplateCompletion)
	}
	if cmd.Flags().Lookup("traffic-split") != nil {
		cmd.RegisterFlagCompletionFunc("traffic-split", TrafficSplitCompletion)
	}
	if cmd.Flags().Lookup("duration") != nil {
		cmd.RegisterFlagCompletionFunc("duration", DurationCompletion)
	}
	if cmd.Flags().Lookup("selector") != nil {
		cmd.RegisterFlagCompletionFunc("selector", SelectorCompletion)
	}

	// Register positional argument completions based on command
	switch cmd.Name() {
	case "status", "start", "stop", "promote", "metrics", "export":
		if cmd.Parent() != nil && cmd.Parent().Name() == "experiment" {
			cmd.ValidArgsFunction = ExperimentIDCompletion
		}
	case "get", "set":
		if cmd.Parent() != nil && cmd.Parent().Name() == "config" {
			cmd.ValidArgsFunction = ConfigKeyCompletion
		}
	case "install", "uninstall", "info":
		if cmd.Parent() != nil && cmd.Parent().Name() == "plugin" {
			cmd.ValidArgsFunction = PluginNameCompletion
		}
	}

	// Recursively register for subcommands
	for _, subCmd := range cmd.Commands() {
		registerCommandCompletions(subCmd)
	}
}