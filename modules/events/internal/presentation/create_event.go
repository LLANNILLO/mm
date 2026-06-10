package presentation

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	createevent "github.com/llannillo/mm/modules/events/internal/application/commands/create_event"
)

type createEventRequest struct {
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Location    *string    `json:"location"`
	StartsAtUtc time.Time  `json:"starts_at_utc"`
	EndsAtUtc   *time.Time `json:"ends_at_utc"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	id, err := h.createEvent.Handle(r.Context(), createevent.Command{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartsAtUtc: req.StartsAtUtc,
		EndsAtUtc:   req.EndsAtUtc,
	})
	if err != nil {
		log.Printf("create event: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create event")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"id": id})
}
