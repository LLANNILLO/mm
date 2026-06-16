package events

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	archivecategory "github.com/llannillo/mm/modules/events/internal/app/commands/archive_category"
	cancelevent "github.com/llannillo/mm/modules/events/internal/app/commands/cancel_event"
	createcategory "github.com/llannillo/mm/modules/events/internal/app/commands/create_category"
	createevent "github.com/llannillo/mm/modules/events/internal/app/commands/create_event"
	createtickettype "github.com/llannillo/mm/modules/events/internal/app/commands/create_ticket_type"
	publishevent "github.com/llannillo/mm/modules/events/internal/app/commands/publish_event"
	renamecategory "github.com/llannillo/mm/modules/events/internal/app/commands/rename_category"
	rescheduleevent "github.com/llannillo/mm/modules/events/internal/app/commands/reschedule_event"
	updateticketprice "github.com/llannillo/mm/modules/events/internal/app/commands/update_ticket_price"
	getcategory "github.com/llannillo/mm/modules/events/internal/app/queries/get_category"
	getevent "github.com/llannillo/mm/modules/events/internal/app/queries/get_event"
	gettickettype "github.com/llannillo/mm/modules/events/internal/app/queries/get_ticket_type"
	listcategories "github.com/llannillo/mm/modules/events/internal/app/queries/list_categories"
	listevents "github.com/llannillo/mm/modules/events/internal/app/queries/list_events"
	listtickettype "github.com/llannillo/mm/modules/events/internal/app/queries/list_ticket_types"
	searchevents "github.com/llannillo/mm/modules/events/internal/app/queries/search_events"
	pg "github.com/llannillo/mm/modules/events/internal/adapters/driven/postgres"
	pgstore "github.com/llannillo/mm/modules/events/internal/adapters/driven/postgres/generated"
	httphandler "github.com/llannillo/mm/modules/events/internal/adapters/driving/http"
	"github.com/llannillo/mm/modules/events/internal/domain"
	"github.com/llannillo/mm/internal/shared"
)

type Module struct {
	handler *httphandler.Handler
}

func New(app shared.App) *Module {
	queries := pgstore.New(app.DB)
	clock := domain.UTCClock{}

	eventRepo := pg.NewEventRepository(queries)
	categoryRepo := pg.NewCategoryRepository(queries)
	ticketTypeRepo := pg.NewTicketTypeRepository(queries)

	eventReader := pg.NewEventReader(queries)
	categoryReader := pg.NewCategoryReader(queries)
	ticketTypeReader := pg.NewTicketTypeReader(queries)

	events := &eventService{
		createEvent:     createevent.NewHandler(eventRepo, clock),
		publishEvent:    publishevent.NewHandler(eventRepo, ticketTypeRepo),
		cancelEvent:     cancelevent.NewHandler(eventRepo, clock),
		rescheduleEvent: rescheduleevent.NewHandler(eventRepo),
		getEvent:        getevent.NewHandler(eventReader),
		listEvents:      listevents.NewHandler(eventReader),
		searchEvents:    searchevents.NewHandler(eventReader),
	}

	categories := &categoryService{
		createCategory:  createcategory.NewHandler(categoryRepo),
		archiveCategory: archivecategory.NewHandler(categoryRepo),
		renameCategory:  renamecategory.NewHandler(categoryRepo),
		getCategory:     getcategory.NewHandler(categoryReader),
		listCategories:  listcategories.NewHandler(categoryReader),
	}

	tickets := &ticketService{
		createTicketType:  createtickettype.NewHandler(ticketTypeRepo),
		updateTicketPrice: updateticketprice.NewHandler(ticketTypeRepo),
		getTicketType:     gettickettype.NewHandler(ticketTypeReader),
		listTicketTypes:   listtickettype.NewHandler(ticketTypeReader),
	}

	return &Module{
		handler: httphandler.NewHandler(events, categories, tickets),
	}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.handler.RegisterRoutes(mux)
}

// -- service facades --

type eventService struct {
	createEvent     *createevent.Handler
	publishEvent    *publishevent.Handler
	cancelEvent     *cancelevent.Handler
	rescheduleEvent *rescheduleevent.Handler
	getEvent        *getevent.Handler
	listEvents      *listevents.Handler
	searchEvents    *searchevents.Handler
}

func (s *eventService) CreateEvent(ctx context.Context, cmd createevent.Command) (uuid.UUID, error) {
	return s.createEvent.Handle(ctx, cmd)
}

func (s *eventService) PublishEvent(ctx context.Context, cmd publishevent.Command) error {
	return s.publishEvent.Handle(ctx, cmd)
}

func (s *eventService) CancelEvent(ctx context.Context, cmd cancelevent.Command) error {
	return s.cancelEvent.Handle(ctx, cmd)
}

func (s *eventService) RescheduleEvent(ctx context.Context, cmd rescheduleevent.Command) error {
	return s.rescheduleEvent.Handle(ctx, cmd)
}

func (s *eventService) GetEvent(ctx context.Context, q getevent.Query) (*getevent.Response, error) {
	return s.getEvent.Handle(ctx, q)
}

func (s *eventService) ListEvents(ctx context.Context) ([]listevents.EventItem, error) {
	return s.listEvents.Handle(ctx)
}

func (s *eventService) SearchEvents(ctx context.Context, q searchevents.Query) (*searchevents.Page[searchevents.EventItem], error) {
	return s.searchEvents.Handle(ctx, q)
}

type categoryService struct {
	createCategory  *createcategory.Handler
	archiveCategory *archivecategory.Handler
	renameCategory  *renamecategory.Handler
	getCategory     *getcategory.Handler
	listCategories  *listcategories.Handler
}

func (s *categoryService) CreateCategory(ctx context.Context, cmd createcategory.Command) (uuid.UUID, error) {
	return s.createCategory.Handle(ctx, cmd)
}

func (s *categoryService) ArchiveCategory(ctx context.Context, cmd archivecategory.Command) error {
	return s.archiveCategory.Handle(ctx, cmd)
}

func (s *categoryService) RenameCategory(ctx context.Context, cmd renamecategory.Command) error {
	return s.renameCategory.Handle(ctx, cmd)
}

func (s *categoryService) GetCategory(ctx context.Context, q getcategory.Query) (*getcategory.Response, error) {
	return s.getCategory.Handle(ctx, q)
}

func (s *categoryService) ListCategories(ctx context.Context) ([]listcategories.CategoryItem, error) {
	return s.listCategories.Handle(ctx)
}

type ticketService struct {
	createTicketType  *createtickettype.Handler
	updateTicketPrice *updateticketprice.Handler
	getTicketType     *gettickettype.Handler
	listTicketTypes   *listtickettype.Handler
}

func (s *ticketService) CreateTicketType(ctx context.Context, cmd createtickettype.Command) (uuid.UUID, error) {
	return s.createTicketType.Handle(ctx, cmd)
}

func (s *ticketService) UpdateTicketPrice(ctx context.Context, cmd updateticketprice.Command) error {
	return s.updateTicketPrice.Handle(ctx, cmd)
}

func (s *ticketService) GetTicketType(ctx context.Context, q gettickettype.Query) (*gettickettype.Response, error) {
	return s.getTicketType.Handle(ctx, q)
}

func (s *ticketService) ListTicketTypes(ctx context.Context, q listtickettype.Query) ([]listtickettype.TicketTypeItem, error) {
	return s.listTicketTypes.Handle(ctx, q)
}
