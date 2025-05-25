# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## CRITICAL: Documentation Placement Rules

**NEVER create documentation at the repository root level!** 
- Phoenix-specific docs go in: `phoenix-platform/docs/`
- Repository governance docs go in: `docs/`
- Service-specific docs go in: `<service>/docs/`
- ONLY exception: This CLAUDE.md file must remain at root

See `docs/DOCUMENTATION_GOVERNANCE.md` for strict enforcement rules.

## Architectural Integrity Guidelines

### Key Principles for Maintaining Structural Integrity
1. Always preserve the existing folder structure
2. Avoid introducing new top-level directories
3. Keep all code within the `phoenix-platform/` subdirectory
4. Follow mono-repo governance rules strictly
5. Do not create files at the repository root
6. Ensure all updates align with existing architectural patterns
7. Maintain clear separation between services
8. Use GitOps for all configuration changes
9. Validate structural changes against mono-repo governance
10. Prioritize code organization and predictability

### Anti-Drift Measures
- Regularly run `make validate-structure` to catch potential architectural deviations
- Review all changes against `docs/MONO_REPO_GOVERNANCE.md`
- Use existing template files and patterns for new implementations
- Consult architecture documentation before making significant changes

### Update Instructions
- Prefer modifying existing files over creating new ones
- If a new component is required, use the existing service template
- Always update documentation to reflect architectural changes
- Ensure new code follows existing patterns and guidelines

## Recent Actions
- Update all relevant .md files in docs based on set of changes

## Phoenix Platform Context (January 2025)

### Project Overview
Phoenix is an observability cost optimization platform that reduces metrics volume by 50-80% through intelligent OpenTelemetry pipeline optimization. It uses A/B testing between baseline and candidate configurations without requiring a service mesh.

### Key Architectural Decisions (ADRs)
1. **No Service Mesh** (ADR-001): Use dual collectors pattern instead
2. **GitOps Mandatory** (ADR-002): All deployments via ArgoCD, no direct kubectl
3. **Visual Pipeline Builder** (ADR-003): Drag-drop as primary configuration interface
4. **Mono-Repo Boundaries** (ADR-004): Strict service separation enforced
5. **Dual Metrics Export** (ADR-005): Export to both Prometheus and New Relic

### Implementation Status
- **Foundation Phase**: 100% Complete
  - Proto definitions for all services
  - Client libraries with examples
  - Validation scripts enforcing architecture
  - Database migrations framework
  - Pre-commit hooks for automated checks

- **Core Services**: 60% Complete
  - Experiment Controller: 80% (state machine, DB integration)
  - Config Generator: 80% (template engine, manifest generation)
  - Pipeline Operator: 85% (full reconciliation, DaemonSet management)
  - API Service: 30% (proto definitions ready, implementation pending)

### Critical Files & Locations
- **Proto Definitions**: `phoenix-platform/api/proto/v1/`
- **Client Libraries**: `phoenix-platform/pkg/clients/`
- **Validation Scripts**: `phoenix-platform/scripts/validate/`
- **Database Migrations**: `phoenix-platform/migrations/`
- **Service Implementations**: `phoenix-platform/services/`
- **Kubernetes Operators**: `phoenix-platform/operators/`
- **Architecture Docs**: `phoenix-platform/docs/adr/`
- **Implementation Plans**: `phoenix-platform/docs/planning/`

### Development Workflow
1. **Before Making Changes**:
   - Run `make validate` to check structure
   - Review relevant ADRs in `docs/adr/`
   - Check `docs/planning/IMPLEMENTATION_CHECKLIST.md`

2. **Proto Changes**:
   - Edit files in `api/proto/v1/`
   - Run `make generate-proto`
   - Update client libraries if needed

3. **Service Development**:
   - Use existing service structure as template
   - Follow interface definitions in proto files
   - Use client libraries for service communication
   - Add to validation scripts if creating new service

4. **Testing**:
   - Unit tests in `<service>/internal/*/test.go`
   - Integration tests in `<service>/test/integration/`
   - E2E tests in `phoenix-platform/test/e2e/`

5. **Validation**:
   - `make validate-structure`: Check mono-repo structure
   - `make validate-imports`: Verify Go import rules
   - `make validate-services`: Check service boundaries
   - `make validate`: Run all checks

