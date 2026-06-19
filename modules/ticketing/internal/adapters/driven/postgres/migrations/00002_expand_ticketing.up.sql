-- +goose Up
-- +goose StatementBegin
CREATE TABLE ticketing.events (
    id            UUID        PRIMARY KEY,
    title         TEXT        NOT NULL,
    description   TEXT,
    location      TEXT,
    starts_at_utc TIMESTAMPTZ NOT NULL,
    ends_at_utc   TIMESTAMPTZ,
    cancelled     BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE TABLE ticketing.ticket_types (
    id                 UUID   PRIMARY KEY,
    event_id           UUID   NOT NULL REFERENCES ticketing.events(id),
    name               TEXT   NOT NULL,
    price              BIGINT NOT NULL,
    currency           TEXT   NOT NULL,
    quantity           BIGINT NOT NULL,
    available_quantity BIGINT NOT NULL
);

CREATE TABLE ticketing.orders (
    id             UUID        PRIMARY KEY,
    customer_id    UUID        NOT NULL REFERENCES ticketing.customers(id),
    status         TEXT        NOT NULL DEFAULT 'pending',
    total_price    BIGINT      NOT NULL DEFAULT 0,
    currency       TEXT        NOT NULL DEFAULT '',
    tickets_issued BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at_utc TIMESTAMPTZ NOT NULL
);

CREATE TABLE ticketing.order_items (
    id             UUID   PRIMARY KEY,
    order_id       UUID   NOT NULL REFERENCES ticketing.orders(id),
    ticket_type_id UUID   NOT NULL REFERENCES ticketing.ticket_types(id),
    quantity       BIGINT NOT NULL,
    unit_price     BIGINT NOT NULL,
    price          BIGINT NOT NULL,
    currency       TEXT   NOT NULL
);

CREATE TABLE ticketing.tickets (
    id             UUID        PRIMARY KEY,
    customer_id    UUID        NOT NULL REFERENCES ticketing.customers(id),
    order_id       UUID        NOT NULL REFERENCES ticketing.orders(id),
    event_id       UUID        NOT NULL REFERENCES ticketing.events(id),
    ticket_type_id UUID        NOT NULL REFERENCES ticketing.ticket_types(id),
    code           TEXT        NOT NULL UNIQUE,
    created_at_utc TIMESTAMPTZ NOT NULL,
    archived       BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE TABLE ticketing.payments (
    id              UUID        PRIMARY KEY,
    order_id        UUID        NOT NULL REFERENCES ticketing.orders(id),
    transaction_id  UUID        NOT NULL,
    amount          BIGINT      NOT NULL,
    currency        TEXT        NOT NULL,
    amount_refunded BIGINT,
    created_at_utc  TIMESTAMPTZ NOT NULL,
    refunded_at_utc TIMESTAMPTZ
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ticketing.payments;
DROP TABLE IF EXISTS ticketing.tickets;
DROP TABLE IF EXISTS ticketing.order_items;
DROP TABLE IF EXISTS ticketing.orders;
DROP TABLE IF EXISTS ticketing.ticket_types;
DROP TABLE IF EXISTS ticketing.events;
-- +goose StatementEnd
