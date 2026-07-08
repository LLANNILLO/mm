-- +goose Up
-- +goose StatementBegin
ALTER TABLE users.users ADD COLUMN identity_id TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users.users DROP COLUMN identity_id;
-- +goose StatementEnd
