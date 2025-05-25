package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/phoenix/platform/pkg/interfaces"
)

// MockEventBus is a mock implementation of EventBus
type MockEventBus struct {
	mock.Mock
}

// Publish mocks the Publish method
func (m *MockEventBus) Publish(ctx context.Context, event interfaces.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// Subscribe mocks the Subscribe method
func (m *MockEventBus) Subscribe(ctx context.Context, filter interfaces.EventFilter) (<-chan interfaces.Event, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan interfaces.Event), args.Error(1)
}

// Unsubscribe mocks the Unsubscribe method
func (m *MockEventBus) Unsubscribe(ctx context.Context, subscriptionID string) error {
	args := m.Called(ctx, subscriptionID)
	return args.Error(0)
}

// PublishBatch mocks the PublishBatch method
func (m *MockEventBus) PublishBatch(ctx context.Context, events []interfaces.Event) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

// Ensure MockEventBus implements EventBus
var _ interfaces.EventBus = (*MockEventBus)(nil)

// MockEvent is a mock implementation of Event
type MockEvent struct {
	mock.Mock
}

func (m *MockEvent) GetID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockEvent) GetType() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockEvent) GetSource() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockEvent) GetTimestamp() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

func (m *MockEvent) GetData() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockEvent) GetMetadata() map[string]string {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]string)
}

// Ensure MockEvent implements Event
var _ interfaces.Event = (*MockEvent)(nil)