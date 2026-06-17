package eventhandlers

import (
	"context"

	"github.com/llannillo/mm/modules/users/internal/domain"
	ticketingapi "github.com/llannillo/mm/modules/ticketing/api"
)

type UserProfileUpdatedHandler struct {
	ticketingAPI ticketingapi.TicketingAPI
}

func NewUserProfileUpdatedHandler(ticketingAPI ticketingapi.TicketingAPI) *UserProfileUpdatedHandler {
	return &UserProfileUpdatedHandler{ticketingAPI: ticketingAPI}
}

func (h *UserProfileUpdatedHandler) Handle(ctx context.Context, e domain.UserProfileUpdatedDomainEvent) error {
	return h.ticketingAPI.UpdateCustomer(ctx, e.UserID, e.FirstName, e.LastName)
}
