package eventhandlers

import (
	"context"

	"github.com/llannillo/mm/modules/ticketing/internal/domain"
	"github.com/llannillo/mm/modules/ticketing/internal/ports/outbound"
)

// TicketCheckedInStatisticsHandler counts a successful check-in in the
// event_statistics materialized view.
type TicketCheckedInStatisticsHandler struct {
	statsRepo outbound.EventStatisticsRepository
}

func NewTicketCheckedInStatisticsHandler(statsRepo outbound.EventStatisticsRepository) *TicketCheckedInStatisticsHandler {
	return &TicketCheckedInStatisticsHandler{statsRepo: statsRepo}
}

func (h *TicketCheckedInStatisticsHandler) Handle(ctx context.Context, e domain.TicketCheckedInDomainEvent) error {
	return h.statsRepo.IncrementAttendeesCheckedIn(ctx, e.EventID)
}

// TicketCheckInDuplicateStatisticsHandler records a duplicate check-in
// attempt in the event_statistics materialized view.
type TicketCheckInDuplicateStatisticsHandler struct {
	statsRepo outbound.EventStatisticsRepository
}

func NewTicketCheckInDuplicateStatisticsHandler(statsRepo outbound.EventStatisticsRepository) *TicketCheckInDuplicateStatisticsHandler {
	return &TicketCheckInDuplicateStatisticsHandler{statsRepo: statsRepo}
}

func (h *TicketCheckInDuplicateStatisticsHandler) Handle(ctx context.Context, e domain.TicketCheckInDuplicateDomainEvent) error {
	return h.statsRepo.AppendDuplicateCheckIn(ctx, e.EventID, e.Code)
}

// TicketCheckInInvalidStatisticsHandler records an invalid check-in attempt
// (wrong customer for the ticket) in the event_statistics materialized view.
type TicketCheckInInvalidStatisticsHandler struct {
	statsRepo outbound.EventStatisticsRepository
}

func NewTicketCheckInInvalidStatisticsHandler(statsRepo outbound.EventStatisticsRepository) *TicketCheckInInvalidStatisticsHandler {
	return &TicketCheckInInvalidStatisticsHandler{statsRepo: statsRepo}
}

func (h *TicketCheckInInvalidStatisticsHandler) Handle(ctx context.Context, e domain.TicketCheckInInvalidDomainEvent) error {
	return h.statsRepo.AppendInvalidCheckIn(ctx, e.EventID, e.Code)
}
