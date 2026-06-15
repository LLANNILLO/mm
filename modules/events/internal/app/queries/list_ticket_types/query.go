package listtickettype

import "github.com/google/uuid"

type Query struct {
	EventID uuid.UUID
}

type TicketTypeItem struct {
	ID       uuid.UUID `json:"id"`
	EventID  uuid.UUID `json:"event_id"`
	Name     string    `json:"name"`
	Price    int64     `json:"price"`
	Currency string    `json:"currency"`
	Quantity int64     `json:"quantity"`
}
