package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending OrderStatus = "pending"
	OrderStatusPaid    OrderStatus = "paid"
)

type Order struct {
	entity
	id            uuid.UUID
	customerID    uuid.UUID
	status        OrderStatus
	totalPrice    int64
	currency      string
	ticketsIssued bool
	createdAtUtc  time.Time
	items         []OrderItem
}

func (o *Order) ID() uuid.UUID           { return o.id }
func (o *Order) CustomerID() uuid.UUID   { return o.customerID }
func (o *Order) Status() OrderStatus     { return o.status }
func (o *Order) TotalPrice() int64       { return o.totalPrice }
func (o *Order) Currency() string        { return o.currency }
func (o *Order) TicketsIssued() bool     { return o.ticketsIssued }
func (o *Order) CreatedAtUtc() time.Time { return o.createdAtUtc }
func (o *Order) Items() []OrderItem      { return o.items }

func NewOrder(customerID uuid.UUID) *Order {
	o := &Order{
		id:            uuid.New(),
		customerID:    customerID,
		status:        OrderStatusPending,
		totalPrice:    0,
		currency:      "",
		ticketsIssued: false,
		createdAtUtc:  time.Now().UTC(),
	}
	o.raise(OrderCreatedDomainEvent{OrderID: o.id})
	return o
}

// RehydrateOrder reconstructs an Order from persisted state without raising
// domain events. Only the OrderRepository may call this.
func RehydrateOrder(
	id, customerID uuid.UUID,
	status OrderStatus,
	totalPrice int64,
	currency string,
	ticketsIssued bool,
	createdAtUtc time.Time,
	items []OrderItem,
) *Order {
	return &Order{
		id:            id,
		customerID:    customerID,
		status:        status,
		totalPrice:    totalPrice,
		currency:      currency,
		ticketsIssued: ticketsIssued,
		createdAtUtc:  createdAtUtc,
		items:         items,
	}
}

func (o *Order) AddItem(ticketType *TicketType, quantity int64) error {
	item := OrderItem{
		id:           uuid.New(),
		orderID:      o.id,
		ticketTypeID: ticketType.ID(),
		quantity:     quantity,
		unitPrice:    ticketType.Price(),
		price:        ticketType.Price() * quantity,
		currency:     ticketType.Currency(),
	}
	o.items = append(o.items, item)
	o.totalPrice += item.price
	o.currency = item.currency
	return nil
}

func (o *Order) IssueTickets() error {
	if o.ticketsIssued {
		return ErrOrderTicketsAlreadyIssued
	}
	o.ticketsIssued = true
	o.raise(OrderTicketsIssuedDomainEvent{OrderID: o.id})
	return nil
}
