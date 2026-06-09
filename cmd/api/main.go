package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/llannillo/mm/modules/events"
)

func main() {
	poolCfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	db, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	events.New(db).RegisterRoutes(mux)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
