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
	ID            uuid.UUID
	CustomerID    uuid.UUID
	Status        OrderStatus
	TotalPrice    int64
	Currency      string
	TicketsIssued bool
	CreatedAtUtc  time.Time
	Items         []OrderItem
}

func NewOrder(customerID uuid.UUID) *Order {
	o := &Order{
		ID:           uuid.New(),
		CustomerID:   customerID,
		Status:       OrderStatusPending,
		TotalPrice:   0,
		Currency:     "",
		TicketsIssued: false,
		CreatedAtUtc: time.Now().UTC(),
	}
	o.raise(OrderCreatedDomainEvent{OrderID: o.ID})
	return o
}

func (o *Order) AddItem(ticketType *TicketType, quantity int64) error {
	item := OrderItem{
		ID:           uuid.New(),
		OrderID:      o.ID,
		TicketTypeID: ticketType.ID,
		Quantity:     quantity,
		UnitPrice:    ticketType.Price,
		Price:        ticketType.Price * quantity,
		Currency:     ticketType.Currency,
	}
	o.Items = append(o.Items, item)
	o.TotalPrice += item.Price
	o.Currency = item.Currency
	return nil
}

func (o *Order) IssueTickets() error {
	if o.TicketsIssued {
		return ErrOrderTicketsAlreadyIssued
	}
	o.TicketsIssued = true
	o.raise(OrderTicketsIssuedDomainEvent{OrderID: o.ID})
	return nil
}
