package seq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

var client = &http.Client{Timeout: 2 * time.Second}

type Handler struct {
	endpoint string
	level    slog.Level
	attrs    []slog.Attr
}

func NewHandler(endpoint string, level slog.Level) *Handler {
	return &Handler{endpoint: endpoint, level: level}
}

func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	event := make(map[string]any, r.NumAttrs()+len(h.attrs)+3)
	event["@t"] = r.Time.UTC().Format(time.RFC3339Nano)
	event["@mt"] = r.Message
	if r.Level != slog.LevelInfo {
		event["@l"] = seqLevel(r.Level)
	}

	for _, a := range h.attrs {
		event[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		event[a.Key] = a.Value.Any()
		return true
	})

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("seq: marshal event: %w", err)
	}

	resp, err := client.Post(
		h.endpoint+"/api/events/raw?clef",
		"application/vnd.serilog.clef",
		bytes.NewReader(data),
	)
	if err != nil {
		return fmt.Errorf("seq: post event: %w", err)
	}
	resp.Body.Close()
	return nil
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(merged, h.attrs)
	copy(merged[len(h.attrs):], attrs)
	return &Handler{endpoint: h.endpoint, level: h.level, attrs: merged}
}

func (h *Handler) WithGroup(_ string) slog.Handler {
	return h
}

func seqLevel(l slog.Level) string {
	switch {
	case l < slog.LevelInfo:
		return "Debug"
	case l < slog.LevelWarn:
		return "Information"
	case l < slog.LevelError:
		return "Warning"
	default:
		return "Error"
	}
}
