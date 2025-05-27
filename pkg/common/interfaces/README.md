# Phoenix Platform Interface Definitions

Interface contracts for Phoenix Platform's 70% cost reduction observability system. Defines clean boundaries for agent-based architecture, A/B testing, and real-time monitoring components.

## Overview

The interfaces are organized by domain:

### Core Domain Interfaces

1. **Experiment Management** (`experiment.go`)
   - `ExperimentService`: Core business logic for experiment lifecycle
   - `ExperimentStore`: Persistence layer for experiments
   - Consumed by: API Service, Controller
   - Implemented by: Experiment Controller, PostgreSQL Store

2. **Pipeline Management** (`pipeline.go`)
   - `PipelineService`: Template-based pipeline deployment
   - `PipelineRenderer`: OpenTelemetry configuration rendering
   - `TaskQueue`: Agent task distribution system
   - Consumed by: API Service, Agent System
   - Implemented by: Template Renderer, PostgreSQL Queue

3. **Monitoring & Metrics** (`monitoring.go`)
   - `MetricsCollector`: Agent metrics collection
   - `CostAnalyzer`: Real-time cost savings calculation
   - `WebSocketHub`: Live update broadcasting
   - Consumed by: Dashboard, API Service
   - Implemented by: Prometheus Integration, WebSocket Service

4. **Agent Management** (`agent.go`)
   - `AgentService`: Agent registration and health monitoring
   - `TaskDistributor`: Task assignment and polling
   - `HeartbeatCollector`: Agent status tracking
   - Consumed by: API Service, Task Queue
   - Implemented by: Agent Controller, PostgreSQL Store

### Infrastructure Interfaces

5. **Event-Driven Communication** (`events.go`)
   - `EventBus`: Asynchronous event publishing/subscription
   - `EventProcessor`: Event handling and processing
   - `WorkflowEngine`: Complex multi-step orchestration
   - Used for: Service decoupling, async operations

6. **Service Communication** (`service.go`)
   - `ServiceRegistry`: Service discovery
   - `ServiceClient`: Inter-service communication
   - `LoadBalancer`: Request distribution
   - `CircuitBreaker`: Fault tolerance
   - Used for: Service mesh, resilience

## Interface Design Principles

### 1. Single Responsibility
Each interface focuses on a specific domain capability without mixing concerns.

### 2. Dependency Inversion
High-level modules depend on interfaces, not concrete implementations.

### 3. Interface Segregation
Interfaces are small and focused rather than large and monolithic.

### 4. Explicit Contracts
All methods have clear input/output types with validation rules.

## Usage Examples

### Implementing an Interface

```go
package controller

import (
    "context"
    "github.com/phoenix/platform/pkg/interfaces"
)

type experimentService struct {
    store interfaces.ExperimentStore
    eventBus interfaces.EventBus
}

func NewExperimentService(store interfaces.ExperimentStore, eventBus interfaces.EventBus) interfaces.ExperimentService {
    return &experimentService{
        store: store,
        eventBus: eventBus,
    }
}

func (s *experimentService) CreateExperiment(ctx context.Context, req *interfaces.CreateExperimentRequest) (*interfaces.Experiment, error) {
    // Implementation
}
```

### Consuming an Interface

```go
package api

import (
    "github.com/phoenix/platform/pkg/interfaces"
)

type apiServer struct {
    experimentSvc interfaces.ExperimentService
    pipelineSvc   interfaces.PipelineService
}

func NewAPIServer(experimentSvc interfaces.ExperimentService, pipelineSvc interfaces.PipelineService) *apiServer {
    return &apiServer{
        experimentSvc: experimentSvc,
        pipelineSvc: pipelineSvc,
    }
}
```

## Testing with Interfaces

Interfaces enable easy mocking for unit tests:

```go
package api_test

import (
    "testing"
    "github.com/stretchr/testify/mock"
    "github.com/phoenix/platform/pkg/interfaces"
)

type mockExperimentService struct {
    mock.Mock
}

func (m *mockExperimentService) CreateExperiment(ctx context.Context, req *interfaces.CreateExperimentRequest) (*interfaces.Experiment, error) {
    args := m.Called(ctx, req)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*interfaces.Experiment), args.Error(1)
}

func TestAPIServer_CreateExperiment(t *testing.T) {
    mockSvc := new(mockExperimentService)
    mockSvc.On("CreateExperiment", mock.Anything, mock.Anything).Return(&interfaces.Experiment{ID: "123"}, nil)
    
    // Test implementation
}
```

## Service Integration Map

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Dashboard     │────▶│   API Gateway   │────▶│ Experiment Svc  │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │                          │
                               ▼                          ▼
                        ┌─────────────────┐     ┌─────────────────┐
                        │  Pipeline Svc   │     │  Monitoring Svc │
                        └─────────────────┘     └─────────────────┘
                               │                          │
                               ▼                          ▼
                        ┌─────────────────┐     ┌─────────────────┐
                        │ Config Generator│     │   Prometheus    │
                        └─────────────────┘     └─────────────────┘
```

## Adding New Interfaces

When adding new interfaces:

1. Define the interface in the appropriate domain file
2. Include comprehensive documentation
3. Define all required types (requests, responses, models)
4. Add validation tags where appropriate
5. Update this README with the new interface
6. Create mock implementations for testing

## Interface Versioning

- Interfaces are versioned through the API version (v1, v2, etc.)
- Breaking changes require a new version
- Deprecation notices should be added to outdated methods
- Maintain backward compatibility where possible

## Best Practices

1. **Keep interfaces small** - 5-10 methods maximum
2. **Use context.Context** - All methods should accept context
3. **Return errors** - All methods that can fail should return error
4. **Use pointer receivers** - For consistency and nil-safety
5. **Document everything** - Every type and method needs documentation
6. **Validate inputs** - Use struct tags for validation rules
7. **Version carefully** - Breaking changes need new versions

## Future Enhancements

- Add OpenAPI generation from interfaces
- Create interface compliance tests
- Add performance benchmarks for implementations
- Generate client SDKs from interfaces