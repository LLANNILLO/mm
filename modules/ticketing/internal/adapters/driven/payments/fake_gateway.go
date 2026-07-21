// Package payments holds payment-provider adapters. FakeGateway is a stand-in
// for a real processor (Stripe, etc.) — there is no external payment
// integration in this project yet.
package payments

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type FakeGateway struct {
	logger *slog.Logger
}

func NewFakeGateway(logger *slog.Logger) *FakeGateway {
	return &FakeGateway{logger: logger}
}

func (g *FakeGateway) Refund(ctx context.Context, transactionID uuid.UUID, amount int64) error {
	g.logger.InfoContext(ctx, "refunding payment via fake gateway", "transaction_id", transactionID, "amount", amount)
	return nil
}
