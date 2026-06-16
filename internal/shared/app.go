package shared

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/llannillo/mm/internal/shared/cache"
)

type App struct {
	Config *Config
	DB     *pgxpool.Pool
	Logger *slog.Logger
	Cache  cache.Service
}
