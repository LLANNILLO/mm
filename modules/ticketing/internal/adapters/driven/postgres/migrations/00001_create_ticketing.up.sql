-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS ticketing;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SCHEMA IF EXISTS ticketing CASCADE;
-- +goose StatementEnd
