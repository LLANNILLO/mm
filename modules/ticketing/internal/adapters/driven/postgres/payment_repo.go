package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/llannillo/mm/internal/shared/outbox"
	store "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres/generated"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

type PaymentRepository struct {
	queries *store.Queries
	uow     *UnitOfWork
}

func NewPaymentRepository(q *store.Queries, uow *UnitOfWork) *PaymentRepository {
	return &PaymentRepository{queries: q, uow: uow}
}

func (r *PaymentRepository) Insert(ctx context.Context, p *domain.Payment) error {
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		createdAtUtc := pgtype.Timestamptz{Time: p.CreatedAtUtc(), Valid: true}

		amountRefunded := pgtype.Int8{}
		if p.AmountRefunded() != nil {
			amountRefunded = pgtype.Int8{Int64: *p.AmountRefunded(), Valid: true}
		}

		refundedAtUtc := pgtype.Timestamptz{}
		if p.RefundedAtUtc() != nil {
			refundedAtUtc = pgtype.Timestamptz{Time: *p.RefundedAtUtc(), Valid: true}
		}

		if err := q.InsertPayment(ctx, store.InsertPaymentParams{
			ID:             p.ID(),
			OrderID:        p.OrderID(),
			TransactionID:  p.TransactionID(),
			Amount:         p.Amount(),
			Currency:       p.Currency(),
			AmountRefunded: amountRefunded,
			CreatedAtUtc:   createdAtUtc,
			RefundedAtUtc:  refundedAtUtc,
		}); err != nil {
			return fmt.Errorf("insert payment: %w", err)
		}

		domainEvents := p.DomainEvents()
		p.ClearDomainEvents()
		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	row, err := r.queries.GetPaymentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("get payment: %w", err)
	}

	var amountRefunded *int64
	if row.AmountRefunded.Valid {
		v := row.AmountRefunded.Int64
		amountRefunded = &v
	}
	var refundedAtUtc *time.Time
	if row.RefundedAtUtc.Valid {
		t := row.RefundedAtUtc.Time
		refundedAtUtc = &t
	}
	return domain.RehydratePayment(row.ID, row.OrderID, row.TransactionID, row.Amount, row.Currency, amountRefunded, row.CreatedAtUtc.Time, refundedAtUtc), nil
}
