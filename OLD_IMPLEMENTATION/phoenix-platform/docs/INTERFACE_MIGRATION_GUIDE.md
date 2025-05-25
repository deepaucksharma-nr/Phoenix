# Interface Migration Guide

This guide helps migrate the Phoenix Platform services to use the new interface-based architecture.

## Overview

The new architecture introduces clean interfaces between modules, enabling:
- Better testability through dependency injection
- Event-driven communication via EventBus
- Service discovery and registry patterns
- Standardized authentication and authorization
- Clear separation of concerns

## Migration Steps

### 1. API Service Migration

The API service has been refactored to use interfaces. Here's how to migrate:

#### Old Structure
```go
// cmd/api/main.go
experimentService := api.NewExperimentService(store, generatorService, logger)
pb.RegisterExperimentServiceServer(grpcServer, experimentService)
```

#### New Structure
```go
// cmd/api/main_new.go
container, err := api.NewContainer(config, logger)
experimentGRPCService := api.NewGRPCExperimentService(
    container.ExperimentService,
    container.EventBus,
    logger,
)
pb.RegisterExperimentServiceServer(grpcServer, experimentGRPCService)
```

#### Migration Checklist
- [x] Create dependency injection container
- [x] Implement gRPC service adapters
- [x] Add event publishing for state changes
- [x] Update authentication to use AuthService interface
- [ ] Replace direct service calls with interface methods
- [ ] Add integration tests using mock implementations

### 2. Controller Service Migration

The Controller service needs to implement the ExperimentService interface:

```go
// Before
type ExperimentController struct {
    // fields
}

// After
type ExperimentController struct {
    interfaces.ExperimentService
    // other fields
}

// Or use the adapter pattern
adapter := adapters.NewExperimentServiceAdapter(controller, eventBus, logger)
```

### 3. Generator Service Migration

The Generator service should implement the ConfigGenerator interface:

```go
// Implement the interface
func (g *Service) GenerateConfig(ctx context.Context, req *interfaces.GenerateConfigRequest) (*interfaces.PipelineConfig, error) {
    // Implementation
}

func (g *Service) ValidateConfig(ctx context.Context, config *interfaces.PipelineConfig) error {
    // Implementation
}
```

### 4. Event-Driven Communication

Replace direct service calls with event publishing:

```go
// Before
generatorService.GeneratePipeline(experimentID)

// After
eventBus.Publish(ctx, interfaces.Event{
    Type: interfaces.EventTypePipelineGenerationRequested,
    Data: map[string]interface{}{
        "experiment_id": experimentID,
    },
})
```

### 5. Service Discovery

Register services with the service registry:

```go
serviceRegistry.Register("experiment", &interfaces.ServiceEndpoint{
    ID:       "experiment-controller-1",
    Name:     "Experiment Controller",
    Endpoint: "controller:8080",
})
```

## Testing with Interfaces

Use mock implementations for testing:

```go
func TestExperimentCreation(t *testing.T) {
    mockService := mocks.NewMockExperimentService(t)
    mockService.On("CreateExperiment", mock.Anything, mock.Anything).
        Return(&interfaces.Experiment{ID: "exp-1"}, nil)
    
    // Test your code using mockService
}
```

## Benefits After Migration

1. **Parallel Development**: Teams can work on different services independently
2. **Better Testing**: Mock implementations enable comprehensive unit tests
3. **Event-Driven Architecture**: Loosely coupled services communicate via events
4. **Service Resilience**: Circuit breakers and retries built into service communication
5. **Standardized Auth**: Consistent authentication and authorization across services

## Common Patterns

### Using the Adapter Pattern
```go
// Wrap existing implementation
adapter := adapters.NewExperimentServiceAdapter(
    existingController,
    eventBus,
    logger,
)
```

### Event Handling
```go
// Subscribe to events
events, _ := eventBus.Subscribe(ctx, interfaces.EventFilter{
    Types: []interfaces.EventType{interfaces.EventTypeExperimentCreated},
})

for event := range events {
    // Handle event
}
```

### Service Communication
```go
// Discover service
endpoints, _ := serviceRegistry.Discover(ctx, "pipeline")

// Create client with circuit breaker
client := interfaces.NewServiceClient(endpoints[0], 
    interfaces.WithCircuitBreaker(5, time.Minute),
    interfaces.WithRetry(3, time.Second),
)
```

## Rollback Plan

If issues arise during migration:

1. The old implementation remains in `main.go`
2. New implementation is in `main_new.go`
3. Switch between them by changing the build target
4. Interfaces are backward compatible with existing protobuf definitions

## Next Steps

1. Complete API service migration (in progress)
2. Migrate Controller service (next priority)
3. Migrate Generator service
4. Add comprehensive integration tests
5. Deploy in staging environment
6. Monitor performance and stability
7. Complete production rollout