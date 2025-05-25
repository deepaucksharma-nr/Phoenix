// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	
	"github.com/phoenix/platform/pkg/adapters"
	"github.com/phoenix/platform/pkg/eventbus"
	"github.com/phoenix/platform/pkg/interfaces"
	"github.com/phoenix/platform/pkg/mocks"
)

// TestInterfaceIntegration demonstrates how the interfaces work together
func TestInterfaceIntegration(t *testing.T) {
	logger := zap.NewNop()
	
	// Create event bus
	eventBus := eventbus.NewMemoryEventBus(logger)
	
	// Create mock store
	mockStore := new(mocks.MockExperimentStore)
	
	// Create experiment service using adapter pattern
	// In real usage, this would wrap the actual controller
	experimentService := createMockExperimentService(mockStore, eventBus, logger)
	
	// Subscribe to experiment events
	ctx := context.Background()
	eventFilter := interfaces.EventFilter{
		Types: []string{
			interfaces.EventTypeExperimentCreated,
			interfaces.EventTypeExperimentStarted,
		},
	}
	
	eventChan, err := eventBus.Subscribe(ctx, eventFilter)
	require.NoError(t, err)
	
	// Create experiment via interface
	createReq := &interfaces.CreateExperimentRequest{
		Name:              "Test Experiment",
		Description:       "Integration test experiment",
		BaselinePipeline:  "process-baseline-v1",
		CandidatePipeline: "process-priority-filter-v1",
		TargetNodes:       []string{"node-1", "node-2"},
		Config: &interfaces.ExperimentConfig{
			Duration: 30 * time.Minute,
			SuccessCriteria: &interfaces.SuccessCriteria{
				MinCardinalityReduction: 50.0,
				CriticalProcessCoverage: 100.0,
			},
		},
	}
	
	// Set up mock expectations
	mockStore.On("CreateExperiment", ctx, mock.Anything).Return(nil)
	
	// Create experiment
	experiment, err := experimentService.CreateExperiment(ctx, createReq)
	require.NoError(t, err)
	assert.NotEmpty(t, experiment.ID)
	assert.Equal(t, "Test Experiment", experiment.Name)
	assert.Equal(t, interfaces.ExperimentStatePending, experiment.State)
	
	// Verify event was published
	select {
	case event := <-eventChan:
		assert.Equal(t, interfaces.EventTypeExperimentCreated, event.GetType())
		eventData := event.GetData().(*interfaces.ExperimentCreatedEvent)
		assert.Equal(t, experiment.ID, eventData.ExperimentID)
	case <-time.After(1 * time.Second):
		t.Fatal("expected experiment created event")
	}
	
	// Start experiment
	mockStore.On("GetExperiment", ctx, experiment.ID).Return(convertToInternal(experiment), nil)
	mockStore.On("UpdateExperimentState", ctx, experiment.ID, mock.Anything).Return(nil)
	
	err = experimentService.StartExperiment(ctx, experiment.ID)
	require.NoError(t, err)
	
	// Verify start event was published
	select {
	case event := <-eventChan:
		assert.Equal(t, interfaces.EventTypeExperimentStarted, event.GetType())
	case <-time.After(1 * time.Second):
		t.Fatal("expected experiment started event")
	}
	
	// Verify all mock expectations were met
	mockStore.AssertExpectations(t)
}

// TestEventDrivenWorkflow demonstrates event-driven communication between services
func TestEventDrivenWorkflow(t *testing.T) {
	logger := zap.NewNop()
	eventBus := eventbus.NewMemoryEventBus(logger)
	
	// Simulate pipeline service subscribing to experiment events
	ctx := context.Background()
	pipelineEventChan, err := eventBus.Subscribe(ctx, interfaces.EventFilter{
		Types: []string{interfaces.EventTypeExperimentStarted},
	})
	require.NoError(t, err)
	
	// Simulate monitoring service subscribing to pipeline events
	monitoringEventChan, err := eventBus.Subscribe(ctx, interfaces.EventFilter{
		Types: []string{interfaces.EventTypePipelineDeployed},
	})
	require.NoError(t, err)
	
	// Start event processing goroutines
	pipelineDeployed := make(chan bool)
	go func() {
		for event := range pipelineEventChan {
			if event.GetType() == interfaces.EventTypeExperimentStarted {
				// Pipeline service would deploy pipelines here
				t.Log("Pipeline service received experiment started event")
				
				// Publish pipeline deployed event
				deployEvent := &interfaces.BaseEvent{
					ID:        "deploy-1",
					Type:      interfaces.EventTypePipelineDeployed,
					Source:    "pipeline-service",
					Timestamp: time.Now(),
					Data: &interfaces.PipelineDeployedEvent{
						PipelineID:     "pipeline-1",
						ExperimentID:   "exp-1",
						NodeCount:      2,
						DeploymentType: "daemonset",
					},
				}
				eventBus.Publish(context.Background(), deployEvent)
				pipelineDeployed <- true
			}
		}
	}()
	
	metricsCollected := make(chan bool)
	go func() {
		for event := range monitoringEventChan {
			if event.GetType() == interfaces.EventTypePipelineDeployed {
				// Monitoring service would start collecting metrics here
				t.Log("Monitoring service received pipeline deployed event")
				metricsCollected <- true
			}
		}
	}()
	
	// Trigger the workflow by publishing experiment started event
	startEvent := &interfaces.BaseEvent{
		ID:        "start-1",
		Type:      interfaces.EventTypeExperimentStarted,
		Source:    "experiment-service",
		Timestamp: time.Now(),
		Data:      map[string]string{"experiment_id": "exp-1"},
	}
	
	err = eventBus.Publish(ctx, startEvent)
	require.NoError(t, err)
	
	// Verify the event chain
	select {
	case <-pipelineDeployed:
		t.Log("Pipeline deployment triggered successfully")
	case <-time.After(2 * time.Second):
		t.Fatal("pipeline deployment was not triggered")
	}
	
	select {
	case <-metricsCollected:
		t.Log("Metrics collection triggered successfully")
	case <-time.After(2 * time.Second):
		t.Fatal("metrics collection was not triggered")
	}
}

