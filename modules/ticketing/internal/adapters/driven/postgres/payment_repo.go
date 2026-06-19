package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	store "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres/generated"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

type PaymentRepository struct {
	queries *store.Queries
}

func NewPaymentRepository(q *store.Queries) *PaymentRepository {
	return &PaymentRepository{queries: q}
}

func (r *PaymentRepository) Insert(ctx context.Context, p *domain.Payment) error {
	createdAtUtc := pgtype.Timestamptz{Time: p.CreatedAtUtc, Valid: true}

	amountRefunded := pgtype.Int8{}
	if p.AmountRefunded != nil {
		amountRefunded = pgtype.Int8{Int64: *p.AmountRefunded, Valid: true}
	}

	refundedAtUtc := pgtype.Timestamptz{}
	if p.RefundedAtUtc != nil {
		refundedAtUtc = pgtype.Timestamptz{Time: *p.RefundedAtUtc, Valid: true}
	}

	err := r.queries.InsertPayment(ctx, store.InsertPaymentParams{
		ID:             p.ID,
		OrderID:        p.OrderID,
		TransactionID:  p.TransactionID,
		Amount:         p.Amount,
		Currency:       p.Currency,
		AmountRefunded: amountRefunded,
		CreatedAtUtc:   createdAtUtc,
		RefundedAtUtc:  refundedAtUtc,
	})
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}
	p.ClearDomainEvents()
	return nil
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	row, err := r.queries.GetPaymentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("get payment: %w", err)
	}

	payment := &domain.Payment{
		ID:            row.ID,
		OrderID:       row.OrderID,
		TransactionID: row.TransactionID,
		Amount:        row.Amount,
		Currency:      row.Currency,
		CreatedAtUtc:  row.CreatedAtUtc.Time,
	}
	if row.AmountRefunded.Valid {
		v := row.AmountRefunded.Int64
		payment.AmountRefunded = &v
	}
	if row.RefundedAtUtc.Valid {
		t := row.RefundedAtUtc.Time
		payment.RefundedAtUtc = &t
	}
	return payment, nil
}
