package searchevents

import (
	"time"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/events/internal/domain"
)

type Query struct {
	Status     domain.EventStatus
	CategoryID *uuid.UUID
	StartsFrom *time.Time
	EndsFrom   *time.Time
	Page       int32
	PageSize   int32
}

type EventItem struct {
	ID          uuid.UUID  `json:"id"`
	CategoryID  uuid.UUID  `json:"category_id"`
	Title       string     `json:"title"`
	StartsAtUtc time.Time  `json:"starts_at_utc"`
	EndsAtUtc   *time.Time `json:"ends_at_utc"`
}

type Page[T any] struct {
	Items      []T   `json:"items"`
	TotalCount int64 `json:"total_count"`
	Page       int32 `json:"page"`
	PageSize   int32 `json:"page_size"`
}
