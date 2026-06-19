package updateuser

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/users/internal/ports/outbound"
)

type Handler struct {
	repo outbound.UserRepository
}

func NewHandler(repo outbound.UserRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	if err := cmd.Validate(); err != nil {
		return err
	}
	user, err := h.repo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	user.UpdateProfile(cmd.FirstName, cmd.LastName)
	if err := h.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}
