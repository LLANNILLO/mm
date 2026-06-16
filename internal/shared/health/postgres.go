package health

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresChecker struct {
	pool *pgxpool.Pool
}

// NewPostgresChecker returns a Checker that pings the database pool.
func NewPostgresChecker(pool *pgxpool.Pool) Checker {
	return &postgresChecker{pool: pool}
}

func (c *postgresChecker) Check(ctx context.Context) error {
	return c.pool.Ping(ctx)
}
