package eventhandlers

import (
	"context"

	"github.com/llannillo/mm/modules/ticketing/internal/domain"
	"github.com/llannillo/mm/modules/ticketing/internal/ports/outbound"
)

// PaymentRefundedHandler calls out to the payment gateway once a payment has
// been fully refunded in our own records.
type PaymentRefundedHandler struct {
	gateway outbound.PaymentGateway
}

func NewPaymentRefundedHandler(gateway outbound.PaymentGateway) *PaymentRefundedHandler {
	return &PaymentRefundedHandler{gateway: gateway}
}

func (h *PaymentRefundedHandler) Handle(ctx context.Context, e domain.PaymentRefundedDomainEvent) error {
	return h.gateway.Refund(ctx, e.TransactionID, e.Amount)
}

// PaymentPartiallyRefundedHandler calls out to the payment gateway once a
// payment has been partially refunded in our own records.
type PaymentPartiallyRefundedHandler struct {
	gateway outbound.PaymentGateway
}

func NewPaymentPartiallyRefundedHandler(gateway outbound.PaymentGateway) *PaymentPartiallyRefundedHandler {
	return &PaymentPartiallyRefundedHandler{gateway: gateway}
}

func (h *PaymentPartiallyRefundedHandler) Handle(ctx context.Context, e domain.PaymentPartiallyRefundedDomainEvent) error {
	return h.gateway.Refund(ctx, e.TransactionID, e.Amount)
}
