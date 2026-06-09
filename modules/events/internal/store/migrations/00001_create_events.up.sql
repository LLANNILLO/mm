-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS events;

CREATE TABLE IF NOT EXISTS events.events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title TEXT NOT NULL,
  description TEXT,
  location TEXT,
  starts_at_utc TIMESTAMPTZ NOT NULL,
  ends_at_utc TIMESTAMPTZ NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'draft',

  -- Optional: ensure ends_at is after starts_at
  CONSTRAINT chk_dates CHECK (ends_at_utc > starts_at_utc),

  -- Optional: restrict status to known values
  CONSTRAINT chk_status CHECK (status IN ('draft', 'published', 'cancelled', 'completed'))
);

CREATE INDEX idx_events_status ON events.events (status);
CREATE INDEX idx_events_starts_at ON events.events (starts_at_utc);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events.events;
DROP SCHEMA IF EXISTS events;
-- +goose StatementEnd
