package presentation

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	archivecategory "github.com/llannillo/mm/modules/events/internal/application/commands/archive_category"
	cancelevent "github.com/llannillo/mm/modules/events/internal/application/commands/cancel_event"
	createcategory "github.com/llannillo/mm/modules/events/internal/application/commands/create_category"
	createevent "github.com/llannillo/mm/modules/events/internal/application/commands/create_event"
	createtickettype "github.com/llannillo/mm/modules/events/internal/application/commands/create_ticket_type"
	publishevent "github.com/llannillo/mm/modules/events/internal/application/commands/publish_event"
	renamecategory "github.com/llannillo/mm/modules/events/internal/application/commands/rename_category"
	rescheduleevent "github.com/llannillo/mm/modules/events/internal/application/commands/reschedule_event"
	updateticketprice "github.com/llannillo/mm/modules/events/internal/application/commands/update_ticket_price"
	getevent "github.com/llannillo/mm/modules/events/internal/application/queries/get-event"
	getcategory "github.com/llannillo/mm/modules/events/internal/application/queries/get_category"
	gettickettype "github.com/llannillo/mm/modules/events/internal/application/queries/get_ticket_type"
	listcategories "github.com/llannillo/mm/modules/events/internal/application/queries/list_categories"
	listevents "github.com/llannillo/mm/modules/events/internal/application/queries/list_events"
	listtickettype "github.com/llannillo/mm/modules/events/internal/application/queries/list_ticket_types"
	searchevents "github.com/llannillo/mm/modules/events/internal/application/queries/search_events"
)

type (
	createEventHandler interface {
		Handle(ctx context.Context, cmd createevent.Command) (uuid.UUID, error)
	}
	getEventHandler interface {
		Handle(ctx context.Context, q getevent.Query) (*getevent.Response, error)
	}
	listEventsHandler interface {
		Handle(ctx context.Context) ([]listevents.EventItem, error)
	}
	searchEventsHandler interface {
		Handle(ctx context.Context, q searchevents.Query) (*searchevents.Page[searchevents.EventItem], error)
	}
	publishEventHandler interface {
		Handle(ctx context.Context, cmd publishevent.Command) error
	}
	cancelEventHandler interface {
		Handle(ctx context.Context, cmd cancelevent.Command) error
	}
	rescheduleEventHandler interface {
		Handle(ctx context.Context, cmd rescheduleevent.Command) error
	}
	createCategoryHandler interface {
		Handle(ctx context.Context, cmd createcategory.Command) (uuid.UUID, error)
	}
	getCategoryHandler interface {
		Handle(ctx context.Context, q getcategory.Query) (*getcategory.Response, error)
	}
	listCategoriesHandler interface {
		Handle(ctx context.Context) ([]listcategories.CategoryItem, error)
	}
	archiveCategoryHandler interface {
		Handle(ctx context.Context, cmd archivecategory.Command) error
	}
	renameCategoryHandler interface {
		Handle(ctx context.Context, cmd renamecategory.Command) error
	}
	createTicketTypeHandler interface {
		Handle(ctx context.Context, cmd createtickettype.Command) (uuid.UUID, error)
	}
	getTicketTypeHandler interface {
		Handle(ctx context.Context, q gettickettype.Query) (*gettickettype.Response, error)
	}
	listTicketTypesHandler interface {
		Handle(ctx context.Context, q listtickettype.Query) ([]listtickettype.TicketTypeItem, error)
	}
	updateTicketPriceHandler interface {
		Handle(ctx context.Context, cmd updateticketprice.Command) error
	}
)

type Deps struct {
	CreateEvent       createEventHandler
	GetEvent          getEventHandler
	ListEvents        listEventsHandler
	SearchEvents      searchEventsHandler
	PublishEvent      publishEventHandler
	CancelEvent       cancelEventHandler
	RescheduleEvent   rescheduleEventHandler
	CreateCategory    createCategoryHandler
	GetCategory       getCategoryHandler
	ListCategories    listCategoriesHandler
	ArchiveCategory   archiveCategoryHandler
	RenameCategory    renameCategoryHandler
	CreateTicketType  createTicketTypeHandler
	GetTicketType     getTicketTypeHandler
	ListTicketTypes   listTicketTypesHandler
	UpdateTicketPrice updateTicketPriceHandler
}

type Handler struct{ deps Deps }

func NewHandler(deps Deps) *Handler {
	return &Handler{deps: deps}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /events", h.createEvent)
	mux.HandleFunc("GET /events", h.listEvents)
	mux.HandleFunc("GET /events/search", h.searchEvents)
	mux.HandleFunc("GET /events/{id}", h.getEvent)
	mux.HandleFunc("POST /events/{id}/publish", h.publishEvent)
	mux.HandleFunc("POST /events/{id}/cancel", h.cancelEvent)
	mux.HandleFunc("PUT /events/{id}/reschedule", h.rescheduleEvent)

	mux.HandleFunc("POST /categories", h.createCategory)
	mux.HandleFunc("GET /categories", h.listCategories)
	mux.HandleFunc("GET /categories/{id}", h.getCategory)
	mux.HandleFunc("POST /categories/{id}/archive", h.archiveCategory)
	mux.HandleFunc("PUT /categories/{id}/name", h.renameCategory)

	mux.HandleFunc("POST /ticket-types", h.createTicketType)
	mux.HandleFunc("GET /ticket-types", h.listTicketTypes)
	mux.HandleFunc("GET /ticket-types/{id}", h.getTicketType)
	mux.HandleFunc("PUT /ticket-types/{id}/price", h.updateTicketTypePrice)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parseUUID(w http.ResponseWriter, raw string) (uuid.UUID, bool) {
	id, err := uuid.Parse(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return uuid.Nil, false
	}
	return id, true
}

func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return v, false
	}
	return v, true
}
