-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS events.categories (
  id UUID PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  is_archived BOOLEAN NOT NULL DEFAULT false
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events.categories;
-- +goose StatementEnd
