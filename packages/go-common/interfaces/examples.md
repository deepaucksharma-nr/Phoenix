# Phoenix Platform Interface Examples

This document provides practical examples of how to use the Phoenix platform interfaces in your services.

## Table of Contents
1. [Creating a Service with Interfaces](#creating-a-service-with-interfaces)
2. [Event-Driven Communication](#event-driven-communication)
3. [Service Discovery](#service-discovery)
4. [Error Handling](#error-handling)
5. [Testing with Mocks](#testing-with-mocks)

## Creating a Service with Interfaces

### Example: Building an API Service

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/phoenix/platform/pkg/interfaces"
    "github.com/phoenix/platform/pkg/adapters"
    "github.com/phoenix/platform/pkg/eventbus"
)

func main() {
    // Initialize dependencies
    logger := zap.NewProduction()
    eventBus := eventbus.NewMemoryEventBus(logger)
    
    // Get experiment service (from controller via adapter)
    experimentController := initExperimentController() // Your initialization
    experimentService := adapters.NewExperimentServiceAdapter(
        experimentController,
        eventBus,
        logger,
    )
    
    // Create API handler
    handler := NewAPIHandler(experimentService, eventBus, logger)
    
    // Start HTTP server
    http.ListenAndServe(":8080", handler)
}

type APIHandler struct {
    experimentSvc interfaces.ExperimentService
    eventBus      interfaces.EventBus
    logger        *zap.Logger
}

func (h *APIHandler) CreateExperiment(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var req interfaces.CreateExperimentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Create experiment via interface
    exp, err := h.experimentSvc.CreateExperiment(r.Context(), &req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Return response
    json.NewEncoder(w).Encode(exp)
}
```

## Event-Driven Communication

### Publishing Events

```go
func (s *ExperimentService) StartExperiment(ctx context.Context, id string) error {
    // Start the experiment
    if err := s.doStart(ctx, id); err != nil {
        return err
    }
    
    // Publish event
    event := &interfaces.BaseEvent{
        ID:        uuid.New().String(),
        Type:      interfaces.EventTypeExperimentStarted,
        Source:    "experiment-service",
        Timestamp: time.Now(),
        Data: &interfaces.ExperimentStateChangedEvent{
            ExperimentID: id,
            FromState:    interfaces.ExperimentStatePending,
            ToState:      interfaces.ExperimentStateRunning,
        },
    }
    
    return s.eventBus.Publish(ctx, event)
}
```

### Subscribing to Events

```go
func (s *PipelineService) Start(ctx context.Context) error {
    // Subscribe to experiment events
    eventChan, err := s.eventBus.Subscribe(ctx, interfaces.EventFilter{
        Types: []string{
            interfaces.EventTypeExperimentStarted,
            interfaces.EventTypeExperimentStopped,
        },
    })
    if err != nil {
        return err
    }
    
    // Process events
    go func() {
        for event := range eventChan {
            switch event.GetType() {
            case interfaces.EventTypeExperimentStarted:
                s.handleExperimentStarted(ctx, event)
            case interfaces.EventTypeExperimentStopped:
                s.handleExperimentStopped(ctx, event)
            }
        }
    }()
    
    return nil
}

func (s *PipelineService) handleExperimentStarted(ctx context.Context, event interfaces.Event) {
    data := event.GetData().(*interfaces.ExperimentStateChangedEvent)
    
    // Deploy pipelines for the experiment
    if err := s.deployPipelines(ctx, data.ExperimentID); err != nil {
        s.logger.Error("failed to deploy pipelines", 
            zap.String("experiment_id", data.ExperimentID),
            zap.Error(err),
        )
        return
    }
    
    // Publish pipeline deployed event
    deployEvent := &interfaces.BaseEvent{
        ID:        uuid.New().String(),
        Type:      interfaces.EventTypePipelineDeployed,
        Source:    "pipeline-service",
        Timestamp: time.Now(),
        Data: &interfaces.PipelineDeployedEvent{
            ExperimentID: data.ExperimentID,
            NodeCount:    10,
        },
    }
    
    s.eventBus.Publish(ctx, deployEvent)
}
```

## Service Discovery

### Registering a Service

```go
func registerService(registry interfaces.ServiceRegistry) error {
    instance := &interfaces.ServiceInstance{
        ID:       uuid.New().String(),
        Name:     "experiment-service",
        Version:  "v1.0.0",
        Address:  getLocalIP(),
        Port:     5050,
        Protocol: "grpc",
        Status:   interfaces.HealthStatusHealthy,
        HealthCheck: &interfaces.HealthCheckConfig{
            Path:     "/health",
            Interval: 10 * time.Second,
            Timeout:  5 * time.Second,
        },
        Metadata: map[string]string{
            "region": "us-east-1",
            "zone":   "a",
        },
    }
    
    return registry.Register(context.Background(), instance)
}
```

### Discovering and Calling Services

```go
func callExperimentService(registry interfaces.ServiceRegistry, client interfaces.ServiceClient) error {
    // Discover service instances
    instances, err := registry.Discover(context.Background(), "experiment-service")
    if err != nil {
        return err
    }
    
    if len(instances) == 0 {
        return fmt.Errorf("no experiment service instances found")
    }
    
    // Use load balancer to select instance
    lb := NewRoundRobinLoadBalancer()
    instance, err := lb.Choose(instances)
    if err != nil {
        return err
    }
    
    // Make service call with retry
    var response interfaces.Experiment
    err = client.CallWithRetry(
        context.Background(),
        fmt.Sprintf("%s:%d", instance.Address, instance.Port),
        "GetExperiment",
        &GetExperimentRequest{ID: "exp-123"},
        &response,
        &interfaces.RetryConfig{
            MaxAttempts:  3,
            InitialDelay: 100 * time.Millisecond,
            MaxDelay:     5 * time.Second,
            Multiplier:   2.0,
        },
    )
    
    return err
}
```

## Error Handling

### Domain-Specific Errors

```go
// Define domain errors
type ExperimentError struct {
    Code    string
    Message string
    Details map[string]interface{}
}

func (e *ExperimentError) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Use in service implementation
func (s *ExperimentService) CreateExperiment(ctx context.Context, req *interfaces.CreateExperimentRequest) (*interfaces.Experiment, error) {
    // Validate request
    if req.Name == "" {
        return nil, &ExperimentError{
            Code:    "INVALID_NAME",
            Message: "Experiment name is required",
            Details: map[string]interface{}{
                "field": "name",
            },
        }
    }
    
    // Check for duplicate
    existing, _ := s.store.GetExperimentByName(ctx, req.Name)
    if existing != nil {
        return nil, &ExperimentError{
            Code:    "DUPLICATE_NAME",
            Message: "Experiment with this name already exists",
            Details: map[string]interface{}{
                "name":        req.Name,
                "existing_id": existing.ID,
            },
        }
    }
    
    // Create experiment
    exp, err := s.doCreate(ctx, req)
    if err != nil {
        return nil, &ExperimentError{
            Code:    "CREATE_FAILED",
            Message: "Failed to create experiment",
            Details: map[string]interface{}{
                "error": err.Error(),
            },
        }
    }
    
    return exp, nil
}
```

## Testing with Mocks

### Unit Testing a Service

```go
func TestExperimentService_CreateExperiment(t *testing.T) {
    // Create mocks
    mockStore := new(mocks.MockExperimentStore)
    mockEventBus := new(mocks.MockEventBus)
    
    // Create service
    service := NewExperimentService(mockStore, mockEventBus, zap.NewNop())
    
    // Set up test data
    req := &interfaces.CreateExperimentRequest{
        Name:              "Test Experiment",
        BaselinePipeline:  "baseline-v1",
        CandidatePipeline: "candidate-v1",
        TargetNodes:       []string{"node-1"},
        Config: &interfaces.ExperimentConfig{
            Duration: 30 * time.Minute,
        },
    }
    
    // Set expectations
    mockStore.On("CreateExperiment", mock.Anything, mock.MatchedBy(func(exp *interfaces.Experiment) bool {
        return exp.Name == "Test Experiment"
    })).Return(nil)
    
    mockEventBus.On("Publish", mock.Anything, mock.MatchedBy(func(event interfaces.Event) bool {
        return event.GetType() == interfaces.EventTypeExperimentCreated
    })).Return(nil)
    
    // Execute
    exp, err := service.CreateExperiment(context.Background(), req)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, exp)
    assert.Equal(t, "Test Experiment", exp.Name)
    assert.Equal(t, interfaces.ExperimentStatePending, exp.State)
    
    // Verify mock expectations
    mockStore.AssertExpectations(t)
    mockEventBus.AssertExpectations(t)
}
```

### Integration Testing with Real Components

```go
func TestExperimentWorkflow_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Create real components
    logger := zap.NewNop()
    eventBus := eventbus.NewMemoryEventBus(logger)
    store := createTestStore(t) // Helper to create test DB
    
    // Create services
    expService := adapters.NewExperimentServiceAdapter(
        controller,
        eventBus,
        logger,
    )
    
    // Subscribe to events
    eventChan, _ := eventBus.Subscribe(context.Background(), interfaces.EventFilter{})
    
    // Create and start experiment
    req := createTestExperimentRequest()
    exp, err := expService.CreateExperiment(context.Background(), req)
    require.NoError(t, err)
    
    err = expService.StartExperiment(context.Background(), exp.ID)
    require.NoError(t, err)
    
    // Verify events
    verifyEventSequence(t, eventChan, []string{
        interfaces.EventTypeExperimentCreated,
        interfaces.EventTypeExperimentStarted,
    })
    
    // Verify final state
    result, err := expService.GetExperiment(context.Background(), exp.ID)
    require.NoError(t, err)
    assert.Equal(t, interfaces.ExperimentStateRunning, result.State)
}
```

## Best Practices

1. **Always use interfaces** for dependencies, not concrete types
2. **Handle context cancellation** in all interface methods
3. **Return meaningful errors** with context about what failed
4. **Mock interfaces for unit tests**, use real implementations for integration tests
5. **Document interface contracts** clearly in comments
6. **Version interfaces carefully** - breaking changes require new versions
7. **Keep interfaces small** - prefer many small interfaces over few large ones