package ticketing

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/llannillo/mm/internal/shared"
	ticketingapp "github.com/llannillo/mm/modules/ticketing/internal/app"
	additemtocart "github.com/llannillo/mm/modules/ticketing/internal/app/commands/add_item_to_cart"
	createcustomer "github.com/llannillo/mm/modules/ticketing/internal/app/commands/create_customer"
	updatecustomer "github.com/llannillo/mm/modules/ticketing/internal/app/commands/update_customer"
	pg "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres"
	pgstore "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres/generated"
	httphandler "github.com/llannillo/mm/modules/ticketing/internal/adapters/driving/http"
	eventsapi "github.com/llannillo/mm/modules/events/api"
	ticketingapi "github.com/llannillo/mm/modules/ticketing/api"
)

const moduleName = "ticketing"

type Module struct {
	handler        *httphandler.Handler
	createCustomer *createcustomer.Handler
	updateCustomer *updatecustomer.Handler
}

func New(app shared.App, eventsAPI eventsapi.EventsAPI) *Module {
	queries := pgstore.New(app.DB)
	customerRepo := pg.NewCustomerRepository(queries)
	cartService := ticketingapp.NewCartService(app.Cache)

	carts := &cartServiceImpl{
		log:          app.Logger,
		addItemToCart: additemtocart.NewHandler(cartService, customerRepo, eventsAPI),
	}

	return &Module{
		handler:        httphandler.NewHandler(carts),
		createCustomer: createcustomer.NewHandler(customerRepo),
		updateCustomer: updatecustomer.NewHandler(customerRepo),
	}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.handler.RegisterRoutes(mux)
}

// CreateCustomer implements ticketingapi.TicketingAPI.
func (m *Module) CreateCustomer(ctx context.Context, id uuid.UUID, email, firstName, lastName string) error {
	return m.createCustomer.Handle(ctx, createcustomer.Command{
		ID:        id,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	})
}

// UpdateCustomer implements ticketingapi.TicketingAPI.
func (m *Module) UpdateCustomer(ctx context.Context, id uuid.UUID, firstName, lastName string) error {
	return m.updateCustomer.Handle(ctx, updatecustomer.Command{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
	})
}

var _ ticketingapi.TicketingAPI = (*Module)(nil)

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
	log           *slog.Logger
	addItemToCart *additemtocart.Handler
}

func (s *cartServiceImpl) AddItemToCart(ctx context.Context, cmd additemtocart.Command) error {
	done := logHandler(s.log, ctx, "AddItemToCart")
	err := s.addItemToCart.Handle(ctx, cmd)
	done(err)
	return err
}
