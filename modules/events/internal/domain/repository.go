package domain

import (
	"context"
)

type EventRepository interface {
	Insert(ctx context.Context, event *Event) error
	// GetEvent(ctx context.Context, id uuid.UUID)
}
