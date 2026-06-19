package domain

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	entity
	ID             uuid.UUID
	OrderID        uuid.UUID
	TransactionID  uuid.UUID
	Amount         int64
	Currency       string
	AmountRefunded *int64
	CreatedAtUtc   time.Time
	RefundedAtUtc  *time.Time
}

func NewPayment(order *Order, transactionID uuid.UUID, amount int64, currency string) *Payment {
	p := &Payment{
		ID:            uuid.New(),
		OrderID:       order.ID,
		TransactionID: transactionID,
		Amount:        amount,
		Currency:      currency,
		CreatedAtUtc:  time.Now().UTC(),
	}
	p.raise(PaymentCreatedDomainEvent{PaymentID: p.ID})
	return p
}

func (p *Payment) Refund(amount int64) error {
	currentRefunded := int64(0)
	if p.AmountRefunded != nil {
		currentRefunded = *p.AmountRefunded
	}

	if currentRefunded >= p.Amount {
		return ErrPaymentAlreadyRefunded
	}

	if currentRefunded+amount > p.Amount {
		return ErrPaymentRefundExceedsAmount
	}

	newRefunded := currentRefunded + amount
	p.AmountRefunded = &newRefunded

	if newRefunded == p.Amount {
		now := time.Now().UTC()
		p.RefundedAtUtc = &now
		p.raise(PaymentRefundedDomainEvent{
			PaymentID:     p.ID,
			TransactionID: p.TransactionID,
			Amount:        amount,
		})
	} else {
		p.raise(PaymentPartiallyRefundedDomainEvent{
			PaymentID:     p.ID,
			TransactionID: p.TransactionID,
			Amount:        amount,
		})
	}

	return nil
}
