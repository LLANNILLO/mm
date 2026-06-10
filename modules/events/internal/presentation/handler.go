package presentation

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	createevent "github.com/llannillo/mm/modules/events/internal/application/commands/create_event"
	getevent "github.com/llannillo/mm/modules/events/internal/application/queries/get-event"
)

type createEventHandler interface {
	Handle(ctx context.Context, cmd createevent.Command) (uuid.UUID, error)
}

type getEventHandler interface {
	Handle(ctx context.Context, q getevent.Query) (*getevent.Response, error)
}

type Handler struct {
	createEvent createEventHandler
	getEvent    getEventHandler
}

func NewHandler(create createEventHandler, get getEventHandler) *Handler {
	return &Handler{createEvent: create, getEvent: get}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
