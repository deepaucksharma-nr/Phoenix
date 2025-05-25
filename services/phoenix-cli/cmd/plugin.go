package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/phoenix-vnext/platform/services/phoenix-cli/internal/output"
	"github.com/phoenix-vnext/platform/services/phoenix-cli/internal/plugin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage CLI plugins",
	Long:  "Install, uninstall, and manage Phoenix CLI plugins to extend functionality",
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed plugins",
	Long:  "Display all installed Phoenix CLI plugins",
	RunE:  runPluginList,
}

var pluginInstallCmd = &cobra.Command{
	Use:   "install <source>",
	Short: "Install a plugin",
	Long:  "Install a Phoenix CLI plugin from a local directory or remote source",
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginInstall,
}

var pluginUninstallCmd = &cobra.Command{
	Use:   "uninstall <plugin-name>",
	Short: "Uninstall a plugin",
	Long:  "Remove an installed Phoenix CLI plugin",
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginUninstall,
}

var pluginInfoCmd = &cobra.Command{
	Use:   "info <plugin-name>",
	Short: "Show plugin information",
	Long:  "Display detailed information about an installed plugin",
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginInfo,
}

var pluginCreateCmd = &cobra.Command{
	Use:   "create <plugin-name>",
	Short: "Create a new plugin template",
	Long:  "Generate a new plugin template with boilerplate code",
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginCreate,
}

func init() {
	// Add plugin subcommands
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginInstallCmd)
	pluginCmd.AddCommand(pluginUninstallCmd)
	pluginCmd.AddCommand(pluginInfoCmd)
	pluginCmd.AddCommand(pluginCreateCmd)

	// Plugin install flags
	pluginInstallCmd.Flags().Bool("force", false, "Force installation even if plugin already exists")

	// Plugin create flags
	pluginCreateCmd.Flags().String("description", "", "Plugin description")
	pluginCreateCmd.Flags().String("author", "", "Plugin author")
	pluginCreateCmd.Flags().String("language", "bash", "Plugin language (bash, go, python)")

	rootCmd.AddCommand(pluginCmd)
}

func getPluginManager() (*plugin.PluginManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	pluginDir := filepath.Join(homeDir, ".phoenix", "plugins")
	pm := plugin.NewPluginManager(pluginDir)

	if err := pm.LoadPlugins(); err != nil {
		return nil, fmt.Errorf("failed to load plugins: %w", err)
	}

	return pm, nil
}

func runPluginList(cmd *cobra.Command, args []string) error {
	pm, err := getPluginManager()
	if err != nil {
		return err
	}

	plugins := pm.ListPlugins()
	if len(plugins) == 0 {
		fmt.Println("No plugins installed")
		return nil
	}

	outputFormat := viper.GetString("output")
	switch outputFormat {
	case "json":
		return output.PrintJSON(map[string]interface{}{
			"plugins": plugins,
		})
	case "yaml":
		return output.PrintYAML(map[string]interface{}{
			"plugins": plugins,
		})
	default:
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tVERSION\tDESCRIPTION\tAUTHOR")
		for _, p := range plugins {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				p.Name,
				p.Version,
				p.Description,
				p.Author,
			)
		}
		return w.Flush()
	}
}

func runPluginInstall(cmd *cobra.Command, args []string) error {
	source := args[0]
	force, _ := cmd.Flags().GetBool("force")

	pm, err := getPluginManager()
	if err != nil {
		return err
	}

	// Check if plugin already exists (unless force is set)
	if !force {
		// Try to extract plugin name from source to check if it exists
		manifestPath := filepath.Join(source, "plugin.json")
		if data, err := os.ReadFile(manifestPath); err == nil {
			var manifest map[string]interface{}
			if json.Unmarshal(data, &manifest) == nil {
				if name, ok := manifest["name"].(string); ok {
					if _, exists := pm.GetPlugin(name); exists {
						return fmt.Errorf("plugin %s is already installed (use --force to overwrite)", name)
					}
				}
			}
		}
	}

	return pm.InstallPlugin(source)
}

