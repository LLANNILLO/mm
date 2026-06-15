package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	cancelevent "github.com/llannillo/mm/modules/events/internal/app/commands/cancel_event"
	createevent "github.com/llannillo/mm/modules/events/internal/app/commands/create_event"
	publishevent "github.com/llannillo/mm/modules/events/internal/app/commands/publish_event"
	rescheduleevent "github.com/llannillo/mm/modules/events/internal/app/commands/reschedule_event"
	getevent "github.com/llannillo/mm/modules/events/internal/app/queries/get_event"
	searchevents "github.com/llannillo/mm/modules/events/internal/app/queries/search_events"
	"github.com/llannillo/mm/modules/events/internal/domain"
)

func (h *Handler) createEvent(w http.ResponseWriter, r *http.Request) {
	type request struct {
		CategoryID  uuid.UUID  `json:"category_id"`
		Title       string     `json:"title"`
		Description *string    `json:"description"`
		Location    *string    `json:"location"`
		StartsAtUtc time.Time  `json:"starts_at_utc"`
		EndsAtUtc   *time.Time `json:"ends_at_utc"`
	}
	req, ok := decodeJSON[request](w, r)
	if !ok {
		return
	}
	id, err := h.events.CreateEvent(r.Context(), createevent.Command{
		CategoryID:  req.CategoryID,
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartsAtUtc: req.StartsAtUtc,
		EndsAtUtc:   req.EndsAtUtc,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (h *Handler) getEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("id"))
	if !ok {
		return
	}
	resp, err := h.events.GetEvent(r.Context(), getevent.Query{ID: id})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) listEvents(w http.ResponseWriter, r *http.Request) {
	items, err := h.events.ListEvents(r.Context())
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) searchEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	status := domain.EventStatus(q.Get("status"))
	if status == "" {
		status = domain.StatusPublished
	}

	query := searchevents.Query{
		Status:   status,
		Page:     1,
		PageSize: 20,
	}

	if raw := q.Get("category-id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid category_id")
			return
		}
		query.CategoryID = &id
	}

	page, err := h.events.SearchEvents(r.Context(), query)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *Handler) publishEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("id"))
	if !ok {
		return
	}
	if err := h.events.PublishEvent(r.Context(), publishevent.Command{EventID: id}); err != nil {
		handleDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) cancelEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("id"))
	if !ok {
		return
	}
	if err := h.events.CancelEvent(r.Context(), cancelevent.Command{EventID: id}); err != nil {
		handleDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) rescheduleEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("id"))
	if !ok {
		return
	}
	type request struct {
		StartsAtUtc time.Time  `json:"starts_at_utc"`
		EndsAtUtc   *time.Time `json:"ends_at_utc"`
	}
	req, ok := decodeJSON[request](w, r)
	if !ok {
		return
	}
	if err := h.events.RescheduleEvent(r.Context(), rescheduleevent.Command{
		EventID:     id,
		StartsAtUtc: req.StartsAtUtc,
		EndsAtUtc:   req.EndsAtUtc,
	}); err != nil {
		handleDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
