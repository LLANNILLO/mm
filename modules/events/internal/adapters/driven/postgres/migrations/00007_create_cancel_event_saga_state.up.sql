-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS events.cancel_event_saga_state (
    event_id        UUID        PRIMARY KEY,
    completed_steps SMALLINT    NOT NULL DEFAULT 0,
    created_at_utc  TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events.cancel_event_saga_state;
-- +goose StatementEnd
