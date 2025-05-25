# Phoenix Platform Interface Contracts

This document describes the interfaces that define contracts between services in the Phoenix Platform.

## Core Interfaces

### ExperimentService
Location: `packages/go-common/interfaces/experiment.go`

Implemented by:
- `projects/controller` - Core experiment management
- `projects/api` - REST API exposure

Consumed by:
- `projects/phoenix-cli` - CLI commands
- `projects/dashboard` - Web UI

### ConfigGenerator
Location: `packages/go-common/interfaces/pipeline.go`

Implemented by:
- `projects/generator` - Pipeline configuration generation

Consumed by:
- `projects/controller` - Experiment controller

### EventBus
Location: `packages/go-common/interfaces/events.go`

Implemented by:
- `packages/go-common/eventbus` - In-memory implementation

Consumed by:
- All services for event propagation

## Boundary Rules

1. **No Cross-Project Imports**: Projects cannot import from other projects
2. **Shared Code in Packages**: All shared code must be in `/packages/*`
3. **Interface-Based Communication**: Services communicate through defined interfaces
4. **Proto/OpenAPI Contracts**: External APIs defined in `/packages/contracts/*`

## Validation

Run `./scripts/validate-boundaries.sh` to check boundary compliance.
