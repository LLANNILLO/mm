package domain

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	entity
	id          uuid.UUID
	title       string
	description *string
	location    *string
	startsAtUtc time.Time
	endsAtUtc   *time.Time
	cancelled   bool
}

func (e *Event) ID() uuid.UUID          { return e.id }
func (e *Event) Title() string          { return e.title }
func (e *Event) Description() *string   { return e.description }
func (e *Event) Location() *string      { return e.location }
func (e *Event) StartsAtUtc() time.Time { return e.startsAtUtc }
func (e *Event) EndsAtUtc() *time.Time  { return e.endsAtUtc }
func (e *Event) Cancelled() bool        { return e.cancelled }

func NewEvent(
	id uuid.UUID,
	title string,
	description *string,
	location *string,
	startsAtUtc time.Time,
	endsAtUtc *time.Time,
) *Event {
	return &Event{
		id:          id,
		title:       title,
		description: description,
		location:    location,
		startsAtUtc: startsAtUtc,
		endsAtUtc:   endsAtUtc,
		cancelled:   false,
	}
}

// RehydrateEvent reconstructs an Event replica from persisted state without
// raising domain events. Only repositories may call this.
func RehydrateEvent(
	id uuid.UUID,
	title string,
	description, location *string,
	startsAtUtc time.Time,
	endsAtUtc *time.Time,
	cancelled bool,
) *Event {
	return &Event{
		id:          id,
		title:       title,
		description: description,
		location:    location,
		startsAtUtc: startsAtUtc,
		endsAtUtc:   endsAtUtc,
		cancelled:   cancelled,
	}
}

func (e *Event) Reschedule(startsAtUtc time.Time, endsAtUtc *time.Time) {
	e.startsAtUtc = startsAtUtc
	e.endsAtUtc = endsAtUtc
	e.raise(EventRescheduledDomainEvent{
		EventID:     e.id,
		StartsAtUtc: startsAtUtc,
		EndsAtUtc:   endsAtUtc,
	})
}

func (e *Event) Cancel() {
	if e.cancelled {
		return
	}
	e.cancelled = true
	e.raise(EventCancelledDomainEvent{EventID: e.id})
}

func (e *Event) PaymentsRefunded() {
	e.raise(EventPaymentsRefundedDomainEvent{EventID: e.id})
}

func (e *Event) TicketsArchived() {
	e.raise(EventTicketsArchivedDomainEvent{EventID: e.id})
}
