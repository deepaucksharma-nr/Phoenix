package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorBlue   = "\033[0;34m"
	colorReset  = "\033[0m"
)

// Migration represents a database migration
type Migration struct {
	Version  string
	Name     string
	FilePath string
	SQL      string
}

func main() {
	var (
		dbURL        = flag.String("db", getEnvOrDefault("DATABASE_URL", "postgres://phoenix:phoenix@localhost:5432/phoenix?sslmode=disable"), "Database URL")
		migrationsDir = flag.String("dir", "migrations", "Migrations directory")
		command      = flag.String("cmd", "up", "Command: up, down, status, create")
		name         = flag.String("name", "", "Migration name (for create command)")
	)
	flag.Parse()

	switch *command {
	case "up":
		runMigrations(*dbURL, *migrationsDir, true)
	case "down":
		runMigrations(*dbURL, *migrationsDir, false)
	case "status":
		showStatus(*dbURL, *migrationsDir)
	case "create":
		createMigration(*migrationsDir, *name)
	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}

func runMigrations(dbURL, migrationsDir string, up bool) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("%sError connecting to database: %v%s", colorRed, err, colorReset)
	}
	defer db.Close()

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		log.Fatalf("%sError creating migrations table: %v%s", colorRed, err, colorReset)
	}

	// Get all migration files
	migrations, err := getMigrations(migrationsDir)
	if err != nil {
		log.Fatalf("%sError reading migrations: %v%s", colorRed, err, colorReset)
	}

	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		log.Fatalf("%sError getting applied migrations: %v%s", colorRed, err, colorReset)
	}

	if up {
		// Apply pending migrations
		pending := getPendingMigrations(migrations, applied)
		if len(pending) == 0 {
			fmt.Printf("%s✓ Database is up to date%s\n", colorGreen, colorReset)
			return
		}

		fmt.Printf("%sFound %d pending migrations%s\n", colorBlue, len(pending), colorReset)
		for _, m := range pending {
			if err := applyMigration(db, m); err != nil {
				log.Fatalf("%sError applying migration %s: %v%s", colorRed, m.Version, err, colorReset)
			}
			fmt.Printf("%s✓ Applied migration %s: %s%s\n", colorGreen, m.Version, m.Name, colorReset)
		}
	} else {
		// Rollback last migration
		if len(applied) == 0 {
			fmt.Printf("%s✓ No migrations to rollback%s\n", colorYellow, colorReset)
			return
		}

		lastApplied := applied[len(applied)-1]
		fmt.Printf("%sRolling back migration %s%s\n", colorYellow, lastApplied, colorReset)
		
		// For now, we don't support down migrations
		fmt.Printf("%s⚠ Down migrations not implemented. Please manually rollback if needed.%s\n", colorYellow, colorReset)
	}
}

func createMigrationsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`
	_, err := db.Exec(query)
	return err
}

func getMigrations(dir string) ([]Migration, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		parts := strings.SplitN(file.Name(), "_", 2)
		if len(parts) != 2 {
			continue
		}

		version := parts[0]
		name := strings.TrimSuffix(parts[1], ".sql")
		filePath := filepath.Join(dir, file.Name())

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", filePath, err)
		}

		migrations = append(migrations, Migration{
			Version:  version,
			Name:     name,
			FilePath: filePath,
			SQL:      string(content),
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func getAppliedMigrations(db *sql.DB) ([]string, error) {
	query := "SELECT version FROM schema_migrations ORDER BY version"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, rows.Err()
}

func getPendingMigrations(all []Migration, applied []string) []Migration {
	appliedMap := make(map[string]bool)
	for _, v := range applied {
		appliedMap[v] = true
	}

	var pending []Migration
	for _, m := range all {
		if !appliedMap[m.Version] {
			pending = append(pending, m)
		}
	}

	return pending
}

func applyMigration(db *sql.DB, m Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration
	if _, err := tx.Exec(m.SQL); err != nil {
		return fmt.Errorf("executing migration: %w", err)
	}

	// Record migration
	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", m.Version); err != nil {
		return fmt.Errorf("recording migration: %w", err)
	}

	return tx.Commit()
}

func showStatus(dbURL, migrationsDir string) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("%sError connecting to database: %v%s", colorRed, err, colorReset)
	}
	defer db.Close()

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		log.Fatalf("%sError creating migrations table: %v%s", colorRed, err, colorReset)
	}

	// Get all migrations
	migrations, err := getMigrations(migrationsDir)
	if err != nil {
		log.Fatalf("%sError reading migrations: %v%s", colorRed, err, colorReset)
	}

	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		log.Fatalf("%sError getting applied migrations: %v%s", colorRed, err, colorReset)
	}

	appliedMap := make(map[string]bool)
	for _, v := range applied {
		appliedMap[v] = true
	}

	fmt.Println("\nMigration Status:")
	fmt.Println("================")
	for _, m := range migrations {
		status := colorRed + "✗ Pending" + colorReset
		if appliedMap[m.Version] {
			status = colorGreen + "✓ Applied" + colorReset
		}
		fmt.Printf("%s %s: %s\n", status, m.Version, m.Name)
	}

	pending := getPendingMigrations(migrations, applied)
	fmt.Printf("\nTotal: %d migrations, %d applied, %d pending\n", 
		len(migrations), len(applied), len(pending))
}

func createMigration(dir, name string) {
	if name == "" {
		log.Fatal("Migration name is required")
	}

	// Generate timestamp-based version
	version := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s.sql", version, strings.ReplaceAll(name, " ", "_"))
	filepath := filepath.Join(dir, filename)

	template := fmt.Sprintf(`-- %s
-- %s

-- Add your migration SQL here

-- To rollback (not automatically supported):
-- Add rollback SQL as comments
`, filename, name)

	if err := ioutil.WriteFile(filepath, []byte(template), 0644); err != nil {
		log.Fatalf("%sError creating migration: %v%s", colorRed, err, colorReset)
	}

	fmt.Printf("%s✓ Created migration: %s%s\n", colorGreen, filepath, colorReset)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}