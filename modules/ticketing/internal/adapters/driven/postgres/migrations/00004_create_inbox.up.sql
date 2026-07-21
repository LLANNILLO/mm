-- +goose Up
-- +goose StatementBegin
-- No payload column and no separate inbox_messages table: unlike the C#
-- reference (MassTransit consumer -> inbox row -> background job -> handler),
-- our EventBus.Publish is a synchronous in-process call, so the integration
-- event is already in hand when the consumer runs — nothing to defer or
-- reconstruct later. This table only tracks "have I already run this
-- consumer for this message", keyed by the outbox_message_id that
-- originated the publish (propagated via ctx, see internal/shared/events).
CREATE TABLE IF NOT EXISTS ticketing.inbox_message_consumers (
    message_id UUID          NOT NULL,
    name       VARCHAR(500)  NOT NULL,
    CONSTRAINT pk_ticketing_inbox_message_consumers PRIMARY KEY (message_id, name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ticketing.inbox_message_consumers;
-- +goose StatementEnd
