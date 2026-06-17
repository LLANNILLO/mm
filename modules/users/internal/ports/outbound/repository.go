package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/users/internal/domain"
)

type UserRepository interface {
	Insert(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, u *domain.User) error
}
