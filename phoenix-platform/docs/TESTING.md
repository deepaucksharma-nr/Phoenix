# Testing Guide

This document describes how to run tests for the Phoenix Platform.

## Test Structure

The Phoenix Platform includes several types of tests:

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test component interactions and workflows
- **End-to-End Tests**: Test complete user scenarios (planned)

## Integration Tests

### Prerequisites

Before running integration tests, ensure you have:

1. **Go 1.21 or later**
2. **PostgreSQL server running** with the following default configuration:
   - Host: `localhost`
   - Port: `5432`
   - User: `phoenix`
   - Password: `phoenix`
   - The test runner will create a test database automatically

### Running Integration Tests

#### Using Make

The simplest way to run integration tests:

```bash
make test-integration
```

#### Using the Test Runner Script

For more control, use the test runner script directly:

```bash
# Run all integration tests
./scripts/run-integration-tests.sh

# Show help
./scripts/run-integration-tests.sh --help
```

#### Using Go Test Directly

If you prefer to run tests manually:

```bash
# Ensure PostgreSQL is available and set environment variables
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=phoenix
export POSTGRES_PASSWORD=phoenix
export TEST_DATABASE_NAME=phoenix_test

# Run tests
go test -tags=integration -v ./test/integration/...
```

### Environment Variables

You can customize the test environment using these variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_HOST` | `localhost` | PostgreSQL server hostname |
| `POSTGRES_PORT` | `5432` | PostgreSQL server port |
| `POSTGRES_USER` | `phoenix` | PostgreSQL username |
| `POSTGRES_PASSWORD` | `phoenix` | PostgreSQL password |
| `TEST_DATABASE_NAME` | `phoenix_test` | Name of the test database |

Example with custom settings:

```bash
POSTGRES_HOST=my-db-server POSTGRES_PASSWORD=secret ./scripts/run-integration-tests.sh
```

## Test Coverage

### Experiment Controller Tests

The integration tests cover:

- **Experiment Lifecycle**: Create → Initialize → Run → Analyze → Complete
- **State Transitions**: Valid and invalid state changes
- **Database Operations**: CRUD operations with PostgreSQL
- **Scheduler Functionality**: Automatic state transitions
- **Error Handling**: Invalid inputs and error recovery
- **Concurrent Operations**: Multiple experiments running in parallel

### Config Generator Tests

The integration tests cover:

- **Configuration Generation**: OTel collector configs and Kubernetes manifests
- **Pipeline Types**: Different processor combinations (filter, aggregate, sample)
- **Variable Substitution**: Environment variable replacement
- **HTTP API**: REST endpoints for config generation
- **Template Engine**: Pipeline template processing

### End-to-End Workflow Tests

The integration tests include:

- **Complete Experiment Flow**: From creation through config generation to completion
- **Service Integration**: Experiment Controller + Config Generator working together
- **Error Recovery**: Cancellation and failure scenarios
- **Parallel Processing**: Multiple experiments with config generation

## Test Database

The integration tests automatically:

1. **Create a test database** before running tests
2. **Run database migrations** to set up the schema
3. **Clean up test data** between test runs
4. **Drop the test database** after tests complete

The test database is completely separate from any development or production databases.

## Troubleshooting

### PostgreSQL Connection Issues

If you see database connection errors:

1. **Check if PostgreSQL is running**:
   ```bash
   # On macOS with Homebrew
   brew services list | grep postgresql
   
   # On Linux
   systemctl status postgresql
   ```

2. **Verify connection settings**:
   ```bash
   psql -h localhost -p 5432 -U phoenix -d postgres
   ```

3. **Create the phoenix user if it doesn't exist**:
   ```sql
   CREATE USER phoenix WITH PASSWORD 'phoenix';
   ALTER USER phoenix CREATEDB;
   ```

### Permission Issues

If you see permission errors with the test script:

```bash
chmod +x ./scripts/run-integration-tests.sh
```

### Go Module Issues

If you see module-related errors:

```bash
go mod tidy
go mod download
```

## Test Results

Successful test runs will show output like:

```
[INFO] Phoenix Platform Integration Test Runner
[INFO] ======================================
[INFO] Checking prerequisites...
[SUCCESS] Go version: 1.21.0
[INFO] Checking PostgreSQL connectivity...
[SUCCESS] PostgreSQL connection successful
[INFO] Setting up test environment...
[SUCCESS] Environment variables set
[INFO] Running integration tests...
[INFO] Building controller...
[INFO] Building generator...
[SUCCESS] Build successful
[INFO] Executing integration tests...

=== RUN   TestExperimentControllerIntegration
=== RUN   TestExperimentControllerIntegration/CreateExperiment
=== RUN   TestExperimentControllerIntegration/ExperimentStateTransitions
=== RUN   TestExperimentControllerIntegration/ExperimentScheduler
=== RUN   TestExperimentControllerIntegration/ListExperiments
=== RUN   TestExperimentControllerIntegration/ExperimentCancellation
--- PASS: TestExperimentControllerIntegration (10.23s)

[SUCCESS] All integration tests passed!
```

## Adding New Tests

When adding new integration tests:

1. **Place test files** in `test/integration/`
2. **Use build tag**: Add `// +build integration` at the top
3. **Follow naming convention**: `*_test.go`
4. **Use testify framework**: For assertions and test structure
5. **Clean up resources**: Ensure tests clean up after themselves
6. **Document test cases**: Add clear descriptions of what each test validates

Example test structure:

```go
// +build integration

package integration

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Setup
    // ...
    
    t.Run("SpecificScenario", func(t *testing.T) {
        // Test implementation
        // ...
        
        // Assertions
        require.NoError(t, err)
        assert.Equal(t, expected, actual)
    })
    
    // Cleanup
    defer CleanupTestData(t)
}
```