# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

Phoenix Platform is a monorepo for an observability cost optimization system designed to reduce metrics cardinality by up to 90% while maintaining critical visibility. The architecture implements strict boundaries and comprehensive validation to prevent architectural drift, especially when using AI-assisted development.

## Critical Architecture Boundaries

### Project Independence
- **NEVER** import between projects in `/projects/*`
- Projects can only import from `/pkg/*` (shared packages)
- Each project maintains its own lifecycle and dependencies

### Database Access
- **NEVER** use direct database drivers (`database/sql`, `pgx`, `mongo-driver`)
- Always use abstractions in `pkg/database/*` or `projects/*/internal/store/*`

### Security
- **NEVER** hardcode secrets, passwords, or API keys
- **NEVER** modify production configurations outside `/deployments/kubernetes/overlays/production/`
- Security-sensitive files require security team approval (see CODEOWNERS)

## Build and Development Commands

### Repository-wide Commands
```bash
# Setup development environment (first time)
make setup

# Start all development services
make dev-up

# Validate entire repository structure
make validate

# Build all projects
make build

# Run all tests
make test

# Format all code
make fmt

# Run linting
make lint

# Run security scans
make security
```

### Project-specific Commands
```bash
# Work with specific project (replace <project> with actual name)
make build-<project>     # Build specific project
make test-<project>      # Test specific project
make lint-<project>      # Lint specific project

# Example for platform-api:
cd projects/platform-api
make build              # Build the service
make test               # Run all tests
make test-unit          # Run unit tests only
make test-integration   # Run integration tests
make run                # Build and run locally
make docker             # Build Docker image
```

### Validation Commands (Run Before Committing)
```bash
# Check architectural boundaries
./tools/analyzers/boundary-check.sh

# Check for AI-generated issues
./tools/analyzers/llm-safety-check.sh

# Enhanced structure validation
./build/scripts/utils/validate-structure-enhanced.sh
```

## Code Architecture

### Repository Structure
```
phoenix/
├── pkg/                    # Shared packages (strict review required)
│   ├── auth/              # Authentication (security review)
│   ├── telemetry/         # Logging, metrics, tracing
│   ├── database/          # Database abstractions
│   └── contracts/         # API contracts and schemas
├── projects/              # Independent micro-projects
│   └── <project-name>/    # Each follows standard structure:
│       ├── cmd/           # Application entrypoints
│       ├── internal/      # Private application code
│       ├── build/         # Docker and build configs
│       └── Makefile       # Project-specific commands
├── build/                 # Shared build infrastructure
│   └── makefiles/         # Reusable Makefile components
├── tools/                 # Development and validation tools
│   └── analyzers/         # Static analysis scripts
└── deployments/           # K8s, Helm, Terraform configs
```

### Project Standard Structure
Every project under `/projects/` follows:
- `cmd/`: Application entrypoints
- `internal/api/`: HTTP/gRPC handlers
- `internal/domain/`: Business logic (entities, services, repositories)
- `internal/infrastructure/`: External dependencies (DB, cache, etc.)

### Shared Package Usage
- Import shared packages: `github.com/phoenix/platform/pkg/<package>`
- Module replacement in go.mod: `replace github.com/phoenix/platform/pkg => ../../pkg`

## AI Safety Configuration

The repository has `.ai-safety` configuration that defines:
- Forbidden patterns and operations
- File modification restrictions
- Import limitations
- Metrics tracking for anomaly detection

Key restrictions:
- Cannot modify CODEOWNERS, .ai-safety, LICENSE, production configs
- Cannot disable tests or remove validation
- Must follow approved templates for code generation

## Testing Requirements

### For New Features
1. Write unit tests in `*_test.go` files
2. Use table-driven tests for Go code
3. Maintain >80% coverage
4. Integration tests go in `tests/integration/`

### Test Patterns
```go
// Go table-driven test pattern
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Pre-commit Validation

The repository uses extensive pre-commit hooks (`.pre-commit-config.yaml`):
- File format validation
- Secret scanning
- License header checking
- Import boundary validation
- LLM safety checks

These run automatically on commit but can be run manually:
```bash
pre-commit run --all-files
```

## Go Workspace

This is a Go workspace monorepo. Key points:
- `go.work` defines the workspace modules
- Each project has its own `go.mod`
- Shared packages in `/pkg` have a separate `go.mod`
- Use `go work sync` after modifying workspace

## Code Review Requirements

CODEOWNERS enforces review requirements:
- `/pkg/` changes require architect review
- Security-sensitive files require security team
- Production deployments require DevOps + security review
- Each project has designated team ownership

## Common Issues and Solutions

### Cross-project Import Violation
**Error**: "Cross-project import detected"
**Solution**: Move shared code to `/pkg/` or duplicate if project-specific

### Direct Database Access
**Error**: "Direct database driver import"
**Solution**: Use `pkg/database` abstractions instead

### Hardcoded Secrets
**Error**: "Potential hardcoded secret detected"
**Solution**: Use environment variables or secret management

### Go Workspace Issues
**Error**: "go.work references non-existent module"
**Solution**: Run `go work sync` or update go.work to match existing projects

## Development Workflow

1. Create feature branch
2. Make changes following architecture boundaries
3. Run validation: `make validate`
4. Run tests: `make test`
5. Check boundaries: `./tools/analyzers/boundary-check.sh`
6. Check AI safety: `./tools/analyzers/llm-safety-check.sh`
7. Commit (pre-commit hooks will run)
8. Create PR (CODEOWNERS will assign reviewers)

## Important Configuration Files

- `.ai-safety`: AI agent boundaries and rules
- `.pre-commit-config.yaml`: Automated validation hooks
- `CODEOWNERS`: Review requirements
- `.golangci.yml`: Go linting configuration
- `go.work`: Go workspace configuration

## Deployment

- Development: `make k8s-deploy-dev`
- Uses Kubernetes with Helm charts
- GitOps workflow with manifest generation
- Production requires multi-team approval

## Recent Architecture Updates

### WebSocket Integration
- Phoenix API now includes WebSocket support on port 8081
- Real-time updates for experiments, metrics, and agent status
- Hub pattern implementation in `pkg/common/websocket/`

### UI-First Experience
- New UI handlers in `projects/phoenix-api/internal/api/ui_handlers.go`
- Dashboard components for live cost monitoring and agent visualization
- React 18 with TypeScript and Vite for development

### Task Queue Pattern
- PostgreSQL-based task queue for agent communication
- Long-polling design for security (agents only make outbound connections)
- Atomic task assignment and status tracking

### Model Extensions
- Added `Variant` field to PipelineDeployment for A/B testing
- DeploymentMetrics includes `MetricsPerSecond` and `CardinalityReduction`
- UpdateDeploymentRequest supports `StatusMessage` and `UpdatedBy` fields
- New deployment statuses: `degraded` and `healthy`

Remember: The structure is designed to be self-validating. When in doubt, run `make validate` to check if your changes follow the architectural rules.