# Phoenix Platform - E2E Tests

## Overview

End-to-end tests for the Phoenix Platform's 70% cost reduction observability system. Tests validate agent-based architecture, A/B experiments, and real-time cost monitoring.

## Test Files

- **`simple_e2e_test.go`** - Basic E2E tests covering:
  - Service health checks
  - Experiment creation
  - Pipeline template listing
  - Configuration generation

- **`experiment_workflow_test.go`** - Comprehensive workflow tests covering:
  - Complete experiment lifecycle
  - Pipeline template validation
  - Error handling scenarios
  - Service integration

## Dependencies

### Go Modules
The E2E tests depend on:
- `github.com/stretchr/testify` - Test assertions
- Phoenix platform packages from `../../pkg`

### Required Services
The tests expect these services to be running:
- **Phoenix API** (default: http://localhost:8080)
- **PostgreSQL** database
- **Prometheus** (optional: http://localhost:9090)
- **Pushgateway** (optional: http://localhost:9091)

### Environment Variables
Configure these in your environment or `.env` file:
```bash
# Service URLs (optional - defaults provided)
API_URL=http://localhost:8080
DATABASE_URL=postgresql://phoenix:phoenix@localhost:5432/phoenix_test?sslmode=disable

# Authentication
JWT_SECRET=test-secret-key

# Optional monitoring
PROMETHEUS_URL=http://localhost:9090
PUSHGATEWAY_URL=http://localhost:9091
```

## Running the Tests

### Prerequisites
1. Ensure Go 1.22+ is installed
2. Start required services (or use docker-compose)
3. Set environment variables
4. Run database migrations

### Run All E2E Tests
```bash
# From project root
make test-e2e

# Or directly
cd tests/e2e
go test -tags e2e -v
```

### Run Specific Test
```bash
# Run simple E2E test
go test -tags e2e -v -run TestSimpleE2E

# Run experiment workflow test
go test -tags e2e -v -run TestExperimentWorkflowE2E
```

### Skip in Short Mode
```bash
# E2E tests are skipped with -short flag
go test -short
```

## Contracts

The E2E tests validate against these contracts:

### OpenAPI Contracts
- `pkg/contracts/openapi/control-api.yaml` - REST API specifications

### Protocol Buffer Contracts
- `pkg/contracts/proto/v1/common.proto` - Common types
- `pkg/contracts/proto/v1/controller.proto` - Controller service
- `pkg/contracts/proto/v1/experiment.proto` - Experiment definitions
- `pkg/contracts/proto/v1/generator.proto` - Generator service

## Test Structure

All E2E tests follow this pattern:
1. **Setup** - Check service health
2. **Execute** - Perform operations
3. **Validate** - Assert expected behavior
4. **Cleanup** - Clean test data (if needed)

## Validation Script

Run the validation script to check all dependencies:
```bash
# Windows
validate_e2e.bat

# Linux/Mac (if available)
./validate_e2e_dependencies.sh
```

## Troubleshooting

### Tests Won't Compile
```bash
# Ensure go.mod is updated
cd tests/e2e
go mod tidy

# Verify workspace includes e2e tests
cat ../../go.work | grep "tests/e2e"
```

### Services Not Ready
```bash
# Check Phoenix API health
curl http://localhost:8080/health

# Test agent authentication
curl -H "X-Agent-Host-ID: test-agent" http://localhost:8080/api/v1/agent/tasks

# Check WebSocket endpoint
websocketd --port=8080 --dir=/tmp echo
```

### Database Connection Issues
```bash
# Verify PostgreSQL is running
docker ps | grep postgres

# Test connection
psql $DATABASE_URL -c "SELECT 1"
```

## CI/CD Integration

The E2E tests are integrated into the CI pipeline:
1. Services are started in Docker
2. Database migrations are applied
3. E2E tests run with proper tags
4. Results are reported

See `.github/workflows/e2e-tests.yml` for CI configuration.
