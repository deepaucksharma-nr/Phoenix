package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/completion"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/plugin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	apiEndpoint string
	outputFormat string
	verbose     bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "phoenix",
	Short: "Phoenix CLI - Manage observability pipeline experiments",
	Long: `Phoenix CLI is a command-line interface for the Phoenix Platform.
	
It allows you to:
  - Create and manage A/B experiments for pipeline optimization
  - Deploy pipeline configurations directly
  - Monitor experiment results and metrics
  - Promote successful configurations to production

For more information, visit: https://phoenix.example.com/docs`,
	Version: "0.1.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	// Register completions for all commands
	completion.RegisterCompletions(rootCmd)
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig, loadPlugins)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.phoenix/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiEndpoint, "api-endpoint", "", "Phoenix API endpoint (default is http://localhost:8080)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Bind flags to viper
	viper.BindPFlag("api.endpoint", rootCmd.PersistentFlags().Lookup("api-endpoint"))
	viper.BindPFlag("output.format", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Search config in home directory with name ".phoenix" (without extension).
		phoenixDir := filepath.Join(home, ".phoenix")
		viper.AddConfigPath(phoenixDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")

		// Create config directory if it doesn't exist
		if err := os.MkdirAll(phoenixDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
		}
	}

	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvPrefix("PHOENIX") // will be uppercased automatically

	// Set defaults
	viper.SetDefault("api.endpoint", "http://localhost:8080")
	viper.SetDefault("output.format", "table")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// loadPlugins discovers and loads CLI plugins
func loadPlugins() {
	// Skip plugin loading if we're running plugin management commands
	// to avoid circular dependencies
	if len(os.Args) > 1 && os.Args[1] == "plugin" {
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to get home directory for plugins: %v\n", err)
		}
		return
	}

	pluginDir := filepath.Join(home, ".phoenix", "plugins")
	pm := plugin.NewPluginManager(pluginDir)

	if err := pm.LoadPlugins(); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to load plugins: %v\n", err)
		}
		return
	}

	// Add plugin commands to root
	pluginCommands := pm.CreatePluginCommands()
	for _, cmd := range pluginCommands {
		rootCmd.AddCommand(cmd)
	}

	if verbose && len(pluginCommands) > 0 {
		fmt.Fprintf(os.Stderr, "Loaded %d plugin(s)\n", len(pluginCommands))
	}
}