package ticketing

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/llannillo/mm/internal/shared"
	"github.com/llannillo/mm/internal/shared/eventbus"
	sharedevents "github.com/llannillo/mm/internal/shared/events"
	"github.com/llannillo/mm/internal/shared/inbox"
	"github.com/llannillo/mm/internal/shared/outbox"
	"github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/payments"
	pg "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres"
	pgstore "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres/generated"
	httphandler "github.com/llannillo/mm/modules/ticketing/internal/adapters/driving/http"
	ticketingapp "github.com/llannillo/mm/modules/ticketing/internal/app"
	additemtocart "github.com/llannillo/mm/modules/ticketing/internal/app/commands/add_item_to_cart"
	cancelevent "github.com/llannillo/mm/modules/ticketing/internal/app/commands/cancel_event"
	checkinticket "github.com/llannillo/mm/modules/ticketing/internal/app/commands/check_in_ticket"
	createcustomer "github.com/llannillo/mm/modules/ticketing/internal/app/commands/create_customer"
	createevent "github.com/llannillo/mm/modules/ticketing/internal/app/commands/create_event"
	createorder "github.com/llannillo/mm/modules/ticketing/internal/app/commands/create_order"
	rescheduleevent "github.com/llannillo/mm/modules/ticketing/internal/app/commands/reschedule_event"
	updatecustomer "github.com/llannillo/mm/modules/ticketing/internal/app/commands/update_customer"
	updatetickettypeprice "github.com/llannillo/mm/modules/ticketing/internal/app/commands/update_ticket_type_price"
	"github.com/llannillo/mm/modules/ticketing/internal/app/consumers"
	eventhandlers "github.com/llannillo/mm/modules/ticketing/internal/app/event_handlers"
	geteventstatistics "github.com/llannillo/mm/modules/ticketing/internal/app/queries/get_event_statistics"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"

	eventsintegrationevents "github.com/llannillo/mm/modules/events/api/integrationevents"
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
	statsRepo := pg.NewEventStatisticsRepository(queries)
	statsReader := pg.NewEventStatisticsReader(queries)

	_ = paymentRepo

	paymentGateway := payments.NewFakeGateway(app.Logger)

	sharedevents.Register(app.Dispatcher, outbox.Idempotent(
		"OrderCreatedHandler", app.DB, schema,
		eventhandlers.NewOrderCreatedHandler(orderRepo).Handle,
	))
	sharedevents.Register(app.Dispatcher, outbox.Idempotent(
		"ArchiveTicketsHandler", app.DB, schema,
		eventhandlers.NewArchiveTicketsHandler(eventRepo).Handle,
	))
	sharedevents.Register(app.Dispatcher, outbox.Idempotent(
		"RefundPaymentsHandler", app.DB, schema,
		eventhandlers.NewRefundPaymentsHandler(eventRepo).Handle,
	))
	sharedevents.Register(app.Dispatcher, outbox.Idempotent(
		"PaymentRefundedHandler", app.DB, schema,
		eventhandlers.NewPaymentRefundedHandler(paymentGateway).Handle,
	))
	sharedevents.Register(app.Dispatcher, outbox.Idempotent(
		"PaymentPartiallyRefundedHandler", app.DB, schema,
		eventhandlers.NewPaymentPartiallyRefundedHandler(paymentGateway).Handle,
	))
	sharedevents.Register(app.Dispatcher, outbox.Idempotent(
		"TicketCreatedStatisticsHandler", app.DB, schema,
		eventhandlers.NewTicketCreatedStatisticsHandler(statsRepo).Handle,
	))
	sharedevents.Register(app.Dispatcher, outbox.Idempotent(
		"TicketCheckedInStatisticsHandler", app.DB, schema,
		eventhandlers.NewTicketCheckedInStatisticsHandler(statsRepo).Handle,
	))
	sharedevents.Register(app.Dispatcher, outbox.Idempotent(
		"TicketCheckInDuplicateStatisticsHandler", app.DB, schema,
		eventhandlers.NewTicketCheckInDuplicateStatisticsHandler(statsRepo).Handle,
	))
	sharedevents.Register(app.Dispatcher, outbox.Idempotent(
		"TicketCheckInInvalidStatisticsHandler", app.DB, schema,
		eventhandlers.NewTicketCheckInInvalidStatisticsHandler(statsRepo).Handle,
	))
	sharedevents.Register(app.Dispatcher, outbox.Idempotent(
		"PaymentsRefundedIntegrationEventPublisher", app.DB, schema,
		eventhandlers.NewPaymentsRefundedIntegrationEventPublisher(app.EventBus).Handle,
	))
	sharedevents.Register(app.Dispatcher, outbox.Idempotent(
		"TicketsArchivedIntegrationEventPublisher", app.DB, schema,
		eventhandlers.NewTicketsArchivedIntegrationEventPublisher(app.EventBus).Handle,
	))

	// Every domain event type this module can raise must be registered here so
	// the worker can decode it, even ones with no handler today:
	//  - EventRescheduledDomainEvent, TicketTypeSoldOutDomainEvent,
	//    OrderTicketsIssuedDomainEvent, TicketArchivedDomainEvent,
	//    PaymentCreatedDomainEvent: in the C# reference these only republish
	//    as integration events for other modules to consume — nothing
	//    consumes them here, so no handler is registered (Dispatch to zero
	//    handlers is a documented no-op).
	//  - EventCancelledDomainEvent is raised by Event.Cancel(), called from
	//    cancel_event's command handler, now reachable via
	//    EventCancellationStartedConsumer below (the cancel-event saga's
	//    "started" step, published by modules/events). ArchiveTicketsHandler
	//    and RefundPaymentsHandler fire from it, and their own completion
	//    events now feed the two IntegrationEventPublisher handlers above,
	//    closing the loop back to the saga.
	//  - TicketCreatedDomainEvent now also feeds TicketCreatedStatisticsHandler
	//    (the event_statistics materialized view's first in-module consumer).
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
	outbox.RegisterType[domain.TicketCheckedInDomainEvent](registry)
	outbox.RegisterType[domain.TicketCheckInDuplicateDomainEvent](registry)
	outbox.RegisterType[domain.TicketCheckInInvalidDomainEvent](registry)
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

	eventbus.Subscribe[usersintegrationevents.UserRegisteredIntegrationEvent](app.EventBus, inbox.Idempotent(
		"UserRegisteredConsumer", app.DB, schema,
		consumers.NewUserRegisteredConsumer(createCustomerHandler).Handle,
	))
	eventbus.Subscribe[usersintegrationevents.UserProfileUpdatedIntegrationEvent](app.EventBus, inbox.Idempotent(
		"UserProfileUpdatedConsumer", app.DB, schema,
		consumers.NewUserProfileUpdatedConsumer(updateCustomerHandler).Handle,
	))

	cancelEventHandler := cancelevent.NewHandler(eventRepo)
	eventbus.Subscribe[eventsintegrationevents.EventCancellationStartedIntegrationEvent](app.EventBus, inbox.Idempotent(
		"EventCancellationStartedConsumer", app.DB, schema,
		consumers.NewEventCancellationStartedConsumer(cancelEventHandler).Handle,
	))

	// Replica-sync commands — see the registry comment above for why these
	// stay unwired until the EDA phase.
	_ = createevent.NewHandler(eventRepo, ticketTypeRepo)
	_ = rescheduleevent.NewHandler(eventRepo)
	_ = updatetickettypeprice.NewHandler(ticketTypeRepo)

	carts := &cartServiceImpl{
		log:           app.Logger,
		addItemToCart: additemtocart.NewHandler(cartService, customerRepo, ticketTypeRepo),
	}

	orders := &orderServiceImpl{
		log:         app.Logger,
		createOrder: createorder.NewHandler(ticketTypeRepo, orderRepo),
	}

	tickets := &ticketServiceImpl{
		log:                app.Logger,
		checkIn:            checkinticket.NewHandler(ticketRepo),
		getEventStatistics: geteventstatistics.NewHandler(statsReader),
	}

	return &Module{
		handler:      httphandler.NewHandler(carts, orders, tickets),
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

type orderServiceImpl struct {
	log         *slog.Logger
	createOrder *createorder.Handler
}

func (s *orderServiceImpl) CreateOrder(ctx context.Context, cmd createorder.Command) (uuid.UUID, error) {
	done := logHandler(s.log, ctx, "CreateOrder")
	id, err := s.createOrder.Handle(ctx, cmd)
	done(err)
	return id, err
}

type ticketServiceImpl struct {
	log                *slog.Logger
	checkIn            *checkinticket.Handler
	getEventStatistics *geteventstatistics.Handler
}

func (s *ticketServiceImpl) CheckIn(ctx context.Context, cmd checkinticket.Command) error {
	done := logHandler(s.log, ctx, "CheckIn")
	err := s.checkIn.Handle(ctx, cmd)
	done(err)
	return err
}

func (s *ticketServiceImpl) GetEventStatistics(ctx context.Context, q geteventstatistics.Query) (*geteventstatistics.Response, error) {
	done := logHandler(s.log, ctx, "GetEventStatistics")
	resp, err := s.getEventStatistics.Handle(ctx, q)
	done(err)
	return resp, err
}
