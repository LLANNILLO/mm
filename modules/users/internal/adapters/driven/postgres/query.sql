-- name: InsertUser :one
INSERT INTO users.users (id, email, first_name, last_name)
VALUES (@id, @email, @first_name, @last_name)
RETURNING *;

-- name: SelectUserForUpdate :one
SELECT id, email, first_name, last_name
FROM users.users
WHERE id = @id
FOR UPDATE;

-- name: SelectUserByID :one
SELECT id, email, first_name, last_name
FROM users.users
WHERE id = @id;

-- name: UpdateUserProfile :exec
UPDATE users.users
SET first_name = @first_name,
    last_name  = @last_name
WHERE id = @id;
