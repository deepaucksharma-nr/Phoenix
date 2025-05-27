package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/migration"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage configuration migrations",
	Long:  "Migrate Phoenix CLI configuration files between versions and manage backups",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Migrate configuration to latest version",
	Long:  "Upgrade configuration file to the latest schema version",
	RunE:  runMigrateUp,
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  "Display current configuration version and available migrations",
	RunE:  runMigrateStatus,
}

var migrateRollbackCmd = &cobra.Command{
	Use:   "rollback <version>",
	Short: "Rollback to a previous version",
	Long:  "Rollback configuration to a previous version using backups",
	Args:  cobra.ExactArgs(1),
	RunE:  runMigrateRollback,
}

var migrateBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage configuration backups",
	Long:  "List and manage configuration file backups",
}

var migrateBackupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available backups",
	Long:  "Show all available configuration backups",
	RunE:  runMigrateBackupList,
}

var migrateValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  "Validate current configuration file structure and values",
	RunE:  runMigrateValidate,
}

var migrateExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export configuration",
	Long:  "Export current configuration in various formats",
	RunE:  runMigrateExport,
}

func init() {
	// Add subcommands
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateRollbackCmd)
	migrateCmd.AddCommand(migrateBackupCmd)
	migrateCmd.AddCommand(migrateValidateCmd)
	migrateCmd.AddCommand(migrateExportCmd)

	// Backup subcommands
	migrateBackupCmd.AddCommand(migrateBackupListCmd)

	// Migration flags
	migrateUpCmd.Flags().Bool("dry-run", false, "Show what would be migrated without making changes")
	migrateUpCmd.Flags().Bool("force", false, "Force migration even if validation fails")

	// Export flags
	migrateExportCmd.Flags().String("format", "yaml", "Export format: yaml, json")
	migrateExportCmd.Flags().String("output", "", "Output file (default: stdout)")

	rootCmd.AddCommand(migrateCmd)
}

func getMigrationManager() (*migration.Manager, error) {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		// Use default config path
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configFile = filepath.Join(home, ".phoenix", "config.yaml")
	}

	migrationsDir := filepath.Join(filepath.Dir(configFile), "migrations")
	statusFile := filepath.Join(filepath.Dir(configFile), "migration_status.json")
	return migration.NewManager(migrationsDir, statusFile), nil
}

func runMigrateUp(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")

	mm, err := getMigrationManager()
	if err != nil {
		return err
	}

	// Check if migration is needed
	needed, currentVersion, err := mm.CheckMigrationNeeded()
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if !needed {
		fmt.Printf("Configuration is already at the latest version (%s)\n", currentVersion)
		return nil
	}

	// Validate current config unless forced
	if !force {
		if err := mm.ValidateConfig(); err != nil {
			fmt.Printf("Warning: Current configuration has validation errors: %v\n", err)
			fmt.Println("Use --force to proceed anyway")
			return nil
		}
	}

	// Perform migration
	fmt.Printf("Migrating configuration from version %s to latest\n", currentVersion)
	return mm.Migrate(dryRun)
}

func runMigrateStatus(cmd *cobra.Command, args []string) error {
	mm, err := getMigrationManager()
	if err != nil {
		return err
	}

	currentVersion, err := mm.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	needed, _, err := mm.CheckMigrationNeeded()
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	outputFormat := viper.GetString("output")

	status := map[string]interface{}{
		"current_version":  currentVersion,
		"latest_version":   "1.0.0", // This should come from the migration manager
		"migration_needed": needed,
		"config_file":      viper.ConfigFileUsed(),
	}

	switch outputFormat {
	case "json":
		return output.PrintJSON(cmd.OutOrStdout(), status)
	case "yaml":
		return output.PrintYAML(cmd.OutOrStdout(), status)
	default:
		fmt.Printf("Configuration Status:\n")
		fmt.Printf("  Current Version: %s\n", currentVersion)
		fmt.Printf("  Latest Version:  1.0.0\n")
		fmt.Printf("  Migration Needed: %v\n", needed)
		fmt.Printf("  Config File: %s\n", viper.ConfigFileUsed())

		if needed {
			fmt.Printf("\nTo migrate: phoenix migrate up\n")
		}

		return nil
	}
}

func runMigrateRollback(cmd *cobra.Command, args []string) error {
	targetVersion := args[0]

	mm, err := getMigrationManager()
	if err != nil {
		return err
	}

	fmt.Printf("Rolling back to version %s...\n", targetVersion)
	return mm.Rollback(targetVersion)
}

func runMigrateBackupList(cmd *cobra.Command, args []string) error {
	mm, err := getMigrationManager()
	if err != nil {
		return err
	}

	backups, err := mm.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	if len(backups) == 0 {
		fmt.Println("No backups found")
		return nil
	}

	outputFormat := viper.GetString("output")

	switch outputFormat {
	case "json":
		return output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"backups": backups,
		})
	case "yaml":
		return output.PrintYAML(cmd.OutOrStdout(), map[string]interface{}{
			"backups": backups,
		})
	default:
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 2, ' ', 0)
		fmt.Fprintln(w, "FILENAME\tVERSION\tTIMESTAMP\tSIZE")

		for _, backup := range backups {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				backup.Filename,
				backup.Version,
				backup.Timestamp.Format("2006-01-02 15:04:05"),
				output.FormatBytes(backup.Size),
			)
		}

		return w.Flush()
	}
}

func runMigrateValidate(cmd *cobra.Command, args []string) error {
	mm, err := getMigrationManager()
	if err != nil {
		return err
	}

	fmt.Println("Validating configuration...")

	if err := mm.ValidateConfig(); err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		return err
	}

	fmt.Println("✅ Configuration is valid")
	return nil
}

func runMigrateExport(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	outputFile, _ := cmd.Flags().GetString("output")

	mm, err := getMigrationManager()
	if err != nil {
		return err
	}

	data, err := mm.ExportConfig(format)
	if err != nil {
		return fmt.Errorf("failed to export config: %w", err)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Configuration exported to: %s\n", outputFile)
	} else {
		cmd.OutOrStdout().Write(data)
	}

	return nil
}
