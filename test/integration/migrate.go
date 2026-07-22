package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// moduleMigrations lists, per module, where its goose migrations live and
// which version table tracks them — mirrors the -dir/-table pairs in
// Taskfile.yml's migrate:* tasks.
var moduleMigrations = []struct {
	dir   string
	table string
}{
	{"../../modules/events/internal/adapters/driven/postgres/migrations", "goose_db_events_versions"},
	{"../../modules/ticketing/internal/adapters/driven/postgres/migrations", "goose_db_ticketing_versions"},
	{"../../modules/users/internal/adapters/driven/postgres/migrations", "goose_db_users_versions"},
}

// migrateAll applies every module's migrations against dsn, using a separate
// goose version table per module so their migration histories don't collide.
func migrateAll(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	for _, m := range moduleMigrations {
		provider, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS(m.dir), goose.WithTableName(m.table))
		if err != nil {
			return fmt.Errorf("new goose provider for %s: %w", m.dir, err)
		}
		if _, err := provider.Up(ctx); err != nil {
			return fmt.Errorf("migrate %s: %w", m.dir, err)
		}
	}
	return nil
}
