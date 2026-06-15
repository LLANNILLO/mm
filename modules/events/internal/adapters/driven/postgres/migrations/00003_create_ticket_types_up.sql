-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS events.ticket_types (
  id UUID PRIMARY KEY,
  event_id UUID NOT NULL,
  name VARCHAR(100) NOT NULL,
  price BIGINT NOT NULL,
  currency VARCHAR(50) NOT NULL,
  quantity BIGINT NOT NULL,

  CONSTRAINT fk_events_ticket_type
  FOREIGN KEY (event_id)
  REFERENCES events.events(id)
  ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events.ticket_types;
-- +goose StatementEnd
