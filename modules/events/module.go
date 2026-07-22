package events

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/llannillo/mm/internal/shared"
	"github.com/llannillo/mm/internal/shared/eventbus"
	sharedevents "github.com/llannillo/mm/internal/shared/events"
	"github.com/llannillo/mm/internal/shared/inbox"
	"github.com/llannillo/mm/internal/shared/outbox"
	eventsapi "github.com/llannillo/mm/modules/events/api"
	"github.com/llannillo/mm/modules/events/api/integrationevents"
	pg "github.com/llannillo/mm/modules/events/internal/adapters/driven/postgres"
	pgstore "github.com/llannillo/mm/modules/events/internal/adapters/driven/postgres/generated"
	httphandler "github.com/llannillo/mm/modules/events/internal/adapters/driving/http"
	archivecategory "github.com/llannillo/mm/modules/events/internal/app/commands/archive_category"
	cancelevent "github.com/llannillo/mm/modules/events/internal/app/commands/cancel_event"
	createcategory "github.com/llannillo/mm/modules/events/internal/app/commands/create_category"
	createevent "github.com/llannillo/mm/modules/events/internal/app/commands/create_event"
	createtickettype "github.com/llannillo/mm/modules/events/internal/app/commands/create_ticket_type"
	publishevent "github.com/llannillo/mm/modules/events/internal/app/commands/publish_event"
	renamecategory "github.com/llannillo/mm/modules/events/internal/app/commands/rename_category"
	rescheduleevent "github.com/llannillo/mm/modules/events/internal/app/commands/reschedule_event"
	updateticketprice "github.com/llannillo/mm/modules/events/internal/app/commands/update_ticket_price"
	eventhandlers "github.com/llannillo/mm/modules/events/internal/app/event_handlers"
	getcategory "github.com/llannillo/mm/modules/events/internal/app/queries/get_category"
	getevent "github.com/llannillo/mm/modules/events/internal/app/queries/get_event"
	gettickettype "github.com/llannillo/mm/modules/events/internal/app/queries/get_ticket_type"
	listcategories "github.com/llannillo/mm/modules/events/internal/app/queries/list_categories"
	listevents "github.com/llannillo/mm/modules/events/internal/app/queries/list_events"
	listtickettype "github.com/llannillo/mm/modules/events/internal/app/queries/list_ticket_types"
	searchevents "github.com/llannillo/mm/modules/events/internal/app/queries/search_events"
	eventcancellation "github.com/llannillo/mm/modules/events/internal/app/sagas/event_cancellation"
	"github.com/llannillo/mm/modules/events/internal/domain"
	ticketingintegrationevents "github.com/llannillo/mm/modules/ticketing/api/integrationevents"
)

const moduleName = "events"

const schema = "events"

type Module struct {
	handler            *httphandler.Handler
	getTicketTypeQuery *gettickettype.Handler
	outboxWorker       *outbox.Worker
}

