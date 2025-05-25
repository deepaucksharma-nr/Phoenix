// +build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// TestMain handles setup and teardown for integration tests
func TestMain(m *testing.M) {
	// Setup
	if err := setupIntegrationTests(); err != nil {
		log.Fatalf("Failed to setup integration tests: %v", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if err := teardownIntegrationTests(); err != nil {
		log.Printf("Failed to cleanup integration tests: %v", err)
	}

	os.Exit(code)
}

func setupIntegrationTests() error {
	log.Println("Setting up integration tests...")

	// Check if PostgreSQL is available
	if err := ensurePostgresAvailable(); err != nil {
		return fmt.Errorf("PostgreSQL not available: %w", err)
	}

	// Create test database if it doesn't exist
	if err := createTestDatabase(); err != nil {
		return fmt.Errorf("failed to create test database: %w", err)
	}

	log.Println("Integration test setup complete")
	return nil
}

func teardownIntegrationTests() error {
	log.Println("Cleaning up integration tests...")

	// Drop test database
	if err := dropTestDatabase(); err != nil {
		log.Printf("Warning: failed to drop test database: %v", err)
	}

	log.Println("Integration test cleanup complete")
	return nil
}

func ensurePostgresAvailable() error {
	// Try to connect to PostgreSQL server
	db, err := sql.Open("postgres", getPostgresConnectionURL())
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}

func createTestDatabase() error {
	// Connect to postgres database to create test database
	db, err := sql.Open("postgres", getPostgresConnectionURL())
	if err != nil {
		return err
	}
	defer db.Close()

	// Check if test database already exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)", getTestDatabaseName()).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		// Create test database
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", getTestDatabaseName()))
		if err != nil {
			return err
		}
		log.Printf("Created test database: %s", getTestDatabaseName())
	} else {
		log.Printf("Test database already exists: %s", getTestDatabaseName())
	}

	return nil
}

func dropTestDatabase() error {
	// Connect to postgres database to drop test database
	db, err := sql.Open("postgres", getPostgresConnectionURL())
	if err != nil {
		return err
	}
	defer db.Close()

	// Drop test database
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", getTestDatabaseName()))
	if err != nil {
		return err
	}

	log.Printf("Dropped test database: %s", getTestDatabaseName())
	return nil
}

func getPostgresConnectionURL() string {
	// Connect to the default postgres database for admin operations
	host := getEnvOrDefault("POSTGRES_HOST", "localhost")
	port := getEnvOrDefault("POSTGRES_PORT", "5432")
	user := getEnvOrDefault("POSTGRES_USER", "phoenix")
	password := getEnvOrDefault("POSTGRES_PASSWORD", "phoenix")
	
	return fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable", user, password, host, port)
}

func getTestDatabaseName() string {
	return getEnvOrDefault("TEST_DATABASE_NAME", "phoenix_test")
}

func getTestDatabaseURL() string {
	host := getEnvOrDefault("POSTGRES_HOST", "localhost")
	port := getEnvOrDefault("POSTGRES_PORT", "5432")
	user := getEnvOrDefault("POSTGRES_USER", "phoenix")
	password := getEnvOrDefault("POSTGRES_PASSWORD", "phoenix")
	dbName := getTestDatabaseName()
	
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Test utilities

// WaitForCondition waits for a condition to be true or times out
func WaitForCondition(condition func() bool, timeout time.Duration, checkInterval time.Duration) bool {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(checkInterval)
	}
	
	return false
}

// RetryOperation retries an operation with backoff
func RetryOperation(operation func() error, maxRetries int, baseDelay time.Duration) error {
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		if err := operation(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		
		// Exponential backoff
		delay := baseDelay * time.Duration(1<<uint(i))
		time.Sleep(delay)
	}
	
	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}

// CleanupTestData removes test data from the database
func CleanupTestData(t *testing.T) {
	db, err := sql.Open("postgres", getTestDatabaseURL())
	if err != nil {
		t.Logf("Warning: failed to connect to test database for cleanup: %v", err)
		return
	}
	defer db.Close()

	// Clean up test experiments
	_, err = db.Exec("DELETE FROM experiments WHERE id LIKE 'test-%' OR id LIKE 'grpc-%'")
	if err != nil {
		t.Logf("Warning: failed to clean up test experiments: %v", err)
	}
}

// CreateTestExperiment creates a test experiment with default values
func CreateTestExperiment(id string) map[string]interface{} {
	return map[string]interface{}{
		"id":                  id,
		"name":               fmt.Sprintf("Test Experiment %s", id),
		"description":        "Integration test experiment",
		"baseline_pipeline":  "process-baseline-v1",
		"candidate_pipeline": "process-priority-filter-v1",
		"target_hosts":       []string{"test-node-1", "test-node-2"},
		"duration":           "5m",
		"success_criteria": map[string]interface{}{
			"min_cardinality_reduction": 50.0,
			"max_cpu_overhead":          10.0,
			"max_memory_overhead":       15.0,
			"critical_process_coverage": 100.0,
		},
	}
}