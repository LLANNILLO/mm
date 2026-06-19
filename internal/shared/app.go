package shared

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/llannillo/mm/internal/shared/cache"
	"github.com/llannillo/mm/internal/shared/eventbus"
	"github.com/llannillo/mm/internal/shared/events"
)

type App struct {
	Config     *Config
	DB         *pgxpool.Pool
	Logger     *slog.Logger
	Cache      cache.Service
	Dispatcher *events.Dispatcher
	EventBus   eventbus.EventBus
}
