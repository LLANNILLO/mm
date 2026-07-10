package consumers

import (
	"context"

	createcustomer "github.com/llannillo/mm/modules/ticketing/internal/app/commands/create_customer"
	usersintegrationevents "github.com/llannillo/mm/modules/users/api/integrationevents"
)

type UserRegisteredConsumer struct {
	createCustomer *createcustomer.Handler
}

func NewUserRegisteredConsumer(h *createcustomer.Handler) *UserRegisteredConsumer {
	return &UserRegisteredConsumer{createCustomer: h}
}

func (c *UserRegisteredConsumer) Handle(ctx context.Context, e usersintegrationevents.UserRegisteredIntegrationEvent) error {
	return c.createCustomer.Handle(ctx, createcustomer.Command{
		ID:        e.UserID,
		Email:     e.Email,
		FirstName: e.FirstName,
		LastName:  e.LastName,
	})
}
