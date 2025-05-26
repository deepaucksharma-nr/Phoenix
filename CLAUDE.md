# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

Phoenix Platform is an observability cost optimization system that reduces metrics cardinality by up to 90% while maintaining critical visibility. The platform uses a lean-core architecture with centralized control plane and lightweight distributed agents.

## Architecture

The platform consists of three main components:

1. **Phoenix API** (`projects/phoenix-api/`) - Monolithic control plane managing experiments, tasks, and analysis
2. **Phoenix Agent** (`projects/phoenix-agent/`) - Lightweight agents that poll for tasks and manage OTel collectors
3. **Dashboard** (`projects/dashboard/`) - React-based web UI with real-time WebSocket updates

Key architectural principles:
- Agents use long-polling (no incoming connections required)
- PostgreSQL serves as the single source of truth with task queue pattern
- Metrics flow: Agent → OTel Collector → Pushgateway → Prometheus → API
- All agent communication is outbound-only for security

## Critical Architecture Boundaries

### Project Independence
- **NEVER** import between projects in `/projects/*`
- Projects can only import from `/pkg/*` (shared packages)
- Each project has its own `go.mod` and lifecycle

### Database Access
- **NEVER** use direct database drivers (`database/sql`, `pgx`, `mongo-driver`)
- Always use abstractions in `pkg/database/*` or `projects/*/internal/store/*`

### Security
- **NEVER** hardcode secrets, passwords, or API keys
- **NEVER** modify production configurations outside `/deployments/kubernetes/overlays/production/`
- Security-sensitive files require security team approval (see CODEOWNERS)

## Build and Development Commands

### Quick Start
```bash
# One-command start (builds and runs everything)
./scripts/run-phoenix.sh

# Run a demo experiment
./scripts/demo-flow.sh

# Access points:
# - Dashboard: http://localhost:3000
# - API: http://localhost:8080
# - Prometheus: http://localhost:9090
```

### Repository-wide Commands
```bash
make setup              # First-time setup
make dev-up             # Start dev services (postgres, redis, prometheus)
make build              # Build all projects
make test               # Run all tests
make validate           # Validate architecture boundaries
make fmt                # Format all code
make lint               # Run linters
```

### Project-specific Commands
```bash
# Pattern: make <action>-<project>
make build-phoenix-api
make test-phoenix-agent
make lint-dashboard

# Or work directly in project:
cd projects/phoenix-api
make run                # Run with hot reload
make test               # Run all tests
make test-unit          # Unit tests only
make test-integration   # Integration tests
make docker             # Build container
```

### Testing Commands
```bash
# Run specific test
go test -v -run TestName ./...

# Run with coverage
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Integration tests (requires services)
make dev-up
go test -tags=integration ./tests/integration/...

# Dashboard tests
cd projects/dashboard
npm test
npm run test:watch     # Watch mode
```

### Validation Commands (Run Before Committing)
```bash
./tools/analyzers/boundary-check.sh        # Check import boundaries
./tools/analyzers/llm-safety-check.sh      # Check for AI issues
make validate                              # Run all validations
```

## Code Architecture

### Repository Structure
```
phoenix/
├── pkg/                    # Shared packages (strict review)
│   ├── auth/              # JWT authentication
│   ├── database/          # DB abstractions
│   ├── models/            # Shared data models
│   └── telemetry/         # Logging, metrics
├── projects/              # Independent services
│   ├── phoenix-api/       # Control plane
│   ├── phoenix-agent/     # Data plane agent
│   ├── phoenix-cli/       # CLI tool
│   └── dashboard/         # React UI
├── tests/                 # Cross-project tests
└── deployments/           # K8s, Docker configs
```

### Phoenix API Structure
```
projects/phoenix-api/
├── cmd/api/              # Main entrypoint
├── internal/
│   ├── api/             # HTTP/WebSocket handlers
│   ├── controller/      # Business logic
│   ├── store/           # Database layer
│   ├── tasks/           # Task queue implementation
│   └── websocket/       # Real-time updates
└── migrations/          # SQL migrations
```

### Task Queue Pattern
The API uses PostgreSQL-based task queue:
1. API creates tasks in `tasks` table
2. Agents poll `/api/v1/agents/tasks` endpoint
3. API assigns tasks atomically
4. Agents report results back
5. API updates experiment status

### WebSocket Events
Real-time updates via WebSocket (port 8081):
- `experiment_started`
- `experiment_phase_changed`
- `metrics_updated`
- `agent_status_changed`

## Go Workspace

This is a Go workspace monorepo:
```bash
go work sync            # Sync workspace modules
go mod tidy -e          # Tidy individual module
go work use ./new-proj  # Add new project
```

Current workspace modules:
- `./pkg`
- `./projects/phoenix-cli`
- `./projects/phoenix-api`
- `./projects/phoenix-agent`
- `./tests/e2e`

## Testing Patterns

### Go Table-Driven Tests
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "test", "TEST", false},
        {"empty input", "", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Test Pattern
```go
//go:build integration
// +build integration

func TestIntegration(t *testing.T) {
    // Requires make dev-up
}
```

## Common Development Tasks

### Add New API Endpoint
1. Define handler in `projects/phoenix-api/internal/api/`
2. Add route in `server.go`
3. Update OpenAPI spec if needed
4. Write handler tests
5. Add integration test

### Modify Database Schema
1. Create migration: `migrate create -ext sql -dir migrations -seq name`
2. Write up/down SQL
3. Test locally: `make migrate`
4. Update models in `internal/models/`
5. Update store methods

### Add WebSocket Event
1. Define event type in `internal/websocket/hub.go`
2. Send from appropriate handler
3. Update dashboard WebSocket handler
4. Document in API docs

## Environment Variables

### Phoenix API
- `DATABASE_URL` - PostgreSQL connection
- `PROMETHEUS_URL` - Prometheus server
- `PUSHGATEWAY_URL` - Pushgateway for metrics
- `JWT_SECRET` - Auth token signing
- `WEBSOCKET_PORT` - WebSocket server port

### Phoenix Agent
- `PHOENIX_API_URL` - API server URL
- `HOST_ID` - Unique host identifier
- `POLL_INTERVAL` - Task polling interval
- `CONFIG_DIR` - OTel configs directory

## Pre-commit Validation

The repo uses pre-commit hooks that run automatically:
- Import boundary validation
- Secret scanning
- Go formatting
- License headers

Run manually: `pre-commit run --all-files`

## Important Files

- `.ai-safety` - AI safety rules and restrictions
- `CODEOWNERS` - Code review requirements
- `go.work` - Go workspace configuration
- `.pre-commit-config.yaml` - Pre-commit hooks
- `deployments/single-vm/` - Simple deployment option

## Deployment

### Local Development
```bash
docker-compose up -d    # Start all services
make dev-up            # Alternative
```

### Production
- Kubernetes deployment in `deployments/kubernetes/`
- Helm charts in `infrastructure/helm/`
- Single VM option in `deployments/single-vm/`

Remember: When in doubt, run `make validate` to ensure your changes follow architectural rules.