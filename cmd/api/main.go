package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/llannillo/mm/internal/shared"
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

	poolCfg, err := pgxpool.ParseConfig(cfg.Database.DSN)
	if err != nil {
		log.Fatal(err)
	}

	db, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := shared.App{Config: cfg, DB: db}

	mux := http.NewServeMux()
	events.New(app).RegisterRoutes(mux)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
