package eventhandlers

import (
	"context"

	"github.com/llannillo/mm/modules/users/internal/domain"
)

func HandleUserRegistered(_ context.Context, _ domain.UserRegisteredDomainEvent) error {
	return nil
}
