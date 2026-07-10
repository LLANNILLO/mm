package domain

import (
	"time"

	"github.com/google/uuid"
)

var (
	ErrEventNotFound         = &DomainError{Code: "event.not_found", Message: "event not found", Kind: KindNotFound}
	ErrEventNotDraft         = &DomainError{Code: "event.not_draft", Message: "event is not in draft status", Kind: KindConflict}
	ErrEventAlreadyCancelled = &DomainError{Code: "event.already_cancelled", Message: "event is already cancelled", Kind: KindConflict}
	ErrEventAlreadyStarted   = &DomainError{Code: "event.already_started", Message: "event has already started", Kind: KindConflict}
	ErrEventStartDateInPast  = &DomainError{Code: "event.start_date_in_past", Message: "start date must be in the future", Kind: KindValidation}
	ErrEventEndBeforeStart   = &DomainError{Code: "event.end_before_start", Message: "end date must be after start date", Kind: KindValidation}
)

type Event struct {
	entity
	id          uuid.UUID
	categoryID  uuid.UUID
	title       string
	description *string
	location    *string
	startsAtUtc time.Time
	endsAtUtc   *time.Time
	status      EventStatus
}

func (e *Event) ID() uuid.UUID          { return e.id }
func (e *Event) CategoryID() uuid.UUID  { return e.categoryID }
func (e *Event) Title() string          { return e.title }
func (e *Event) Description() *string   { return e.description }
func (e *Event) Location() *string      { return e.location }
func (e *Event) StartsAtUtc() time.Time { return e.startsAtUtc }
func (e *Event) EndsAtUtc() *time.Time  { return e.endsAtUtc }
func (e *Event) Status() EventStatus    { return e.status }

func NewEvent(
	categoryID uuid.UUID,
	title string,
	description, location *string,
	startsAtUtc time.Time,
	endsAtUtc *time.Time,
	now time.Time,
) (*Event, error) {
	if !startsAtUtc.After(now) {
		return nil, ErrEventStartDateInPast
	}
	if endsAtUtc != nil && !endsAtUtc.After(startsAtUtc) {
		return nil, ErrEventEndBeforeStart
	}

	e := &Event{
		id:          uuid.New(),
		categoryID:  categoryID,
		title:       title,
		description: description,
		location:    location,
		startsAtUtc: startsAtUtc,
		endsAtUtc:   endsAtUtc,
		status:      StatusDraft,
	}
	e.raise(EventCreatedDomainEvent{EventID: e.id})
	return e, nil
}

// RehydrateEvent reconstructs an Event from persisted state without
// re-running creation invariants or raising domain events. Only repositories
// may call this.
func RehydrateEvent(
	id, categoryID uuid.UUID,
	title string,
	description, location *string,
	startsAtUtc time.Time,
	endsAtUtc *time.Time,
	status EventStatus,
) *Event {
	return &Event{
		id:          id,
		categoryID:  categoryID,
		title:       title,
		description: description,
		location:    location,
		startsAtUtc: startsAtUtc,
		endsAtUtc:   endsAtUtc,
		status:      status,
	}
}

func (e *Event) Publish() error {
	if e.status != StatusDraft {
		return ErrEventNotDraft
	}
	e.status = StatusPublished
	e.raise(EventPublishedDomainEvent{EventID: e.id})
	return nil
}

func (e *Event) Cancel(now time.Time) error {
	if e.status == StatusCancelled {
		return ErrEventAlreadyCancelled
	}
	if now.After(e.startsAtUtc) {
		return ErrEventAlreadyStarted
	}
	e.status = StatusCancelled
	e.raise(EventCancelledDomainEvent{EventID: e.id})
	return nil
}

func (e *Event) Reschedule(startsAt time.Time, endsAt *time.Time) error {
	sameStart := e.startsAtUtc.Equal(startsAt)
	sameEnd := (e.endsAtUtc == nil && endsAt == nil) ||
		(e.endsAtUtc != nil && endsAt != nil && e.endsAtUtc.Equal(*endsAt))
	if sameStart && sameEnd {
		return nil
	}
	e.startsAtUtc = startsAt
	e.endsAtUtc = endsAt
	e.raise(EventRescheduledDomainEvent{EventID: e.id, StartsAtUtc: startsAt, EndsAtUtc: endsAt})
	return nil
}
