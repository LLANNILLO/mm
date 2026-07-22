package createevent

import (
	"time"

	"github.com/google/uuid"
)

type Command struct {
	EventID     uuid.UUID
	Title       string
	Description *string
	Location    *string
	StartsAtUtc time.Time
	EndsAtUtc   *time.Time
}
