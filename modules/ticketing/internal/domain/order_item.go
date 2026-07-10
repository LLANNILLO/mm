package domain

import "github.com/google/uuid"

// OrderItem is a child value object of the Order aggregate. It is only ever
// constructed by Order.AddItem or rehydrated by the OrderRepository — never
// built directly from outside this package.
type OrderItem struct {
	id           uuid.UUID
	orderID      uuid.UUID
	ticketTypeID uuid.UUID
	quantity     int64
	unitPrice    int64
	price        int64
	currency     string
}

func (i OrderItem) ID() uuid.UUID           { return i.id }
func (i OrderItem) OrderID() uuid.UUID      { return i.orderID }
func (i OrderItem) TicketTypeID() uuid.UUID { return i.ticketTypeID }
func (i OrderItem) Quantity() int64         { return i.quantity }
func (i OrderItem) UnitPrice() int64        { return i.unitPrice }
func (i OrderItem) Price() int64            { return i.price }
func (i OrderItem) Currency() string        { return i.currency }

// RehydrateOrderItem reconstructs an OrderItem from persisted state. Only the
// OrderRepository may call this.
func RehydrateOrderItem(id, orderID, ticketTypeID uuid.UUID, quantity, unitPrice, price int64, currency string) OrderItem {
	return OrderItem{
		id:           id,
		orderID:      orderID,
		ticketTypeID: ticketTypeID,
		quantity:     quantity,
		unitPrice:    unitPrice,
		price:        price,
		currency:     currency,
	}
}
