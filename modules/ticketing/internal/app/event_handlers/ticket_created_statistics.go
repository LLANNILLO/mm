package eventhandlers

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/ticketing/internal/domain"
	"github.com/llannillo/mm/modules/ticketing/internal/ports/outbound"
)

// TicketCreatedStatisticsHandler projects a sold ticket into the
// event_statistics materialized view, creating the row on first sale.
type TicketCreatedStatisticsHandler struct {
	statsRepo outbound.EventStatisticsRepository
}

func NewTicketCreatedStatisticsHandler(statsRepo outbound.EventStatisticsRepository) *TicketCreatedStatisticsHandler {
	return &TicketCreatedStatisticsHandler{statsRepo: statsRepo}
}

func (h *TicketCreatedStatisticsHandler) Handle(ctx context.Context, e domain.TicketCreatedDomainEvent) error {
	if err := h.statsRepo.EnsureRow(ctx, e.EventID); err != nil {
		return fmt.Errorf("ensure event statistics row: %w", err)
	}
	if err := h.statsRepo.IncrementTicketsSold(ctx, e.EventID); err != nil {
		return fmt.Errorf("increment tickets sold: %w", err)
	}
	return nil
}