func New(app shared.App) *Module {
	queries := pgstore.New(app.DB)
	clock := domain.UTCClock{}

	sharedevents.Register(app.Dispatcher, outbox.Idempotent(
		"HandleEventRescheduled", app.DB, schema,
		eventhandlers.HandleEventRescheduled,
	))
	sharedevents.Register(app.Dispatcher, outbox.Idempotent(
		"EventCancelledHandler", app.DB, schema,
		eventhandlers.NewEventCancelledHandler(app.EventBus).Handle,
	))

	registry := outbox.NewTypeRegistry()
	outbox.RegisterType[domain.EventCreatedDomainEvent](registry)
	outbox.RegisterType[domain.EventPublishedDomainEvent](registry)
	outbox.RegisterType[domain.EventCancelledDomainEvent](registry)
	outbox.RegisterType[domain.EventRescheduledDomainEvent](registry)
	outbox.RegisterType[domain.CategoryCreatedDomainEvent](registry)
	outbox.RegisterType[domain.CategoryArchivedDomainEvent](registry)
	outbox.RegisterType[domain.CategoryNameChangedDomainEvent](registry)
	outbox.RegisterType[domain.TicketTypeCreatedDomainEvent](registry)
	outbox.RegisterType[domain.TicketTypePriceChangedDomainEvent](registry)

	outboxWorker := outbox.NewWorker(
		app.DB, schema, moduleName, app.Dispatcher, registry,
		outbox.Config{
			IntervalSeconds: app.Config.Events.Outbox.IntervalSeconds,
			BatchSize:       app.Config.Events.Outbox.BatchSize,
		},
		app.Logger,
	)

	uow := pg.NewUnitOfWork(app.DB)
	eventRepo := pg.NewEventRepository(queries, uow)
	categoryRepo := pg.NewCategoryRepository(queries, uow)
	ticketTypeRepo := pg.NewTicketTypeRepository(queries, uow)

	eventReader := pg.NewEventReader(queries)
	categoryReader := pg.NewCategoryReader(queries)
	ticketTypeReader := pg.NewTicketTypeReader(queries)

	sagaRepo := pg.NewCancelEventSagaRepository(queries)

	eventbus.Subscribe[integrationevents.EventCanceledIntegrationEvent](app.EventBus, inbox.Idempotent(
		"CancelEventSagaStartedHandler", app.DB, schema,
		eventcancellation.NewStartedHandler(sagaRepo, app.EventBus).Handle,
	))
	eventbus.Subscribe[ticketingintegrationevents.EventPaymentsRefundedIntegrationEvent](app.EventBus, inbox.Idempotent(
		"CancelEventSagaPaymentsRefundedHandler", app.DB, schema,
		eventcancellation.NewPaymentsRefundedHandler(sagaRepo, app.EventBus).Handle,
	))
	eventbus.Subscribe[ticketingintegrationevents.EventTicketsArchivedIntegrationEvent](app.EventBus, inbox.Idempotent(
		"CancelEventSagaTicketsArchivedHandler", app.DB, schema,
		eventcancellation.NewTicketsArchivedHandler(sagaRepo, app.EventBus).Handle,
	))

	events := &eventService{
		log:             app.Logger,
		createEvent:     createevent.NewHandler(eventRepo, clock),
		publishEvent:    publishevent.NewHandler(eventRepo, ticketTypeRepo),
		cancelEvent:     cancelevent.NewHandler(eventRepo, clock),
		rescheduleEvent: rescheduleevent.NewHandler(eventRepo),
		getEvent:        getevent.NewHandler(eventReader),
		listEvents:      listevents.NewHandler(eventReader),
		searchEvents:    searchevents.NewHandler(eventReader),
	}

	categories := &categoryService{
		log:             app.Logger,
		createCategory:  createcategory.NewHandler(categoryRepo),
		archiveCategory: archivecategory.NewHandler(categoryRepo),
		renameCategory:  renamecategory.NewHandler(categoryRepo),
		getCategory:     getcategory.NewHandler(categoryReader),
		listCategories:  listcategories.NewHandler(categoryReader),
	}

	getTicketTypeHandler := gettickettype.NewHandler(ticketTypeReader)

	tickets := &ticketService{
		log:               app.Logger,
		createTicketType:  createtickettype.NewHandler(ticketTypeRepo),
		updateTicketPrice: updateticketprice.NewHandler(ticketTypeRepo),
		getTicketType:     getTicketTypeHandler,
		listTicketTypes:   listtickettype.NewHandler(ticketTypeReader),
	}

	return &Module{
		handler:            httphandler.NewHandler(events, categories, tickets),
		getTicketTypeQuery: getTicketTypeHandler,
		outboxWorker:       outboxWorker,
	}
}

// RunOutbox polls and dispatches this module's outbox messages until ctx is
// cancelled. Meant to be launched in its own goroutine at startup.
func (m *Module) RunOutbox(ctx context.Context) {
	m.outboxWorker.Run(ctx)
}

