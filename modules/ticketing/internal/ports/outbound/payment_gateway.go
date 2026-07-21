package outbound

import (
	"context"

	"github.com/google/uuid"
)

// PaymentGateway is the boundary to whatever external payment provider
// processes refunds. transactionID identifies the original charge.
type PaymentGateway interface {
	Refund(ctx context.Context, transactionID uuid.UUID, amount int64) error
}
