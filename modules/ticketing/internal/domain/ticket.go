package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	entity
	ID           uuid.UUID
	CustomerID   uuid.UUID
	OrderID      uuid.UUID
	EventID      uuid.UUID
	TicketTypeID uuid.UUID
	Code         string
	CreatedAtUtc time.Time
	Archived     bool
}

func NewTicket(order *Order, ticketType *TicketType) *Ticket {
	t := &Ticket{
		ID:           uuid.New(),
		CustomerID:   order.CustomerID,
		OrderID:      order.ID,
		EventID:      ticketType.EventID,
		TicketTypeID: ticketType.ID,
		Code:         fmt.Sprintf("tc_%s", uuid.New().String()),
		CreatedAtUtc: time.Now().UTC(),
		Archived:     false,
	}
	t.raise(TicketCreatedDomainEvent{TicketID: t.ID})
	return t
}

func (t *Ticket) Archive() {
	if t.Archived {
		return
	}
	t.Archived = true
	t.raise(TicketArchivedDomainEvent{TicketID: t.ID, Code: t.Code})
}
