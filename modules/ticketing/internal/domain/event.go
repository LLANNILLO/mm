package domain

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	entity
	ID          uuid.UUID
	Title       string
	Description *string
	Location    *string
	StartsAtUtc time.Time
	EndsAtUtc   *time.Time
	Cancelled   bool
}

func NewEvent(
	id uuid.UUID,
	title string,
	description *string,
	location *string,
	startsAtUtc time.Time,
	endsAtUtc *time.Time,
) *Event {
	return &Event{
		ID:          id,
		Title:       title,
		Description: description,
		Location:    location,
		StartsAtUtc: startsAtUtc,
		EndsAtUtc:   endsAtUtc,
		Cancelled:   false,
	}
}

func (e *Event) Reschedule(startsAtUtc time.Time, endsAtUtc *time.Time) {
	e.StartsAtUtc = startsAtUtc
	e.EndsAtUtc = endsAtUtc
	e.raise(EventRescheduledDomainEvent{
		EventID:     e.ID,
		StartsAtUtc: startsAtUtc,
		EndsAtUtc:   endsAtUtc,
	})
}

func (e *Event) Cancel() {
	if e.Cancelled {
		return
	}
	e.Cancelled = true
	e.raise(EventCancelledDomainEvent{EventID: e.ID})
}

func (e *Event) PaymentsRefunded() {
	e.raise(EventPaymentsRefundedDomainEvent{EventID: e.ID})
}

func (e *Event) TicketsArchived() {
	e.raise(EventTicketsArchivedDomainEvent{EventID: e.ID})
}
