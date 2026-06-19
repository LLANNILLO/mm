package domain

import "github.com/google/uuid"

type TicketType struct {
	entity
	ID                uuid.UUID
	EventID           uuid.UUID
	Name              string
	Price             int64
	Currency          string
	Quantity          int64
	AvailableQuantity int64
}

func NewTicketType(
	id uuid.UUID,
	eventID uuid.UUID,
	name string,
	price int64,
	currency string,
	quantity int64,
) *TicketType {
	return &TicketType{
		ID:                id,
		EventID:           eventID,
		Name:              name,
		Price:             price,
		Currency:          currency,
		Quantity:          quantity,
		AvailableQuantity: quantity,
	}
}

func (t *TicketType) UpdatePrice(price int64) {
	t.Price = price
}

func (t *TicketType) UpdateQuantity(qty int64) error {
	if qty > t.AvailableQuantity {
		return ErrTicketTypeInsufficientQuantity
	}
	t.AvailableQuantity -= qty
	if t.AvailableQuantity == 0 {
		t.raise(TicketTypeSoldOutDomainEvent{TicketTypeID: t.ID})
	}
	return nil
}
