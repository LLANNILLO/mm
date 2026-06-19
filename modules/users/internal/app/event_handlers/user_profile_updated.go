package eventhandlers

import (
	"context"

	"github.com/llannillo/mm/internal/shared/eventbus"
	"github.com/llannillo/mm/modules/users/internal/domain"
	usersapi "github.com/llannillo/mm/modules/users/api"
)

type UserProfileUpdatedHandler struct {
	eventBus eventbus.EventBus
}

func NewUserProfileUpdatedHandler(eventBus eventbus.EventBus) *UserProfileUpdatedHandler {
	return &UserProfileUpdatedHandler{eventBus: eventBus}
}

func (h *UserProfileUpdatedHandler) Handle(ctx context.Context, e domain.UserProfileUpdatedDomainEvent) error {
	return h.eventBus.Publish(ctx, usersapi.UserProfileUpdatedIntegrationEvent{
		UserID:    e.UserID,
		FirstName: e.FirstName,
		LastName:  e.LastName,
	})
}
