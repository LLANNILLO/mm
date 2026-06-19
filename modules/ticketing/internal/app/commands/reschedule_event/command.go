package rescheduleevent

import (
	"time"

	"github.com/google/uuid"
)

type Command struct {
	EventID     uuid.UUID
	StartsAtUtc time.Time
	EndsAtUtc   *time.Time
}
