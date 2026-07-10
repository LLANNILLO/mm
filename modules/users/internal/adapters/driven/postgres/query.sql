-- name: InsertUser :one
INSERT INTO users.users (id, email, first_name, last_name, identity_id)
VALUES (@id, @email, @first_name, @last_name, @identity_id)
RETURNING *;

-- name: SelectUserForUpdate :one
SELECT id, email, first_name, last_name, identity_id
FROM users.users
WHERE id = @id
FOR UPDATE;

-- name: SelectUserByID :one
SELECT id, email, first_name, last_name, identity_id
FROM users.users
WHERE id = @id;

-- name: UpdateUserProfile :exec
UPDATE users.users
SET first_name = @first_name,
    last_name  = @last_name
WHERE id = @id;

-- name: InsertUserRole :exec
INSERT INTO users.user_roles (user_id, role_name)
VALUES (@user_id, @role_name);
