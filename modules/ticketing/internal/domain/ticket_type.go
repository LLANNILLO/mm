package domain

import "github.com/google/uuid"

type TicketType struct {
	entity
	id                uuid.UUID
	eventID           uuid.UUID
	name              string
	price             int64
	currency          string
	quantity          int64
	availableQuantity int64
}

func (t *TicketType) ID() uuid.UUID            { return t.id }
func (t *TicketType) EventID() uuid.UUID       { return t.eventID }
func (t *TicketType) Name() string             { return t.name }
func (t *TicketType) Price() int64             { return t.price }
func (t *TicketType) Currency() string         { return t.currency }
func (t *TicketType) Quantity() int64          { return t.quantity }
func (t *TicketType) AvailableQuantity() int64 { return t.availableQuantity }

func NewTicketType(
	id uuid.UUID,
	eventID uuid.UUID,
	name string,
	price int64,
	currency string,
	quantity int64,
) *TicketType {
	return &TicketType{
		id:                id,
		eventID:           eventID,
		name:              name,
		price:             price,
		currency:          currency,
		quantity:          quantity,
		availableQuantity: quantity,
	}
}

// RehydrateTicketType reconstructs a TicketType replica from persisted state
// without raising domain events. Only repositories may call this.
func RehydrateTicketType(
	id, eventID uuid.UUID,
	name string,
	price int64,
	currency string,
	quantity, availableQuantity int64,
) *TicketType {
	return &TicketType{
		id:                id,
		eventID:           eventID,
		name:              name,
		price:             price,
		currency:          currency,
		quantity:          quantity,
		availableQuantity: availableQuantity,
	}
}

func (t *TicketType) UpdatePrice(price int64) {
	t.price = price
}

func (t *TicketType) UpdateQuantity(qty int64) error {
	if qty > t.availableQuantity {
		return ErrTicketTypeInsufficientQuantity
	}
	t.availableQuantity -= qty
	if t.availableQuantity == 0 {
		t.raise(TicketTypeSoldOutDomainEvent{TicketTypeID: t.id})
	}
	return nil
}
