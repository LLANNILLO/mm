package domain

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface {
	domainEvent()
}

// Event domain events

type EventCreatedDomainEvent struct{ EventID uuid.UUID }
type EventPublishedDomainEvent struct{ EventID uuid.UUID }
type EventCancelledDomainEvent struct{ EventID uuid.UUID }
type EventRescheduledDomainEvent struct {
	EventID     uuid.UUID
	StartsAtUtc time.Time
	EndsAtUtc   *time.Time
}

func (EventCreatedDomainEvent) domainEvent()     {}
func (EventPublishedDomainEvent) domainEvent()   {}
func (EventCancelledDomainEvent) domainEvent()   {}
func (EventRescheduledDomainEvent) domainEvent() {}

// Category domain events

type CategoryCreatedDomainEvent struct{ CategoryID uuid.UUID }
type CategoryArchivedDomainEvent struct{ CategoryID uuid.UUID }
type CategoryNameChangedDomainEvent struct {
	CategoryID uuid.UUID
	Name       string
}

func (CategoryCreatedDomainEvent) domainEvent()     {}
func (CategoryArchivedDomainEvent) domainEvent()    {}
func (CategoryNameChangedDomainEvent) domainEvent() {}

// TicketType domain events

type TicketTypeCreatedDomainEvent struct{ TicketTypeID uuid.UUID }
type TicketTypePriceChangedDomainEvent struct {
	TicketTypeID uuid.UUID
	Price        int64
}

func (TicketTypeCreatedDomainEvent) domainEvent()      {}
func (TicketTypePriceChangedDomainEvent) domainEvent() {}
