package shared

import (
	"context"
	"log/slog"
	"os"

	"github.com/llannillo/mm/internal/shared/seq"
	"github.com/lmittmann/tint"
)

func NewLogger(env string, cfg LoggingConfig) *slog.Logger {
	level := parseLevel(cfg.Level)

	var handlers []slog.Handler

	if env == "development" {
		handlers = append(handlers, tint.NewHandler(os.Stderr, &tint.Options{Level: level}))
	} else {
		handlers = append(handlers, slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	}

	if cfg.Seq.Endpoint != "" {
		handlers = append(handlers, seq.NewHandler(cfg.Seq.Endpoint, level))
	}

	return slog.New(newMultiHandler(handlers...))
}

func parseLevel(s string) slog.Level {
	var l slog.Level
	if err := l.UnmarshalText([]byte(s)); err != nil {
		return slog.LevelInfo
	}
	return l
}

type multiHandler []slog.Handler

func newMultiHandler(handlers ...slog.Handler) slog.Handler {
	if len(handlers) == 1 {
		return handlers[0]
	}
	return multiHandler(handlers)
}

func (m multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m {
		if h.Enabled(ctx, r.Level) {
			h.Handle(ctx, r) //nolint:errcheck
		}
	}
	return nil
}

func (m multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make(multiHandler, len(m))
	for i, h := range m {
		handlers[i] = h.WithAttrs(attrs)
	}
	return handlers
}

func (m multiHandler) WithGroup(name string) slog.Handler {
	handlers := make(multiHandler, len(m))
	for i, h := range m {
		handlers[i] = h.WithGroup(name)
	}
	return handlers
}
