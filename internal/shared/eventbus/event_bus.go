package eventbus

import (
	"context"
	"fmt"
	"reflect"
)

// IntegrationEvent is the shared marker interface for integration events across modules.
type IntegrationEvent interface {
	IsIntegrationEvent()
}

type handlerFunc func(ctx context.Context, event IntegrationEvent) error

// EventBus routes integration events to their registered handlers.
type EventBus interface {
	Publish(ctx context.Context, event IntegrationEvent) error
	subscribe(t reflect.Type, h handlerFunc)
}

type inMemoryBus struct {
	handlers map[reflect.Type][]handlerFunc
}

func NewInMemoryEventBus() EventBus {
	return &inMemoryBus{
		handlers: make(map[reflect.Type][]handlerFunc),
	}
}

func (b *inMemoryBus) subscribe(t reflect.Type, h handlerFunc) {
	b.handlers[t] = append(b.handlers[t], h)
}

// Publish delivers the event to every handler registered for its concrete type,
// in registration order. Stops on first error, wrapping it with the event type name.
func (b *inMemoryBus) Publish(ctx context.Context, event IntegrationEvent) error {
	t := reflect.TypeOf(event)
	for _, h := range b.handlers[t] {
		if err := h(ctx, event); err != nil {
			return fmt.Errorf("handle %s: %w", t.Name(), err)
		}
	}
	return nil
}

// Subscribe wires a typed handler for integration event type T.
func Subscribe[T IntegrationEvent](bus EventBus, h func(ctx context.Context, event T) error) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	bus.subscribe(t, func(ctx context.Context, event IntegrationEvent) error {
		return h(ctx, event.(T))
	})
}
