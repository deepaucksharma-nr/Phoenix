# Phoenix Platform - Build and Run Guide

## Implementation Status

See [Implementation Status](docs/IMPLEMENTATION_STATUS.md) for detailed component completion status.

**Quick Summary:**
- API Service: 30% complete (basic structure)
- Dashboard: 25% complete (React setup)
- Controller: 80% complete (fully functional core)
- Generator: 80% complete (template engine and API)
- Pipeline Operator: 85% complete (full Kubernetes operator)
- LoadSim Operator: 10% complete (basic structure)
- Testing: 60% coverage (integration tests implemented)

## Quick Start Guide

### Prerequisites

- Go 1.21+
- Node.js 18+ and npm
- Docker and Docker Compose
- Make

### Initial Setup

1. **Clone and navigate to the project**
   ```bash
   cd phoenix-platform
   ```

2. **Create missing package-lock.json**
   ```bash
   cd dashboard && npm install && cd ..
   ```

3. **Download Go dependencies**
   ```bash
   go mod download
   ```

4. **Create .env file**
   ```bash
   cat > .env << EOF
   DATABASE_URL=postgres://phoenix:phoenix@localhost:5432/phoenix?sslmode=disable
   JWT_SECRET=development-secret-change-me
   NEW_RELIC_API_KEY=your-api-key-here
   GIT_TOKEN=your-git-token-here
   EOF
   ```

## Building Components

### Option 1: Build What's Currently Implemented

```bash
# Build all implemented components
make build-api build-controller build-generator build-operators build-simulator build-dashboard

# Or build them individually:

# API Service
CGO_ENABLED=0 go build -o build/phoenix-api ./cmd/api

# Experiment Controller (fully implemented!)
CGO_ENABLED=0 go build -o build/experiment-controller ./cmd/controller

# Config Generator (NEW - template engine implemented!)
CGO_ENABLED=0 go build -o build/config-generator ./cmd/generator

# Pipeline Operator (NEW - full Kubernetes operator!)
CGO_ENABLED=0 go build -o build/pipeline-operator ./operators/pipeline/cmd

# LoadSim Operator
CGO_ENABLED=0 go build -o build/loadsim-operator ./operators/loadsim/cmd

# Process Simulator  
CGO_ENABLED=0 go build -o build/process-simulator ./cmd/simulator

# Dashboard (requires npm install first)
cd dashboard && npm run build
```

### Option 2: Docker Build (Implemented Components Only)

```bash
# Build API Docker image
docker build -f docker/api/Dockerfile -t phoenix/api:latest .

# Build Simulator Docker image
docker build -f docker/simulator/Dockerfile -t phoenix/simulator:latest .

# Build Dashboard Docker image
docker build -f docker/dashboard/Dockerfile -t phoenix/dashboard:latest .
```

## Running the Platform

### Local Development Mode

1. **Start infrastructure services**
   ```bash
   docker-compose up -d postgres prometheus grafana
   ```

2. **Run services locally**
   ```bash
   # API service (terminal 1)
   go run cmd/api/main.go
   
   # Experiment Controller (terminal 2)
   DATABASE_URL=postgres://phoenix:phoenix@localhost:5432/phoenix?sslmode=disable \
   go run cmd/controller/main.go
   
   # Config Generator (terminal 3)
   go run cmd/generator/main.go
   
   # Pipeline Operator (terminal 4 - requires Kubernetes cluster)
   go run operators/pipeline/cmd/main.go
   ```

3. **Run dashboard development server**
   ```bash
   cd dashboard
   npm run dev
   ```

4. **Access services**
   - API: http://localhost:8080
   - Experiment Controller Health: http://localhost:8081/health
   - Config Generator: http://localhost:8082
   - Dashboard: http://localhost:5173 (Vite dev server)
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3001 (admin/admin)

### Docker Compose Mode

```bash
# Start all implemented services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

## Testing What's Available

### 1. Test API Health
```bash
curl http://localhost:8080/health
```

### 2. Test API Metrics
```bash
curl http://localhost:8080/metrics
```

### 3. Test Experiment Controller
```bash
# Check health endpoint
curl http://localhost:8081/health

# Check metrics
curl http://localhost:8081/metrics

# Test gRPC service (requires grpcurl)
grpcurl -plaintext localhost:50051 list

# Create an experiment via gRPC
grpcurl -plaintext -d '{
  "name": "Test Experiment",
  "description": "Testing the controller"
}' localhost:50051 phoenix.v1.ExperimentService/CreateExperiment
```

### 4. Run Process Simulator
```bash
go run cmd/simulator/main.go --profile realistic --duration 5m
```

### 5. Access Dashboard
Open http://localhost:3000 in your browser

## Implementation TODOs

To get the full platform working, the following components need to be implemented:

### 1. ~~Experiment Controller~~ ✅ COMPLETED!
The Experiment Controller has been fully implemented with:
- Complete state machine for experiment lifecycle
- PostgreSQL integration with automatic migrations
- gRPC service with health checks
- Automatic state transitions and timeout handling
- Scheduler for periodic reconciliation

Next integration steps:
- Connect to Config Generator service (when implemented)
- Integrate with Pipeline Operator for K8s deployments
- Add Prometheus metrics collection from running experiments

### 2. Config Generator (`cmd/generator/main.go`)
```go
// Minimal implementation needed:
// - Generate OTel collector configs
// - Create Kubernetes manifests
// - Git integration for config storage
```

### 3. Pipeline Operator (`operators/pipeline/cmd/main.go`)
```go
// Minimal implementation needed:
// - Watch PhoenixProcessPipeline CRDs
// - Deploy/update OTel collectors
// - Manage ConfigMaps
```

### 4. LoadSim Operator (`operators/loadsim/cmd/main.go`)
```go
// Minimal implementation needed:
// - Watch LoadSimulationJob CRDs
// - Create simulator Jobs
// - Clean up completed jobs
```

### 5. Create Missing Dockerfiles

Each missing component needs a Dockerfile in its respective `docker/` directory.

## Troubleshooting

### Common Issues

1. **npm ci fails**
   - Run `npm install` instead to generate package-lock.json
   
2. **Go module errors**
   - Run `go mod tidy` to clean up dependencies
   
3. **Port conflicts**
   - Check if ports 3000, 5050, 8080, 9090, 3001 are in use
   - Modify docker-compose.yaml or use different ports

4. **Database connection errors**
   - Ensure PostgreSQL is running
   - Check DATABASE_URL in .env

### Verifying Services

```bash
# Check running containers
docker-compose ps

# Check API logs
docker-compose logs api

# Check database connection
docker exec -it phoenix-platform-postgres-1 psql -U phoenix -d phoenix
```

## Next Steps

1. **For Development**: Focus on implementing missing components one by one
2. **For Testing**: Use the API and simulator that are already implemented
3. **For UI Development**: The dashboard can be developed independently

## Minimal Working Setup

If you just want to see something working:

```bash
# 1. Start PostgreSQL
docker-compose up -d postgres

# 2. Run the API
go run cmd/api/main.go

# 3. Test the API
curl http://localhost:8080/health

# 4. Run the dashboard
cd dashboard && npm install && npm run dev
```

This will give you a working API and dashboard to start experimenting with.