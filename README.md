# Evently — Go

Go implementation of **Evently**, a modular monolith reference application.

The architecture and domain logic are based on Milan Jovanovic's
[Modular Monolith Architecture](https://www.milanjovanovic.tech/modular-monolith-architecture) course,
translated from C# to Go while preserving the same structural boundaries and design decisions.

## Stack

- **Go 1.22+** — `net/http` with the native `ServeMux` (method + path pattern matching)
- **PostgreSQL 16** — one schema per module
- **SQLC** — type-safe query generation (no ORM)
- **golang-migrate** — database migrations
- **Docker / Docker Compose** — local development environment

## Architecture

The project follows a **modular monolith** structure. Each module is a self-contained unit:

```
modules/
└── events/
    ├── module.go          ← only exported symbol; public API of the module
    └── internal/
        ├── handler/       ← HTTP handlers
        └── store/         ← SQLC-generated queries + migrations
```

The `internal/` directory inside each module is enforced by the Go compiler —
nothing outside `modules/events/` can import its internals.
`module.go` is the only entry point: it wires dependencies and registers routes.

## Getting started

```bash
# Start Postgres
docker compose up evently.database -d

# Run migrations (adjust connection string as needed)
migrate -path modules/events/internal/store/migrations \
        -database "$DATABASE_URL" up

# Generate SQLC code
cd modules/events/internal/store && sqlc generate

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

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/events` | Create a new event |
| `GET` | `/events/{id}` | Get an event by ID |

## Credit

Original course and C# reference implementation by
[Milan Jovanovic](https://www.milanjovanovic.tech/modular-monolith-architecture).
