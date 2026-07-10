package registeruser

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/users/internal/domain"
	"github.com/llannillo/mm/modules/users/internal/ports/outbound"
)

type Handler struct {
	identity outbound.IdentityProvider
	repo     outbound.UserRepository
}

func NewHandler(identity outbound.IdentityProvider, repo outbound.UserRepository) *Handler {
	return &Handler{identity: identity, repo: repo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (uuid.UUID, error) {
	if err := cmd.Validate(); err != nil {
		return uuid.Nil, err
	}
	identityID, err := h.identity.RegisterUser(ctx, cmd.Email, cmd.Password, cmd.FirstName, cmd.LastName)
	if err != nil {
		return uuid.Nil, fmt.Errorf("register user in identity provider: %w", err)
	}
	user := domain.NewUser(cmd.Email, cmd.FirstName, cmd.LastName, identityID)
	if err := h.repo.Insert(ctx, user); err != nil {
		return uuid.Nil, fmt.Errorf("register user: %w", err)
	}
	return user.ID(), nil
}
