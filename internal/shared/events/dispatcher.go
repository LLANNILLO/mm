package events

import (
	"context"
	"fmt"
	"reflect"
)

// DomainEvent is the shared marker interface for domain events across all modules.
type DomainEvent interface {
	IsDomainEvent()
}

type handlerFunc func(ctx context.Context, event DomainEvent) error

// Dispatcher routes domain events to their registered handlers.
type Dispatcher struct {
	handlers map[reflect.Type][]handlerFunc
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[reflect.Type][]handlerFunc),
	}
}

// Register wires a typed handler for domain event type T.
func Register[T DomainEvent](d *Dispatcher, h func(ctx context.Context, event T) error) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	d.handlers[t] = append(d.handlers[t], func(ctx context.Context, event DomainEvent) error {
		return h(ctx, event.(T))
	})
}

// Dispatch sends each event to all handlers registered for its concrete type.
// Events with no registered handlers are silently ignored.
func (d *Dispatcher) Dispatch(ctx context.Context, domainEvents []DomainEvent) error {
	for _, e := range domainEvents {
		t := reflect.TypeOf(e)
		for _, h := range d.handlers[t] {
			if err := h(ctx, e); err != nil {
				return fmt.Errorf("handle %s: %w", t.Name(), err)
			}
		}
	}
	return nil
}
