package consumers

import (
	"context"

	updatecustomer "github.com/llannillo/mm/modules/ticketing/internal/app/commands/update_customer"
	usersintegrationevents "github.com/llannillo/mm/modules/users/api/integrationevents"
)

type UserProfileUpdatedConsumer struct {
	updateCustomer *updatecustomer.Handler
}

func NewUserProfileUpdatedConsumer(h *updatecustomer.Handler) *UserProfileUpdatedConsumer {
	return &UserProfileUpdatedConsumer{updateCustomer: h}
}

func (c *UserProfileUpdatedConsumer) Handle(ctx context.Context, e usersintegrationevents.UserProfileUpdatedIntegrationEvent) error {
	return c.updateCustomer.Handle(ctx, updatecustomer.Command{
		ID:        e.UserID,
		FirstName: e.FirstName,
		LastName:  e.LastName,
	})
}
