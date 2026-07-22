// Package integration boots the real application — real Postgres, real
// Valkey, wired through internal/bootstrap the same way cmd/api does — behind
// an httptest.Server, and drives it over HTTP. Containers are started once in
// TestMain and shared across every test in the package.
//
// Keycloak is intentionally not wired up yet: cfg.Authentication.IssuerURL is
// left empty, so bootstrap.Build skips the OIDC verifier and no auth
// middleware is applied. Endpoints that require a real access token stay out
// of scope until the Keycloak container is added.
package integration

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/llannillo/mm/internal/bootstrap"
	"github.com/llannillo/mm/internal/shared"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	baseURL    string
	httpClient = &http.Client{Timeout: 10 * time.Second}
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:18-alpine",
		postgres.WithDatabase("evently"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		log.Printf("start postgres container: %v", err)
		return 1
	}
	defer pgContainer.Terminate(ctx) //nolint:errcheck

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("postgres connection string: %v", err)
		return 1
	}

	if err := migrateAll(ctx, dsn); err != nil {
		log.Printf("run migrations: %v", err)
		return 1
	}

	redisContainer, err := redis.Run(ctx, "valkey/valkey:9-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp").WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Printf("start valkey container: %v", err)
		return 1
	}
	defer redisContainer.Terminate(ctx) //nolint:errcheck

	redisHost, err := redisContainer.Host(ctx)
	if err != nil {
		log.Printf("valkey host: %v", err)
		return 1
	}
	redisPort, err := redisContainer.MappedPort(ctx, "6379/tcp")
	if err != nil {
		log.Printf("valkey port: %v", err)
		return 1
	}

	cfg := &shared.Config{
		Database:  shared.DatabaseConfig{DSN: dsn},
		Logging:   shared.LoggingConfig{Level: "error"},
		Cache:     shared.CacheConfig{Address: fmt.Sprintf("%s:%s", redisHost, redisPort.Port())},
		Events:    shared.EventsConfig{Outbox: shared.OutboxConfig{IntervalSeconds: 1, BatchSize: 10}},
		Ticketing: shared.TicketingConfig{Outbox: shared.OutboxConfig{IntervalSeconds: 1, BatchSize: 10}},
		Users:     shared.UsersConfig{Outbox: shared.OutboxConfig{IntervalSeconds: 1, BatchSize: 10}},
	}

	logger := shared.NewLogger("test", cfg.Logging)

	app, err := bootstrap.Build(ctx, cfg, logger)
	if err != nil {
		log.Printf("build app: %v", err)
		return 1
	}
	defer app.Close()

	server := httptest.NewServer(app.Handler)
	defer server.Close()

	baseURL = server.URL

	return m.Run()
}
