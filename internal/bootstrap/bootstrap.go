// Package bootstrap wires the full application — database, cache, auth, all
// modules, and the HTTP middleware chain — from a shared.Config. cmd/api uses
// it at startup; the integration test harness uses it to boot the real app
// against testcontainers-backed infrastructure, so production and tests never
// drift apart.
package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/llannillo/mm/internal/shared"
	"github.com/llannillo/mm/internal/shared/cache"
	"github.com/llannillo/mm/internal/shared/eventbus"
	sharedevents "github.com/llannillo/mm/internal/shared/events"
	"github.com/llannillo/mm/internal/shared/health"
	"github.com/llannillo/mm/internal/shared/middleware"
	"github.com/llannillo/mm/modules/events"
	"github.com/llannillo/mm/modules/ticketing"
	"github.com/llannillo/mm/modules/users"
	"github.com/valkey-io/valkey-go"
)

// App is the fully wired application.
type App struct {
	Handler http.Handler
	DB      *pgxpool.Pool
	Valkey  valkey.Client

	runOutbox func(ctx context.Context)
}

// RunOutbox starts every module's outbox worker in its own goroutine.
// Non-blocking; workers run until ctx is cancelled.
func (a *App) RunOutbox(ctx context.Context) {
	a.runOutbox(ctx)
}

// Close releases infrastructure connections.
func (a *App) Close() {
	a.DB.Close()
	if a.Valkey != nil {
		a.Valkey.Close()
	}
}

// Build connects to infrastructure and wires every module into a single HTTP
// handler.
func Build(ctx context.Context, cfg *shared.Config, logger *slog.Logger) (*App, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	db, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	var cacheService cache.Service
	var valkeyClient valkey.Client
	if cfg.Cache.Address != "" {
		valkeyClient, err = valkey.NewClient(valkey.ClientOption{
			InitAddress: []string{cfg.Cache.Address},
		})
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("connect to Valkey: %w", err)
		}
		cacheService = cache.NewService(valkeyClient)
		logger.Info("connected to Valkey", "address", cfg.Cache.Address)
	}

	var tokenVerifier *oidc.IDTokenVerifier
	if cfg.Authentication.IssuerURL != "" {
		provider, err := oidc.NewProvider(ctx, cfg.Authentication.IssuerURL)
		if err != nil {
			db.Close()
			if valkeyClient != nil {
				valkeyClient.Close()
			}
			return nil, fmt.Errorf("initialize OIDC provider: %w", err)
		}
		tokenVerifier = provider.Verifier(&oidc.Config{
			ClientID:        cfg.Authentication.Audience,
			SkipIssuerCheck: true,
		})
		logger.Info("OIDC provider initialized", "issuer", cfg.Authentication.IssuerURL)
	}

	sharedApp := shared.App{
		Config:     cfg,
		DB:         db,
		Logger:     logger,
		Cache:      cacheService,
		Dispatcher: sharedevents.NewDispatcher(),
		EventBus:   eventbus.NewInMemoryEventBus(),
	}

	checkers := map[string]health.Checker{
		"postgres": health.NewPostgresChecker(db),
	}
	if valkeyClient != nil {
		checkers["valkey"] = health.NewValkeyChecker(valkeyClient)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /health", health.NewHandler(checkers))

	eventsModule := events.New(sharedApp)
	ticketingModule := ticketing.New(sharedApp)
	usersModule := users.New(sharedApp)
	eventsModule.RegisterRoutes(mux)
	ticketingModule.RegisterRoutes(mux)
	usersModule.RegisterRoutes(mux)

	var handler http.Handler = mux
	handler = middleware.Recovery(logger)(middleware.RequestLogging(logger)(handler))
	if tokenVerifier != nil {
		handler = middleware.Authentication(tokenVerifier, usersModule.PermissionService())(handler)
	}

	return &App{
		Handler: handler,
		DB:      db,
		Valkey:  valkeyClient,
		runOutbox: func(ctx context.Context) {
			go eventsModule.RunOutbox(ctx)
			go ticketingModule.RunOutbox(ctx)
			go usersModule.RunOutbox(ctx)
		},
	}, nil
}
