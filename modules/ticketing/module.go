package ticketing

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/llannillo/mm/internal/shared"
	"github.com/llannillo/mm/internal/shared/eventbus"
	"github.com/llannillo/mm/internal/shared/outbox"
	pg "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres"
	pgstore "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres/generated"
	httphandler "github.com/llannillo/mm/modules/ticketing/internal/adapters/driving/http"
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
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
	usersintegrationevents "github.com/llannillo/mm/modules/users/api/integrationevents"
)

const moduleName = "ticketing"

const schema = "ticketing"

type Module struct {
	handler      *httphandler.Handler
	outboxWorker *outbox.Worker
}

func New(app shared.App) *Module {
	queries := pgstore.New(app.DB)
	uow := pg.NewUnitOfWork(app.DB)

	customerRepo := pg.NewCustomerRepository(queries)
	eventRepo := pg.NewEventRepository(queries, uow)
	ticketTypeRepo := pg.NewTicketTypeRepository(queries, uow)
	orderRepo := pg.NewOrderRepository(queries, uow)
	ticketRepo := pg.NewTicketRepository(queries, uow)
	paymentRepo := pg.NewPaymentRepository(queries, uow)

	_ = ticketRepo
	_ = paymentRepo

	// No intra-module domain event handlers exist yet for Order/Ticket/Payment
	// flows (order fulfillment isn't wired end-to-end — see create_order,
	// create_event etc. below, also discarded via `_`). The outbox still
	// records and marks these events processed; Dispatch is a documented
	// no-op for types with zero registered handlers. Register the types now
	// so the worker can decode them once handlers land.
	registry := outbox.NewTypeRegistry()
	outbox.RegisterType[domain.EventCancelledDomainEvent](registry)
	outbox.RegisterType[domain.EventRescheduledDomainEvent](registry)
	outbox.RegisterType[domain.EventPaymentsRefundedDomainEvent](registry)
	outbox.RegisterType[domain.EventTicketsArchivedDomainEvent](registry)
	outbox.RegisterType[domain.TicketTypeSoldOutDomainEvent](registry)
	outbox.RegisterType[domain.OrderCreatedDomainEvent](registry)
	outbox.RegisterType[domain.OrderTicketsIssuedDomainEvent](registry)
	outbox.RegisterType[domain.TicketCreatedDomainEvent](registry)
	outbox.RegisterType[domain.TicketArchivedDomainEvent](registry)
	outbox.RegisterType[domain.PaymentCreatedDomainEvent](registry)
	outbox.RegisterType[domain.PaymentRefundedDomainEvent](registry)
	outbox.RegisterType[domain.PaymentPartiallyRefundedDomainEvent](registry)

	outboxWorker := outbox.NewWorker(
		app.DB, schema, moduleName, app.Dispatcher, registry,
		outbox.Config{
			IntervalSeconds: app.Config.Ticketing.Outbox.IntervalSeconds,
			BatchSize:       app.Config.Ticketing.Outbox.BatchSize,
		},
		app.Logger,
	)

	cartService := ticketingapp.NewCartService(app.Cache)

	createCustomerHandler := createcustomer.NewHandler(customerRepo)
	updateCustomerHandler := updatecustomer.NewHandler(customerRepo)

	eventbus.Subscribe[usersintegrationevents.UserRegisteredIntegrationEvent](app.EventBus, consumers.NewUserRegisteredConsumer(createCustomerHandler).Handle)
	eventbus.Subscribe[usersintegrationevents.UserProfileUpdatedIntegrationEvent](app.EventBus, consumers.NewUserProfileUpdatedConsumer(updateCustomerHandler).Handle)

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
		handler:      httphandler.NewHandler(carts),
		outboxWorker: outboxWorker,
	}
}

// RunOutbox polls and dispatches this module's outbox messages until ctx is
// cancelled. Meant to be launched in its own goroutine at startup.
func (m *Module) RunOutbox(ctx context.Context) {
	m.outboxWorker.Run(ctx)
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
