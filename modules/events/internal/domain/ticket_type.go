package domain

import "github.com/google/uuid"

var (
	ErrTicketTypeNotFound    = &DomainError{Code: "ticket_type.not_found", Message: "ticket type not found", Kind: KindNotFound}
	ErrTicketTypeInvalidPrice = &DomainError{Code: "ticket_type.invalid_price", Message: "price must be zero or positive", Kind: KindValidation}
	ErrTicketTypeNameEmpty   = &DomainError{Code: "ticket_type.name_empty", Message: "ticket type name cannot be empty", Kind: KindValidation}
	ErrEventHasNoTickets     = &DomainError{Code: "event.no_tickets", Message: "event has no ticket types", Kind: KindConflict}
)

type TicketType struct {
	entity
	ID       uuid.UUID
	EventID  uuid.UUID
	Name     string
	Price    int64
	Currency string
	Quantity int64
}

func NewTicketType(eventID uuid.UUID, name string, price int64, currency string, quantity int64) (*TicketType, error) {
	if name == "" {
		return nil, ErrTicketTypeNameEmpty
	}
	if price < 0 {
		return nil, ErrTicketTypeInvalidPrice
	}
	tt := &TicketType{
		ID:       uuid.New(),
		EventID:  eventID,
		Name:     name,
		Price:    price,
		Currency: currency,
		Quantity: quantity,
	}
	tt.raise(TicketTypeCreatedDomainEvent{TicketTypeID: tt.ID})
	return tt, nil
}

func (tt *TicketType) UpdatePrice(price int64) error {
	if price < 0 {
		return ErrTicketTypeInvalidPrice
	}
	if tt.Price == price {
		return nil
	}
	tt.Price = price
	tt.raise(TicketTypePriceChangedDomainEvent{TicketTypeID: tt.ID, Price: price})
	return nil
}
