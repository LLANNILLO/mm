package eventhandlers

import (
	"context"

	"github.com/llannillo/mm/modules/ticketing/internal/domain"
	"github.com/llannillo/mm/modules/ticketing/internal/ports/outbound"
)

// ArchiveTicketsHandler archives every ticket for a cancelled event.
type ArchiveTicketsHandler struct {
	eventRepo outbound.EventRepository
}

func NewArchiveTicketsHandler(eventRepo outbound.EventRepository) *ArchiveTicketsHandler {
	return &ArchiveTicketsHandler{eventRepo: eventRepo}
}

func (h *ArchiveTicketsHandler) Handle(ctx context.Context, e domain.EventCancelledDomainEvent) error {
	return h.eventRepo.ArchiveTickets(ctx, e.EventID)
}

// RefundPaymentsHandler refunds every payment for a cancelled event.
type RefundPaymentsHandler struct {
	eventRepo outbound.EventRepository
}

func NewRefundPaymentsHandler(eventRepo outbound.EventRepository) *RefundPaymentsHandler {
	return &RefundPaymentsHandler{eventRepo: eventRepo}
}

func (h *RefundPaymentsHandler) Handle(ctx context.Context, e domain.EventCancelledDomainEvent) error {
	return h.eventRepo.RefundPayments(ctx, e.EventID)
}
