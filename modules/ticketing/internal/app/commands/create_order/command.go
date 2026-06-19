package createorder

import "github.com/google/uuid"

type Command struct {
	CustomerID  uuid.UUID
	TicketTypes []OrderItem
}

type OrderItem struct {
	TicketTypeID uuid.UUID
	Quantity     int64
}
