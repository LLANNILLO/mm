# Evently — Go

Go implementation of **Evently**, a modular monolith reference application.

The architecture and domain logic are based on Milan Jovanovic's
[Modular Monolith Architecture](https://www.milanjovanovic.tech/modular-monolith-architecture) course,
translated from C# to Go while adapting the architecture to idiomatic Go patterns.

## Stack

- **Go 1.25+** — `net/http` with the native `ServeMux` (method + path pattern matching)
- **PostgreSQL 16** — one schema per module
- **SQLC** — type-safe query generation (no ORM)
- **Goose** — database migrations per module
- **pgx/v5** — PostgreSQL driver + connection pooling
- **Redis** — cart session storage (in-memory cache per customer)
- **Docker / Docker Compose** — local development environment

## Architecture

The project follows a **Modular Monolith** with **Hexagonal Architecture (Ports & Adapters)** and **CQRS** inside each module.

Each module is a self-contained unit with:
- A single public entry point (`module.go`)
- A public `api/` package exposing only integration event types
- An `internal/` directory enforcing compiler-level encapsulation

Nothing outside a module can access its internals — Go's `internal/` package rules guarantee this at compile time.

### Module structure

```
modules/<module>/
├── api/
│   ├── api.go                         ← public API interface (if module is consumed sync)
│   └── integration_events.go          ← cross-module event contracts (async bus)
├── module.go                          ← wiring + DI (only exported entry point)
└── internal/
    ├── domain/                        ← entities, value objects, domain events, business rules
    │
    ├── ports/
    │   ├── inbound/                   ← service interfaces (called by driving adapters)
    │   └── outbound/                  ← repository interfaces (implemented by driven adapters)
    │
    ├── app/
    │   ├── commands/                  ← write side: one package per use case
    │   ├── queries/                   ← read side: one package per use case
    │   ├── consumers/                 ← integration event consumers (bus subscribers)
    │   └── event_handlers/            ← domain event handlers (post-persist side effects)
    │
    └── adapters/
        ├── driving/
        │   └── http/                  ← HTTP handler (calls inbound service interfaces)
        └── driven/
            └── postgres/              ← SQLC queries, Goose migrations, repo implementations
                ├── migrations/
                ├── sqlc.yaml
                ├── query.sql
                └── generated/
```

### Dependency flow

```
adapters/driving/http
        │
        ▼
  ports/inbound          (service interfaces)
        │
        ▼
   app/commands           app/queries
        │
        ▼
      domain              ports/outbound
                                │
                                ▼
                      adapters/driven/postgres
```

### Inter-module communication

Modules communicate through two mechanisms, both defined in `modules/<module>/api/`:

**Synchronous (in-process):** A module exposes an interface in `api/api.go`. Other modules depend on the interface, not the concrete implementation. Used when the caller needs a result inline (e.g. validating a resource exists before writing).

**Asynchronous (event bus):** A module publishes Integration Events to a shared in-memory `EventBus`. Other modules subscribe consumers at startup. Integration event types are defined in `api/integration_events.go` of the publishing module.

```
Users module raises UserRegisteredDomainEvent
        │
        ▼
UserRegisteredDomainEventHandler (users/internal/app/event_handlers/)
        │  publishes
        ▼
EventBus.Publish(UserRegisteredIntegrationEvent)   ← defined in users/api/
        │
        ▼
UserRegisteredConsumer (ticketing/internal/app/consumers/)
        │  calls
        ▼
CreateCustomerCommand → ticketing.customers table
```

The `EventBus` lives in `internal/shared/eventbus/`. It is in-memory and synchronous today — swappable for a real broker (NATS, Redis Streams) without touching domain logic.

### Key design decisions

**CQRS** — commands (write) and queries (read) are strictly separated. Commands go through domain logic and repositories. Queries project directly from the database into read-model DTOs, bypassing the domain entirely.

**Hexagonal ports** — `ports/inbound/` defines what each driving adapter can call. `ports/outbound/` defines what the application needs from the outside world. Neither the domain nor the application layer knows about HTTP or PostgreSQL.

**No ORM** — SQLC generates type-safe Go code from `.sql` files. The generated code lives alongside its migrations inside `adapters/driven/postgres/`.

**Domain events vs. integration events** — Domain events are internal to a module (raised by aggregates, dispatched post-persist via `events.Dispatcher`). Integration events are the module's public async contract — what it announces to the rest of the system. These two are intentionally separate types.

**Prices in minor units** — All monetary values are stored as `int64` (cents). `5000` = $50.00.

## Database

Each module owns its own PostgreSQL schema. Cross-schema queries are prohibited — modules may only read their own schema.

### Schema: `events`

| Table | Description |
|---|---|
| `events.categories` | Event categories |
| `events.events` | Events (draft → published → cancelled) |
| `events.ticket_types` | Ticket types belonging to an event |

### Schema: `users`

| Table | Description |
|---|---|
| `users.users` | Registered users |

### Schema: `ticketing`

| Table | Description |
|---|---|
| `ticketing.customers` | Customer replica (synced from Users via integration event) |
| `ticketing.events` | Event replica (synced from Events — future chapter) |
| `ticketing.ticket_types` | TicketType replica with available quantity tracking |
| `ticketing.orders` | Purchase orders |
| `ticketing.order_items` | Line items per order |
| `ticketing.tickets` | Issued tickets (one per order item unit) |
| `ticketing.payments` | Payments linked to orders |

## Getting started

```bash
# Start infrastructure (Postgres + Redis)
docker compose up evently.database evently.cache -d

# Run migrations for each module
goose -dir modules/events/internal/adapters/driven/postgres/migrations \
  postgres "$DATABASE_URL" up

goose -dir modules/users/internal/adapters/driven/postgres/migrations \
  postgres "$DATABASE_URL" up

goose -dir modules/ticketing/internal/adapters/driven/postgres/migrations \
  postgres "$DATABASE_URL" up

# Regenerate SQLC (only needed after editing query.sql)
sqlc generate --file modules/events/internal/adapters/driven/postgres/sqlc.yaml
sqlc generate --file modules/users/internal/adapters/driven/postgres/sqlc.yaml
sqlc generate --file modules/ticketing/internal/adapters/driven/postgres/sqlc.yaml

# Run the API
DATABASE_URL="postgres://postgres:postgres@localhost:5432/evently?sslmode=disable" \
REDIS_URL="redis://localhost:6379" \
  go run ./cmd/api
```

Or run everything with Docker:

```bash
docker compose up --build
```

The API will be available at `http://localhost:5000`.

## Endpoints

### Users

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/users/register` | Register a new user |
| `GET` | `/users/{id}/profile` | Get user profile |
| `PUT` | `/users/{id}/profile` | Update user profile |

#### `POST /users/register`
```json
{
  "email": "string",
  "first_name": "string",
  "last_name": "string"
}
```
Response `201`: `{ "id": "uuid" }`

#### `GET /users/{id}/profile`
Response `200`:
```json
{
  "id": "uuid",
  "email": "string",
  "first_name": "string",
  "last_name": "string"
}
```

#### `PUT /users/{id}/profile`
```json
{
  "first_name": "string",
  "last_name": "string"
}
```
Response `204`

---

### Events

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/events` | Create a new event |
| `GET` | `/events` | List all events |
| `GET` | `/events/search` | Search events |
| `GET` | `/events/{id}` | Get an event by ID (includes ticket types) |
| `POST` | `/events/{id}/publish` | Publish an event |
| `POST` | `/events/{id}/cancel` | Cancel an event |
| `PUT` | `/events/{id}/reschedule` | Reschedule an event |

#### `POST /events`
```json
{
  "category_id": "uuid",
  "title": "string",
  "description": "string | null",
  "location": "string | null",
  "starts_at_utc": "2025-01-01T10:00:00Z",
  "ends_at_utc": "2025-01-01T18:00:00Z | null"
}
```

#### `GET /events/search`
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | string | no | `draft`, `published`, `cancelled` (default: `published`) |
| `category-id` | uuid | no | Filter by category |

#### `PUT /events/{id}/reschedule`
```json
{
  "starts_at_utc": "2025-06-01T10:00:00Z",
  "ends_at_utc": "2025-06-01T18:00:00Z | null"
}
```

---

### Categories

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/categories` | Create a category |
| `GET` | `/categories` | List all categories |
| `GET` | `/categories/{id}` | Get a category by ID |
| `POST` | `/categories/{id}/archive` | Archive a category |
| `PUT` | `/categories/{id}/name` | Rename a category |

#### `POST /categories`
```json
{ "name": "string" }
```

#### `PUT /categories/{id}/name`
```json
{ "name": "string" }
```

---

### Ticket Types

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/ticket-types` | Create a ticket type for an event |
| `GET` | `/ticket-types` | List ticket types (query param: `event_id`) |
| `GET` | `/ticket-types/{id}` | Get a ticket type by ID |
| `PUT` | `/ticket-types/{id}/price` | Update price |

#### `POST /ticket-types`
```json
{
  "event_id": "uuid",
  "name": "string",
  "price": 5000,
  "currency": "USD",
  "quantity": 100
}
```

> Prices are stored in minor units (cents). `5000` = $50.00.

#### `PUT /ticket-types/{id}/price`
```json
{ "price": 7500 }
```

---

### Carts (Ticketing)

| Method | Path | Description |
|--------|------|-------------|
| `PUT` | `/carts/add` | Add a ticket type to a customer's cart |

#### `PUT /carts/add`
```json
{
  "customer_id": "uuid",
  "ticket_type_id": "uuid",
  "quantity": 2
}
```
Response `200`. The cart is stored in Redis keyed by `customer_id`.

> **Note:** `customer_id` is the same UUID as the user's `id`. A customer record is created automatically when a user registers, via the `UserRegisteredIntegrationEvent` flow.

---

## Error responses

All errors follow a consistent shape:

```json
{ "error": "message", "code": "ERROR_CODE" }
```

| HTTP Status | Meaning |
|-------------|---------|
| `400` | Validation error |
| `404` | Resource not found |
| `409` | Business rule conflict (e.g. publishing an event with no tickets) |
| `500` | Internal server error |

## Credit

Original course and C# reference implementation by
[Milan Jovanovic](https://www.milanjovanovic.tech/modular-monolith-architecture).
