package updateticketprice

import "github.com/google/uuid"

type Command struct {
	TicketTypeID uuid.UUID
	Price        int64
}
