-- name: InsertCustomer :exec
INSERT INTO ticketing.customers (id, email, first_name, last_name)
VALUES ($1, $2, $3, $4);

-- name: SelectCustomerByID :one
SELECT id, email, first_name, last_name
FROM ticketing.customers
WHERE id = $1;

-- name: UpdateCustomer :exec
UPDATE ticketing.customers
SET first_name = $1, last_name = $2
WHERE id = $3;
