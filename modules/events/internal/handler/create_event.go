package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/events/internal/store"
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.queries.CreateEvent(r.Context(), store.CreateEventParams{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartsAtUtc: req.StartsAtUtc,
		EndsAtUtc:   *req.EndsAtUtc,
		Status:      "draft",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": id})
}
