package ticketing

import (
	"net/http"

	"github.com/llannillo/mm/internal/shared"
	httphandler "github.com/llannillo/mm/modules/ticketing/internal/adapters/driving/http"
)

const moduleName = "ticketing"

type Module struct {
	handler *httphandler.Handler
}

func New(app shared.App) *Module {
	return &Module{
		handler: httphandler.NewHandler(),
	}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.handler.RegisterRoutes(mux)
}
