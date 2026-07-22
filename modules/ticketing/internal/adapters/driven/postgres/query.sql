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

-- name: DecrementTicketTypeQuantity :one
UPDATE ticketing.ticket_types
SET available_quantity = available_quantity - $1
WHERE id = $2
RETURNING available_quantity;

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

-- name: GetOrderItemsByOrderID :many
SELECT id, order_id, ticket_type_id, quantity, unit_price, price, currency
FROM ticketing.order_items
WHERE order_id = $1;

-- name: UpdateOrderTicketsIssued :exec
UPDATE ticketing.orders
SET tickets_issued = TRUE
WHERE id = $1;

-- Tickets

-- name: InsertTicket :exec
INSERT INTO ticketing.tickets (id, customer_id, order_id, event_id, ticket_type_id, code, created_at_utc, archived, used_at_utc)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: GetTicketByID :one
SELECT id, customer_id, order_id, event_id, ticket_type_id, code, created_at_utc, archived, used_at_utc
FROM ticketing.tickets
WHERE id = $1;

-- name: GetTicketsByEventID :many
SELECT id, customer_id, order_id, event_id, ticket_type_id, code, created_at_utc, archived, used_at_utc
FROM ticketing.tickets
WHERE event_id = $1 AND archived = FALSE;

-- name: UpdateTicketArchived :exec
UPDATE ticketing.tickets
SET archived = TRUE
WHERE id = $1;

-- name: UpdateTicketCheckedIn :exec
UPDATE ticketing.tickets
SET used_at_utc = $1
WHERE id = $2;

-- Event Statistics

-- name: EnsureEventStatisticsRow :exec
INSERT INTO ticketing.event_statistics (event_id)
VALUES ($1)
ON CONFLICT (event_id) DO NOTHING;

-- name: IncrementEventStatisticsTicketsSold :exec
UPDATE ticketing.event_statistics
SET tickets_sold = tickets_sold + 1
WHERE event_id = $1;

-- name: IncrementEventStatisticsAttendeesCheckedIn :exec
UPDATE ticketing.event_statistics
SET attendees_checked_in = attendees_checked_in + 1
WHERE event_id = $1;

-- name: AppendEventStatisticsDuplicateCheckIn :exec
UPDATE ticketing.event_statistics
SET duplicate_check_in_tickets = array_append(duplicate_check_in_tickets, sqlc.arg(ticket_code)::text)
WHERE event_id = sqlc.arg(event_id);

-- name: AppendEventStatisticsInvalidCheckIn :exec
UPDATE ticketing.event_statistics
SET invalid_check_in_tickets = array_append(invalid_check_in_tickets, sqlc.arg(ticket_code)::text)
WHERE event_id = sqlc.arg(event_id);

-- name: GetEventStatisticsByEventID :one
SELECT event_id, tickets_sold, attendees_checked_in, duplicate_check_in_tickets, invalid_check_in_tickets
FROM ticketing.event_statistics
WHERE event_id = $1;

-- Payments

-- name: InsertPayment :exec
INSERT INTO ticketing.payments (id, order_id, transaction_id, amount, currency, amount_refunded, created_at_utc, refunded_at_utc)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetPaymentByID :one
SELECT id, order_id, transaction_id, amount, currency, amount_refunded, created_at_utc, refunded_at_utc
FROM ticketing.payments
WHERE id = $1;

-- name: GetPaymentsByEventID :many
SELECT DISTINCT p.id, p.order_id, p.transaction_id, p.amount, p.currency, p.amount_refunded, p.created_at_utc, p.refunded_at_utc
FROM ticketing.payments p
JOIN ticketing.orders o ON o.id = p.order_id
JOIN ticketing.order_items oi ON oi.order_id = o.id
JOIN ticketing.ticket_types tt ON tt.id = oi.ticket_type_id
WHERE tt.event_id = $1
  AND (p.amount_refunded IS NULL OR p.amount_refunded < p.amount);

-- name: UpdatePaymentRefund :exec
UPDATE ticketing.payments
SET amount_refunded = $1, refunded_at_utc = $2
WHERE id = $3;
