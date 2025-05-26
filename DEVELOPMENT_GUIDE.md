# Phoenix Platform Development Guide

This guide helps you set up a development environment for the Phoenix Platform.

## Prerequisites

### Required Tools
- **Go** 1.24+ ([install](https://golang.org/doc/install))
- **Node.js** 18+ ([install](https://nodejs.org/))
- **Docker** 20.10+ ([install](https://docs.docker.com/get-docker/))
- **Docker Compose** 2.0+ (included with Docker Desktop)
- **PostgreSQL** 15+ (or use Docker)
- **Make** (usually pre-installed)

### Recommended Tools
- **golangci-lint** - Go linter
- **goreman** - Process manager
- **air** - Hot reload for Go
- **protoc** - Protocol buffer compiler

## 🚀 Quick Setup

```bash
# Clone repository
git clone https://github.com/phoenix/platform.git
cd phoenix

# Install all dependencies and tools
make setup

# Start development environment
make dev-up

# Verify installation
make test
```

## 📁 Repository Structure

```
phoenix/
├── pkg/                    # Shared Go packages
├── projects/              # Microservices
│   ├── phoenix-api/      # Control plane API
│   ├── phoenix-agent/    # Data plane agent
│   ├── phoenix-cli/      # CLI tool
│   └── dashboard/        # React UI
├── docs/                 # Documentation
├── scripts/              # Utility scripts
└── tests/               # Integration tests
```

## 🛠️ Development Workflow

### 1. Environment Setup

Create a `.env` file:

```bash
# Database
DATABASE_URL=postgres://phoenix:phoenix@localhost:5432/phoenix_dev?sslmode=disable

# Services
PHOENIX_API_URL=http://localhost:8080
PROMETHEUS_URL=http://localhost:9090
PUSHGATEWAY_URL=http://localhost:9091

# Development
LOG_LEVEL=debug
ENABLE_HOT_RELOAD=true
```

### 2. Database Setup

```bash
# Start PostgreSQL
docker-compose up -d postgres

# Run migrations
make migrate

# Seed test data (optional)
make seed
```

### 3. Running Services

#### Option A: All Services (Recommended)
```bash
# Start everything with hot reload
make dev-up

# Or use goreman
goreman start
```

#### Option B: Individual Services
```bash
# Terminal 1: Phoenix API
cd projects/phoenix-api
make run

# Terminal 2: Phoenix Agent
cd projects/phoenix-agent
make run

# Terminal 3: Dashboard
cd projects/dashboard
npm run dev
```

### 4. Making Changes

1. **Create feature branch**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make changes**
   - Follow existing code patterns
   - Add tests for new features
   - Update documentation

3. **Run checks**
   ```bash
   make lint        # Run linters
   make test        # Run tests
   make validate    # Check boundaries
   ```

4. **Commit changes**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

## 🧪 Testing

### Unit Tests
```bash
# Test everything
make test

# Test specific project
make test-phoenix-api
make test-phoenix-agent

# With coverage
make test-coverage
```

### Integration Tests
```bash
# Run integration tests
make test-integration

# Run specific test
go test -tags=integration ./tests/integration/... -run TestName
```

### E2E Tests
```bash
# Start test environment
make test-env-up

# Run E2E tests
make test-e2e

# Clean up
make test-env-down
```

## 🔧 Common Tasks

### Building

```bash
# Build all projects
make build

# Build specific project
make build-phoenix-api

# Build Docker images
make docker-build
```

### Database Operations

```bash
# Create new migration
make migration-create name=add_new_table

# Run migrations
make migrate

# Rollback migration
make migrate-down
```

### Code Generation

```bash
# Generate protobuf code
make generate

# Generate mocks
make mocks

# Update OpenAPI spec
make openapi
```

## 🐛 Debugging

### VS Code Configuration

`.vscode/launch.json`:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Phoenix API",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/projects/phoenix-api/cmd/api",
      "env": {
        "LOG_LEVEL": "debug"
      }
    }
  ]
}
```

### Common Issues

**Port already in use?**
```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>
```

**Database connection failed?**
```bash
# Check PostgreSQL
docker-compose ps postgres

# View logs
docker-compose logs postgres
```

**Dependencies out of sync?**
```bash
# Update Go modules
go work sync
go mod tidy

# Update Node modules
cd projects/dashboard
npm install
```

## 📏 Code Standards

### Go Code
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` and `goimports`
- Write table-driven tests
- Keep functions small and focused

### TypeScript/React
- Use functional components
- Follow React hooks best practices
- Use TypeScript strict mode
- Write component tests

### General
- Write clear commit messages
- Keep PRs focused and small
- Update documentation
- Add tests for new features

## 🚀 Advanced Topics

### Performance Profiling
```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.

# Memory profiling
go test -memprofile=mem.prof -bench=.

# Analyze profile
go tool pprof cpu.prof
```

### Load Testing
```bash
# Run load tests
make load-test

# Custom load test
hey -n 10000 -c 100 http://localhost:8080/api/v1/experiments
```

## 📚 Resources

- [Architecture Documentation](docs/architecture/PLATFORM_ARCHITECTURE.md)
- [API Documentation](docs/api/)
- [Contributing Guidelines](CONTRIBUTING.md)
- [Phoenix CLI](projects/phoenix-cli/README.md)

## 💬 Getting Help

- Check [Documentation](docs/)
- Search [GitHub Issues](https://github.com/phoenix/platform/issues)
- Join [Discord Community](https://discord.gg/phoenix)
- Read [FAQ](docs/FAQ.md)