package createtickettype

import (
	"github.com/google/uuid"
	"github.com/llannillo/mm/internal/shared/validation"
)

type Command struct {
	EventID  uuid.UUID
	Name     string
	Price    int64
	Currency string
	Quantity int64
}

func (c Command) Validate() error {
	return validation.New().
		Custom("event_id", c.EventID == uuid.Nil, "event_id is required").
		Required("name", c.Name).
		Custom("price", c.Price <= 0, "price must be greater than 0").
		Required("currency", c.Currency).
		Custom("quantity", c.Quantity <= 0, "quantity must be greater than 0").
		Err()
}