### Essential Development Commands

```bash
# Setup and dependencies
make deps                    # Install Go and npm dependencies
make setup-hooks            # Setup git pre-commit hooks

# Code generation and validation
make generate               # Generate protobuf code and CRDs
make generate-proto         # Generate only protobuf code
make validate              # Run all validation checks
make validate-structure    # Check mono-repo structure
make validate-imports      # Validate Go import rules

# Building
make build                 # Build all components
make build-api            # Build specific service
make build-dashboard      # Build frontend dashboard
make docker               # Build all Docker images

# Testing
make test                 # Run all tests (unit + integration)
make test-unit           # Unit tests only
make test-integration    # Integration tests only
make test-e2e           # End-to-end tests
make test-dashboard     # Dashboard tests with coverage
make coverage           # Generate test coverage report

# Code quality
make fmt                 # Format Go and frontend code
make lint               # Run linters (Go + frontend)
make verify             # Run all pre-commit checks

# Local development
make dev                # Start local development environment
make dev-down          # Stop local development environment
make dev-logs          # Show development environment logs
make dev-status        # Show development environment status

# Kubernetes development
make cluster-up        # Start local kind cluster
make cluster-down      # Stop local kind cluster
make deploy           # Deploy to Kubernetes
make undeploy         # Remove from Kubernetes
make port-forward     # Forward ports for local access

# Utilities
make clean            # Clean build artifacts
make help             # Show all available targets
```

### Project Structure and Architecture

**Core Services** (all in `phoenix-platform/cmd/`):
- `api/` - Main API service (HTTP/REST endpoints)
- `api-gateway/` - HTTP to gRPC gateway with auth middleware  
- `controller/` - Experiment controller with state machine
- `control-service/` - Control plane gRPC service
- `generator/` - Configuration generator for OTel pipelines
- `simulator/` - Process simulator for testing

**Frontend Dashboard** (`phoenix-platform/dashboard/`):
- React/TypeScript with Vite build system
- Material-UI components and React Flow for pipeline builder
- Vitest for testing, ESLint/Prettier for code quality

**Kubernetes Operators** (`phoenix-platform/operators/`):
- `pipeline/` - Manages OTel collector DaemonSets
- `loadsim/` - Manages load simulation jobs

**Protocol Definitions** (`phoenix-platform/api/proto/`):
- Centralized protobuf definitions for all services
- Generated Go code for gRPC clients/servers

### Technology Stack
- **Backend**: Go 1.21, gRPC, PostgreSQL, Kubernetes client-go
- **Frontend**: React 18, TypeScript, Material-UI, React Flow, Vitest
- **Infrastructure**: Kubernetes, Helm, ArgoCD, Prometheus, Grafana
- **Code Quality**: golangci-lint, pre-commit hooks, ESLint, Prettier

### Important Notes
1. **Never bypass validation scripts** - they enforce architectural integrity
2. **All configuration changes must go through GitOps** - no direct kubectl
3. **Service boundaries are strict** - no cross-service imports allowed
4. **Use proto contracts** - all service communication via defined APIs
5. **Follow existing patterns** - consistency is critical

### Critical Development Practices

**Before Making Any Changes:**
1. Run `make validate` to ensure structural integrity
2. Review the Makefile to understand available commands
3. Check existing tests and follow established patterns
4. Always run `make setup-hooks` after cloning to install git hooks

**Code Quality Requirements:**
- All Go code must pass `golangci-lint` (enforced by pre-commit)
- Frontend code must pass ESLint and Prettier formatting
- All changes must pass mono-repo structure validation
- No commits allowed without passing pre-commit hooks

**Testing Strategy:**
- Unit tests: Individual service/component testing
- Integration tests: Service-to-service communication  
- E2E tests: Full workflow testing in Kubernetes
- Dashboard tests: Component and store testing with Vitest

### References
- Technical Spec: `phoenix-platform/docs/TECHNICAL_SPECIFICATION.md`
- Implementation Roadmap: `phoenix-platform/docs/planning/NEXT_STEPS_ACTION_PLAN.md`
- Project Status: `phoenix-platform/docs/planning/PROJECT_STATUS.md`
- Mono-Repo Rules: `docs/MONO_REPO_GOVERNANCE.md`