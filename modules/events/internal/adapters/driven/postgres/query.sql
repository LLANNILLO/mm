-- name: InsertEvent :one
INSERT INTO events.events (id, category_id, title, description, location, starts_at_utc, ends_at_utc, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;

-- name: SCountSearchEvents :one
SELECT COUNT(*)
FROM events.events
WHERE
  status = @status AND
  (sqlc.narg(category_id)::uuid IS NULL OR category_id = sqlc.narg(category_id)::uuid) AND
  (sqlc.narg(starts_date)::timestamptz IS NULL OR starts_at_utc >= sqlc.narg(starts_date)::timestamptz) AND
  (sqlc.narg(ends_date)::timestamptz IS NULL OR ends_at_utc >= sqlc.narg(ends_date)::timestamptz);

-- name: SSearchEvents :many
SELECT id, category_id, title, description, location, starts_at_utc, ends_at_utc
FROM events.events
WHERE
  status = @status AND
  (sqlc.narg(category_id)::uuid IS NULL OR category_id = sqlc.narg(category_id)::uuid) AND
  (sqlc.narg(starts_date)::timestamptz IS NULL OR starts_at_utc >= sqlc.narg(starts_date)::timestamptz) AND
  (sqlc.narg(ends_date)::timestamptz IS NULL OR ends_at_utc >= sqlc.narg(ends_date)::timestamptz)
ORDER BY starts_at_utc
LIMIT @limit_count OFFSET @offset_count;

-- name: SelectEventByID :one
SELECT id, category_id, title, description, location, starts_at_utc, ends_at_utc
FROM events.events
WHERE id = $1;

-- name: SelectEventForUpdate :one
SELECT id, category_id, title, description, location, starts_at_utc, ends_at_utc, status
FROM events.events
WHERE id = $1;

-- name: SelectEvents :many
SELECT id, category_id, title, description, location, starts_at_utc, ends_at_utc
FROM events.events;

-- name: UCancelEvent :exec
UPDATE events.events
SET status = 'cancelled'
WHERE id = $1;

-- name: UPublishEvent :exec
UPDATE events.events
SET status = 'published'
WHERE id = $1;

-- name: URescheduleEvent :exec
UPDATE events.events
SET starts_at_utc = @starts_date, ends_at_utc = @ends_date
WHERE id = $1;

-- name: UpdateEvent :exec
UPDATE events.events
SET
  status = COALESCE(@status::text, status),
  starts_at_utc = COALESCE(@starts_date::timestamptz, starts_at_utc),
  ends_at_utc = COALESCE(@ends_date::timestamptz, ends_at_utc)
WHERE id = $1;



-- name: InsertCategory :one
INSERT INTO events.categories (id, name, is_archived)
VALUES($1, $2, false)
RETURNING id;

-- name: SelectCategories :many
SELECT id, name, is_archived
FROM events.categories;

-- name: SelectCategoryByID :one
SELECT id, name, is_archived
FROM events.categories
WHERE id = $1;

-- name: UpdateCategoryArchived :exec
UPDATE events.categories
SET is_archived = true
WHERE id = $1;

-- name: UpdateCategoryName :exec
UPDATE events.categories
SET name = @name
WHERE id = $1;



-- name: InsertTicketType :one
INSERT INTO events.ticket_types (id, event_id, name, price, currency, quantity)
VALUES($1,$2,$3,$4,$5,$6)
RETURNING id;

-- name: SelectTicketTypesByEventID :many
SELECT id, event_id, name, price, currency, quantity
FROM events.ticket_types
WHERE event_id = $1;

-- name: SExistsTicketTypeByEventID :one
SELECT EXISTS(
  SELECT 1 FROM events.ticket_types
  WHERE event_id = $1
);

-- name: SelectTicketTypeByID :one
SELECT id, event_id, name, price, currency, quantity
FROM events.ticket_types
WHERE id = $1;

-- name: UpdateTicketTypePrice :exec
UPDATE events.ticket_types
SET price = @price
WHERE id = $1;

-- Cancel Event Saga

-- name: StartCancelEventSaga :exec
INSERT INTO events.cancel_event_saga_state (event_id)
VALUES ($1)
ON CONFLICT (event_id) DO NOTHING;

-- name: MarkCancelEventSagaStepComplete :one
UPDATE events.cancel_event_saga_state
SET completed_steps = completed_steps | sqlc.arg(step)::smallint
WHERE event_id = sqlc.arg(event_id)
RETURNING completed_steps;

-- name: DeleteCancelEventSagaState :exec
DELETE FROM events.cancel_event_saga_state WHERE event_id = $1;

