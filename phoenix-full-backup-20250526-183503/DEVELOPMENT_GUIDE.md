# Phoenix Platform Development Guide

## Overview

This guide provides comprehensive information for developers working on the Phoenix Platform after the migration.

## Architecture Overview

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Phoenix CLI   │────▶│   API Gateway    │────▶│   Controller    │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                                │                          │
                                ▼                          ▼
                        ┌──────────────────┐     ┌─────────────────┐
                        │    Generator     │     │   Operators     │
                        └──────────────────┘     └─────────────────┘
```

## Module Structure

### Core Modules
- `packages/go-common` - Shared Go packages
- `packages/contracts` - Protocol buffer definitions
- `services/api` - API Gateway service
- `services/controller` - Experiment Controller
- `services/generator` - Configuration Generator
- `services/phoenix-cli` - Command Line Interface

### Import Conventions
```go
// Always use the phoenix platform prefix
import (
    "github.com/phoenix/platform/packages/go-common/auth"
    "github.com/phoenix/platform/packages/go-common/store"
    pb "github.com/phoenix/platform/packages/contracts/proto/v1"
)
```

## Development Workflow

### 1. Setting Up Your Environment

```bash
# Clone the repository
git clone <repository-url>
cd Phoenix

# Install dependencies
go work sync
go mod download

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate protobuf files
cd packages/contracts
bash generate.sh
```

### 2. Building Services

#### Build Everything
```bash
# Using make (if available)
make build

# Or manually
go work sync
cd services/api && go build -o bin/api ./cmd/main.go
cd ../controller && go build -o bin/controller ./cmd/controller/main.go
cd ../generator && go build -o bin/generator ./cmd/generator/main.go
cd ../phoenix-cli && go build -o bin/phoenix .
```

#### Build Specific Service
```bash
cd services/<service-name>
go build -o bin/<service-name> ./cmd/...
```

### 3. Running Services

#### Local Development
```bash
# Terminal 1: API Gateway
cd services/api
./bin/api --config=configs/dev.yaml

# Terminal 2: Controller
cd services/controller
./bin/controller --config=configs/dev.yaml

# Terminal 3: Generator
cd services/generator
./bin/generator --config=configs/dev.yaml
```

#### Using Docker Compose
```bash
cd infrastructure/docker/compose
docker-compose -f docker-compose.dev.yml up
```

### 4. Testing

#### Unit Tests
```bash
# Run all unit tests
go test ./...

# Run tests for specific package
go test ./services/api/...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Integration Tests
```bash
cd tests/integration
go test -v ./...

# Run specific test
go test -v -run TestExperimentWorkflow
```

### 5. Code Standards

#### Go Code Style
- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and small
- Handle errors explicitly

#### Example:
```go
// CreateExperiment creates a new optimization experiment
func (s *ExperimentService) CreateExperiment(ctx context.Context, req *pb.CreateExperimentRequest) (*pb.Experiment, error) {
    // Validate request
    if err := validateExperimentRequest(req); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    // Business logic here
    experiment := &pb.Experiment{
        Id:          uuid.New().String(),
        Name:        req.Name,
        Status:      pb.ExperimentStatus_PENDING,
        CreatedAt:   timestamppb.Now(),
    }
    
    // Store in database
    if err := s.store.CreateExperiment(ctx, experiment); err != nil {
        return nil, fmt.Errorf("failed to store experiment: %w", err)
    }
    
    return experiment, nil
}
```

### 6. Adding New Features

#### Adding a New Service
1. Create service structure:
```bash
mkdir -p services/my-service/{cmd,internal,api,configs,docs}
cd services/my-service
go mod init github.com/phoenix/platform/services/my-service
```

2. Add to workspace:
```bash
cd ../..
go work use ./services/my-service
```

3. Implement service following existing patterns

#### Adding New CLI Command
1. Create command file in `services/phoenix-cli/cmd/`
2. Register command in root command
3. Implement command logic
4. Add tests

### 7. Debugging

#### Enable Debug Logging
```bash
# Set log level
export LOG_LEVEL=debug

# Run service with debug
./bin/api --log-level=debug
```

#### Using Delve
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug a service
dlv debug ./cmd/main.go -- --config=configs/dev.yaml
```

### 8. Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

## Common Issues and Solutions

### Module Not Found
```bash
# Solution: Sync workspace
go work sync
```

### Import Cycle
```bash
# Solution: Move shared types to packages/go-common/models
```

### Proto Generation Failed
```bash
# Solution: Ensure protoc is installed
bash scripts/install-protoc.sh
```

## Best Practices

1. **Always use the workspace** - Work from the repository root
2. **Keep modules independent** - Services should not import from each other
3. **Use shared packages** - Common code goes in packages/go-common
4. **Write tests** - Aim for >80% coverage
5. **Document APIs** - Use OpenAPI/Swagger for REST, proto comments for gRPC
6. **Handle errors gracefully** - Never panic in production code
7. **Use structured logging** - Use the common logger from packages/go-common/telemetry

## Contributing

1. Create feature branch: `git checkout -b feature/my-feature`
2. Make changes following code standards
3. Add tests for new functionality
4. Run linters: `golangci-lint run`
5. Commit with clear messages
6. Push and create pull request

## Resources

- [Go Style Guide](https://golang.org/doc/effective_go.html)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [gRPC Documentation](https://grpc.io/docs/languages/go/)
- [Docker Compose](https://docs.docker.com/compose/)

## Support

For questions or issues:
1. Check existing documentation
2. Review code examples in tests
3. Look at similar implementations in other services
4. Create an issue with details

Happy coding! 🚀