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

-- Events

-- name: InsertEvent :exec
INSERT INTO ticketing.events (id, title, description, location, starts_at_utc, ends_at_utc, cancelled)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetEventByID :one
SELECT id, title, description, location, starts_at_utc, ends_at_utc, cancelled
FROM ticketing.events
WHERE id = $1;

-- name: UpdateEventSchedule :exec
UPDATE ticketing.events
SET starts_at_utc = $1, ends_at_utc = $2
WHERE id = $3;

-- name: CancelEvent :exec
UPDATE ticketing.events
SET cancelled = TRUE
WHERE id = $1;

-- TicketTypes

-- name: InsertTicketType :exec
INSERT INTO ticketing.ticket_types (id, event_id, name, price, currency, quantity, available_quantity)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetTicketTypeByID :one
SELECT id, event_id, name, price, currency, quantity, available_quantity
FROM ticketing.ticket_types
WHERE id = $1;

-- name: UpdateTicketTypePrice :exec
UPDATE ticketing.ticket_types
SET price = $1
WHERE id = $2;

-- name: DecrementTicketTypeQuantity :exec
UPDATE ticketing.ticket_types
SET available_quantity = available_quantity - $1
WHERE id = $2;

-- Orders

-- name: InsertOrder :exec
INSERT INTO ticketing.orders (id, customer_id, status, total_price, currency, tickets_issued, created_at_utc)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: InsertOrderItem :exec
INSERT INTO ticketing.order_items (id, order_id, ticket_type_id, quantity, unit_price, price, currency)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetOrderByID :one
SELECT id, customer_id, status, total_price, currency, tickets_issued, created_at_utc
FROM ticketing.orders
WHERE id = $1;

-- Tickets

-- name: InsertTicket :exec
INSERT INTO ticketing.tickets (id, customer_id, order_id, event_id, ticket_type_id, code, created_at_utc, archived)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetTicketByID :one
SELECT id, customer_id, order_id, event_id, ticket_type_id, code, created_at_utc, archived
FROM ticketing.tickets
WHERE id = $1;

-- Payments

-- name: InsertPayment :exec
INSERT INTO ticketing.payments (id, order_id, transaction_id, amount, currency, amount_refunded, created_at_utc, refunded_at_utc)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetPaymentByID :one
SELECT id, order_id, transaction_id, amount, currency, amount_refunded, created_at_utc, refunded_at_utc
FROM ticketing.payments
WHERE id = $1;
