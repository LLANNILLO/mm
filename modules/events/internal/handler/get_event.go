package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type getEventResponse struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Location    *string    `json:"location"`
	StartsAtUtc time.Time  `json:"starts_at_utc"`
	EndsAtUtc   *time.Time `json:"ends_at_utc"`
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	row, err := h.queries.GetEvent(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"event": getEventResponse{
		ID:          row.ID,
		Title:       row.Title,
		Description: row.Description,
		Location:    row.Location,
		StartsAtUtc: row.StartsAtUtc,
		EndsAtUtc:   &row.EndsAtUtc,
	}})
}