func runPluginUninstall(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	pm, err := getPluginManager()
	if err != nil {
		return err
	}

	return pm.UninstallPlugin(pluginName)
}

func runPluginInfo(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	pm, err := getPluginManager()
	if err != nil {
		return err
	}

	plugin, ok := pm.GetPlugin(pluginName)
	if !ok {
		return fmt.Errorf("plugin %s is not installed", pluginName)
	}

	outputFormat := viper.GetString("output")
	switch outputFormat {
	case "json":
		return output.PrintJSON(plugin)
	case "yaml":
		return output.PrintYAML(plugin)
	default:
		fmt.Printf("Name:        %s\n", plugin.Name)
		fmt.Printf("Version:     %s\n", plugin.Version)
		fmt.Printf("Description: %s\n", plugin.Description)
		fmt.Printf("Author:      %s\n", plugin.Author)
		fmt.Printf("Path:        %s\n", plugin.Path)
		fmt.Printf("Executable:  %s\n", plugin.Executable)
		if len(plugin.Commands) > 0 {
			fmt.Printf("Commands:    %v\n", plugin.Commands)
		}
		return nil
	}
}

func runPluginCreate(cmd *cobra.Command, args []string) error {
	pluginName := args[0]
	description, _ := cmd.Flags().GetString("description")
	author, _ := cmd.Flags().GetString("author")
	language, _ := cmd.Flags().GetString("language")

	// Validate plugin name
	if err := plugin.ValidatePluginName(pluginName); err != nil {
		return err
	}

	// Set defaults
	if description == "" {
		description = fmt.Sprintf("Phoenix CLI plugin: %s", pluginName)
	}
	if author == "" {
		author = "Unknown"
	}

	// Create plugin directory
	pluginDir := filepath.Join(".", pluginName)
	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		return fmt.Errorf("directory %s already exists", pluginDir)
	}

	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Create plugin manifest
	manifest := map[string]interface{}{
		"name":        pluginName,
		"version":     "1.0.0",
		"description": description,
		"author":      author,
		"commands":    []string{pluginName},
		"executable":  getExecutableName(language),
	}

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to create plugin manifest: %w", err)
	}

	manifestPath := filepath.Join(pluginDir, "plugin.json")
	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		return fmt.Errorf("failed to write plugin manifest: %w", err)
	}

	// Create executable template
	executablePath := filepath.Join(pluginDir, getExecutableName(language))
	executableContent := getExecutableTemplate(language, pluginName, description)

	if err := os.WriteFile(executablePath, []byte(executableContent), 0755); err != nil {
		return fmt.Errorf("failed to create executable: %w", err)
	}

	// Create README
	readmePath := filepath.Join(pluginDir, "README.md")
	readmeContent := getReadmeTemplate(pluginName, description, language)

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}

	fmt.Printf("Plugin %s created successfully in %s\n", pluginName, pluginDir)
	fmt.Printf("To install: phoenix plugin install %s\n", pluginDir)
	return nil
}

func getExecutableName(language string) string {
	switch language {
	case "go":
		return "main"
	case "python":
		return "main.py"
	default:
		return "main.sh"
	}
}

func getExecutableTemplate(language, name, description string) string {
	switch language {
	case "go":
		return fmt.Sprintf(`package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("Phoenix CLI Plugin: %s\n", %q)
	fmt.Printf("Description: %s\n", %q)
	fmt.Printf("Arguments: %%v\n", os.Args[1:])
	
	// Add your plugin logic here
	fmt.Println("Plugin executed successfully!")
}
`, name, description)

	case "python":
		return fmt.Sprintf(`#!/usr/bin/env python3
"""
Phoenix CLI Plugin: %s
Description: %s
"""

import sys
import os

def main():
    print(f"Phoenix CLI Plugin: %s")
    print(f"Description: %s")
    print(f"Arguments: {sys.argv[1:]}")
    
    # Plugin environment variables
    plugin_name = os.getenv('PHOENIX_PLUGIN_NAME', 'unknown')
    plugin_version = os.getenv('PHOENIX_PLUGIN_VERSION', 'unknown')
    
    print(f"Plugin Name: {plugin_name}")
    print(f"Plugin Version: {plugin_version}")
    
    # Add your plugin logic here
    print("Plugin executed successfully!")

if __name__ == "__main__":
    main()
`, name, description, name, description)

	default: // bash
		return fmt.Sprintf(`#!/bin/bash
# Phoenix CLI Plugin: %s
# Description: %s

set -e

echo "Phoenix CLI Plugin: %s"
echo "Description: %s"
echo "Arguments: $@"

# Plugin environment variables
echo "Plugin Name: ${PHOENIX_PLUGIN_NAME:-unknown}"
echo "Plugin Version: ${PHOENIX_PLUGIN_VERSION:-unknown}"

# Add your plugin logic here
echo "Plugin executed successfully!"
`, name, description, name, description)
	}
}

