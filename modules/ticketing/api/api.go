package ticketingapi

import (
	"context"

	"github.com/google/uuid"
)

type TicketingAPI interface {
	CreateCustomer(ctx context.Context, id uuid.UUID, email, firstName, lastName string) error
	UpdateCustomer(ctx context.Context, id uuid.UUID, firstName, lastName string) error
}
