package eventhandlers

import (
	"context"

	"github.com/llannillo/mm/modules/users/internal/domain"
)

func HandleUserProfileUpdated(_ context.Context, _ domain.UserProfileUpdatedDomainEvent) error {
	return nil
}
