package events

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	createevent "github.com/llannillo/mm/modules/events/internal/application/commands/create_event"
	getevent "github.com/llannillo/mm/modules/events/internal/application/queries/get-event"
	"github.com/llannillo/mm/modules/events/internal/infrastructure/persistence"
	"github.com/llannillo/mm/modules/events/internal/presentation"
	"github.com/llannillo/mm/modules/events/internal/store"
)

type Module struct {
	handler *presentation.Handler
}

func New(db *pgxpool.Pool) *Module {
	queries := store.New(db)

	repo := persistence.NewEventRepository(queries)
	reader := persistence.NewEventReader(queries)

	return &Module{
		handler: presentation.NewHandler(
			createevent.NewHandler(repo),
			getevent.NewHandler(reader),
		),
	}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /events", m.handler.Create)
	mux.HandleFunc("GET /events/{id}", m.handler.Get)
}
