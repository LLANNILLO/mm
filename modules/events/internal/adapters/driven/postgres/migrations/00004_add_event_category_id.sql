-- +goose Up
-- +goose StatementBegin
ALTER TABLE events.events 
  ADD COLUMN category_id UUID NOT NULL,
  ADD CONSTRAINT fk_category_event
  FOREIGN KEY (category_id)
  REFERENCES events.categories(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events.events DROP COLUMN category_id;
-- +goose StatementEnd
