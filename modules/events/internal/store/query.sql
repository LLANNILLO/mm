-- name: CreateEvent :one
INSERT INTO events.events (id, title, description, location, starts_at_utc, ends_at_utc, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;

-- name: GetEvent :one
SELECT id, title, description, location, starts_at_utc, ends_at_utc
FROM events.events
WHERE id = $1;
