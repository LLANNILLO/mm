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

	"github.com/llannillo/mm/internal/bootstrap"
	"github.com/llannillo/mm/internal/shared"
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

	app, err := bootstrap.Build(context.Background(), cfg, logger)
	if err != nil {
		logger.Error("failed to build application", "error", err)
		log.Fatal(err)
	}
	defer app.Close()

	app.RunOutbox(ctx)

	server := &http.Server{Addr: ":8080", Handler: app.Handler}

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
