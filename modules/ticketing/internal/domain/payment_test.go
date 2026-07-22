package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPayment(t *testing.T, amount int64) *Payment {
	t.Helper()

	order := NewOrder(uuid.New())
	return NewPayment(order, uuid.New(), amount, "USD")
}

func TestNewPayment_RaisesDomainEvent(t *testing.T) {
	p := newTestPayment(t, 1000)

	domainEvent := assertDomainEventPublished[PaymentCreatedDomainEvent](t, p)
	assert.Equal(t, p.ID(), domainEvent.PaymentID)
}

func TestPayment_Refund_RaisesDomainEvent_WhenFullyRefunded(t *testing.T) {
	p := newTestPayment(t, 1000)
	p.ClearDomainEvents()

	err := p.Refund(1000)

	require.NoError(t, err)
	require.NotNil(t, p.AmountRefunded())
	assert.Equal(t, int64(1000), *p.AmountRefunded())
	assert.NotNil(t, p.RefundedAtUtc())
	domainEvent := assertDomainEventPublished[PaymentRefundedDomainEvent](t, p)
	assert.Equal(t, p.ID(), domainEvent.PaymentID)
	assert.Equal(t, int64(1000), domainEvent.Amount)
}

func TestPayment_Refund_RaisesDomainEvent_WhenPartiallyRefunded(t *testing.T) {
	p := newTestPayment(t, 1000)
	p.ClearDomainEvents()

	err := p.Refund(400)

	require.NoError(t, err)
	assert.Nil(t, p.RefundedAtUtc())
	domainEvent := assertDomainEventPublished[PaymentPartiallyRefundedDomainEvent](t, p)
	assert.Equal(t, p.ID(), domainEvent.PaymentID)
	assert.Equal(t, int64(400), domainEvent.Amount)
}

func TestPayment_Refund_ReturnsError_WhenAlreadyFullyRefunded(t *testing.T) {
	p := newTestPayment(t, 1000)
	require.NoError(t, p.Refund(1000))

	err := p.Refund(1)

	assert.ErrorIs(t, err, ErrPaymentAlreadyRefunded)
}

func TestPayment_Refund_ReturnsError_WhenExceedsAmount(t *testing.T) {
	p := newTestPayment(t, 1000)

	err := p.Refund(1001)

	assert.ErrorIs(t, err, ErrPaymentRefundExceedsAmount)
}

func TestPayment_Refund_ReturnsError_WhenCumulativeExceedsAmount(t *testing.T) {
	p := newTestPayment(t, 1000)
	require.NoError(t, p.Refund(600))

	err := p.Refund(500)

	assert.ErrorIs(t, err, ErrPaymentRefundExceedsAmount)
}
