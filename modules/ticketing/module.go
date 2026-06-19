package ticketing

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/llannillo/mm/internal/shared"
	"github.com/llannillo/mm/internal/shared/eventbus"
	ticketingapp "github.com/llannillo/mm/modules/ticketing/internal/app"
	additemtocart "github.com/llannillo/mm/modules/ticketing/internal/app/commands/add_item_to_cart"
	cancelevent "github.com/llannillo/mm/modules/ticketing/internal/app/commands/cancel_event"
	createcustomer "github.com/llannillo/mm/modules/ticketing/internal/app/commands/create_customer"
	createevent "github.com/llannillo/mm/modules/ticketing/internal/app/commands/create_event"
	createorder "github.com/llannillo/mm/modules/ticketing/internal/app/commands/create_order"
	rescheduleevent "github.com/llannillo/mm/modules/ticketing/internal/app/commands/reschedule_event"
	updatecustomer "github.com/llannillo/mm/modules/ticketing/internal/app/commands/update_customer"
	updatetickettypeprice "github.com/llannillo/mm/modules/ticketing/internal/app/commands/update_ticket_type_price"
	"github.com/llannillo/mm/modules/ticketing/internal/app/consumers"
	pg "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres"
	pgstore "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres/generated"
	httphandler "github.com/llannillo/mm/modules/ticketing/internal/adapters/driving/http"
	usersapi "github.com/llannillo/mm/modules/users/api"
)

const moduleName = "ticketing"

type Module struct {
	handler *httphandler.Handler
}

func New(app shared.App) *Module {
	queries := pgstore.New(app.DB)

	customerRepo := pg.NewCustomerRepository(queries)
	eventRepo := pg.NewEventRepository(queries)
	ticketTypeRepo := pg.NewTicketTypeRepository(queries)
	orderRepo := pg.NewOrderRepository(app.DB, queries)
	ticketRepo := pg.NewTicketRepository(queries)
	paymentRepo := pg.NewPaymentRepository(queries)

	_ = ticketRepo
	_ = paymentRepo

	cartService := ticketingapp.NewCartService(app.Cache)

	createCustomerHandler := createcustomer.NewHandler(customerRepo)
	updateCustomerHandler := updatecustomer.NewHandler(customerRepo)

	eventbus.Subscribe[usersapi.UserRegisteredIntegrationEvent](app.EventBus, consumers.NewUserRegisteredConsumer(createCustomerHandler).Handle)
	eventbus.Subscribe[usersapi.UserProfileUpdatedIntegrationEvent](app.EventBus, consumers.NewUserProfileUpdatedConsumer(updateCustomerHandler).Handle)

	_ = createevent.NewHandler(eventRepo, ticketTypeRepo)
	_ = cancelevent.NewHandler(eventRepo)
	_ = rescheduleevent.NewHandler(eventRepo)
	_ = createorder.NewHandler(ticketTypeRepo, orderRepo)
	_ = updatetickettypeprice.NewHandler(ticketTypeRepo)

	carts := &cartServiceImpl{
		log:           app.Logger,
		addItemToCart: additemtocart.NewHandler(cartService, customerRepo, ticketTypeRepo),
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
	log           *slog.Logger
	addItemToCart *additemtocart.Handler
}

func (s *cartServiceImpl) AddItemToCart(ctx context.Context, cmd additemtocart.Command) error {
	done := logHandler(s.log, ctx, "AddItemToCart")
	err := s.addItemToCart.Handle(ctx, cmd)
	done(err)
	return err
}
