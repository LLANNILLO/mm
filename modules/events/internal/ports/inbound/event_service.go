package inbound

import (
	"context"

	"github.com/google/uuid"
	cancelevent "github.com/llannillo/mm/modules/events/internal/app/commands/cancel_event"
	createevent "github.com/llannillo/mm/modules/events/internal/app/commands/create_event"
	publishevent "github.com/llannillo/mm/modules/events/internal/app/commands/publish_event"
	rescheduleevent "github.com/llannillo/mm/modules/events/internal/app/commands/reschedule_event"
	getevent "github.com/llannillo/mm/modules/events/internal/app/queries/get_event"
	listevents "github.com/llannillo/mm/modules/events/internal/app/queries/list_events"
	searchevents "github.com/llannillo/mm/modules/events/internal/app/queries/search_events"
)

type EventService interface {
	CreateEvent(ctx context.Context, cmd createevent.Command) (uuid.UUID, error)
	PublishEvent(ctx context.Context, cmd publishevent.Command) error
	CancelEvent(ctx context.Context, cmd cancelevent.Command) error
	RescheduleEvent(ctx context.Context, cmd rescheduleevent.Command) error
	GetEvent(ctx context.Context, q getevent.Query) (*getevent.Response, error)
	ListEvents(ctx context.Context) ([]listevents.EventItem, error)
	SearchEvents(ctx context.Context, q searchevents.Query) (*searchevents.Page[searchevents.EventItem], error)
}
