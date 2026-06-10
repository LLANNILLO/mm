package getevent

import (
	"time"

	"github.com/google/uuid"
)

type Response struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Location    *string    `json:"location"`
	StartsAtUtc time.Time  `json:"starts_at_utc"`
	EndsAtUtc   *time.Time `json:"ends_at_utc"`
}