// TestServiceDiscovery demonstrates service discovery pattern
func TestServiceDiscovery(t *testing.T) {
	// This would typically use a real service registry like Consul or etcd
	// For now, we'll use a mock to demonstrate the pattern
	
	mockRegistry := new(mocks.MockServiceRegistry)
	
	// Register experiment service
	experimentInstance := &interfaces.ServiceInstance{
		ID:       "exp-service-1",
		Name:     "experiment-service",
		Version:  "v1.0.0",
		Address:  "localhost",
		Port:     5050,
		Protocol: "grpc",
		Status:   interfaces.HealthStatusHealthy,
		HealthCheck: &interfaces.HealthCheckConfig{
			Path:     "/health",
			Interval: 10 * time.Second,
			Timeout:  5 * time.Second,
		},
	}
	
	ctx := context.Background()
	mockRegistry.On("Register", ctx, experimentInstance).Return(nil)
	
	err := mockRegistry.Register(ctx, experimentInstance)
	require.NoError(t, err)
	
	// Discover service
	mockRegistry.On("Discover", ctx, "experiment-service").Return(
		[]*interfaces.ServiceInstance{experimentInstance}, nil,
	)
	
	instances, err := mockRegistry.Discover(ctx, "experiment-service")
	require.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "localhost", instances[0].Address)
	assert.Equal(t, 5050, instances[0].Port)
	
	mockRegistry.AssertExpectations(t)
}

// Helper functions

func createMockExperimentService(store interfaces.ExperimentStore, eventBus interfaces.EventBus, logger *zap.Logger) interfaces.ExperimentService {
	// In a real implementation, this would use the actual adapter
	// For testing, we'll create a simple mock implementation
	return &mockExperimentService{
		store:    store,
		eventBus: eventBus,
		logger:   logger,
	}
}

type mockExperimentService struct {
	store    interfaces.ExperimentStore
	eventBus interfaces.EventBus
	logger   *zap.Logger
}

func (m *mockExperimentService) CreateExperiment(ctx context.Context, req *interfaces.CreateExperimentRequest) (*interfaces.Experiment, error) {
	exp := &interfaces.Experiment{
		ID:                "exp-123",
		Name:              req.Name,
		Description:       req.Description,
		State:             interfaces.ExperimentStatePending,
		BaselinePipeline:  req.BaselinePipeline,
		CandidatePipeline: req.CandidatePipeline,
		TargetNodes:       req.TargetNodes,
		Config:            req.Config,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	
	if err := m.store.CreateExperiment(ctx, exp); err != nil {
		return nil, err
	}
	
	// Publish event
	event := &interfaces.BaseEvent{
		ID:        "evt-1",
		Type:      interfaces.EventTypeExperimentCreated,
		Source:    "experiment-service",
		Timestamp: time.Now(),
		Data: &interfaces.ExperimentCreatedEvent{
			ExperimentID: exp.ID,
			Name:         exp.Name,
			Config:       exp.Config,
		},
	}
	m.eventBus.Publish(ctx, event)
	
	return exp, nil
}

func (m *mockExperimentService) GetExperiment(ctx context.Context, id string) (*interfaces.Experiment, error) {
	return m.store.GetExperiment(ctx, id)
}

func (m *mockExperimentService) UpdateExperiment(ctx context.Context, id string, req *interfaces.UpdateExperimentRequest) (*interfaces.Experiment, error) {
	exp, err := m.store.GetExperiment(ctx, id)
	if err != nil {
		return nil, err
	}
	
	if req.Description != nil {
		exp.Description = *req.Description
	}
	
	return exp, m.store.UpdateExperiment(ctx, exp)
}

func (m *mockExperimentService) DeleteExperiment(ctx context.Context, id string) error {
	return m.store.DeleteExperiment(ctx, id)
}

func (m *mockExperimentService) ListExperiments(ctx context.Context, filter *interfaces.ExperimentFilter) (*interfaces.ExperimentList, error) {
	exps, err := m.store.ListExperiments(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &interfaces.ExperimentList{Experiments: exps}, nil
}

func (m *mockExperimentService) StartExperiment(ctx context.Context, id string) error {
	exp, err := m.store.GetExperiment(ctx, id)
	if err != nil {
		return err
	}
	
	exp.State = interfaces.ExperimentStateRunning
	if err := m.store.UpdateExperimentState(ctx, id, exp.State); err != nil {
		return err
	}
	
	// Publish event
	event := &interfaces.BaseEvent{
		ID:        "evt-2",
		Type:      interfaces.EventTypeExperimentStarted,
		Source:    "experiment-service",
		Timestamp: time.Now(),
		Data:      map[string]string{"experiment_id": id},
	}
	m.eventBus.Publish(ctx, event)
	
	return nil
}

func (m *mockExperimentService) StopExperiment(ctx context.Context, id string) error {
	return m.store.UpdateExperimentState(ctx, id, interfaces.ExperimentStateCancelled)
}

func (m *mockExperimentService) GetExperimentResults(ctx context.Context, id string) (*interfaces.ExperimentResults, error) {
	return &interfaces.ExperimentResults{}, nil
}

func (m *mockExperimentService) PromoteExperiment(ctx context.Context, id string) error {
	return nil
}

func convertToInternal(exp *interfaces.Experiment) interface{} {
	// This would use the actual adapter conversion
	return exp
}