package outbox

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/llannillo/mm/internal/shared/events"
)

// Idempotent decorates a domain event handler so it runs at most once per
// (outbox message, handler name) pair. Needed because a single domain event
// can fan out to multiple handlers (e.g. archiving tickets AND refunding
// payments on the same EventCancelledDomainEvent) — if one handler fails, the
// outbox worker retries the whole message, and without this guard, handlers
// that already succeeded would re-run and duplicate their side effects.
//
// name must be stable and unique per handler within the module (it's part of
// the tracking row's key) — pass an explicit identifier at the Register call
// site rather than deriving it from the handler value.
//
// Idempotent must only run inside the outbox worker: it requires the message
// id to be present in ctx via WithMessageID, and errors loudly if it's not,
// rather than silently skipping the idempotency check.
func Idempotent[T events.DomainEvent](
	name string,
	pool *pgxpool.Pool,
	schema string,
	inner func(ctx context.Context, event T) error,
) func(ctx context.Context, event T) error {
	return func(ctx context.Context, event T) error {
		messageID, ok := MessageIDFromContext(ctx)
		if !ok {
			return fmt.Errorf(
				"outbox: missing message id in context for handler %q — Idempotent must be called from the outbox worker",
				name,
			)
		}

		var alreadyProcessed bool
		existsSQL := fmt.Sprintf(
			`SELECT EXISTS(SELECT 1 FROM %s.outbox_message_consumers WHERE outbox_message_id = $1 AND name = $2)`,
			schema,
		)
		if err := pool.QueryRow(ctx, existsSQL, messageID, name).Scan(&alreadyProcessed); err != nil {
			return fmt.Errorf("outbox: check consumer %q: %w", name, err)
		}
		if alreadyProcessed {
			return nil
		}

		if err := inner(ctx, event); err != nil {
			return err
		}

		insertSQL := fmt.Sprintf(
			`INSERT INTO %s.outbox_message_consumers (outbox_message_id, name) VALUES ($1, $2)`,
			schema,
		)
		if _, err := pool.Exec(ctx, insertSQL, messageID, name); err != nil {
			return fmt.Errorf("outbox: record consumer %q: %w", name, err)
		}

		return nil
	}
}
