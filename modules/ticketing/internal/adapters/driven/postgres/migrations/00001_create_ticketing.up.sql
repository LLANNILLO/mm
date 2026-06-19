-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS ticketing;

CREATE TABLE ticketing.customers (
    id         UUID         PRIMARY KEY,
    email      VARCHAR(300) NOT NULL,
    first_name VARCHAR(200) NOT NULL,
    last_name  VARCHAR(200) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SCHEMA IF EXISTS ticketing CASCADE;
-- +goose StatementEnd
