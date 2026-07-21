package inbound

import (
	"context"

	"github.com/google/uuid"
	createorder "github.com/llannillo/mm/modules/ticketing/internal/app/commands/create_order"
)

type OrderService interface {
	CreateOrder(ctx context.Context, cmd createorder.Command) (uuid.UUID, error)
}
