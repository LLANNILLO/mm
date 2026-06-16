package shared

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Config *Config
	DB     *pgxpool.Pool
	Logger *slog.Logger
}
