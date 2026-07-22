-- +goose Up
-- +goose StatementBegin
ALTER TABLE ticketing.tickets ADD COLUMN used_at_utc TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS ticketing.event_statistics (
    event_id                    UUID    PRIMARY KEY,
    tickets_sold                INT     NOT NULL DEFAULT 0,
    attendees_checked_in        INT     NOT NULL DEFAULT 0,
    duplicate_check_in_tickets  TEXT[]  NOT NULL DEFAULT '{}',
    invalid_check_in_tickets    TEXT[]  NOT NULL DEFAULT '{}'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ticketing.event_statistics;
ALTER TABLE ticketing.tickets DROP COLUMN IF EXISTS used_at_utc;
-- +goose StatementEnd
