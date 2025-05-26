# Phoenix Platform Quick Start Guide

## Post-Migration Quick Start

This guide helps you get started with the Phoenix Platform after the migration.

### 1. Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Protocol Buffer compiler (protoc)
- Make

### 2. First-Time Setup

```bash
# 1. Install protoc if not already installed
bash scripts/install-protoc.sh

# 2. Generate protocol buffers
cd packages/contracts
bash generate.sh
cd ../..

# 3. Sync Go workspace
go work sync

# 4. Build Phoenix CLI
cd services/phoenix-cli
go build -o bin/phoenix .
export PATH=$PATH:$(pwd)/bin

# 5. Verify CLI installation
phoenix version
```

### 3. Building Services

```bash
# Build all services at once (if Makefile exists)
make build-all

# Or build individually:
cd services/api && go build -o bin/api ./cmd/main.go
cd ../generator && go build -o bin/generator ./cmd/generator/main.go
cd ../controller && go build -o bin/controller ./cmd/controller/main.go
```

### 4. Running Services Locally

#### Using Docker Compose
```bash
cd infrastructure/docker/compose
docker-compose -f docker-compose.dev.yml up
```

#### Running Individually
```bash
# Terminal 1: API Gateway
cd services/api
./bin/api

# Terminal 2: Config Generator
cd services/generator
./bin/generator

# Terminal 3: Experiment Controller
cd services/controller
./bin/controller
```

### 5. Using Phoenix CLI

```bash
# Login to Phoenix
phoenix auth login

# List experiments
phoenix experiment list

# Create a new experiment
phoenix experiment create \
  --name "cost-optimization-test" \
  --baseline "process-baseline-v1" \
  --candidate "process-intelligent-v1" \
  --duration 30m

# Check experiment status
phoenix experiment status <experiment-id>

# View pipeline configurations
phoenix pipeline list
```

### 6. Development Workflow

```bash
# 1. Make changes to code

# 2. Run tests
go test ./...

# 3. Lint code
golangci-lint run

# 4. Build affected services
cd services/<service-name>
go build -o bin/<service-name> ./cmd/...

# 5. Run integration tests
cd tests/integration
go test -v
```

### 7. Common Tasks

#### Adding a New Service
```bash
# 1. Create service directory
mkdir -p services/my-service/{cmd,internal,api,configs}

# 2. Initialize Go module
cd services/my-service
go mod init github.com/phoenix/platform/services/my-service

# 3. Add to workspace
cd ../..
go work use ./services/my-service
```

#### Updating Dependencies
```bash
# Update a specific service
cd services/api
go get -u ./...
go mod tidy

# Update all modules
go work sync
```

### 8. Troubleshooting

#### Module Issues
```bash
# If you see module errors
go clean -modcache
go work sync
```

#### Build Issues
```bash
# Clean and rebuild
make clean
make build-all
```

#### Proto Generation Issues
```bash
# Ensure protoc plugins are in PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Regenerate protos
cd packages/contracts
bash generate.sh
```

### 9. Useful Commands

```bash
# Check which services are running
docker-compose ps

# View service logs
docker-compose logs -f <service-name>

# Run specific tests
go test -v -run TestName ./path/to/package

# Format code
go fmt ./...

# Check for issues
go vet ./...
```

### 10. Next Steps

1. Explore the Phoenix CLI commands: `phoenix --help`
2. Review service documentation in each service's `docs/` directory
3. Check out example experiments in `examples/`
4. Join the development workflow by creating feature branches

## Support

For issues or questions:
1. Check the MIGRATION_SUMMARY.md
2. Review service-specific README files
3. Look at integration test examples
4. Check the architectural documentation

Happy coding with Phoenix Platform! 🚀