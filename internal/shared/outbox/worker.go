package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/llannillo/mm/internal/shared/events"
)

const (
	defaultInterval  = 10 * time.Second
	defaultBatchSize = 20
)

// Config controls how often a Worker polls and how many messages it claims
// per tick. Zero values fall back to sane defaults.
type Config struct {
	IntervalSeconds int
	BatchSize       int
}

// Worker polls schema.outbox_messages for unprocessed rows and dispatches
// each one through dispatcher. It is the only place domain events get
// dispatched — repositories only write outbox rows, never call the
// dispatcher directly.
type Worker struct {
	pool       *pgxpool.Pool
	schema     string
	moduleName string
	dispatcher *events.Dispatcher
	registry   *TypeRegistry
	interval   time.Duration
	batchSize  int
	logger     *slog.Logger
}

func NewWorker(
	pool *pgxpool.Pool,
	schema, moduleName string,
	dispatcher *events.Dispatcher,
	registry *TypeRegistry,
	cfg Config,
	logger *slog.Logger,
) *Worker {
	interval := time.Duration(cfg.IntervalSeconds) * time.Second
	if interval <= 0 {
		interval = defaultInterval
	}
	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	return &Worker{
		pool:       pool,
		schema:     schema,
		moduleName: moduleName,
		dispatcher: dispatcher,
		registry:   registry,
		interval:   interval,
		batchSize:  batchSize,
		logger:     logger,
	}
}

// Run polls for pending outbox messages every interval until ctx is
// cancelled. Intended to be launched in its own goroutine.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				w.logger.ErrorContext(ctx, "outbox: process batch failed", "module", w.moduleName, "error", err)
			}
		}
	}
}

// ProcessOnce claims and dispatches a single batch immediately, without
// waiting for the next tick. Useful for manually flushing the outbox (ops
// tooling, tests) outside the normal polling cadence.
func (w *Worker) ProcessOnce(ctx context.Context) error {
	return w.processBatch(ctx)
}

type pendingMessage struct {
	ID      uuid.UUID
	Type    string
	Content []byte
}

func (w *Worker) processBatch(ctx context.Context) error {
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	selectSQL := fmt.Sprintf(
		`SELECT id, type, content
		 FROM %s.outbox_messages
		 WHERE processed_on_utc IS NULL
		 ORDER BY occurred_on_utc
		 LIMIT $1
		 FOR UPDATE SKIP LOCKED`,
		w.schema,
	)

	rows, err := tx.Query(ctx, selectSQL, w.batchSize)
	if err != nil {
		return fmt.Errorf("select pending messages: %w", err)
	}

	var messages []pendingMessage
	for rows.Next() {
		var m pendingMessage
		if err := rows.Scan(&m.ID, &m.Type, &m.Content); err != nil {
			rows.Close()
			return fmt.Errorf("scan pending message: %w", err)
		}
		messages = append(messages, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate pending messages: %w", err)
	}

	if len(messages) == 0 {
		return tx.Commit(ctx)
	}

	w.logger.InfoContext(ctx, "outbox: processing messages", "module", w.moduleName, "count", len(messages))

	updateSQL := fmt.Sprintf(
		`UPDATE %s.outbox_messages SET processed_on_utc = $1, error = $2 WHERE id = $3`,
		w.schema,
	)

	for _, m := range messages {
		handleErr := w.handle(ctx, m)
		if handleErr != nil {
			w.logger.ErrorContext(ctx, "outbox: handle message failed",
				"module", w.moduleName, "message_id", m.ID, "type", m.Type, "error", handleErr)
		}

		var errText *string
		if handleErr != nil {
			s := handleErr.Error()
			errText = &s
		}
		if _, err := tx.Exec(ctx, updateSQL, time.Now().UTC(), errText, m.ID); err != nil {
			return fmt.Errorf("mark message %s processed: %w", m.ID, err)
		}
	}

	return tx.Commit(ctx)
}

func (w *Worker) handle(ctx context.Context, m pendingMessage) error {
	domainEvent, err := w.registry.Decode(m.Type, m.Content)
	if err != nil {
		return err
	}
	return w.dispatcher.Dispatch(events.WithMessageID(ctx, m.ID), []events.DomainEvent{domainEvent})
}
