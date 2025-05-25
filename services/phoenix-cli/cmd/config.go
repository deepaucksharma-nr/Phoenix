package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/phoenix-vnext/platform/services/phoenix-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// configCmd represents the config command group
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Phoenix CLI configuration",
	Long: `Manage Phoenix CLI configuration settings.

Configuration is stored in ~/.phoenix/config.yaml by default.`,
}

// configGetCmd gets a configuration value
var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long: `Get a configuration value by key.

Examples:
  # Get API endpoint
  phoenix config get api.endpoint

  # Get output format
  phoenix config get output.format`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

// configSetCmd sets a configuration value
var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Examples:
  # Set API endpoint
  phoenix config set api.endpoint https://phoenix.company.com

  # Set default output format
  phoenix config set output.format json

  # Set namespace
  phoenix config set defaults.namespace production`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

// configListCmd lists all configuration values
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long:  `List all configuration values.`,
	RunE:  runConfigList,
}

// configResetCmd resets configuration to defaults
var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Long: `Reset configuration to default values.

This will remove all custom configuration and restore defaults.`,
	RunE: runConfigReset,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configResetCmd)
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]
	
	value := viper.Get(key)
	if value == nil {
		return fmt.Errorf("configuration key '%s' not found", key)
	}
	
	switch outputFormat {
	case "json":
		fmt.Printf("{\"%s\": \"%v\"}\n", key, value)
	case "yaml":
		fmt.Printf("%s: %v\n", key, value)
	default:
		fmt.Printf("%v\n", value)
	}
	
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]
	
	cfg := config.New()
	
	// Handle special keys
	switch key {
	case "api.endpoint":
		return cfg.SetAPIEndpoint(value)
	case "output.format":
		if value != "table" && value != "json" && value != "yaml" {
			return fmt.Errorf("invalid output format: %s (must be table, json, or yaml)", value)
		}
		return cfg.SetOutputFormat(value)
	case "auth.token":
		return fmt.Errorf("cannot set auth token directly, use 'phoenix auth login' instead")
	default:
		// For other keys, use viper directly
		viper.Set(key, value)
		if err := viper.WriteConfig(); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}
	}
	
	fmt.Printf("Configuration updated: %s = %s\n", key, value)
	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	settings := viper.AllSettings()
	
	switch outputFormat {
	case "json":
		data, _ := json.MarshalIndent(settings, "", "  ")
		fmt.Println(string(data))
	case "yaml":
		data, _ := yaml.Marshal(settings)
		fmt.Print(string(data))
	default:
		// Table format
		fmt.Println("Current Configuration:")
		fmt.Println("=====================")
		printSettings("", settings)
	}
	
	return nil
}

func printSettings(prefix string, settings map[string]interface{}) {
	for key, value := range settings {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		
		switch v := value.(type) {
		case map[string]interface{}:
			printSettings(fullKey, v)
		case string:
			// Mask auth token
			if strings.Contains(fullKey, "token") && v != "" {
				fmt.Printf("%-30s %s\n", fullKey+":", "****")
			} else {
				fmt.Printf("%-30s %s\n", fullKey+":", v)
			}
		default:
			fmt.Printf("%-30s %v\n", fullKey+":", v)
		}
	}
}

func runConfigReset(cmd *cobra.Command, args []string) error {
	fmt.Print("Are you sure you want to reset all configuration to defaults? [y/N]: ")
	
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" {
		fmt.Println("Reset cancelled.")
		return nil
	}
	
	// Clear all settings
	for _, key := range viper.AllKeys() {
		viper.Set(key, nil)
	}
	
	// Set defaults
	viper.Set("api.endpoint", "http://localhost:8080")
	viper.Set("output.format", "table")
	
	// Write config
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to reset configuration: %w", err)
	}
	
	fmt.Println("Configuration reset to defaults.")
	return nil
}