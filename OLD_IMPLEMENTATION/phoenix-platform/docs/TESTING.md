# Testing Guide

This document provides comprehensive guidance for testing the Phoenix Platform, including unit tests, integration tests, end-to-end tests, and dashboard testing.

## Overview

The Phoenix Platform follows a comprehensive testing strategy to ensure code quality and reliability:

- **Unit Tests**: Test individual components in isolation (services, utilities, functions)
- **Integration Tests**: Test component interactions and service communication
- **End-to-End Tests**: Test complete user workflows in a Kubernetes environment
- **Dashboard Tests**: Test React components and state management with Vitest
- **Performance Tests**: Test system performance under load (coming soon)

## Quick Start

```bash
# Run all tests
make test

# Run specific test types
make test-unit           # Unit tests only
make test-integration    # Integration tests only
make test-e2e           # End-to-end tests
make test-dashboard     # Dashboard tests with coverage

# Generate coverage report
make coverage
```

## Unit Testing

### Guidelines

1. **Test Coverage**: Aim for minimum 80% coverage on business logic
2. **Test Naming**: Use descriptive names following `Test<Function>_<Scenario>_<ExpectedResult>`
3. **Test Structure**: Follow AAA pattern (Arrange, Act, Assert)
4. **Mocking**: Use interfaces for dependencies to enable mocking

### Running Unit Tests

```bash
# Run all unit tests
make test-unit

# Run tests for specific package
go test -v ./pkg/store/...

# Run with coverage
go test -v -race -coverprofile=coverage.out ./pkg/...
go tool cover -html=coverage.out -o coverage.html
```

### Example Unit Test

```go
package store

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/mock"
)

func TestExperimentStore_Create_Success(t *testing.T) {
    // Arrange
    mockDB := new(MockDatabase)
    store := NewExperimentStore(mockDB)
    
    experiment := &Experiment{
        Name: "test-experiment",
        Type: "process-optimization",
    }
    
    mockDB.On("Insert", mock.Anything).Return(nil)
    
    // Act
    err := store.Create(experiment)
    
    // Assert
    require.NoError(t, err)
    assert.NotEmpty(t, experiment.ID)
    mockDB.AssertExpectations(t)
}
```

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

## Dashboard Testing

The Phoenix dashboard uses Vitest for testing React components and TypeScript code.

### Running Dashboard Tests

```bash
# Run all dashboard tests
make test-dashboard

# Run tests in watch mode
cd dashboard && npm test

# Run with coverage
cd dashboard && npm run test:coverage

# Run UI test interface
cd dashboard && npm run test:ui
```

### Test Structure

Dashboard tests are organized by feature:

```
dashboard/src/
├── components/
│   ├── Auth/__tests__/
│   ├── Metrics/__tests__/
│   └── PipelineBuilder/__tests__/
├── hooks/__tests__/
├── store/__tests__/
└── test/
    ├── setup.ts        # Test configuration
    └── utils.tsx       # Test utilities
```

### Example Component Test

```typescript
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MetricCard } from '../MetricCard';

describe('MetricCard', () => {
  it('displays metric value and handles click', async () => {
    const handleClick = vi.fn();
    
    render(
      <MetricCard
        title="Cost Reduction"
        value="65%"
        trend="up"
        onClick={handleClick}
      />
    );
    
    expect(screen.getByText('Cost Reduction')).toBeInTheDocument();
    expect(screen.getByText('65%')).toBeInTheDocument();
    
    fireEvent.click(screen.getByRole('button'));
    expect(handleClick).toHaveBeenCalledOnce();
  });
});
```

### Example Store Test

```typescript
import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, beforeEach } from 'vitest';
import { useExperimentStore } from '../useExperimentStore';

describe('useExperimentStore', () => {
  beforeEach(() => {
    const { result } = renderHook(() => useExperimentStore());
    act(() => {
      result.current.clearExperiments();
    });
  });

  it('adds and updates experiments', () => {
    const { result } = renderHook(() => useExperimentStore());
    
    act(() => {
      result.current.addExperiment({
        id: '1',
        name: 'Test Experiment',
        status: 'running'
      });
    });
    
    expect(result.current.experiments).toHaveLength(1);
    expect(result.current.experiments[0].status).toBe('running');
  });
});
```

## End-to-End Testing

E2E tests validate complete workflows in a real Kubernetes environment.

### Prerequisites

1. **Kind cluster**: Local Kubernetes cluster
2. **Phoenix deployed**: All services running
3. **Test data**: Sample experiments and pipelines

### Running E2E Tests

```bash
# Setup test environment
make cluster-up
make deploy

# Run E2E tests
make test-e2e

# Cleanup
make cluster-down
```

### E2E Test Example

```go
// +build e2e

package e2e

import (
    "testing"
    "time"
    "github.com/stretchr/testify/require"
)

func TestCompleteExperimentWorkflow(t *testing.T) {
    // Create experiment via API
    experiment := createExperiment(t, "e2e-test-experiment")
    
    // Wait for initialization
    waitForState(t, experiment.ID, "initialized", 30*time.Second)
    
    // Verify OTel collectors deployed
    collectors := getDeployedCollectors(t, experiment.ID)
    require.Len(t, collectors, 2) // baseline and candidate
    
    // Wait for completion
    waitForState(t, experiment.ID, "completed", 5*time.Minute)
    
    // Verify metrics collected
    metrics := getExperimentMetrics(t, experiment.ID)
    require.NotEmpty(t, metrics.CostReduction)
}
```

## CI/CD Integration

Tests are automatically run in CI/CD pipelines:

### GitHub Actions

```yaml
name: Test
on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: make test-unit
      
  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: phoenix
    steps:
      - uses: actions/checkout@v3
      - run: make test-integration
      
  dashboard-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - run: cd dashboard && npm ci
      - run: make test-dashboard
```

## Test Coverage Requirements

| Component | Minimum Coverage | Current |
|-----------|-----------------|---------|
| Core Services | 80% | TBD |
| API Handlers | 70% | TBD |
| Dashboard | 70% | TBD |
| Utilities | 90% | TBD |

## Adding New Tests

### Integration Tests

1. **Place test files** in `test/integration/`
2. **Use build tag**: Add `// +build integration` at the top
3. **Follow naming convention**: `*_test.go`
4. **Use testify framework**: For assertions and test structure
5. **Clean up resources**: Ensure tests clean up after themselves
6. **Document test cases**: Add clear descriptions of what each test validates

### Unit Tests

1. **Co-locate with code**: Place in same package as code being tested
2. **Use table-driven tests**: For testing multiple scenarios
3. **Mock external dependencies**: Use interfaces and mock implementations
4. **Test edge cases**: Include error scenarios and boundary conditions

### Dashboard Tests

1. **Use Testing Library**: Follow testing-library best practices
2. **Test user interactions**: Focus on user behavior, not implementation
3. **Mock API calls**: Use MSW or vi.mock for API mocking
4. **Test accessibility**: Include ARIA and keyboard navigation tests

## Best Practices

1. **Fast Tests**: Keep unit tests under 100ms each
2. **Isolated Tests**: No dependencies between tests
3. **Descriptive Names**: Test names should explain what and why
4. **Deterministic**: Same result every time
5. **Readable**: Tests serve as documentation
6. **Maintainable**: Refactor tests along with code

## Debugging Failed Tests

### Verbose Output
```bash
go test -v -run TestName ./pkg/...
```

### Debug Logging
```go
t.Logf("Debug: experiment state = %v", experiment.State)
```

### Interactive Debugging
```bash
# Install Delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug specific test
dlv test ./pkg/store -- -test.run TestExperimentStore
```