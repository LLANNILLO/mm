package domain

type EventStatus string

const (
	StatusDraft     EventStatus = "draft"
	StatusPublished EventStatus = "published"
	StatusCancelled EventStatus = "cancelled"
	StatusCompleted EventStatus = "completed"
)
