package events

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/llannillo/mm/modules/events/internal/handler"
	"github.com/llannillo/mm/modules/events/internal/store"
)

type Module struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Module {
	return &Module{
		db: db,
	}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	h := handler.New(store.New(m.db))
	mux.HandleFunc("POST /events", h.Create)
	mux.HandleFunc("GET /events/{id}", h.Get)
}
