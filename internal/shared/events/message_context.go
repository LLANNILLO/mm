package events

import (
	"context"

	"github.com/google/uuid"
)

type messageIDKey struct{}

// WithMessageID attaches the id of the message currently being dispatched to
// ctx — set by outbox.Worker before calling Dispatch, and read by both
// outbox.Idempotent (sending side) and inbox.Idempotent (receiving side) to
// key their tracking rows. Lives here, not in outbox, because it's a
// property of "which message is being dispatched right now", relevant to the
// whole event flow rather than specific to writing outbox rows.
func WithMessageID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, messageIDKey{}, id)
}

// MessageIDFromContext retrieves the id set by WithMessageID.
func MessageIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(messageIDKey{}).(uuid.UUID)
	return id, ok
}
