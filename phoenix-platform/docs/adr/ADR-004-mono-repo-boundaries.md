# ADR-004: Strict Mono-Repository Boundaries

## Status
Accepted

## Context
The Phoenix platform uses a mono-repository structure with multiple services. Without clear boundaries, services can become tightly coupled, making maintenance and evolution difficult.

## Decision
Enforce STRICT boundaries between services in the mono-repository through technical and organizational measures.

## Rationale
1. **Maintainability**: Clear boundaries prevent spaghetti dependencies
2. **Team Autonomy**: Teams can work independently on services
3. **Testability**: Services can be tested in isolation
4. **Deployment**: Services can be deployed independently
5. **Evolution**: Services can evolve without breaking others

## Implementation

### Directory Structure
```
phoenix-platform/
├── cmd/                    # Service entry points ONLY
│   ├── api/               # OWNS: API logic
│   ├── controller/        # OWNS: State machine, orchestration
│   └── generator/         # OWNS: Config generation
├── pkg/                   # Shared packages (minimal)
│   ├── auth/             # Shared authentication
│   ├── models/           # Shared data models
│   └── clients/          # Service clients
├── internal/             # Private to Phoenix (not per-service)
└── <service>/internal/   # Private to specific service
```

### Import Rules
```go
// ALLOWED
import "phoenix-platform/pkg/auth"           // Shared packages
import "phoenix-platform/cmd/api/internal"   // Within same service

// FORBIDDEN
import "phoenix-platform/cmd/controller/internal"  // Cross-service internal
import "../../../controller"                       // Relative imports
```

### Communication Rules
1. **Services communicate via APIs only** (gRPC/REST)
2. **No shared database access** between services
3. **No direct file system sharing** between services
4. **Events via message bus** (when implemented)

## Enforcement

### Technical Enforcement
```bash
# validate-imports.go checks all imports
# Pre-commit hooks block violations
# CI/CD fails on boundary violations
```

### Static Analysis Rules
```yaml
import_rules:
  - deny: "cmd/.*/internal from different cmd/"
  - deny: "relative imports beyond module"
  - allow: "pkg/* from anywhere"
  - allow: "internal/* from same module"
```

## Consequences
### Positive
- Clean architecture
- Independent deployability
- Clear ownership
- Easier onboarding

### Negative
- Some code duplication
- More boilerplate for communication
- Requires discipline

## Service Responsibilities

| Service | Owns | Does NOT Own |
|---------|------|--------------|
| API | HTTP/gRPC endpoints, auth | Business logic, state |
| Controller | Experiment state, orchestration | API endpoints, config generation |
| Generator | Config templates, optimization | Deployment, state management |
| Operators | Kubernetes reconciliation | Business logic, API |

## Alternatives Considered
1. **Microservices**: Too much operational overhead initially
2. **Monolith**: Would become unmaintainable
3. **Shared Libraries**: Lead to tight coupling
4. **No Boundaries**: Results in spaghetti architecture

## References
- Mono-repo governance in MONO_REPO_GOVERNANCE.md
- Static analysis rules in STATIC_ANALYSIS_RULES.md
- Service specifications in TECHNICAL_SPEC_*.md