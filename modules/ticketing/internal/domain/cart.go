package domain

import "github.com/google/uuid"

type Cart struct {
	CustomerID uuid.UUID  `json:"customer_id"`
	Items      []CartItem `json:"items"`
}

type CartItem struct {
	TicketTypeID uuid.UUID `json:"ticket_type_id"`
	Quantity     int64     `json:"quantity"`
	Price        int64     `json:"price"`
	Currency     string    `json:"currency"`
}
