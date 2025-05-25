package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigVersion represents a configuration schema version
type ConfigVersion struct {
	Version     string `yaml:"version" json:"version"`
	Description string `yaml:"description" json:"description"`
}

// Migration represents a configuration migration
type Migration struct {
	FromVersion string                                  `yaml:"from_version" json:"from_version"`
	ToVersion   string                                  `yaml:"to_version" json:"to_version"`
	Description string                                  `yaml:"description" json:"description"`
	Transform   func(map[string]interface{}) error      `yaml:"-" json:"-"`
}

// MigrationManager manages configuration migrations
type MigrationManager struct {
	configPath   string
	backupDir    string
	migrations   []Migration
	currentVersion string
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(configPath string) *MigrationManager {
	backupDir := filepath.Join(filepath.Dir(configPath), "backups")
	return &MigrationManager{
		configPath:     configPath,
		backupDir:      backupDir,
		migrations:     getBuiltinMigrations(),
		currentVersion: "1.0.0",
	}
}

// GetCurrentVersion returns the current config version
func (mm *MigrationManager) GetCurrentVersion() (string, error) {
	config, err := mm.loadConfig()
	if err != nil {
		return "", err
	}

	if version, ok := config["config_version"].(string); ok {
		return version, nil
	}

	// If no version is present, assume it's the legacy format
	return "0.1.0", nil
}

// CheckMigrationNeeded checks if migration is needed
func (mm *MigrationManager) CheckMigrationNeeded() (bool, string, error) {
	currentVersion, err := mm.GetCurrentVersion()
	if err != nil {
		return false, "", err
	}

	return currentVersion != mm.currentVersion, currentVersion, nil
}

// Migrate performs configuration migration
func (mm *MigrationManager) Migrate(dryRun bool) error {
	currentVersion, err := mm.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion == mm.currentVersion {
		return fmt.Errorf("configuration is already at the latest version (%s)", mm.currentVersion)
	}

	// Find migration path
	migrationPath, err := mm.findMigrationPath(currentVersion, mm.currentVersion)
	if err != nil {
		return fmt.Errorf("failed to find migration path: %w", err)
	}

	if len(migrationPath) == 0 {
		return fmt.Errorf("no migration path found from %s to %s", currentVersion, mm.currentVersion)
	}

	fmt.Printf("Migration path: %s", currentVersion)
	for _, migration := range migrationPath {
		fmt.Printf(" -> %s", migration.ToVersion)
	}
	fmt.Println()

	if dryRun {
		fmt.Println("Dry run mode: No changes will be made")
		return nil
	}

	// Create backup
	if err := mm.createBackup(); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Apply migrations
	config, err := mm.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	for _, migration := range migrationPath {
		fmt.Printf("Applying migration: %s -> %s (%s)\n", 
			migration.FromVersion, migration.ToVersion, migration.Description)
		
		if err := migration.Transform(config); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	// Update version
	config["config_version"] = mm.currentVersion
	config["last_migration"] = time.Now().Format(time.RFC3339)

	// Save updated config
	if err := mm.saveConfig(config); err != nil {
		return fmt.Errorf("failed to save migrated config: %w", err)
	}

	fmt.Printf("Migration completed successfully: %s -> %s\n", currentVersion, mm.currentVersion)
	return nil
}

// Rollback rolls back to a previous configuration version
func (mm *MigrationManager) Rollback(targetVersion string) error {
	// Find the most recent backup for the target version
	backupFile, err := mm.findBackup(targetVersion)
	if err != nil {
		return fmt.Errorf("failed to find backup for version %s: %w", targetVersion, err)
	}

	// Create backup of current config
	if err := mm.createBackup(); err != nil {
		return fmt.Errorf("failed to create backup before rollback: %w", err)
	}

	// Restore from backup
	backupData, err := os.ReadFile(backupFile)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	if err := os.WriteFile(mm.configPath, backupData, 0644); err != nil {
		return fmt.Errorf("failed to restore config: %w", err)
	}

	fmt.Printf("Rolled back to version %s\n", targetVersion)
	return nil
}

// ListBackups lists available configuration backups
func (mm *MigrationManager) ListBackups() ([]BackupInfo, error) {
	if _, err := os.Stat(mm.backupDir); os.IsNotExist(err) {
		return []BackupInfo{}, nil
	}

	entries, err := os.ReadDir(mm.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			// Parse filename: config_backup_YYYYMMDD_HHMMSS_v1.0.0.yaml
			parts := strings.Split(entry.Name(), "_")
			var version string
			if len(parts) >= 4 {
				versionPart := parts[len(parts)-1]
				version = strings.TrimSuffix(versionPart, ".yaml")
				if strings.HasPrefix(version, "v") {
					version = version[1:] // Remove 'v' prefix
				}
			}

			backups = append(backups, BackupInfo{
				Filename:  entry.Name(),
				Version:   version,
				Timestamp: info.ModTime(),
				Size:      info.Size(),
			})
		}
	}

	return backups, nil
}

// BackupInfo represents information about a configuration backup
type BackupInfo struct {
	Filename  string    `json:"filename"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Size      int64     `json:"size"`
}

// findMigrationPath finds the sequence of migrations needed
func (mm *MigrationManager) findMigrationPath(from, to string) ([]Migration, error) {
	var path []Migration
	currentVersion := from

	// Simple linear migration path for now
	for currentVersion != to {
		found := false
		for _, migration := range mm.migrations {
			if migration.FromVersion == currentVersion {
				path = append(path, migration)
				currentVersion = migration.ToVersion
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("no migration found from version %s", currentVersion)
		}
	}

	return path, nil
}

// loadConfig loads the configuration file
func (mm *MigrationManager) loadConfig() (map[string]interface{}, error) {
	data, err := os.ReadFile(mm.configPath)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

// saveConfig saves the configuration file
func (mm *MigrationManager) saveConfig(config map[string]interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(mm.configPath, data, 0644)
}

// createBackup creates a backup of the current configuration
func (mm *MigrationManager) createBackup() error {
	if err := os.MkdirAll(mm.backupDir, 0755); err != nil {
		return err
	}

	currentVersion, _ := mm.GetCurrentVersion()
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(mm.backupDir, fmt.Sprintf("config_backup_%s_v%s.yaml", timestamp, currentVersion))

	data, err := os.ReadFile(mm.configPath)
	if err != nil {
		return err
	}

	return os.WriteFile(backupFile, data, 0644)
}

// findBackup finds a backup file for a specific version
func (mm *MigrationManager) findBackup(version string) (string, error) {
	backups, err := mm.ListBackups()
	if err != nil {
		return "", err
	}

	// Find the most recent backup for the target version
	var latestBackup *BackupInfo
	for _, backup := range backups {
		if backup.Version == version {
			if latestBackup == nil || backup.Timestamp.After(latestBackup.Timestamp) {
				latestBackup = &backup
			}
		}
	}

	if latestBackup == nil {
		return "", fmt.Errorf("no backup found for version %s", version)
	}

	return filepath.Join(mm.backupDir, latestBackup.Filename), nil
}

// getBuiltinMigrations returns the built-in migration definitions
func getBuiltinMigrations() []Migration {
	return []Migration{
		{
			FromVersion: "0.1.0",
			ToVersion:   "0.2.0",
			Description: "Migrate from legacy config format to structured format",
			Transform:   migrate_0_1_0_to_0_2_0,
		},
		{
			FromVersion: "0.2.0",
			ToVersion:   "1.0.0",
			Description: "Add new authentication and plugin configuration",
			Transform:   migrate_0_2_0_to_1_0_0,
		},
	}
}

// Migration functions

func migrate_0_1_0_to_0_2_0(config map[string]interface{}) error {
	// Migrate from flat structure to nested structure
	
	// Migrate API configuration
	apiConfig := make(map[string]interface{})
	if endpoint, ok := config["api_endpoint"]; ok {
		apiConfig["endpoint"] = endpoint
		delete(config, "api_endpoint")
	}
	if token, ok := config["api_token"]; ok {
		apiConfig["token"] = token
		delete(config, "api_token")
	}
	if len(apiConfig) > 0 {
		config["api"] = apiConfig
	}

	// Migrate output configuration
	outputConfig := make(map[string]interface{})
	if format, ok := config["output_format"]; ok {
		outputConfig["format"] = format
		delete(config, "output_format")
	}
	if len(outputConfig) > 0 {
		config["output"] = outputConfig
	}

	// Add default values
	if config["default_namespace"] == nil {
		config["default_namespace"] = "default"
	}

	return nil
}

func migrate_0_2_0_to_1_0_0(config map[string]interface{}) error {
	// Add authentication section if missing
	if config["auth"] == nil {
		authConfig := make(map[string]interface{})
		
		// Move existing token to auth section
		if api, ok := config["api"].(map[string]interface{}); ok {
			if token, ok := api["token"]; ok {
				authConfig["token"] = token
				delete(api, "token")
			}
		}
		
		authConfig["method"] = "jwt"
		config["auth"] = authConfig
	}

	// Add plugins configuration
	if config["plugins"] == nil {
		pluginConfig := map[string]interface{}{
			"enabled":    true,
			"auto_load":  true,
			"cache_ttl":  "5m",
		}
		config["plugins"] = pluginConfig
	}

	// Add new feature flags
	if config["features"] == nil {
		featureConfig := map[string]interface{}{
			"auto_completion": true,
			"metrics_export":  true,
			"plugin_system":   true,
		}
		config["features"] = featureConfig
	}

	// Migrate legacy boolean flags
	if debug, ok := config["debug"].(bool); ok {
		if config["logging"] == nil {
			config["logging"] = map[string]interface{}{
				"level":   "info",
				"debug":   debug,
				"verbose": debug,
			}
		}
		delete(config, "debug")
	}

	return nil
}

// ValidateConfig validates configuration structure and values
func (mm *MigrationManager) ValidateConfig() error {
	config, err := mm.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	version, err := mm.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get config version: %w", err)
	}

	// Validate based on version
	switch version {
	case "1.0.0":
		return mm.validateV1Config(config)
	case "0.2.0":
		return mm.validateV0_2Config(config)
	default:
		return fmt.Errorf("unknown config version: %s", version)
	}
}

func (mm *MigrationManager) validateV1Config(config map[string]interface{}) error {
	// Check required sections
	requiredSections := []string{"api", "auth"}
	for _, section := range requiredSections {
		if config[section] == nil {
			return fmt.Errorf("missing required section: %s", section)
		}
	}

	// Validate API section
	if api, ok := config["api"].(map[string]interface{}); ok {
		if api["endpoint"] == nil {
			return fmt.Errorf("api.endpoint is required")
		}
	}

	// Validate auth section
	if auth, ok := config["auth"].(map[string]interface{}); ok {
		if method, ok := auth["method"].(string); ok {
			if method != "jwt" && method != "token" {
				return fmt.Errorf("invalid auth method: %s", method)
			}
		}
	}

	return nil
}

func (mm *MigrationManager) validateV0_2Config(config map[string]interface{}) error {
	// Basic validation for v0.2.0 format
	if config["api"] == nil {
		return fmt.Errorf("missing api configuration")
	}

	return nil
}

// ExportConfig exports configuration in a specific format
func (mm *MigrationManager) ExportConfig(format string) ([]byte, error) {
	config, err := mm.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	switch format {
	case "yaml":
		return yaml.Marshal(config)
	case "json":
		return json.MarshalIndent(config, "", "  ")
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}