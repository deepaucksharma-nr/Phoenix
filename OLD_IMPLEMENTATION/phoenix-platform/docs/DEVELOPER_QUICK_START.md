# Phoenix Platform Developer Quick Start Guide

This guide will help you get the Phoenix Platform up and running on your local development environment.

## Prerequisites

- Go 1.24+ installed
- Docker and Docker Compose installed
- PostgreSQL 14+ (or Docker)
- Make utility
- Git

## Quick Start (5 minutes)

### 1. Clone and Build

```bash
# Clone the repository (if not already done)
git clone <repository-url>
cd phoenix-platform

# Build all services
make build-controller build-generator

# Verify builds
ls -la build/
```

### 2. Start PostgreSQL

```bash
# Using Docker (recommended)
docker run --name phoenix-db \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  -d postgres:14

# Wait for PostgreSQL to be ready
sleep 5

# Create test database
docker exec phoenix-db psql -U postgres -c "CREATE DATABASE phoenix_test"
```

### 3. Run Integration Tests

```bash
# Set database URL (optional, defaults to localhost)
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/phoenix_test?sslmode=disable"

# Run integration tests
make test-integration
```

### 4. Start Services

```bash
# Terminal 1: Start Experiment Controller
./build/experiment-controller

# Terminal 2: Start Config Generator
./build/config-generator

# Terminal 3: Test the services
curl http://localhost:8082/health  # Generator health check
curl http://localhost:8081/metrics  # Controller metrics
```

## Development Workflow

### Running Services with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Testing Experiment Creation

```bash
# Using grpcurl (install: brew install grpcurl)
grpcurl -plaintext -d '{
  "name": "Test Experiment",
  "description": "My first experiment",
  "baseline_pipeline": "process-baseline-v1",
  "candidate_pipeline": "process-priority-filter-v1",
  "target_nodes": {"node1": "host1", "node2": "host2"}
}' localhost:50051 phoenix.v1.ExperimentService/CreateExperiment
```

### Config Generator API

```bash
# List available templates
curl http://localhost:8082/templates

# Generate configuration
curl -X POST http://localhost:8082/generate \
  -H "Content-Type: application/json" \
  -d '{
    "template": "process-baseline-v1",
    "variables": {
      "EXPERIMENT_ID": "test-123",
      "NAMESPACE": "phoenix-system"
    }
  }'
```

## Project Structure

```
phoenix-platform/
├── cmd/
│   ├── controller/        # Experiment controller service
│   │   ├── main.go
│   │   └── internal/
│   │       ├── controller/
│   │       ├── grpc/
│   │       └── store/
│   └── generator/         # Config generator service
│       └── main.go
├── pkg/
│   ├── api/              # Proto-generated code
│   ├── models/           # Data models
│   ├── generator/        # Generator logic
│   └── interfaces/       # Shared interfaces
├── pipelines/
│   └── templates/        # Pipeline templates
├── scripts/              # Utility scripts
├── test/
│   └── integration/      # Integration tests
└── build/               # Compiled binaries
```

## Common Development Tasks

### Add a New Pipeline Template

1. Create template file in `pipelines/templates/`
2. Use YAML format with Go template syntax
3. Restart generator service

### Run Specific Tests

```bash
# Run only controller tests
go test ./cmd/controller/...

# Run with verbose output
go test -v ./cmd/controller/internal/controller/...

# Run integration tests with specific test
go test -tags=integration -v -run TestExperimentLifecycle ./cmd/controller/internal/controller/
```

### View Service Logs

```bash
# Controller logs (if running directly)
./build/experiment-controller 2>&1 | tee controller.log

# With Docker Compose
docker-compose logs -f experiment-controller
```

### Database Operations

```bash
# Connect to database
docker exec -it phoenix-db psql -U postgres -d phoenix_test

# View experiments
SELECT id, name, status, created_at FROM experiments;

# Clean test data
DELETE FROM experiments WHERE id LIKE 'test-%';
```

## Debugging Tips

### Port Already in Use

```bash
# Find process using port
lsof -i :8082  # Generator port
lsof -i :50051 # Controller gRPC port

# Kill process
kill -9 <PID>
```

### Database Connection Issues

```bash
# Test PostgreSQL connection
pg_isready -h localhost -p 5432

# Check Docker container
docker ps | grep phoenix-db
docker logs phoenix-db
```

### Build Issues

```bash
# Clean build cache
go clean -cache
make clean

# Update dependencies
go mod tidy
go mod download
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/phoenix?sslmode=disable` | Controller database |
| `GRPC_PORT` | `:50051` | Controller gRPC port |
| `HTTP_PORT` | `:8081` | Controller metrics port |
| `GENERATOR_PORT` | `:8082` | Generator HTTP port |
| `LOG_LEVEL` | `info` | Logging level |

## Next Steps

1. **Explore the API**: Use grpcurl or Postman to interact with services
2. **Create Experiments**: Test the A/B testing workflow
3. **Monitor Metrics**: Access Prometheus metrics at http://localhost:8081/metrics
4. **Customize Pipelines**: Create your own pipeline templates
5. **Run Load Tests**: Use the simulator to generate test data

## Troubleshooting

### Integration Tests Fail

1. Ensure PostgreSQL is running
2. Check database connectivity
3. Verify migrations ran successfully
4. Check test database exists

### Services Won't Start

1. Check all ports are free
2. Verify binaries are built
3. Check PostgreSQL connection
4. Review service logs for errors

### Generator Template Issues

1. Validate YAML syntax
2. Check template variables
3. Ensure templates directory exists
4. Review generator logs

## Getting Help

- Check `docs/` directory for detailed documentation
- Review integration tests for usage examples
- Examine service logs for error details
- Use `--help` flag on binaries for options

## Contributing

1. Create feature branch
2. Add tests for new functionality
3. Ensure all tests pass
4. Update documentation
5. Submit pull request

Happy developing with Phoenix Platform! 🚀