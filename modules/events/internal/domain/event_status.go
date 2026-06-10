package domain

type EventStatus string

const (
	DraftStatus     EventStatus = "draft"
	PublishedStatus EventStatus = "published"
	CanelledStatus  EventStatus = "cancelled"
	CompletedStatus EventStatus = "completed"
)
