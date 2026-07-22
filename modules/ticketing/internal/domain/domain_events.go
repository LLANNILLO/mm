package domain

import (
	"time"

	"github.com/google/uuid"
)

// Event domain events

type EventCancelledDomainEvent struct{ EventID uuid.UUID }
type EventRescheduledDomainEvent struct {
	EventID     uuid.UUID
	StartsAtUtc time.Time
	EndsAtUtc   *time.Time
}
type EventPaymentsRefundedDomainEvent struct{ EventID uuid.UUID }
type EventTicketsArchivedDomainEvent struct{ EventID uuid.UUID }

func (EventCancelledDomainEvent) IsDomainEvent()        {}
func (EventRescheduledDomainEvent) IsDomainEvent()      {}
func (EventPaymentsRefundedDomainEvent) IsDomainEvent() {}
func (EventTicketsArchivedDomainEvent) IsDomainEvent()  {}

// TicketType domain events

type TicketTypeSoldOutDomainEvent struct{ TicketTypeID uuid.UUID }

func (TicketTypeSoldOutDomainEvent) IsDomainEvent() {}

// Order domain events

type OrderCreatedDomainEvent struct{ OrderID uuid.UUID }
type OrderTicketsIssuedDomainEvent struct{ OrderID uuid.UUID }

func (OrderCreatedDomainEvent) IsDomainEvent()       {}
func (OrderTicketsIssuedDomainEvent) IsDomainEvent() {}

// Ticket domain events

type TicketCreatedDomainEvent struct {
	TicketID uuid.UUID
	EventID  uuid.UUID
}
type TicketArchivedDomainEvent struct {
	TicketID uuid.UUID
	Code     string
}
type TicketCheckedInDomainEvent struct {
	TicketID uuid.UUID
	EventID  uuid.UUID
}
type TicketCheckInDuplicateDomainEvent struct {
	TicketID uuid.UUID
	EventID  uuid.UUID
	Code     string
}
type TicketCheckInInvalidDomainEvent struct {
	TicketID uuid.UUID
	EventID  uuid.UUID
	Code     string
}

func (TicketCreatedDomainEvent) IsDomainEvent()          {}
func (TicketArchivedDomainEvent) IsDomainEvent()         {}
func (TicketCheckedInDomainEvent) IsDomainEvent()        {}
func (TicketCheckInDuplicateDomainEvent) IsDomainEvent() {}
func (TicketCheckInInvalidDomainEvent) IsDomainEvent()   {}

// Payment domain events

type PaymentCreatedDomainEvent struct{ PaymentID uuid.UUID }
type PaymentRefundedDomainEvent struct {
	PaymentID     uuid.UUID
	TransactionID uuid.UUID
	Amount        int64
}
type PaymentPartiallyRefundedDomainEvent struct {
	PaymentID     uuid.UUID
	TransactionID uuid.UUID
	Amount        int64
}

func (PaymentCreatedDomainEvent) IsDomainEvent()           {}
func (PaymentRefundedDomainEvent) IsDomainEvent()          {}
func (PaymentPartiallyRefundedDomainEvent) IsDomainEvent() {}
