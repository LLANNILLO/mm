package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestEvent creates a draft Event starting one hour after now, with no end date.
func newTestEvent(t *testing.T, now time.Time) *Event {
	t.Helper()

	title := "Concert"
	e, err := NewEvent(uuid.New(), title, nil, nil, now.Add(time.Hour), nil, now)
	require.NoError(t, err)

	return e
}

func TestNewEvent_ReturnsError(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		startsAtUtc time.Time
		endsAtUtc   *time.Time
		wantErr     error
	}{
		{
			name:        "start date is in the past",
			startsAtUtc: now.Add(-time.Minute),
			endsAtUtc:   nil,
			wantErr:     ErrEventStartDateInPast,
		},
		{
			name:        "end date precedes start date",
			startsAtUtc: now.Add(time.Hour),
			endsAtUtc:   ptr(now.Add(time.Minute)),
			wantErr:     ErrEventEndBeforeStart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEvent(uuid.New(), "Concert", nil, nil, tt.startsAtUtc, tt.endsAtUtc, now)

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNewEvent_RaisesDomainEvent_WhenCreated(t *testing.T) {
	now := time.Now()

	e := newTestEvent(t, now)

	domainEvent := assertDomainEventPublished[EventCreatedDomainEvent](t, e)
	assert.Equal(t, e.ID(), domainEvent.EventID)
}

func TestEvent_Publish_ReturnsError_WhenNotDraft(t *testing.T) {
	now := time.Now()
	e := newTestEvent(t, now)
	require.NoError(t, e.Publish())

	err := e.Publish()

	assert.ErrorIs(t, err, ErrEventNotDraft)
}

func TestEvent_Publish_RaisesDomainEvent_WhenPublished(t *testing.T) {
	now := time.Now()
	e := newTestEvent(t, now)

	err := e.Publish()

	require.NoError(t, err)
	domainEvent := assertDomainEventPublished[EventPublishedDomainEvent](t, e)
	assert.Equal(t, e.ID(), domainEvent.EventID)
}

func TestEvent_Cancel_RaisesDomainEvent_WhenCancelled(t *testing.T) {
	now := time.Now()
	e := newTestEvent(t, now)

	err := e.Cancel(now)

	require.NoError(t, err)
	domainEvent := assertDomainEventPublished[EventCancelledDomainEvent](t, e)
	assert.Equal(t, e.ID(), domainEvent.EventID)
}

func TestEvent_Cancel_ReturnsError_WhenAlreadyCancelled(t *testing.T) {
	now := time.Now()
	e := newTestEvent(t, now)
	require.NoError(t, e.Cancel(now))

	err := e.Cancel(now)

	assert.ErrorIs(t, err, ErrEventAlreadyCancelled)
}

func TestEvent_Cancel_ReturnsError_WhenAlreadyStarted(t *testing.T) {
	now := time.Now()
	e := newTestEvent(t, now)

	err := e.Cancel(now.Add(2 * time.Hour))

	assert.ErrorIs(t, err, ErrEventAlreadyStarted)
}

func TestEvent_Reschedule_RaisesDomainEvent_WhenChanged(t *testing.T) {
	now := time.Now()
	e := newTestEvent(t, now)
	newStart := now.Add(2 * time.Hour)
	newEnd := now.Add(3 * time.Hour)

	err := e.Reschedule(newStart, &newEnd)

	require.NoError(t, err)
	domainEvent := assertDomainEventPublished[EventRescheduledDomainEvent](t, e)
	assert.Equal(t, e.ID(), domainEvent.EventID)
	assert.True(t, newStart.Equal(domainEvent.StartsAtUtc))
	require.NotNil(t, domainEvent.EndsAtUtc)
	assert.True(t, newEnd.Equal(*domainEvent.EndsAtUtc))
}

func TestEvent_Reschedule_DoesNotRaiseDomainEvent_WhenUnchanged(t *testing.T) {
	now := time.Now()
	e := newTestEvent(t, now)
	e.ClearDomainEvents()

	err := e.Reschedule(e.StartsAtUtc(), e.EndsAtUtc())

	require.NoError(t, err)
	assert.Empty(t, e.DomainEvents())
}

func ptr[T any](v T) *T { return &v }
