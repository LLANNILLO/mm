package registeruser

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/users/internal/domain"
	"github.com/llannillo/mm/modules/users/internal/ports/outbound"
)

type Handler struct {
	repo outbound.UserRepository
}

func NewHandler(repo outbound.UserRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (uuid.UUID, error) {
	if err := cmd.Validate(); err != nil {
		return uuid.Nil, err
	}
	user := domain.NewUser(cmd.Email, cmd.FirstName, cmd.LastName)
	if err := h.repo.Insert(ctx, user); err != nil {
		return uuid.Nil, fmt.Errorf("register user: %w", err)
	}
	return user.ID, nil
}
