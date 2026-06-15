package inbound

import (
	"context"

	"github.com/google/uuid"
	createtickettype "github.com/llannillo/mm/modules/events/internal/app/commands/create_ticket_type"
	updateticketprice "github.com/llannillo/mm/modules/events/internal/app/commands/update_ticket_price"
	gettickettype "github.com/llannillo/mm/modules/events/internal/app/queries/get_ticket_type"
	listtickettype "github.com/llannillo/mm/modules/events/internal/app/queries/list_ticket_types"
)

type TicketService interface {
	CreateTicketType(ctx context.Context, cmd createtickettype.Command) (uuid.UUID, error)
	UpdateTicketPrice(ctx context.Context, cmd updateticketprice.Command) error
	GetTicketType(ctx context.Context, q gettickettype.Query) (*gettickettype.Response, error)
	ListTicketTypes(ctx context.Context, q listtickettype.Query) ([]listtickettype.TicketTypeItem, error)
}
