package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

func main() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := shared.LoadConfig(env, []string{"events", "users", "ticketing"})
	if err != nil {
		log.Fatal(err)
	}

	logger := shared.NewLogger(env, cfg.Logging)

	poolCfg, err := pgxpool.ParseConfig(cfg.Database.DSN)
	if err != nil {
		logger.Error("failed to parse database config", "error", err)
		log.Fatal(err)
	}

	db, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		log.Fatal(err)
	}
	defer db.Close()

	var cacheService cache.Service
	var valkeyClient valkey.Client
	if cfg.Cache.Address != "" {
		valkeyClient, err = valkey.NewClient(valkey.ClientOption{
			InitAddress: []string{cfg.Cache.Address},
		})
		if err != nil {
			logger.Error("failed to connect to Valkey", "error", err)
			log.Fatal(err)
		}
		defer valkeyClient.Close()
		cacheService = cache.NewService(valkeyClient)
		logger.Info("connected to Valkey", "address", cfg.Cache.Address)
	}

	var tokenVerifier *oidc.IDTokenVerifier
	if cfg.Authentication.IssuerURL != "" {
		provider, err := oidc.NewProvider(context.Background(), cfg.Authentication.IssuerURL)
		if err != nil {
			logger.Error("failed to initialize OIDC provider", "issuer", cfg.Authentication.IssuerURL, "error", err)
			log.Fatal(err)
		}
		tokenVerifier = provider.Verifier(&oidc.Config{
				ClientID:        cfg.Authentication.Audience,
				SkipIssuerCheck: true,
			})
		logger.Info("OIDC provider initialized", "issuer", cfg.Authentication.IssuerURL)
	}

	app := shared.App{Config: cfg, DB: db, Logger: logger, Cache: cacheService, Dispatcher: sharedevents.NewDispatcher(), EventBus: eventbus.NewInMemoryEventBus()}

	checkers := map[string]health.Checker{
		"postgres": health.NewPostgresChecker(db),
	}
	if valkeyClient != nil {
		checkers["valkey"] = health.NewValkeyChecker(valkeyClient)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /health", health.NewHandler(checkers))
	eventsModule := events.New(app)
	ticketingModule := ticketing.New(app)
	usersModule := users.New(app)
	eventsModule.RegisterRoutes(mux)
	ticketingModule.RegisterRoutes(mux)
	usersModule.RegisterRoutes(mux)

	go eventsModule.RunOutbox(ctx)
	go ticketingModule.RunOutbox(ctx)
	go usersModule.RunOutbox(ctx)

	handler := middleware.Recovery(logger)(middleware.RequestLogging(logger)(mux))
	if tokenVerifier != nil {
		handler = middleware.Authentication(tokenVerifier, usersModule.PermissionService())(handler)
	}

	server := &http.Server{Addr: ":8080", Handler: handler}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
		}
	}()

	logger.Info("listening on :8080")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
