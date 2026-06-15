package getevent

import (
	"time"

	"github.com/google/uuid"
)

type TicketTypeItem struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Price    int64     `json:"price"`
	Currency string    `json:"currency"`
	Quantity int64     `json:"quantity"`
}

type Response struct {
	ID          uuid.UUID        `json:"id"`
	CategoryID  uuid.UUID        `json:"category_id"`
	Title       string           `json:"title"`
	Description *string          `json:"description"`
	Location    *string          `json:"location"`
	StartsAtUtc time.Time        `json:"starts_at_utc"`
	EndsAtUtc   *time.Time       `json:"ends_at_utc"`
	TicketTypes []TicketTypeItem `json:"ticket_types"`
}
