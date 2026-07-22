package inbound

import (
	"context"

	checkinticket "github.com/llannillo/mm/modules/ticketing/internal/app/commands/check_in_ticket"
	geteventstatistics "github.com/llannillo/mm/modules/ticketing/internal/app/queries/get_event_statistics"
)

type TicketService interface {
	CheckIn(ctx context.Context, cmd checkinticket.Command) error
	GetEventStatistics(ctx context.Context, q geteventstatistics.Query) (*geteventstatistics.Response, error)
}
