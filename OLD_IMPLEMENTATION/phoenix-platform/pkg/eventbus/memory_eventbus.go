package eventbus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"github.com/google/uuid"
	
	"github.com/phoenix/platform/pkg/interfaces"
)

// Ensure time is used
var _ = time.Time{}

// MemoryEventBus is an in-memory implementation of EventBus
// Suitable for development and testing, not for production
type MemoryEventBus struct {
	logger        *zap.Logger
	subscribers   map[string]*subscription
	subscribersMu sync.RWMutex
	closed        bool
	closedMu      sync.RWMutex
}

type subscription struct {
	id      string
	filter  interfaces.EventFilter
	channel chan interfaces.Event
	cancel  context.CancelFunc
}

// NewMemoryEventBus creates a new in-memory event bus
func NewMemoryEventBus(logger *zap.Logger) interfaces.EventBus {
	return &MemoryEventBus{
		logger:      logger,
		subscribers: make(map[string]*subscription),
	}
}

// Publish sends an event to all matching subscribers
func (eb *MemoryEventBus) Publish(ctx context.Context, event interfaces.Event) error {
	eb.closedMu.RLock()
	if eb.closed {
		eb.closedMu.RUnlock()
		return fmt.Errorf("event bus is closed")
	}
	eb.closedMu.RUnlock()

	eb.logger.Debug("publishing event",
		zap.String("event_id", event.GetID()),
		zap.String("event_type", event.GetType()),
		zap.String("source", event.GetSource()),
	)

	eb.subscribersMu.RLock()
	defer eb.subscribersMu.RUnlock()

	// Send to all matching subscribers
	for _, sub := range eb.subscribers {
		if eb.matchesFilter(event, sub.filter) {
			select {
			case sub.channel <- event:
				eb.logger.Debug("event sent to subscriber",
					zap.String("subscription_id", sub.id),
					zap.String("event_id", event.GetID()),
				)
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Channel full, log and continue
				eb.logger.Warn("subscriber channel full, dropping event",
					zap.String("subscription_id", sub.id),
					zap.String("event_id", event.GetID()),
				)
			}
		}
	}

	return nil
}

// Subscribe creates a subscription to events matching the filter
func (eb *MemoryEventBus) Subscribe(ctx context.Context, filter interfaces.EventFilter) (<-chan interfaces.Event, error) {
	eb.closedMu.RLock()
	if eb.closed {
		eb.closedMu.RUnlock()
		return nil, fmt.Errorf("event bus is closed")
	}
	eb.closedMu.RUnlock()

	// Create subscription
	subID := uuid.New().String()
	ch := make(chan interfaces.Event, 100) // Buffer of 100 events
	
	// Create cancellable context
	subCtx, cancel := context.WithCancel(ctx)
	
	sub := &subscription{
		id:      subID,
		filter:  filter,
		channel: ch,
		cancel:  cancel,
	}

	eb.subscribersMu.Lock()
	eb.subscribers[subID] = sub
	eb.subscribersMu.Unlock()

	eb.logger.Info("created subscription",
		zap.String("subscription_id", subID),
		zap.Any("filter", filter),
	)

	// Start goroutine to handle context cancellation
	go func() {
		<-subCtx.Done()
		eb.Unsubscribe(context.Background(), subID)
		close(ch)
	}()

	return ch, nil
}

// Unsubscribe removes a subscription
func (eb *MemoryEventBus) Unsubscribe(ctx context.Context, subscriptionID string) error {
	eb.subscribersMu.Lock()
	defer eb.subscribersMu.Unlock()

	sub, exists := eb.subscribers[subscriptionID]
	if !exists {
		return fmt.Errorf("subscription %s not found", subscriptionID)
	}

	// Cancel the subscription context
	sub.cancel()
	
	// Remove from map
	delete(eb.subscribers, subscriptionID)

	eb.logger.Info("removed subscription", zap.String("subscription_id", subscriptionID))
	return nil
}

// PublishBatch sends multiple events atomically
func (eb *MemoryEventBus) PublishBatch(ctx context.Context, events []interfaces.Event) error {
	eb.closedMu.RLock()
	if eb.closed {
		eb.closedMu.RUnlock()
		return fmt.Errorf("event bus is closed")
	}
	eb.closedMu.RUnlock()

	// Publish each event
	for _, event := range events {
		if err := eb.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", event.GetID(), err)
		}
	}

	return nil
}

// Close shuts down the event bus
func (eb *MemoryEventBus) Close() error {
	eb.closedMu.Lock()
	eb.closed = true
	eb.closedMu.Unlock()

	// Cancel all subscriptions
	eb.subscribersMu.Lock()
	for _, sub := range eb.subscribers {
		sub.cancel()
	}
	eb.subscribers = make(map[string]*subscription)
	eb.subscribersMu.Unlock()

	eb.logger.Info("event bus closed")
	return nil
}

// matchesFilter checks if an event matches a subscription filter
func (eb *MemoryEventBus) matchesFilter(event interfaces.Event, filter interfaces.EventFilter) bool {
	// Check event type
	if len(filter.Types) > 0 {
		matched := false
		for _, t := range filter.Types {
			if t == event.GetType() {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check source
	if len(filter.Sources) > 0 {
		matched := false
		for _, s := range filter.Sources {
			if s == event.GetSource() {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check time range
	eventTime := event.GetTimestamp()
	if filter.StartTime != nil && eventTime.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && eventTime.After(*filter.EndTime) {
		return false
	}

	// Check metadata
	if len(filter.Metadata) > 0 {
		eventMeta := event.GetMetadata()
		for k, v := range filter.Metadata {
			if eventMeta[k] != v {
				return false
			}
		}
	}

	return true
}

// Ensure MemoryEventBus implements EventBus
var _ interfaces.EventBus = (*MemoryEventBus)(nil)