package additemtocart

import (
	"github.com/google/uuid"
	"github.com/llannillo/mm/internal/shared/validation"
)

type Command struct {
	CustomerID   uuid.UUID
	TicketTypeID uuid.UUID
	Quantity     int64
}

func (c Command) Validate() error {
	return validation.New().
		Custom("customer_id", c.CustomerID == uuid.Nil, "customer_id is required").
		Custom("ticket_type_id", c.TicketTypeID == uuid.Nil, "ticket_type_id is required").
		Custom("quantity", c.Quantity <= 0, "quantity must be greater than 0").
		Err()
}
