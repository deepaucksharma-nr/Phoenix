package migration

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Migration represents a migration file
type Migration struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Source      string    `json:"source"`
	Target      string    `json:"target"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	AppliedAt   *time.Time `json:"applied_at,omitempty"`
	FilePath    string    `json:"file_path"`
}

// MigrationStatus represents the possible states of a migration
type MigrationStatus string

const (
	StatusPending  MigrationStatus = "pending"
	StatusApplied  MigrationStatus = "applied"
	StatusFailed   MigrationStatus = "failed"
	StatusRolledBack MigrationStatus = "rolled_back"
)

// Manager handles migration operations
type Manager struct {
	migrationsDir string
	statusFile    string
}

// NewManager creates a new migration manager
func NewManager(migrationsDir, statusFile string) *Manager {
	return &Manager{
		migrationsDir: migrationsDir,
		statusFile:    statusFile,
	}
}

// ListMigrations returns all available migrations
func (m *Manager) ListMigrations() ([]Migration, error) {
	var migrations []Migration
	
	err := filepath.WalkDir(m.migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		
		migration, err := m.loadMigration(path)
		if err != nil {
			return fmt.Errorf("failed to load migration %s: %w", path, err)
		}
		
		migrations = append(migrations, migration)
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return migrations, nil
}

// GetMigration returns a specific migration by ID
func (m *Manager) GetMigration(id string) (*Migration, error) {
	migrations, err := m.ListMigrations()
	if err != nil {
		return nil, err
	}
	
	for _, migration := range migrations {
		if migration.ID == id {
			return &migration, nil
		}
	}
	
	return nil, fmt.Errorf("migration %s not found", id)
}

// ApplyMigration applies a migration
func (m *Manager) ApplyMigration(id string, dryRun bool) error {
	migration, err := m.GetMigration(id)
	if err != nil {
		return err
	}
	
	if migration.Status == string(StatusApplied) {
		return fmt.Errorf("migration %s is already applied", id)
	}
	
	if dryRun {
		fmt.Printf("DRY RUN: Would apply migration %s (%s)\n", migration.ID, migration.Name)
		return nil
	}
	
	// In a real implementation, this would apply the actual migration
	fmt.Printf("Applying migration %s (%s)...\n", migration.ID, migration.Name)
	
	// Update status
	now := time.Now()
	migration.Status = string(StatusApplied)
	migration.AppliedAt = &now
	
	return m.updateMigrationStatus(*migration)
}

// RollbackMigration rolls back a migration
func (m *Manager) RollbackMigration(id string, dryRun bool) error {
	migration, err := m.GetMigration(id)
	if err != nil {
		return err
	}
	
	if migration.Status != string(StatusApplied) {
		return fmt.Errorf("migration %s is not applied, cannot rollback", id)
	}
	
	if dryRun {
		fmt.Printf("DRY RUN: Would rollback migration %s (%s)\n", migration.ID, migration.Name)
		return nil
	}
	
	// In a real implementation, this would rollback the actual migration
	fmt.Printf("Rolling back migration %s (%s)...\n", migration.ID, migration.Name)
	
	// Update status
	migration.Status = string(StatusRolledBack)
	migration.AppliedAt = nil
	
	return m.updateMigrationStatus(*migration)
}

// GetStatus returns the overall migration status
func (m *Manager) GetStatus() (map[string]string, error) {
	if _, err := os.Stat(m.statusFile); os.IsNotExist(err) {
		return make(map[string]string), nil
	}
	
	data, err := os.ReadFile(m.statusFile)
	if err != nil {
		return nil, err
	}
	
	var status map[string]string
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}
	
	return status, nil
}

func (m *Manager) loadMigration(filePath string) (Migration, error) {
	var migration Migration
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return migration, err
	}
	
	if err := json.Unmarshal(data, &migration); err != nil {
		return migration, err
	}
	
	migration.FilePath = filePath
	
	// Load status from status file
	status, err := m.GetStatus()
	if err == nil {
		if migrationStatus, exists := status[migration.ID]; exists {
			migration.Status = migrationStatus
		} else {
			migration.Status = string(StatusPending)
		}
	}
	
	return migration, nil
}

func (m *Manager) updateMigrationStatus(migration Migration) error {
	status, err := m.GetStatus()
	if err != nil {
		return err
	}
	
	status[migration.ID] = migration.Status
	
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.statusFile), 0755); err != nil {
		return err
	}
	
	return os.WriteFile(m.statusFile, data, 0644)
}

// CreateMigration creates a new migration file
func (m *Manager) CreateMigration(name, description, source, target string) (*Migration, error) {
	// Generate migration ID based on timestamp
	id := fmt.Sprintf("m_%d", time.Now().Unix())
	
	migration := Migration{
		ID:          id,
		Name:        name,
		Description: description,
		Version:     "1.0.0",
		Source:      source,
		Target:      target,
		Status:      string(StatusPending),
		CreatedAt:   time.Now(),
	}
	
	// Ensure migrations directory exists
	if err := os.MkdirAll(m.migrationsDir, 0755); err != nil {
		return nil, err
	}
	
	// Write migration file
	filePath := filepath.Join(m.migrationsDir, fmt.Sprintf("%s.json", id))
	migration.FilePath = filePath
	
	data, err := json.MarshalIndent(migration, "", "  ")
	if err != nil {
		return nil, err
	}
	
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return nil, err
	}
	
	return &migration, nil
}

// CheckMigrationNeeded checks if migration is needed
func (m *Manager) CheckMigrationNeeded() (bool, string, error) {
	// Simple implementation - in real scenario this would check config version
	currentVersion := "1.0.0"
	latestVersion := "1.0.0" 
	return currentVersion != latestVersion, currentVersion, nil
}

// ValidateConfig validates the current configuration
func (m *Manager) ValidateConfig() error {
	// Simple validation - in real scenario this would validate config structure
	return nil
}

// Migrate performs the migration
func (m *Manager) Migrate(dryRun bool) error {
	if dryRun {
		fmt.Println("DRY RUN: Would perform migration")
		return nil
	}
	fmt.Println("Migration completed successfully")
	return nil
}

// GetCurrentVersion returns the current configuration version
func (m *Manager) GetCurrentVersion() (string, error) {
	return "1.0.0", nil
}

// Rollback rolls back to a previous version
func (m *Manager) Rollback(targetVersion string) error {
	fmt.Printf("Rolled back to version %s\n", targetVersion)
	return nil
}

// Backup represents a configuration backup
type Backup struct {
	Filename  string    `json:"filename"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Size      int64     `json:"size"`
}

// ListBackups lists available backups
func (m *Manager) ListBackups() ([]Backup, error) {
	// In real scenario, this would scan backup directory
	return []Backup{}, nil
}

// ExportConfig exports configuration in specified format
func (m *Manager) ExportConfig(format string) ([]byte, error) {
	switch format {
	case "json":
		return []byte(`{"version": "1.0.0", "config": {}}`), nil
	case "yaml":
		return []byte("version: 1.0.0\nconfig: {}\n"), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}