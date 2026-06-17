package eventhandlers

import (
	"context"
	"fmt"

	getuser "github.com/llannillo/mm/modules/users/internal/app/queries/get_user"
	"github.com/llannillo/mm/modules/users/internal/domain"
	ticketingapi "github.com/llannillo/mm/modules/ticketing/api"
)

type UserRegisteredHandler struct {
	getUserQuery *getuser.Handler
	ticketingAPI ticketingapi.TicketingAPI
}

func NewUserRegisteredHandler(getUserQuery *getuser.Handler, ticketingAPI ticketingapi.TicketingAPI) *UserRegisteredHandler {
	return &UserRegisteredHandler{
		getUserQuery: getUserQuery,
		ticketingAPI: ticketingAPI,
	}
}

func (h *UserRegisteredHandler) Handle(ctx context.Context, e domain.UserRegisteredDomainEvent) error {
	user, err := h.getUserQuery.Handle(ctx, getuser.Query{UserID: e.UserID})
	if err != nil {
		return fmt.Errorf("get user for customer creation: %w", err)
	}
	return h.ticketingAPI.CreateCustomer(ctx, user.ID, user.Email, user.FirstName, user.LastName)
}
