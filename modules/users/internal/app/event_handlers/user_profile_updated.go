package eventhandlers

import (
	"context"

	"github.com/llannillo/mm/internal/shared/eventbus"
	usersintegrationevents "github.com/llannillo/mm/modules/users/api/integrationevents"
	"github.com/llannillo/mm/modules/users/internal/domain"
)

type UserProfileUpdatedHandler struct {
	eventBus eventbus.EventBus
}

func NewUserProfileUpdatedHandler(eventBus eventbus.EventBus) *UserProfileUpdatedHandler {
	return &UserProfileUpdatedHandler{eventBus: eventBus}
}

func (h *UserProfileUpdatedHandler) Handle(ctx context.Context, e domain.UserProfileUpdatedDomainEvent) error {
	return h.eventBus.Publish(ctx, usersintegrationevents.UserProfileUpdatedIntegrationEvent{
		UserID:    e.UserID,
		FirstName: e.FirstName,
		LastName:  e.LastName,
	})
}
