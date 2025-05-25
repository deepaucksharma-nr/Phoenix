# Phoenix Platform Local Development Guide

This guide explains how to set up and run the Phoenix Platform locally for development.

## Prerequisites

- Docker and Docker Compose (v2.0+)
- Go 1.21+
- Node.js 18+ and npm
- Make
- Git

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

### Accessing Service Shells

```bash
# Execute shell in a container
./scripts/dev-environment.sh exec experiment-controller

# Or using docker-compose directly
docker-compose -f docker-compose.dev.yml exec experiment-controller /bin/sh
```

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

## Development Tips

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

## Testing

### Unit Tests
```bash
# Run all unit tests
make test

# Run tests for specific service
cd cmd/experiment-controller
go test ./...
```

### Integration Tests
```bash
# Ensure services are running
./scripts/dev-environment.sh up

# Run integration tests
./scripts/dev-environment.sh test
```

### Manual Testing
Use the provided API examples or import the Postman collection from `docs/postman/`

## Cleanup

```bash
# Stop services and keep data
./scripts/dev-environment.sh down

# Stop services and remove all data
./scripts/dev-environment.sh clean
```

## Next Steps

- Review the [API Documentation](./api-reference.md)
- Check the [Architecture Guide](./architecture.md)
- Read about [Contributing](../CONTRIBUTING.md)