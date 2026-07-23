package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	store "github.com/llannillo/mm/modules/events/internal/adapters/driven/postgres/generated"
	gettickettype "github.com/llannillo/mm/modules/events/internal/app/queries/get_ticket_type"
	listtickettype "github.com/llannillo/mm/modules/events/internal/app/queries/list_ticket_types"
	"github.com/llannillo/mm/modules/events/internal/domain"
)

type TicketTypeReader struct {
	queries *store.Queries
}

func NewTicketTypeReader(q *store.Queries) *TicketTypeReader {
	return &TicketTypeReader{queries: q}
}

func (r *TicketTypeReader) GetTicketType(ctx context.Context, id uuid.UUID) (*gettickettype.Response, error) {
	row, err := r.queries.SelectTicketTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTicketTypeNotFound
		}
		return nil, fmt.Errorf("get ticket type: %w", err)
	}
	return &gettickettype.Response{
		ID:       row.ID,
		EventID:  row.EventID,
		Name:     row.Name,
		Price:    row.Price,
		Currency: row.Currency,
		Quantity: row.Quantity,
	}, nil
}

func (r *TicketTypeReader) ListTicketTypes(ctx context.Context, eventID uuid.UUID) ([]listtickettype.TicketTypeItem, error) {
	rows, err := r.queries.SelectTicketTypesByEventID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("list ticket types: %w", err)
	}
	items := make([]listtickettype.TicketTypeItem, len(rows))
	for i, row := range rows {
		items[i] = listtickettype.TicketTypeItem{
			ID:       row.ID,
			EventID:  row.EventID,
			Name:     row.Name,
			Price:    row.Price,
			Currency: row.Currency,
			Quantity: row.Quantity,
		}
	}
	return items, nil
}
