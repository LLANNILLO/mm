package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	entity
	id           uuid.UUID
	customerID   uuid.UUID
	orderID      uuid.UUID
	eventID      uuid.UUID
	ticketTypeID uuid.UUID
	code         string
	createdAtUtc time.Time
	archived     bool
	usedAtUtc    *time.Time
}

func (t *Ticket) ID() uuid.UUID           { return t.id }
func (t *Ticket) CustomerID() uuid.UUID   { return t.customerID }
func (t *Ticket) OrderID() uuid.UUID      { return t.orderID }
func (t *Ticket) EventID() uuid.UUID      { return t.eventID }
func (t *Ticket) TicketTypeID() uuid.UUID { return t.ticketTypeID }
func (t *Ticket) Code() string            { return t.code }
func (t *Ticket) CreatedAtUtc() time.Time { return t.createdAtUtc }
func (t *Ticket) Archived() bool          { return t.archived }
func (t *Ticket) UsedAtUtc() *time.Time   { return t.usedAtUtc }

func NewTicket(order *Order, ticketType *TicketType) *Ticket {
	t := &Ticket{
		id:           uuid.New(),
		customerID:   order.CustomerID(),
		orderID:      order.ID(),
		eventID:      ticketType.EventID(),
		ticketTypeID: ticketType.ID(),
		code:         fmt.Sprintf("tc_%s", uuid.New().String()),
		createdAtUtc: time.Now().UTC(),
		archived:     false,
	}
	t.raise(TicketCreatedDomainEvent{TicketID: t.id, EventID: t.eventID})
	return t
}

// RehydrateTicket reconstructs a Ticket from persisted state without raising
// domain events. Only the TicketRepository may call this.
func RehydrateTicket(
	id, customerID, orderID, eventID, ticketTypeID uuid.UUID,
	code string,
	createdAtUtc time.Time,
	archived bool,
	usedAtUtc *time.Time,
) *Ticket {
	return &Ticket{
		id:           id,
		customerID:   customerID,
		orderID:      orderID,
		eventID:      eventID,
		ticketTypeID: ticketTypeID,
		code:         code,
		createdAtUtc: createdAtUtc,
		archived:     archived,
		usedAtUtc:    usedAtUtc,
	}
}

func (t *Ticket) Archive() {
	if t.archived {
		return
	}
	t.archived = true
	t.raise(TicketArchivedDomainEvent{TicketID: t.id, Code: t.code})
}

// CheckIn marks the ticket as used by customerID. Unlike Archive, it raises a
// domain event on every outcome — including the two failure paths — because
// the materialized event_statistics view counts invalid/duplicate attempts,
// not just successes. Callers must persist the ticket even when CheckIn
// returns an error, or the event never reaches the outbox.
func (t *Ticket) CheckIn(customerID uuid.UUID) error {
	if t.customerID != customerID {
		t.raise(TicketCheckInInvalidDomainEvent{TicketID: t.id, EventID: t.eventID, Code: t.code})
		return ErrTicketCheckInInvalid
	}
	if t.usedAtUtc != nil {
		t.raise(TicketCheckInDuplicateDomainEvent{TicketID: t.id, EventID: t.eventID, Code: t.code})
		return ErrTicketAlreadyCheckedIn
	}
	now := time.Now().UTC()
	t.usedAtUtc = &now
	t.raise(TicketCheckedInDomainEvent{TicketID: t.id, EventID: t.eventID})
	return nil
}
