package gettickettype

import "github.com/google/uuid"

type Query struct {
	ID uuid.UUID
}

type Response struct {
	ID       uuid.UUID `json:"id"`
	EventID  uuid.UUID `json:"event_id"`
	Name     string    `json:"name"`
	Price    int64     `json:"price"`
	Currency string    `json:"currency"`
	Quantity int64     `json:"quantity"`
}
