package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/users/internal/ports/inbound"
)

type Handler struct {
	users inbound.UserService
}

func NewHandler(users inbound.UserService) *Handler {
	return &Handler{users: users}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users/register", h.registerUser)
	mux.HandleFunc("GET /users/{id}/profile", h.getUserProfile)
	mux.HandleFunc("PUT /users/{id}/profile", h.updateUserProfile)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func parseUUID(w http.ResponseWriter, raw string) (uuid.UUID, bool) {
	id, err := uuid.Parse(raw)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return uuid.Nil, false
	}
	return id, true
}

func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return v, false
	}
	return v, true
}
