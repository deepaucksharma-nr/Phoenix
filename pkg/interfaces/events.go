package interfaces

import (
	"context"
	"time"
)

// Event represents a domain event in the system
type Event interface {
	GetID() string
	GetType() string
	GetSource() string
	GetTimestamp() time.Time
	GetMetadata() map[string]string
	GetPayload() interface{}
}

// EventFilter defines criteria for filtering events
type EventFilter struct {
	Types     []string          // Event types to match
	Sources   []string          // Event sources to match
	StartTime *time.Time        // Earliest event time
	EndTime   *time.Time        // Latest event time
	Metadata  map[string]string // Metadata key-value pairs to match
}

// EventBus defines the interface for event publishing and subscription
type EventBus interface {
	// Publish sends an event to all matching subscribers
	Publish(ctx context.Context, event Event) error
	
	// PublishBatch sends multiple events atomically
	PublishBatch(ctx context.Context, events []Event) error
	
	// Subscribe creates a subscription to events matching the filter
	Subscribe(ctx context.Context, filter EventFilter) (<-chan Event, error)
	
	// Unsubscribe removes a subscription
	Unsubscribe(ctx context.Context, subscriptionID string) error
	
	// Close shuts down the event bus
	Close() error
}

// BaseEvent provides a basic implementation of the Event interface
type BaseEvent struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Source    string            `json:"source"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata"`
	Payload   interface{}       `json:"payload"`
}

func (e *BaseEvent) GetID() string                { return e.ID }
func (e *BaseEvent) GetType() string              { return e.Type }
func (e *BaseEvent) GetSource() string            { return e.Source }
func (e *BaseEvent) GetTimestamp() time.Time      { return e.Timestamp }
func (e *BaseEvent) GetMetadata() map[string]string { return e.Metadata }
func (e *BaseEvent) GetPayload() interface{}      { return e.Payload }