/*
func getReadmeTemplate(name, description, language string) string {
	return fmt.Sprintf(`# %s

%s

## Installation

` + "```bash\nphoenix plugin install .\n```" + `

## Usage

` + "```bash\nphoenix %s [arguments]\n```" + `

## Development

This plugin is written in %s. To modify:

1. Edit the main executable file
2. Update the plugin.json manifest if needed
3. Reinstall the plugin: ` + "`phoenix plugin install . --force`"

## Plugin Structure

- ` + "`plugin.json`" + ` - Plugin manifest with metadata
- ` + "`%s`" + ` - Main executable
- ` + "`README.md`" + ` - This file

## Environment Variables

When executed, the plugin has access to:
- ` + "`PHOENIX_PLUGIN_NAME`" + ` - The plugin name
- ` + "`PHOENIX_PLUGIN_VERSION`" + ` - The plugin version

## API Integration

To interact with the Phoenix API from your plugin, you can:

1. Use the phoenix CLI commands as subprocesses
2. Make direct HTTP requests to the API
3. Use the Phoenix API token from the user's configuration

Example API call:
` + "```bash\n# Get auth token from phoenix config\nAPI_TOKEN=$(phoenix config get auth_token)\nAPI_URL=$(phoenix config get api_url)\n\n# Make API request\ncurl -H \"Authorization: Bearer $API_TOKEN\" \\\n     \"$API_URL/api/v1/experiments\"\n```" + `
`, name, description, name, language, getExecutableName(language))
}
*/
// getReadmeTemplate generates a README.md template for plugins
func getReadmeTemplate(name, description, language string) string {
	executableName := getExecutableName(language)
	
	template := `# %s

%s

## Installation

` + "```bash" + `
phoenix plugin install .
` + "```" + `

## Usage

` + "```bash" + `
phoenix %s [arguments]
` + "```" + `

## Development

This plugin is written in %s. To modify:

1. Edit the main executable file
2. Update the plugin.json manifest if needed
3. Reinstall the plugin: ` + "`" + `phoenix plugin install . --force` + "`" + `

## Plugin Structure

- ` + "`" + `plugin.json` + "`" + ` - Plugin manifest with metadata
- ` + "`" + `%s` + "`" + ` - Main executable
- ` + "`" + `README.md` + "`" + ` - This file

## Environment Variables

When executed, the plugin has access to:
- ` + "`" + `PHOENIX_PLUGIN_NAME` + "`" + ` - The plugin name
- ` + "`" + `PHOENIX_PLUGIN_VERSION` + "`" + ` - The plugin version

## API Integration

To interact with the Phoenix API from your plugin, you can:

1. Use the phoenix CLI commands as subprocesses
2. Make direct HTTP requests to the API
3. Use the Phoenix API token from the user's configuration

Example API call:
` + "```bash" + `
# Get auth token from phoenix config
API_TOKEN=$(phoenix config get auth_token)
API_URL=$(phoenix config get api_url)

# Make API request
curl -H "Authorization: Bearer $API_TOKEN" \
     "$API_URL/api/v1/experiments"
` + "```" + `
`
	
	return fmt.Sprintf(template, name, description, name, language, executableName)
}