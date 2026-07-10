package users

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/llannillo/mm/internal/shared"
	"github.com/llannillo/mm/internal/shared/auth"
	"github.com/llannillo/mm/internal/shared/cache"
	sharedevents "github.com/llannillo/mm/internal/shared/events"
	usersapi "github.com/llannillo/mm/modules/users/api"
	kc "github.com/llannillo/mm/modules/users/internal/adapters/driven/keycloak"
	pg "github.com/llannillo/mm/modules/users/internal/adapters/driven/postgres"
	pgstore "github.com/llannillo/mm/modules/users/internal/adapters/driven/postgres/generated"
	httphandler "github.com/llannillo/mm/modules/users/internal/adapters/driving/http"
	registeruser "github.com/llannillo/mm/modules/users/internal/app/commands/register_user"
	updateuser "github.com/llannillo/mm/modules/users/internal/app/commands/update_user"
	eventhandlers "github.com/llannillo/mm/modules/users/internal/app/event_handlers"
	getuser "github.com/llannillo/mm/modules/users/internal/app/queries/get_user"
	getuserperms "github.com/llannillo/mm/modules/users/internal/app/queries/get_user_permissions"
	"github.com/llannillo/mm/modules/users/internal/domain"
	"github.com/llannillo/mm/modules/users/internal/ports/inbound"
)

const moduleName = "users"

const permissionsTTL = 5 * time.Minute

type Module struct {
	handler      *httphandler.Handler
	getUserQuery *getuser.Handler
	permSvc      *permissionService
}

func New(app shared.App) *Module {
	queries := pgstore.New(app.DB)

	userRepo := pg.NewUserRepository(queries, app.Dispatcher)
	userReader := pg.NewUserReader(queries)

	getUserHandler := getuser.NewHandler(userReader)

	sharedevents.Register(app.Dispatcher, eventhandlers.NewUserRegisteredHandler(getUserHandler, app.EventBus).Handle)
	sharedevents.Register(app.Dispatcher, eventhandlers.NewUserProfileUpdatedHandler(app.EventBus).Handle)

	keycloakClient := kc.NewClient(kc.Config{
		AdminURL:                 app.Config.Users.Keycloak.AdminURL,
		TokenURL:                 app.Config.Users.Keycloak.TokenURL,
		ConfidentialClientID:     app.Config.Users.Keycloak.ConfidentialClientID,
		ConfidentialClientSecret: app.Config.Users.Keycloak.ConfidentialClientSecret,
	})

	users := &userService{
		log:          app.Logger,
		registerUser: registeruser.NewHandler(keycloakClient, userRepo),
		updateUser:   updateuser.NewHandler(userRepo),
		getUser:      getUserHandler,
	}

	permSvc := &permissionService{
		handler: getuserperms.NewHandler(pg.NewPermissionsReader(app.DB)),
		cache:   app.Cache,
	}

	return &Module{
		handler:      httphandler.NewHandler(users),
		getUserQuery: getUserHandler,
		permSvc:      permSvc,
	}
}

// PermissionService returns the auth.PermissionService backed by Valkey cache + Postgres.
func (m *Module) PermissionService() auth.PermissionService {
	return m.permSvc
}

// GetUser implements usersapi.UsersAPI — allows other modules to query users.
// Returns nil, nil when the user does not exist.
func (m *Module) GetUser(ctx context.Context, id uuid.UUID) (*usersapi.UserResponse, error) {
	resp, err := m.getUserQuery.Handle(ctx, getuser.Query{UserID: id})
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &usersapi.UserResponse{
		ID:        resp.ID,
		Email:     resp.Email,
		FirstName: resp.FirstName,
		LastName:  resp.LastName,
	}, nil
}

var _ usersapi.UsersAPI = (*Module)(nil)

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

// -- permission service --

type cachedPermissions struct {
	UserID      uuid.UUID `json:"user_id"`
	Permissions []string  `json:"permissions"`
}

type permissionService struct {
	handler *getuserperms.Handler
	cache   cache.Service
}

var _ auth.PermissionService = (*permissionService)(nil)

func (s *permissionService) GetUserPermissions(ctx context.Context, identityID string) (uuid.UUID, []string, error) {
	cacheKey := "permissions:" + identityID

	if s.cache != nil {
		var hit cachedPermissions
		if err := s.cache.Get(ctx, cacheKey, &hit); err == nil {
			return hit.UserID, hit.Permissions, nil
		}
	}

	result, err := s.handler.Handle(ctx, getuserperms.Query{IdentityID: identityID})
	if err != nil {
		return uuid.Nil, nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, cachedPermissions{
			UserID:      result.UserID,
			Permissions: result.Permissions,
		}, permissionsTTL)
	}

	return result.UserID, result.Permissions, nil
}
