# Phoenix Platform Development Guide

This guide covers the complete development workflow, from setting up your local environment to contributing code following best practices.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Architecture Overview](#architecture-overview)
4. [Project Structure](#project-structure)
5. [Development Workflow](#development-workflow)
6. [Services](#services)
7. [API Examples](#api-examples)
8. [Coding Standards](#coding-standards)
9. [Testing Guidelines](#testing-guidelines)
10. [Debugging](#debugging)
11. [Troubleshooting](#troubleshooting)
12. [Contributing](#contributing)

## Prerequisites

- Docker and Docker Compose (v2.0+)
- Go 1.21+
- Node.js 18+ and npm
- Make
- Git
- Kubernetes 1.28+ (kind or minikube for local development) - optional

## Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/phoenix-platform/phoenix.git
   cd phoenix/phoenix-platform
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start the development environment**
   ```bash
   ./scripts/dev-environment.sh up
   ```

4. **Run database migrations**
   ```bash
   ./scripts/dev-environment.sh migrate
   ```

5. **Access the services**
   - API Gateway: http://localhost:8080
   - Dashboard: http://localhost:5173
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000 (admin/admin)

## Architecture Overview

The local development environment runs all Phoenix Platform services in Docker containers:

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Dashboard     │────▶│  API Gateway    │────▶│    Services     │
│  (React/Vite)   │     │  (REST/gRPC)    │     │   (gRPC APIs)   │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                │                         │
                                │                         ▼
                                │                ┌─────────────────┐
                                │                │   PostgreSQL    │
                                │                │     Redis       │
                                │                │   Prometheus    │
                                └───────────────▶│    Grafana      │
                                                 └─────────────────┘
```

## Project Structure

```
phoenix-platform/
├── cmd/                    # Application entry points
│   ├── api-gateway/       # API Gateway server
│   ├── control-service/   # Control service
│   ├── controller/        # Experiment controller
│   ├── generator/         # Config generator
│   └── simulator/         # Process simulator
├── pkg/                   # Public packages
│   ├── api/              # API business logic
│   ├── auth/             # Authentication
│   ├── generator/        # Config generation
│   ├── interfaces/       # Service interfaces
│   └── models/           # Data models
├── internal/              # Private packages
├── operators/             # Kubernetes operators
│   ├── pipeline/         # Pipeline operator
│   └── loadsim/          # Load simulation operator
├── dashboard/             # React frontend
├── pipelines/            # Pipeline templates
├── k8s/                  # Kubernetes manifests
├── helm/                 # Helm charts
├── migrations/           # Database migrations
├── scripts/              # Development scripts
└── docs/                 # Documentation
```

## Development Workflow

### Running Services

```bash
# Start all services
./scripts/dev-environment.sh up

# Stop all services
./scripts/dev-environment.sh down

# View logs
./scripts/dev-environment.sh logs

# Follow logs
./scripts/dev-environment.sh logs-f

# Restart a specific service
./scripts/dev-environment.sh restart experiment-controller
```

### Building and Testing

```bash
# Build all services
make build

# Run unit tests
make test

# Run integration tests
./scripts/dev-environment.sh test

# Validate code structure
make validate

# Generate proto code
make generate-proto
```

### Database Management

```bash
# Run migrations
./scripts/dev-environment.sh migrate

# Connect to PostgreSQL
docker-compose -f docker-compose.dev.yml exec postgres psql -U phoenix

# View experiment data
SELECT * FROM experiments;
```

### Working with Kubernetes (Optional)

1. **Local Kubernetes Cluster**
   ```bash
   make cluster-up
   ```

2. **Deploy to Local Cluster**
   ```bash
   make deploy-dev
   ```

3. **Port Forwarding**
   ```bash
   # API
   kubectl port-forward svc/phoenix-api 8080:8080

   # Dashboard
   kubectl port-forward svc/phoenix-dashboard 3000:80
   ```

## Services

### Core Services

1. **API Gateway** (Port 8080)
   - REST API endpoints
   - Routes to gRPC services
   - Authentication/authorization
   - Request logging and metrics

2. **Experiment Controller** (Port 50051)
   - Manages experiment lifecycle
   - State machine implementation
   - Database persistence

3. **Config Generator** (Port 50052)
   - Generates OTel configurations
   - Template management
   - GitOps integration

4. **Control Service** (Port 50053)
   - Traffic management
   - Drift detection
   - Control signals

### Supporting Services

- **PostgreSQL**: Primary database
- **Redis**: Caching and session storage
- **Prometheus**: Metrics collection
- **Grafana**: Metrics visualization
- **NATS**: Event bus (future use)

## API Examples

### Create an Experiment

```bash
curl -X POST http://localhost:8080/api/v1/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Cost Optimization Test",
    "description": "Test 30% cost reduction",
    "baseline_pipeline_id": "baseline-v1",
    "candidate_pipeline_id": "optimized-v1",
    "traffic_percentage": 10,
    "target_services": ["service-a", "service-b"]
  }'
```

### Generate Configuration

```bash
curl -X POST http://localhost:8080/api/v1/generate \
  -H "Content-Type: application/json" \
  -d '{
    "pipeline_id": "current-pipeline",
    "goals": [{
      "type": 1,
      "target_value": 0.3,
      "priority": 10
    }],
    "constraints": [{
      "type": 1,
      "metric": "minimum_retention",
      "min_value": 0.95
    }]
  }'
```

### Execute Control Signal

```bash
curl -X POST http://localhost:8080/api/v1/control/signals \
  -H "Content-Type: application/json" \
  -d '{
    "experiment_id": "exp-123",
    "type": "traffic_split",
    "action": {
      "baseline_pipeline_id": "baseline-v1",
      "candidate_pipeline_id": "candidate-v1",
      "candidate_percentage": 25
    },
    "reason": "Increase traffic to candidate"
  }'
```

## Coding Standards

### Go Code

1. **Style Guide**
   - Follow [Effective Go](https://golang.org/doc/effective_go.html)
   - Use `gofmt` and `golangci-lint`
   - Package names should be lowercase, single-word

2. **Project Conventions**
   ```go
   // Package comment
   // Package api provides the core business logic for experiments.
   package api

   // Exported types need comments
   // ExperimentService handles experiment lifecycle management.
   type ExperimentService struct {
       store Store
       log   *zap.Logger
   }
   ```

3. **Error Handling**
   ```go
   // Wrap errors with context
   if err != nil {
       return fmt.Errorf("failed to create experiment: %w", err)
   }
   ```

### TypeScript/React Code

1. **Style Guide**
   - Use TypeScript strict mode
   - Follow React hooks best practices
   - Use functional components

2. **Component Structure**
   ```typescript
   interface Props {
       experiment: Experiment;
       onUpdate: (exp: Experiment) => void;
   }

   export const ExperimentCard: React.FC<Props> = ({ experiment, onUpdate }) => {
       // Component logic
   };
   ```

### Commit Messages

Follow conventional commits:
```
feat: add experiment comparison view
fix: resolve memory leak in collector
docs: update pipeline configuration guide
chore: upgrade dependencies
```

## Testing

For comprehensive testing guidance including unit tests, integration tests, E2E tests, and dashboard testing, see [TESTING.md](TESTING.md).

Quick reference:
```bash
# Run all tests
make test

# Run specific test types
make test-unit
make test-integration
make test-e2e
make test-dashboard

# Generate coverage report
make coverage
```

## Debugging

### Service Health Checks

All services expose health endpoints:

```bash
# API Gateway
curl http://localhost:8080/health

# Experiment Controller
curl http://localhost:8081/health

# Config Generator
curl http://localhost:8082/health

# Control Service
curl http://localhost:8083/health
```

### API Debugging

1. **Enable Debug Logging**
   ```bash
   LOG_LEVEL=debug go run cmd/api/main.go
   ```

2. **Use Delve Debugger**
   ```bash
   dlv debug cmd/api/main.go
   ```

3. **Inspect gRPC Calls**
   ```bash
   grpcurl -plaintext localhost:50051 list
   ```

### Accessing Service Shells

```bash
# Execute shell in a container
./scripts/dev-environment.sh exec experiment-controller

# Or using docker-compose directly
docker-compose -f docker-compose.dev.yml exec experiment-controller /bin/sh
```

### Kubernetes Debugging

1. **View Pod Logs**
   ```bash
   kubectl logs -f deployment/phoenix-api
   ```

2. **Exec into Pod**
   ```bash
   kubectl exec -it deployment/phoenix-api -- /bin/sh
   ```

3. **Describe Resources**
   ```bash
   kubectl describe phoenixexperiment my-experiment
   ```

### Dashboard Debugging

1. **React Developer Tools**
   - Install browser extension
   - Inspect component props and state

2. **Network Tab**
   - Monitor API calls
   - Check request/response payloads

## Troubleshooting

### Common Issues

1. **Port conflicts**
   - Check if ports are already in use: `lsof -i :8080`
   - Stop conflicting services or change ports in docker-compose.dev.yml

2. **Database connection errors**
   - Ensure PostgreSQL is healthy: `docker-compose ps postgres`
   - Check DATABASE_URL in .env matches container configuration

3. **Service discovery issues**
   - Services communicate using container names
   - Ensure all services are on the same Docker network

### Development Tips

1. **Hot Reload**: The dashboard supports hot reload. Changes to React code are reflected immediately.

2. **Proto Changes**: After modifying proto files:
   ```bash
   make generate-proto
   make build
   ./scripts/dev-environment.sh restart <affected-service>
   ```

3. **Database Schema Changes**:
   - Add migration file to `migrations/`
   - Run `./scripts/dev-environment.sh migrate`

4. **Monitoring**: 
   - View metrics in Prometheus: http://localhost:9090
   - Import dashboards in Grafana from `configs/monitoring/grafana/dashboards/`

## Contributing

### Pre-commit Checks

1. **Run Linters**
   ```bash
   make lint
   ```

2. **Format Code**
   ```bash
   make fmt
   ```

3. **Run Tests**
   ```bash
   make test
   ```

4. **Validate Structure**
   ```bash
   make validate
   ```

### Pull Request Process

1. Create feature branch from `main`
2. Make changes following coding standards
3. Add/update tests
4. Update documentation
5. Run `make pre-commit`
6. Submit PR with clear description

### Code Review Checklist

- [ ] Tests pass
- [ ] Code follows style guide
- [ ] Documentation updated
- [ ] No security vulnerabilities
- [ ] Performance impact considered
- [ ] Backwards compatibility maintained

## Cleanup

```bash
# Stop services and keep data
./scripts/dev-environment.sh down

# Stop services and remove all data
./scripts/dev-environment.sh clean
```

## Next Steps

- Review the [API Documentation](./API_REFERENCE.md)
- Check the [Architecture Guide](./ARCHITECTURE.md)
- Read about [Testing](./TESTING.md)
- See [Troubleshooting Guide](./TROUBLESHOOTING.md)