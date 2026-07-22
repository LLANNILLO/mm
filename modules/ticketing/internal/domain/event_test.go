package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func newTestEvent(t *testing.T) *Event {
	t.Helper()

	return NewEvent(uuid.New(), "Concert", nil, nil, time.Now().Add(time.Hour), nil)
}

func TestEvent_Reschedule_RaisesDomainEvent(t *testing.T) {
	e := newTestEvent(t)
	newStart := time.Now().Add(2 * time.Hour)
	newEnd := time.Now().Add(3 * time.Hour)

	e.Reschedule(newStart, &newEnd)

	domainEvent := assertDomainEventPublished[EventRescheduledDomainEvent](t, e)
	assert.Equal(t, e.ID(), domainEvent.EventID)
	assert.True(t, newStart.Equal(domainEvent.StartsAtUtc))
}

func TestEvent_Cancel_RaisesDomainEvent_WhenNotCancelled(t *testing.T) {
	e := newTestEvent(t)

	e.Cancel()

	assert.True(t, e.Cancelled())
	domainEvent := assertDomainEventPublished[EventCancelledDomainEvent](t, e)
	assert.Equal(t, e.ID(), domainEvent.EventID)
}

func TestEvent_Cancel_DoesNotRaiseDomainEvent_WhenAlreadyCancelled(t *testing.T) {
	e := newTestEvent(t)
	e.Cancel()
	e.ClearDomainEvents()

	e.Cancel()

	assertNoDomainEventPublished[EventCancelledDomainEvent](t, e)
}

func TestEvent_PaymentsRefunded_RaisesDomainEvent(t *testing.T) {
	e := newTestEvent(t)

	e.PaymentsRefunded()

	domainEvent := assertDomainEventPublished[EventPaymentsRefundedDomainEvent](t, e)
	assert.Equal(t, e.ID(), domainEvent.EventID)
}

func TestEvent_TicketsArchived_RaisesDomainEvent(t *testing.T) {
	e := newTestEvent(t)

	e.TicketsArchived()

	domainEvent := assertDomainEventPublished[EventTicketsArchivedDomainEvent](t, e)
	assert.Equal(t, e.ID(), domainEvent.EventID)
}
