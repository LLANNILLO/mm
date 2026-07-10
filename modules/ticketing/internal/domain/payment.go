package domain

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	entity
	id             uuid.UUID
	orderID        uuid.UUID
	transactionID  uuid.UUID
	amount         int64
	currency       string
	amountRefunded *int64
	createdAtUtc   time.Time
	refundedAtUtc  *time.Time
}

func (p *Payment) ID() uuid.UUID             { return p.id }
func (p *Payment) OrderID() uuid.UUID        { return p.orderID }
func (p *Payment) TransactionID() uuid.UUID  { return p.transactionID }
func (p *Payment) Amount() int64             { return p.amount }
func (p *Payment) Currency() string          { return p.currency }
func (p *Payment) AmountRefunded() *int64    { return p.amountRefunded }
func (p *Payment) CreatedAtUtc() time.Time   { return p.createdAtUtc }
func (p *Payment) RefundedAtUtc() *time.Time { return p.refundedAtUtc }

func NewPayment(order *Order, transactionID uuid.UUID, amount int64, currency string) *Payment {
	p := &Payment{
		id:            uuid.New(),
		orderID:       order.ID(),
		transactionID: transactionID,
		amount:        amount,
		currency:      currency,
		createdAtUtc:  time.Now().UTC(),
	}
	p.raise(PaymentCreatedDomainEvent{PaymentID: p.id})
	return p
}

// RehydratePayment reconstructs a Payment from persisted state without
// raising domain events. Only the PaymentRepository may call this.
func RehydratePayment(
	id, orderID, transactionID uuid.UUID,
	amount int64,
	currency string,
	amountRefunded *int64,
	createdAtUtc time.Time,
	refundedAtUtc *time.Time,
) *Payment {
	return &Payment{
		id:             id,
		orderID:        orderID,
		transactionID:  transactionID,
		amount:         amount,
		currency:       currency,
		amountRefunded: amountRefunded,
		createdAtUtc:   createdAtUtc,
		refundedAtUtc:  refundedAtUtc,
	}
}

func (p *Payment) Refund(amount int64) error {
	currentRefunded := int64(0)
	if p.amountRefunded != nil {
		currentRefunded = *p.amountRefunded
	}

	if currentRefunded >= p.amount {
		return ErrPaymentAlreadyRefunded
	}

	if currentRefunded+amount > p.amount {
		return ErrPaymentRefundExceedsAmount
	}

	newRefunded := currentRefunded + amount
	p.amountRefunded = &newRefunded

	if newRefunded == p.amount {
		now := time.Now().UTC()
		p.refundedAtUtc = &now
		p.raise(PaymentRefundedDomainEvent{
			PaymentID:     p.id,
			TransactionID: p.transactionID,
			Amount:        amount,
		})
	} else {
		p.raise(PaymentPartiallyRefundedDomainEvent{
			PaymentID:     p.id,
			TransactionID: p.transactionID,
			Amount:        amount,
		})
	}

	return nil
}
