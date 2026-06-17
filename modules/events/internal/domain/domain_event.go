package domain

import (
	"time"

	"github.com/google/uuid"
)

// Event domain events

type EventCreatedDomainEvent struct{ EventID uuid.UUID }
type EventPublishedDomainEvent struct{ EventID uuid.UUID }
type EventCancelledDomainEvent struct{ EventID uuid.UUID }
type EventRescheduledDomainEvent struct {
	EventID     uuid.UUID
	StartsAtUtc time.Time
	EndsAtUtc   *time.Time
}

func (EventCreatedDomainEvent) IsDomainEvent()     {}
func (EventPublishedDomainEvent) IsDomainEvent()   {}
func (EventCancelledDomainEvent) IsDomainEvent()   {}
func (EventRescheduledDomainEvent) IsDomainEvent() {}

// Category domain events

type CategoryCreatedDomainEvent struct{ CategoryID uuid.UUID }
type CategoryArchivedDomainEvent struct{ CategoryID uuid.UUID }
type CategoryNameChangedDomainEvent struct {
	CategoryID uuid.UUID
	Name       string
}

func (CategoryCreatedDomainEvent) IsDomainEvent()     {}
func (CategoryArchivedDomainEvent) IsDomainEvent()    {}
func (CategoryNameChangedDomainEvent) IsDomainEvent() {}

// TicketType domain events

type TicketTypeCreatedDomainEvent struct{ TicketTypeID uuid.UUID }
type TicketTypePriceChangedDomainEvent struct {
	TicketTypeID uuid.UUID
	Price        int64
}

func (TicketTypeCreatedDomainEvent) IsDomainEvent()      {}
func (TicketTypePriceChangedDomainEvent) IsDomainEvent() {}
