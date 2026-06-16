package shared

import "github.com/jackc/pgx/v5/pgxpool"

type App struct {
	Config *Config
	DB     *pgxpool.Pool
}
