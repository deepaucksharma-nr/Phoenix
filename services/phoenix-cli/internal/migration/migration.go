package migration

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Version represents a configuration version
type Version struct {
	Version   string    `json:"version" yaml:"version"`
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
}

// Manager handles configuration migrations
type Manager struct {
	configPath string
}

// NewManager creates a new migration manager
func NewManager(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
	}
}

// GetCurrentVersion gets the current configuration version
func (m *Manager) GetCurrentVersion() (string, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("failed to parse config file: %w", err)
	}

	version, ok := config["version"].(string)
	if !ok {
		return "0.0.0", nil // Default version if not specified
	}

	return version, nil
}

// Backup creates a backup of the current configuration
func (m *Manager) Backup() (string, error) {
	backupDir := filepath.Join(filepath.Dir(m.configPath), "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	backupFile := filepath.Join(backupDir, fmt.Sprintf("config-%s.yaml", timestamp))

	src, err := os.Open(m.configPath)
	if err != nil {
		return "", fmt.Errorf("failed to open config file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(backupFile)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy config file: %w", err)
	}

	return backupFile, nil
}

// Migrate performs the migration to the latest version
func (m *Manager) Migrate() error {
	currentVersion, err := m.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Create backup before migration
	backupPath, err := m.Backup()
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Load current configuration
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply migrations based on version
	switch currentVersion {
	case "0.0.0":
		// Migrate from no version to v1
		config["version"] = "1.0.0"
		config["updated_at"] = time.Now().Format(time.RFC3339)
	}

	// Write updated configuration
	updatedData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Configuration migrated successfully. Backup saved to: %s\n", backupPath)
	return nil
}

// History returns migration history
func (m *Manager) History() ([]Version, error) {
	// For now, return a simple history
	// In a real implementation, this would read from a history file
	return []Version{
		{Version: "1.0.0", UpdatedAt: time.Now()},
	}, nil
}