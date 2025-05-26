# Phoenix Platform E2E Demo Guide

## Overview
This guide demonstrates a complete end-to-end flow of the Phoenix Platform after migration, showing how all components work together.

## Architecture Flow
```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│   Client    │────▶│  API Service │────▶│  PostgreSQL  │
└─────────────┘     └──────────────┘     └──────────────┘
                            │
                            ▼
                    ┌──────────────┐
                    │  Controller  │
                    └──────────────┘
                            │
                            ▼
                    ┌──────────────┐
                    │  Generator   │
                    └──────────────┘
```

## Quick Start

### Option 1: Local Development (Fastest)
```bash
# Run the E2E demo
./scripts/run-e2e-demo.sh

# Choose option 1 for local execution
```

### Option 2: Docker Compose (Production-like)
```bash
# Run with Docker
./scripts/run-e2e-demo.sh

# Choose option 2 for Docker execution
```

### Option 3: Manual Testing
```bash
# Build services
make -f Makefile.e2e build-services

# Start services
make -f Makefile.e2e start-e2e

# Run tests
make -f Makefile.e2e test-e2e

# View logs
make -f Makefile.e2e logs-e2e

# Stop services
make -f Makefile.e2e stop-e2e
```

## Service Endpoints

### API Service (Port 8080)
- **Health**: `GET http://localhost:8080/health`
- **Create Experiment**: `POST http://localhost:8080/api/v1/experiments`
- **Get Experiment**: `GET http://localhost:8080/api/v1/experiments/{id}`
- **List Experiments**: `GET http://localhost:8080/api/v1/experiments`
- **Start Experiment**: `POST http://localhost:8080/api/v1/experiments/{id}/start`
- **Stop Experiment**: `POST http://localhost:8080/api/v1/experiments/{id}/stop`
- **List Pipelines**: `GET http://localhost:8080/api/v1/pipelines`
- **Metrics**: `GET http://localhost:8080/metrics`

### Controller Service (Port 8082)
- **Health**: `GET http://localhost:8082/health`
- **Metrics**: `GET http://localhost:8082/metrics`

### Generator Service (Port 8083)
- **Health**: `GET http://localhost:8083/health`
- **Generate Config**: `POST http://localhost:8083/generate`
- **List Templates**: `GET http://localhost:8083/templates`
- **Metrics**: `GET http://localhost:8083/metrics`

### Dashboard (Port 3000)
- **UI**: `http://localhost:3000`

## Example API Calls

### Create an Experiment
```bash
curl -X POST http://localhost:8080/api/v1/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cost-optimization-test",
    "description": "Test cardinality reduction strategies",
    "baseline_pipeline": "baseline-v1",
    "candidate_pipeline": "optimized-v1",
    "target_selector": {
      "app": "my-app",
      "env": "staging"
    },
    "duration": "2h",
    "traffic_split": {
      "baseline": 50,
      "candidate": 50
    }
  }'
```

### Check Experiment Status
```bash
curl http://localhost:8080/api/v1/experiments/exp-123456
```

### Generate Pipeline Configuration
```bash
curl -X POST http://localhost:8083/generate \
  -H "Content-Type: application/json" \
  -d '{
    "experiment_id": "exp-123456",
    "type": "baseline",
    "parameters": {
      "memory_limit": "512Mi",
      "batch_size": 1000
    }
  }'
```

## Validation Steps

1. **Service Health**
   - All services respond with `{"status":"healthy"}`
   - No errors in service logs

2. **API Functionality**
   - Can create experiments
   - Can retrieve experiment details
   - Experiment IDs are generated correctly

3. **Controller Processing**
   - Controller logs show periodic checks
   - Would create Kubernetes resources (mocked in demo)

4. **Config Generation**
   - Generator returns valid pipeline configs
   - Templates are listed correctly

5. **Metrics Exposure**
   - All services expose Prometheus metrics
   - Metrics endpoints return valid data

## Troubleshooting

### Service Won't Start
```bash
# Check if ports are already in use
lsof -i :8080
lsof -i :8082
lsof -i :8083

# Kill existing processes
pkill -f "go run ./cmd"
```

### Docker Issues
```bash
# Clean up containers and volumes
docker-compose -f docker-compose.e2e.yml down -v

# Rebuild images
docker-compose -f docker-compose.e2e.yml build --no-cache
```

### Database Connection Issues
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Test connection
docker exec -it phoenix_postgres_1 psql -U phoenix -d phoenix
```

## Next Steps

1. **Add Real Database Integration**
   - Implement actual database models
   - Add migrations
   - Wire up real queries

2. **Implement Kubernetes Integration**
   - Add K8s client to controller
   - Create actual CRDs
   - Deploy to real cluster

3. **Complete Proto Integration**
   - Generate proto code
   - Implement gRPC services
   - Add service-to-service communication

4. **Production Readiness**
   - Add authentication
   - Implement proper logging
   - Add comprehensive tests
   - Set up CI/CD

## Success Criteria

✅ All core services start successfully
✅ Health endpoints respond
✅ Can create and retrieve experiments via API
✅ Config generator produces valid configurations
✅ Prometheus metrics are exposed
✅ No import errors or missing dependencies
✅ Services can find shared packages

The migration is successful when all these criteria are met!