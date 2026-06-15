package domain

import (
	"time"

	"github.com/google/uuid"
)

var (
	ErrEventNotFound        = &DomainError{Code: "event.not_found", Message: "event not found", Kind: KindNotFound}
	ErrEventNotDraft        = &DomainError{Code: "event.not_draft", Message: "event is not in draft status", Kind: KindConflict}
	ErrEventAlreadyCancelled = &DomainError{Code: "event.already_cancelled", Message: "event is already cancelled", Kind: KindConflict}
	ErrEventAlreadyStarted  = &DomainError{Code: "event.already_started", Message: "event has already started", Kind: KindConflict}
	ErrEventStartDateInPast = &DomainError{Code: "event.start_date_in_past", Message: "start date must be in the future", Kind: KindValidation}
	ErrEventEndBeforeStart  = &DomainError{Code: "event.end_before_start", Message: "end date must be after start date", Kind: KindValidation}
)

type Event struct {
	entity
	ID          uuid.UUID
	CategoryID  uuid.UUID
	Title       string
	Description *string
	Location    *string
	StartsAtUtc time.Time
	EndsAtUtc   *time.Time
	Status      EventStatus
}

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
		ID:          uuid.New(),
		CategoryID:  categoryID,
		Title:       title,
		Description: description,
		Location:    location,
		StartsAtUtc: startsAtUtc,
		EndsAtUtc:   endsAtUtc,
		Status:      StatusDraft,
	}
	e.raise(EventCreatedDomainEvent{EventID: e.ID})
	return e, nil
}

func (e *Event) Publish() error {
	if e.Status != StatusDraft {
		return ErrEventNotDraft
	}
	e.Status = StatusPublished
	e.raise(EventPublishedDomainEvent{EventID: e.ID})
	return nil
}

func (e *Event) Cancel(now time.Time) error {
	if e.Status == StatusCancelled {
		return ErrEventAlreadyCancelled
	}
	if now.After(e.StartsAtUtc) {
		return ErrEventAlreadyStarted
	}
	e.Status = StatusCancelled
	e.raise(EventCancelledDomainEvent{EventID: e.ID})
	return nil
}

func (e *Event) Reschedule(startsAt time.Time, endsAt *time.Time) error {
	sameStart := e.StartsAtUtc.Equal(startsAt)
	sameEnd := (e.EndsAtUtc == nil && endsAt == nil) ||
		(e.EndsAtUtc != nil && endsAt != nil && e.EndsAtUtc.Equal(*endsAt))
	if sameStart && sameEnd {
		return nil
	}
	e.StartsAtUtc = startsAt
	e.EndsAtUtc = endsAt
	e.raise(EventRescheduledDomainEvent{EventID: e.ID, StartsAtUtc: startsAt, EndsAtUtc: endsAt})
	return nil
}
