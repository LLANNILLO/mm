package domain

import "github.com/google/uuid"

var (
	ErrTicketTypeNotFound     = &DomainError{Code: "ticket_type.not_found", Message: "ticket type not found", Kind: KindNotFound}
	ErrTicketTypeInvalidPrice = &DomainError{Code: "ticket_type.invalid_price", Message: "price must be zero or positive", Kind: KindValidation}
	ErrTicketTypeNameEmpty    = &DomainError{Code: "ticket_type.name_empty", Message: "ticket type name cannot be empty", Kind: KindValidation}
	ErrEventHasNoTickets      = &DomainError{Code: "event.no_tickets", Message: "event has no ticket types", Kind: KindConflict}
)

type TicketType struct {
	entity
	id       uuid.UUID
	eventID  uuid.UUID
	name     string
	price    int64
	currency string
	quantity int64
}

func (tt *TicketType) ID() uuid.UUID      { return tt.id }
func (tt *TicketType) EventID() uuid.UUID { return tt.eventID }
func (tt *TicketType) Name() string       { return tt.name }
func (tt *TicketType) Price() int64       { return tt.price }
func (tt *TicketType) Currency() string   { return tt.currency }
func (tt *TicketType) Quantity() int64    { return tt.quantity }

func NewTicketType(eventID uuid.UUID, name string, price int64, currency string, quantity int64) (*TicketType, error) {
	if name == "" {
		return nil, ErrTicketTypeNameEmpty
	}
	if price < 0 {
		return nil, ErrTicketTypeInvalidPrice
	}
	tt := &TicketType{
		id:       uuid.New(),
		eventID:  eventID,
		name:     name,
		price:    price,
		currency: currency,
		quantity: quantity,
	}
	tt.raise(TicketTypeCreatedDomainEvent{TicketTypeID: tt.id})
	return tt, nil
}

// RehydrateTicketType reconstructs a TicketType from persisted state without
// re-running creation invariants or raising domain events. Only repositories
// may call this.
func RehydrateTicketType(id, eventID uuid.UUID, name string, price int64, currency string, quantity int64) *TicketType {
	return &TicketType{id: id, eventID: eventID, name: name, price: price, currency: currency, quantity: quantity}
}

func (tt *TicketType) UpdatePrice(price int64) error {
	if price < 0 {
		return ErrTicketTypeInvalidPrice
	}
	if tt.price == price {
		return nil
	}
	tt.price = price
	tt.raise(TicketTypePriceChangedDomainEvent{TicketTypeID: tt.id, Price: price})
	return nil
}
