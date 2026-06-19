package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
}
