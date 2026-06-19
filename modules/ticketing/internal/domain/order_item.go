package domain

import "github.com/google/uuid"

type OrderItem struct {
	ID           uuid.UUID
	OrderID      uuid.UUID
	TicketTypeID uuid.UUID
	Quantity     int64
	UnitPrice    int64
	Price        int64
	Currency     string
}
