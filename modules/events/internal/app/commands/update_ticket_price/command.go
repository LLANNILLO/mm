package updateticketprice

import (
	"github.com/google/uuid"
	"github.com/llannillo/mm/internal/shared/validation"
)

type Command struct {
	TicketTypeID uuid.UUID
	Price        int64
}

func (c Command) Validate() error {
	return validation.New().
		Custom("ticket_type_id", c.TicketTypeID == uuid.Nil, "ticket_type_id is required").
		Custom("price", c.Price <= 0, "price must be greater than 0").
		Err()
}
