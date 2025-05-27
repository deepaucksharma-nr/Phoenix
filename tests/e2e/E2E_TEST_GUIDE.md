# Phoenix Platform E2E Test Guide

## Overview

This directory contains comprehensive end-to-end tests for the Phoenix Platform. These tests validate the entire system working together, including:

- Experiment lifecycle management
- Pipeline deployment and validation
- Agent integration and task execution
- Real-time metrics and analysis
- WebSocket communication
- Cost analysis and optimization
- Load and stress testing

## Test Structure

### Test Files

1. **simple_e2e_test.go** - Basic smoke tests to verify services are running
2. **experiment_workflow_test.go** - Tests the complete experiment workflow
3. **comprehensive_e2e_test.go** - Full platform testing including all features
4. **rest_api_test.go** - REST API endpoint testing

### Test Suites in Comprehensive Test

- **ExperimentLifecycle**: Complete experiment creation, execution, and analysis
- **PipelineDeployment**: Pipeline management and deployment workflows
- **AgentIntegration**: Agent registration, task polling, and metric reporting
- **MetricsAndAnalysis**: Metrics collection and KPI analysis
- **WebSocketRealTime**: Real-time updates via WebSocket
- **CostAnalysis**: Cost calculation and optimization recommendations
- **LoadAndStress**: High-volume concurrent operations

## Running Tests

### Prerequisites

- Go 1.19+
- Docker and Docker Compose
- PostgreSQL (via Docker)
- Make

### Quick Start

```bash
# Run all E2E tests
./run_e2e_tests.sh

# Run specific test suite
./run_e2e_tests.sh simple
./run_e2e_tests.sh workflow
./run_e2e_tests.sh comprehensive

# Run with custom configuration
export PHOENIX_API_URL=http://localhost:8080
export TEST_TIMEOUT=15m
./run_e2e_tests.sh
```

### Manual Test Execution

```bash
# Start infrastructure
docker-compose -f ../../docker-compose-infra.yml up -d

# Build services
cd ../../projects/phoenix-api && make build
cd ../../projects/phoenix-agent && make build

# Run migrations
cd ../../projects/phoenix-api
go run cmd/api/main.go migrate up

# Start services
./api &
cd ../phoenix-agent && ./phoenix-agent &

# Run tests
cd ../../tests/e2e
go test -v -tags=e2e -timeout=10m ./...
```

## Test Configuration

### Environment Variables

- `PHOENIX_API_URL` - Phoenix API URL (default: http://localhost:8080)
- `DATABASE_URL` - PostgreSQL connection string
- `TEST_TIMEOUT` - Test execution timeout (default: 10m)
- `CLEANUP` - Auto-cleanup after tests (default: true)

### Test Tags

Tests use the `e2e` build tag to separate from unit tests:

```go
// +build e2e
```

## Writing New E2E Tests

### Test Structure Example

```go
func TestNewFeatureE2E(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    // Setup
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    apiURL := getEnvOrDefault("PHOENIX_API_URL", "http://localhost:8080")
    
    // Wait for service
    waitForService(t, apiURL+"/health", 30*time.Second)
    
    // Test implementation
    t.Run("SubTest", func(t *testing.T) {
        // Test logic
    })
}
```

### Best Practices

1. **Isolation**: Each test should be independent and not rely on state from other tests
2. **Cleanup**: Always clean up resources (experiments, deployments) after tests
3. **Timeouts**: Use appropriate timeouts for async operations
4. **Assertions**: Use clear assertions with helpful error messages
5. **Logging**: Log important steps for debugging failed tests

## Debugging Failed Tests

### Common Issues

1. **Service Not Ready**
   - Increase wait timeouts in `waitForService()`
   - Check service logs for startup errors

2. **Database Connection**
   - Verify PostgreSQL is running: `docker ps`
   - Check migrations: `go run cmd/api/main.go migrate status`

3. **Port Conflicts**
   - Ensure ports 8080, 5432, 9090 are available
   - Stop conflicting services

4. **WebSocket Connection**
   - Verify WebSocket endpoint is accessible
   - Check for proxy/firewall issues

### Debug Mode

Enable verbose logging:

```bash
export DEBUG=true
export LOG_LEVEL=debug
./run_e2e_tests.sh
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: E2E Tests
on: [push, pull_request]

jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: phoenix
          POSTGRES_DB: phoenix_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.19'
    
    - name: Run E2E Tests
      env:
        DATABASE_URL: postgres://postgres:phoenix@localhost/phoenix_test?sslmode=disable
      run: |
        cd tests/e2e
        ./run_e2e_tests.sh
```

## Performance Benchmarks

Expected performance metrics:

- Experiment creation: < 100ms
- Pipeline deployment: < 500ms
- Metrics ingestion: > 10,000 metrics/second
- WebSocket latency: < 50ms
- API response time (p99): < 200ms

## Test Coverage

Current E2E test coverage:

- ✅ Experiment CRUD operations
- ✅ Experiment state transitions
- ✅ Pipeline template management
- ✅ Pipeline deployment and rollback
- ✅ Agent registration and heartbeat
- ✅ Task polling and execution
- ✅ Metrics collection and storage
- ✅ KPI calculation and analysis
- ✅ Real-time WebSocket updates
- ✅ Cost analysis and projections
- ✅ Load testing with concurrent operations
- ✅ Error handling and edge cases

## Contributing

When adding new E2E tests:

1. Follow the existing test structure
2. Add appropriate documentation
3. Ensure tests are idempotent
4. Update this guide if adding new test categories
5. Run full test suite before submitting PR