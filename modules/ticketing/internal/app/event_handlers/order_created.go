package eventhandlers

import (
	"context"

	"github.com/llannillo/mm/modules/ticketing/internal/domain"
	"github.com/llannillo/mm/modules/ticketing/internal/ports/outbound"
)

// OrderCreatedHandler issues one Ticket per unit of quantity across the
// order's items once the order itself is durably recorded.
type OrderCreatedHandler struct {
	orderRepo outbound.OrderRepository
}

func NewOrderCreatedHandler(orderRepo outbound.OrderRepository) *OrderCreatedHandler {
	return &OrderCreatedHandler{orderRepo: orderRepo}
}

func (h *OrderCreatedHandler) Handle(ctx context.Context, e domain.OrderCreatedDomainEvent) error {
	return h.orderRepo.IssueTickets(ctx, e.OrderID)
}
