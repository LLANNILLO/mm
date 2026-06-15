package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/events/internal/ports/inbound"
)

type Handler struct {
	events     inbound.EventService
	categories inbound.CategoryService
	tickets    inbound.TicketService
}

func NewHandler(events inbound.EventService, categories inbound.CategoryService, tickets inbound.TicketService) *Handler {
	return &Handler{events: events, categories: categories, tickets: tickets}
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
