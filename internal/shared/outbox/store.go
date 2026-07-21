package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/llannillo/mm/internal/shared/events"
)

// InsertMessages persists domainEvents as rows in schema.outbox_messages,
// using tx so the write commits atomically with the aggregate change that
// raised them. Dispatch happens later, out of band, in the module's Worker.
func InsertMessages(ctx context.Context, tx pgx.Tx, schema string, domainEvents []events.DomainEvent) error {
	if len(domainEvents) == 0 {
		return nil
	}

	sql := fmt.Sprintf(
		`INSERT INTO %s.outbox_messages (id, type, content, occurred_on_utc) VALUES ($1, $2, $3, $4)`,
		schema,
	)

	for _, e := range domainEvents {
		content, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("outbox: marshal %T: %w", e, err)
		}

		if _, err := tx.Exec(ctx, sql, uuid.New(), reflect.TypeOf(e).Name(), content, time.Now().UTC()); err != nil {
			return fmt.Errorf("outbox: insert message %T: %w", e, err)
		}
	}

	return nil
}
