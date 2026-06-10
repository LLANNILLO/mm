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
	Status      EventStatus
}

func NewEvent(title string, description, location *string, startsAtUtc time.Time, endsAtUtc *time.Time) *Event {
	e := &Event{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		Location:    location,
		StartsAtUtc: startsAtUtc,
		EndsAtUtc:   endsAtUtc,
		Status:      DraftStatus,
	}
	e.raise(EventCreatedDomainEvent{EventID: e.ID})
	return e
}
