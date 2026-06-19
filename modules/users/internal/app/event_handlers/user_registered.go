package eventhandlers

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/internal/shared/eventbus"
	getuser "github.com/llannillo/mm/modules/users/internal/app/queries/get_user"
	"github.com/llannillo/mm/modules/users/internal/domain"
	usersapi "github.com/llannillo/mm/modules/users/api"
)

type UserRegisteredHandler struct {
	getUserQuery *getuser.Handler
	eventBus     eventbus.EventBus
}

func NewUserRegisteredHandler(getUserQuery *getuser.Handler, eventBus eventbus.EventBus) *UserRegisteredHandler {
	return &UserRegisteredHandler{
		getUserQuery: getUserQuery,
		eventBus:     eventBus,
	}
}

func (h *UserRegisteredHandler) Handle(ctx context.Context, e domain.UserRegisteredDomainEvent) error {
	user, err := h.getUserQuery.Handle(ctx, getuser.Query{UserID: e.UserID})
	if err != nil {
		return fmt.Errorf("get user for customer creation: %w", err)
	}
	return h.eventBus.Publish(ctx, usersapi.UserRegisteredIntegrationEvent{
		UserID:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	})
}
