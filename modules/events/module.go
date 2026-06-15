package events

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	archivecategory "github.com/llannillo/mm/modules/events/internal/application/commands/archive_category"
	cancelevent "github.com/llannillo/mm/modules/events/internal/application/commands/cancel_event"
	createcategory "github.com/llannillo/mm/modules/events/internal/application/commands/create_category"
	createevent "github.com/llannillo/mm/modules/events/internal/application/commands/create_event"
	createtickettype "github.com/llannillo/mm/modules/events/internal/application/commands/create_ticket_type"
	publishevent "github.com/llannillo/mm/modules/events/internal/application/commands/publish_event"
	renamecategory "github.com/llannillo/mm/modules/events/internal/application/commands/rename_category"
	rescheduleevent "github.com/llannillo/mm/modules/events/internal/application/commands/reschedule_event"
	updateticketprice "github.com/llannillo/mm/modules/events/internal/application/commands/update_ticket_price"
	getevent "github.com/llannillo/mm/modules/events/internal/application/queries/get-event"
	getcategory "github.com/llannillo/mm/modules/events/internal/application/queries/get_category"
	gettickettype "github.com/llannillo/mm/modules/events/internal/application/queries/get_ticket_type"
	listcategories "github.com/llannillo/mm/modules/events/internal/application/queries/list_categories"
	listevents "github.com/llannillo/mm/modules/events/internal/application/queries/list_events"
	listtickettype "github.com/llannillo/mm/modules/events/internal/application/queries/list_ticket_types"
	searchevents "github.com/llannillo/mm/modules/events/internal/application/queries/search_events"
	"github.com/llannillo/mm/modules/events/internal/domain"
	"github.com/llannillo/mm/modules/events/internal/infrastructure/persistence"
	store "github.com/llannillo/mm/modules/events/internal/infrastructure/store/generated"
	"github.com/llannillo/mm/modules/events/internal/presentation"
)

type Module struct {
	handler *presentation.Handler
}

func New(db *pgxpool.Pool) *Module {
	queries := store.New(db)
	clock := domain.UTCClock{}

	eventRepo := persistence.NewEventRepository(queries)
	categoryRepo := persistence.NewCategoryRepository(queries)
	ticketTypeRepo := persistence.NewTicketTypeRepository(queries)

	eventReader := persistence.NewEventReader(queries)
	categoryReader := persistence.NewCategoryReader(queries)
	ticketTypeReader := persistence.NewTicketTypeReader(queries)

	return &Module{
		handler: presentation.NewHandler(presentation.Deps{
			CreateEvent:     createevent.NewHandler(eventRepo, clock),
			GetEvent:        getevent.NewHandler(eventReader),
			ListEvents:      listevents.NewHandler(eventReader),
			SearchEvents:    searchevents.NewHandler(eventReader),
			PublishEvent:    publishevent.NewHandler(eventRepo, ticketTypeRepo),
			CancelEvent:     cancelevent.NewHandler(eventRepo, clock),
			RescheduleEvent: rescheduleevent.NewHandler(eventRepo),

			CreateCategory:  createcategory.NewHandler(categoryRepo),
			GetCategory:     getcategory.NewHandler(categoryReader),
			ListCategories:  listcategories.NewHandler(categoryReader),
			ArchiveCategory: archivecategory.NewHandler(categoryRepo),
			RenameCategory:  renamecategory.NewHandler(categoryRepo),

			CreateTicketType:  createtickettype.NewHandler(ticketTypeRepo),
			GetTicketType:     gettickettype.NewHandler(ticketTypeReader),
			ListTicketTypes:   listtickettype.NewHandler(ticketTypeReader),
			UpdateTicketPrice: updateticketprice.NewHandler(ticketTypeRepo),
		}),
	}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.handler.RegisterRoutes(mux)
}
