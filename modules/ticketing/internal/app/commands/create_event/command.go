package createevent

import (
	"time"

	"github.com/google/uuid"
)

type Command struct {
	EventID     uuid.UUID
	Title       string
	Description *string
	Location    *string
	StartsAtUtc time.Time
	EndsAtUtc   *time.Time
	TicketTypes []TicketTypeData
}

type TicketTypeData struct {
	ID       uuid.UUID
	EventID  uuid.UUID
	Name     string
	Price    int64
	Currency string
	Quantity int64
}
