package outbox

import (
	"context"

	"github.com/google/uuid"
)

type messageIDKey struct{}

// WithMessageID attaches the outbox message id being processed to ctx, so
// Idempotent-decorated handlers downstream can key their tracking row on it.
// Set exclusively by Worker before calling Dispatcher.Dispatch.
func WithMessageID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, messageIDKey{}, id)
}

// MessageIDFromContext retrieves the id set by WithMessageID.
func MessageIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(messageIDKey{}).(uuid.UUID)
	return id, ok
}
