package ticketing

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/llannillo/mm/internal/shared"
	additemtocart "github.com/llannillo/mm/modules/ticketing/internal/app/commands/add_item_to_cart"
	ticketingapp "github.com/llannillo/mm/modules/ticketing/internal/app"
	httphandler "github.com/llannillo/mm/modules/ticketing/internal/adapters/driving/http"
	eventsapi "github.com/llannillo/mm/modules/events/api"
	usersapi "github.com/llannillo/mm/modules/users/api"
)

const moduleName = "ticketing"

type Module struct {
	handler *httphandler.Handler
}

func New(app shared.App, eventsAPI eventsapi.EventsAPI, usersAPI usersapi.UsersAPI) *Module {
	cartService := ticketingapp.NewCartService(app.Cache)

	carts := &cartServiceImpl{
		log:          app.Logger,
		addItemToCart: additemtocart.NewHandler(cartService, eventsAPI, usersAPI),
	}

	return &Module{
		handler: httphandler.NewHandler(carts),
	}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.handler.RegisterRoutes(mux)
}

func logHandler(log *slog.Logger, ctx context.Context, name string) func(error) {
	log.InfoContext(ctx, "Processing request", "module", moduleName, "request", name)
	return func(err error) {
		if err != nil {
			log.ErrorContext(ctx, "Completed request with error", "module", moduleName, "request", name, "error", err)
			return
		}
		log.InfoContext(ctx, "Completed request", "module", moduleName, "request", name)
	}
}

type cartServiceImpl struct {
	log          *slog.Logger
	addItemToCart *additemtocart.Handler
}

func (s *cartServiceImpl) AddItemToCart(ctx context.Context, cmd additemtocart.Command) error {
	done := logHandler(s.log, ctx, "AddItemToCart")
	err := s.addItemToCart.Handle(ctx, cmd)
	done(err)
	return err
}
