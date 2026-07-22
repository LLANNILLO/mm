package domain

import (
	"testing"

	sharedevents "github.com/llannillo/mm/internal/shared/events"
	"github.com/stretchr/testify/require"
)

// domainEventSource is implemented by entity and lets tests assert on raised
// domain events without depending on any concrete aggregate.
type domainEventSource interface {
	DomainEvents() []sharedevents.DomainEvent
}

// assertDomainEventPublished fails the test unless exactly one event of type T
// was raised on src, and returns it for further assertions on its fields.
func assertDomainEventPublished[T sharedevents.DomainEvent](t *testing.T, src domainEventSource) T {
	t.Helper()

	var found []T
	for _, e := range src.DomainEvents() {
		if v, ok := e.(T); ok {
			found = append(found, v)
		}
	}

	require.Lenf(t, found, 1, "expected exactly one %T to be published, got %d", *new(T), len(found))
	return found[0]
}

// assertNoDomainEventPublished fails the test if any event of type T was raised on src.
func assertNoDomainEventPublished[T sharedevents.DomainEvent](t *testing.T, src domainEventSource) {
	t.Helper()

	for _, e := range src.DomainEvents() {
		if _, ok := e.(T); ok {
			require.Failf(t, "unexpected domain event", "expected no %T to be published", *new(T))
		}
	}
}
