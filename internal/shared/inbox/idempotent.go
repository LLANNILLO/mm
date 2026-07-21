// Package inbox protects cross-module integration-event consumers from
// re-running when the sending module retries a failed outbox message.
//
// Our eventbus.EventBus is a synchronous in-process call, not a broker —
// there's no message redelivery to guard against. The real risk is that the
// sending and receiving modules commit to independent Postgres transactions:
// if the process crashes between the receiver committing its side effect and
// the sender committing its own outbox_message_consumers row, the sender's
// next retry re-publishes and re-invokes the receiver's consumer with no
// protection of its own. Idempotent closes that gap on the receiving side,
// mirroring outbox.Idempotent but keyed on inbox_message_consumers.
package inbox

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/llannillo/mm/internal/shared/events"
	"github.com/llannillo/mm/internal/shared/eventbus"
)

// Idempotent decorates an integration-event consumer so it runs at most once
// per (message, consumer name) pair. name must be stable and unique per
// consumer within the module — pass an explicit identifier at the Subscribe
// call site rather than deriving it from the handler value.
//
// Idempotent relies on the outbox message id already present in ctx (set by
// the sending module's outbox.Worker and propagated unchanged through
// EventBus.Publish) and errors loudly if it's missing, rather than silently
// skipping the idempotency check.
func Idempotent[T eventbus.IntegrationEvent](
	name string,
	pool *pgxpool.Pool,
	schema string,
	inner func(ctx context.Context, event T) error,
) func(ctx context.Context, event T) error {
	return func(ctx context.Context, event T) error {
		messageID, ok := events.MessageIDFromContext(ctx)
		if !ok {
			return fmt.Errorf(
				"inbox: missing message id in context for consumer %q — Idempotent must run downstream of an outbox worker dispatch",
				name,
			)
		}

		var alreadyProcessed bool
		existsSQL := fmt.Sprintf(
			`SELECT EXISTS(SELECT 1 FROM %s.inbox_message_consumers WHERE message_id = $1 AND name = $2)`,
			schema,
		)
		if err := pool.QueryRow(ctx, existsSQL, messageID, name).Scan(&alreadyProcessed); err != nil {
			return fmt.Errorf("inbox: check consumer %q: %w", name, err)
		}
		if alreadyProcessed {
			return nil
		}

		if err := inner(ctx, event); err != nil {
			return err
		}

		insertSQL := fmt.Sprintf(
			`INSERT INTO %s.inbox_message_consumers (message_id, name) VALUES ($1, $2)`,
			schema,
		)
		if _, err := pool.Exec(ctx, insertSQL, messageID, name); err != nil {
			return fmt.Errorf("inbox: record consumer %q: %w", name, err)
		}

		return nil
	}
}