// GetTicketType implements eventsapi.EventsAPI — allows other modules to query ticket types.
// Returns nil, nil when the ticket type does not exist.
func (m *Module) GetTicketType(ctx context.Context, id uuid.UUID) (*eventsapi.TicketTypeResponse, error) {
	resp, err := m.getTicketTypeQuery.Handle(ctx, gettickettype.Query{ID: id})
	if err != nil {
		if errors.Is(err, domain.ErrTicketTypeNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &eventsapi.TicketTypeResponse{
		ID:       resp.ID,
		EventID:  resp.EventID,
		Name:     resp.Name,
		Price:    resp.Price,
		Currency: resp.Currency,
		Quantity: resp.Quantity,
	}, nil
}

var _ eventsapi.EventsAPI = (*Module)(nil)

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.handler.RegisterRoutes(mux)
}

// logHandler logs the start of a request and returns a func to call on completion.
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

// -- event service --

type eventService struct {
	log             *slog.Logger
	createEvent     *createevent.Handler
	publishEvent    *publishevent.Handler
	cancelEvent     *cancelevent.Handler
	rescheduleEvent *rescheduleevent.Handler
	getEvent        *getevent.Handler
	listEvents      *listevents.Handler
	searchEvents    *searchevents.Handler
}

func (s *eventService) CreateEvent(ctx context.Context, cmd createevent.Command) (uuid.UUID, error) {
	done := logHandler(s.log, ctx, "CreateEvent")
	result, err := s.createEvent.Handle(ctx, cmd)
	done(err)
	return result, err
}

func (s *eventService) PublishEvent(ctx context.Context, cmd publishevent.Command) error {
	done := logHandler(s.log, ctx, "PublishEvent")
	err := s.publishEvent.Handle(ctx, cmd)
	done(err)
	return err
}

func (s *eventService) CancelEvent(ctx context.Context, cmd cancelevent.Command) error {
	done := logHandler(s.log, ctx, "CancelEvent")
	err := s.cancelEvent.Handle(ctx, cmd)
	done(err)
	return err
}

func (s *eventService) RescheduleEvent(ctx context.Context, cmd rescheduleevent.Command) error {
	done := logHandler(s.log, ctx, "RescheduleEvent")
	err := s.rescheduleEvent.Handle(ctx, cmd)
	done(err)
	return err
}

func (s *eventService) GetEvent(ctx context.Context, q getevent.Query) (*getevent.Response, error) {
	done := logHandler(s.log, ctx, "GetEvent")
	result, err := s.getEvent.Handle(ctx, q)
	done(err)
	return result, err
}

func (s *eventService) ListEvents(ctx context.Context) ([]listevents.EventItem, error) {
	done := logHandler(s.log, ctx, "ListEvents")
	result, err := s.listEvents.Handle(ctx)
	done(err)
	return result, err
}

func (s *eventService) SearchEvents(ctx context.Context, q searchevents.Query) (*searchevents.Page[searchevents.EventItem], error) {
	done := logHandler(s.log, ctx, "SearchEvents")
	result, err := s.searchEvents.Handle(ctx, q)
	done(err)
	return result, err
}

// -- category service --

type categoryService struct {
	log             *slog.Logger
	createCategory  *createcategory.Handler
	archiveCategory *archivecategory.Handler
	renameCategory  *renamecategory.Handler
	getCategory     *getcategory.Handler
	listCategories  *listcategories.Handler
}

func (s *categoryService) CreateCategory(ctx context.Context, cmd createcategory.Command) (uuid.UUID, error) {
	done := logHandler(s.log, ctx, "CreateCategory")
	result, err := s.createCategory.Handle(ctx, cmd)
	done(err)
	return result, err
}

func (s *categoryService) ArchiveCategory(ctx context.Context, cmd archivecategory.Command) error {
	done := logHandler(s.log, ctx, "ArchiveCategory")
	err := s.archiveCategory.Handle(ctx, cmd)
	done(err)
	return err
}

func (s *categoryService) RenameCategory(ctx context.Context, cmd renamecategory.Command) error {
	done := logHandler(s.log, ctx, "RenameCategory")
	err := s.renameCategory.Handle(ctx, cmd)
	done(err)
	return err
}

func (s *categoryService) GetCategory(ctx context.Context, q getcategory.Query) (*getcategory.Response, error) {
	done := logHandler(s.log, ctx, "GetCategory")
	result, err := s.getCategory.Handle(ctx, q)
	done(err)
	return result, err
}

func (s *categoryService) ListCategories(ctx context.Context) ([]listcategories.CategoryItem, error) {
	done := logHandler(s.log, ctx, "ListCategories")
	result, err := s.listCategories.Handle(ctx)
	done(err)
	return result, err
}

// -- ticket service --

type ticketService struct {
	log               *slog.Logger
	createTicketType  *createtickettype.Handler
	updateTicketPrice *updateticketprice.Handler
	getTicketType     *gettickettype.Handler
	listTicketTypes   *listtickettype.Handler
}

func (s *ticketService) CreateTicketType(ctx context.Context, cmd createtickettype.Command) (uuid.UUID, error) {
	done := logHandler(s.log, ctx, "CreateTicketType")
	result, err := s.createTicketType.Handle(ctx, cmd)
	done(err)
	return result, err
}

func (s *ticketService) UpdateTicketPrice(ctx context.Context, cmd updateticketprice.Command) error {
	done := logHandler(s.log, ctx, "UpdateTicketPrice")
	err := s.updateTicketPrice.Handle(ctx, cmd)
	done(err)
	return err
}

func (s *ticketService) GetTicketType(ctx context.Context, q gettickettype.Query) (*gettickettype.Response, error) {
	done := logHandler(s.log, ctx, "GetTicketType")
	result, err := s.getTicketType.Handle(ctx, q)
	done(err)
	return result, err
}

func (s *ticketService) ListTicketTypes(ctx context.Context, q listtickettype.Query) ([]listtickettype.TicketTypeItem, error) {
	done := logHandler(s.log, ctx, "ListTicketTypes")
	result, err := s.listTicketTypes.Handle(ctx, q)
	done(err)
	return result, err
}
