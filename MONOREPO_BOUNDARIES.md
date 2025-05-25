# Phoenix Platform Monorepo Boundaries

## Overview

The Phoenix Platform uses a strict monorepo structure with enforced boundaries to maintain clean architecture and prevent coupling between services.

## Directory Structure

```
phoenix/
├── packages/              # Shared packages (no project imports allowed)
│   ├── go-common/        # Go shared libraries
│   │   ├── auth/         # Authentication utilities
│   │   ├── interfaces/   # Service interface definitions
│   │   ├── store/        # Database abstractions
│   │   ├── eventbus/     # Event propagation
│   │   └── utils/        # Common utilities
│   └── contracts/        # API contracts
│       ├── proto/        # gRPC/Protobuf definitions
│       └── openapi/      # REST API specifications
├── projects/             # Independent services (no cross-imports)
│   ├── api/             # API Gateway
│   ├── controller/      # Experiment Controller
│   ├── generator/       # Config Generator
│   ├── dashboard/       # Web UI
│   └── ...             # Other services
└── infrastructure/      # Deployment configurations
```

## Boundary Rules

### 1. No Cross-Project Imports ❌
Projects in `/projects/*` **CANNOT** import from other projects.

```go
// ❌ FORBIDDEN
import "github.com/phoenix/platform/projects/controller/internal/types"

// ✅ ALLOWED
import "github.com/phoenix/platform/packages/go-common/interfaces"
```

### 2. Packages Cannot Import Projects ❌
Packages in `/packages/*` **CANNOT** import from `/projects/*`.

```go
// ❌ FORBIDDEN (in packages/)
import "github.com/phoenix/platform/projects/api/models"

// ✅ ALLOWED (in packages/)
import "github.com/phoenix/platform/packages/go-common/utils"
```

### 3. Shared Code in Packages ✅
All shared code **MUST** be in `/packages/*`.

- Interface definitions → `/packages/go-common/interfaces/`
- Common types → `/packages/go-common/models/`
- Utilities → `/packages/go-common/utils/`
- Proto definitions → `/packages/contracts/proto/`

### 4. Interface-Based Communication ✅
Services communicate through well-defined interfaces.

```go
// In packages/go-common/interfaces/experiment.go
type ExperimentService interface {
    CreateExperiment(ctx context.Context, req *CreateExperimentRequest) (*Experiment, error)
    // ...
}

// In projects/controller/
type Controller struct {
    // Implements ExperimentService
}

// In projects/api/
type APIHandler struct {
    expService interfaces.ExperimentService // Uses interface
}
```

## Go Module Configuration

### Project go.mod
Each project has its own `go.mod` with replace directives:

```go
module github.com/phoenix/platform/projects/myservice

go 1.21

require (
    github.com/phoenix/platform/packages/go-common v0.0.0
    github.com/phoenix/platform/packages/contracts v0.0.0
)

replace github.com/phoenix/platform/packages/go-common => ../../packages/go-common
replace github.com/phoenix/platform/packages/contracts => ../../packages/contracts
```

### Workspace go.work
The root `go.work` file manages the workspace:

```go
go 1.21

use (
    ./packages/go-common
    ./packages/contracts
    ./projects/api
    ./projects/controller
    // ... all other projects
)
```

## Validation and Enforcement

### Automated Validation
Run boundary validation:
```bash
./scripts/validate-boundaries.sh
```

### Continuous Enforcement
Add to pre-commit hooks:
```bash
./scripts/enforce-boundaries.sh
```

### CI/CD Integration
The boundary validation runs in CI to prevent violations from being merged.

## Benefits

1. **Clear Dependencies**: Easy to understand service dependencies
2. **Maintainability**: Changes to shared code are explicit
3. **Testability**: Services can be tested in isolation
4. **Team Autonomy**: Teams can work on projects independently
5. **Architectural Integrity**: Prevents architectural drift

## Migration Support

When migrating code:
1. Identify shared components
2. Move to appropriate `/packages/*` location
3. Update imports in all consumers
4. Run boundary validation
5. Update interface definitions if needed

## Common Patterns

### Service Communication
```go
// Define interface in packages/go-common/interfaces/
type MyService interface {
    DoSomething(ctx context.Context, input Input) (Output, error)
}

// Implement in projects/myservice/
type serviceImpl struct {
    // implementation
}

// Use in projects/consumer/
func NewConsumer(svc interfaces.MyService) *Consumer {
    return &Consumer{service: svc}
}
```

### Shared Types
```go
// In packages/go-common/models/
type SharedModel struct {
    ID   string
    Name string
}

// Used by multiple projects
import "github.com/phoenix/platform/packages/go-common/models"
```

## Troubleshooting

### Import Violation Error
If you see "Cross-project import detected":
1. Move shared code to `/packages/*`
2. Update imports to use the package
3. Run `./scripts/update-imports.sh`

### Missing Replace Directive
If you see "missing replace directive":
1. Add to project's `go.mod`:
   ```
   replace github.com/phoenix/platform/packages/go-common => ../../packages/go-common
   replace github.com/phoenix/platform/packages/contracts => ../../packages/contracts
   ```

### Interface Not Found
If interfaces aren't found:
1. Ensure interface is in `/packages/go-common/interfaces/`
2. Run `go mod tidy` in the project
3. Check replace directives are correct