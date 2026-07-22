package checkinticket

import "github.com/google/uuid"

type Command struct {
	TicketID   uuid.UUID
	CustomerID uuid.UUID
}
