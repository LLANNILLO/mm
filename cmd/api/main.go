package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/llannillo/mm/internal/shared"
	"github.com/llannillo/mm/internal/shared/middleware"
	"github.com/llannillo/mm/modules/events"
)

func main() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	cfg, err := shared.LoadConfig(env, []string{"events"})
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

	app := shared.App{Config: cfg, DB: db, Logger: logger}

	mux := http.NewServeMux()
	events.New(app).RegisterRoutes(mux)

	handler := middleware.Recovery(logger)(middleware.RequestLogging(logger)(mux))

	logger.Info("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
