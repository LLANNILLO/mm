-- +goose Up
-- +goose StatementBegin
ALTER TABLE events.events 
  ADD COLUMN IF NOT EXISTS category_id UUID NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'fk_category_event'
        AND table_name = 'events'
        AND table_schema = 'events'
    ) THEN
        ALTER TABLE events.events
        ADD CONSTRAINT fk_category_event
        FOREIGN KEY (category_id)
        REFERENCES events.categories(id);
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events.events DROP CONSTRAINT IF EXISTS fk_category_event;
ALTER TABLE events.events DROP COLUMN category_id;
-- +goose StatementEnd
