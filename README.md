# Evently — Go

Go implementation of **Evently**, a modular monolith reference application.

The architecture and domain logic are based on Milan Jovanovic's
[Modular Monolith Architecture](https://www.milanjovanovic.tech/modular-monolith-architecture) course,
translated from C# to Go while adapting the architecture to idiomatic Go patterns.

## Stack

- **Go 1.22+** — `net/http` with the native `ServeMux` (method + path pattern matching)
- **PostgreSQL 18** — one schema per module
- **SQLC** — type-safe query generation (no ORM)
- **Goose** — database migrations
- **pgx/v5** — PostgreSQL driver
- **Docker / Docker Compose** — local development environment

## Architecture

The project follows a **Modular Monolith** with **Hexagonal Architecture (Ports & Adapters)** and **CQRS** inside each module.

Each module is a self-contained unit with a single public entry point (`module.go`) and an `events.go` file that exposes integration events for inter-module communication. Nothing outside a module can access its internals — enforced by the Go compiler via the `internal/` directory.

### Module structure

```
modules/events/
├── events.go                          ← public API: integration events for inter-module communication
├── module.go                          ← wiring + service facades (only exported entry point)
└── internal/
    ├── domain/                        ← entities, value objects, domain events, business rules
    │
    ├── ports/
    │   ├── inbound/                   ← service interfaces (EventService, CategoryService, TicketService)
    │   └── outbound/                  ← repository interfaces (EventRepository, CategoryRepository...)
    │
    ├── app/
    │   ├── commands/                  ← write side: one package per use case
    │   │   ├── create_event/
    │   │   ├── publish_event/
    │   │   ├── cancel_event/
    │   │   ├── reschedule_event/
    │   │   ├── create_category/
    │   │   ├── archive_category/
    │   │   ├── rename_category/
    │   │   ├── create_ticket_type/
    │   │   └── update_ticket_price/
    │   └── queries/                   ← read side: one package per use case
    │       ├── get_event/
    │       ├── list_events/
    │       ├── search_events/
    │       ├── get_category/
    │       ├── list_categories/
    │       ├── get_ticket_type/
    │       └── list_ticket_types/
    │
    └── adapters/
        ├── driving/
        │   └── http/                  ← HTTP handler (calls inbound service interfaces)
        └── driven/
            └── postgres/              ← SQLC queries, Goose migrations, repo + reader implementations
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
   app/queries            (reader interfaces defined per query, implemented by adapters)
        │
        ▼
      domain              ports/outbound
                                │
                                ▼
                      adapters/driven/postgres
```

### Key design decisions

**CQRS** — commands (write) and queries (read) are strictly separated. Commands go through domain logic and repositories. Queries bypass the domain entirely, projecting directly from the database into read-model DTOs.

**Hexagonal ports** — `ports/inbound/` defines what each driving adapter (HTTP today, gRPC or CLI tomorrow) can call. `ports/outbound/` defines what the application needs from the outside world. Neither the domain nor the application layer knows about HTTP or PostgreSQL.

**No ORM** — SQLC generates type-safe Go code from `.sql` files. The generated code lives alongside its migrations inside `adapters/driven/postgres/`.

**Consumer-side reader interfaces** — each query handler defines its own minimal reader interface, decoupling query handlers from concrete infrastructure.

**Integration events vs. domain** — `events.go` exposes only integration event types for inter-module communication. The domain model stays private to the module.

## Getting started

```bash
# Start Postgres
docker compose up evently.database -d

# Run migrations
cd modules/events/internal/adapters/driven/postgres
goose postgres "$DATABASE_URL" up

# Regenerate SQLC (only needed after editing query.sql)
sqlc generate

# Run the API
DATABASE_URL="postgres://postgres:postgres@localhost:5432/evently?sslmode=disable" \
  go run ./cmd/api
```

Or run everything with Docker:

```bash
docker compose up --build
```

The API will be available at `http://localhost:5000`.

## Endpoints

### Events

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/events` | Create a new event |
| `GET` | `/events` | List all events |
| `GET` | `/events/search` | Search events (query params: `status`, `category-id`) |
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
| `GET` | `/ticket-types` | List ticket types for an event (query param: `event_id`) |
| `GET` | `/ticket-types/{id}` | Get a ticket type by ID |
| `PUT` | `/ticket-types/{id}/price` | Update a ticket type price |

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

> Prices are stored in cents (e.g. `5000` = $50.00).

#### `PUT /ticket-types/{id}/price`
```json
{ "price": 7500 }
```

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
