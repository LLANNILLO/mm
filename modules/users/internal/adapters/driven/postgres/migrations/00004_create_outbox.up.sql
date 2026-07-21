-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users.outbox_messages (
    id              UUID        PRIMARY KEY,
    type            TEXT        NOT NULL,
    content         JSONB       NOT NULL,
    occurred_on_utc TIMESTAMPTZ NOT NULL,
    processed_on_utc TIMESTAMPTZ,
    error           TEXT
);

CREATE INDEX IF NOT EXISTS idx_users_outbox_messages_pending
    ON users.outbox_messages (occurred_on_utc)
    WHERE processed_on_utc IS NULL;

-- No FK to outbox_messages(id): the idempotency check/insert runs on a
-- separate pool connection from the worker's batch transaction, which still
-- holds a FOR UPDATE lock on that row — an FK here would force Postgres to
-- take a FOR KEY SHARE lock on the same row to validate it, deadlocking
-- against the worker's own open transaction.
CREATE TABLE IF NOT EXISTS users.outbox_message_consumers (
    outbox_message_id UUID          NOT NULL,
    name              VARCHAR(500)  NOT NULL,
    CONSTRAINT pk_users_outbox_message_consumers PRIMARY KEY (outbox_message_id, name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users.outbox_message_consumers;
DROP TABLE IF EXISTS users.outbox_messages;
-- +goose StatementEnd
