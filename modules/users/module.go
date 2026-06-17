package users

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	sharedevents "github.com/llannillo/mm/internal/shared/events"
	"github.com/llannillo/mm/internal/shared"
	eventhandlers "github.com/llannillo/mm/modules/users/internal/app/event_handlers"
	registeruser "github.com/llannillo/mm/modules/users/internal/app/commands/register_user"
	updateuser "github.com/llannillo/mm/modules/users/internal/app/commands/update_user"
	getuser "github.com/llannillo/mm/modules/users/internal/app/queries/get_user"
	pg "github.com/llannillo/mm/modules/users/internal/adapters/driven/postgres"
	pgstore "github.com/llannillo/mm/modules/users/internal/adapters/driven/postgres/generated"
	httphandler "github.com/llannillo/mm/modules/users/internal/adapters/driving/http"
	"github.com/llannillo/mm/modules/users/internal/ports/inbound"
)

const moduleName = "users"

type Module struct {
	handler *httphandler.Handler
}

func New(app shared.App) *Module {
	queries := pgstore.New(app.DB)

	sharedevents.Register(app.Dispatcher, eventhandlers.HandleUserRegistered)
	sharedevents.Register(app.Dispatcher, eventhandlers.HandleUserProfileUpdated)

	userRepo := pg.NewUserRepository(queries, app.Dispatcher)
	userReader := pg.NewUserReader(queries)

	users := &userService{
		log:          app.Logger,
		registerUser: registeruser.NewHandler(userRepo),
		updateUser:   updateuser.NewHandler(userRepo),
		getUser:      getuser.NewHandler(userReader),
	}

	return &Module{
		handler: httphandler.NewHandler(users),
	}
}

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

// -- user service --

type userService struct {
	log          *slog.Logger
	registerUser *registeruser.Handler
	updateUser   *updateuser.Handler
	getUser      *getuser.Handler
}

var _ inbound.UserService = (*userService)(nil)

func (s *userService) RegisterUser(ctx context.Context, cmd registeruser.Command) (uuid.UUID, error) {
	done := logHandler(s.log, ctx, "RegisterUser")
	result, err := s.registerUser.Handle(ctx, cmd)
	done(err)
	return result, err
}

func (s *userService) GetUser(ctx context.Context, q getuser.Query) (*getuser.Response, error) {
	done := logHandler(s.log, ctx, "GetUser")
	result, err := s.getUser.Handle(ctx, q)
	done(err)
	return result, err
}

func (s *userService) UpdateUser(ctx context.Context, cmd updateuser.Command) error {
	done := logHandler(s.log, ctx, "UpdateUser")
	err := s.updateUser.Handle(ctx, cmd)
	done(err)
	return err
}
