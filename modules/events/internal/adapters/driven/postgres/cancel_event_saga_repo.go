package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	store "github.com/llannillo/mm/modules/events/internal/adapters/driven/postgres/generated"
	"github.com/llannillo/mm/modules/events/internal/ports/outbound"
)

// CancelEventSagaRepository writes directly through store.Queries — no
// UnitOfWork. Each call is a single statement invoked from an
// outbox/inbox.Idempotent-wrapped handler, which already guarantees
// at-least-once, deduplicated delivery.
type CancelEventSagaRepository struct {
	queries *store.Queries
}

func NewCancelEventSagaRepository(q *store.Queries) *CancelEventSagaRepository {
	return &CancelEventSagaRepository{queries: q}
}

func (r *CancelEventSagaRepository) Start(ctx context.Context, eventID uuid.UUID) error {
	if err := r.queries.StartCancelEventSaga(ctx, eventID); err != nil {
		return fmt.Errorf("start cancel event saga: %w", err)
	}
	return nil
}

func (r *CancelEventSagaRepository) MarkStepComplete(ctx context.Context, eventID uuid.UUID, step outbound.Step) (outbound.Step, error) {
	completed, err := r.queries.MarkCancelEventSagaStepComplete(ctx, store.MarkCancelEventSagaStepCompleteParams{
		Step:    int16(step),
		EventID: eventID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("cancel event saga %s not found", eventID)
		}
		return 0, fmt.Errorf("mark cancel event saga step complete: %w", err)
	}
	return outbound.Step(completed), nil
}

func (r *CancelEventSagaRepository) Delete(ctx context.Context, eventID uuid.UUID) error {
	if err := r.queries.DeleteCancelEventSagaState(ctx, eventID); err != nil {
		return fmt.Errorf("delete cancel event saga state: %w", err)
	}
	return nil
}
