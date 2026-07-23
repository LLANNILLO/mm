# Evently — Go

Go implementation of **Evently**, a modular monolith reference application.

The architecture and domain logic are based on Milan Jovanovic's
[Modular Monolith Architecture](https://www.milanjovanovic.tech/modular-monolith-architecture) course,
translated from C# to Go while adapting the architecture to idiomatic Go patterns.

## Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25+ |
| HTTP | `net/http` with native `ServeMux` (method + path patterns) |
| Database | PostgreSQL 16 — one schema per module |
| Query generation | SQLC — type-safe, no ORM |
| Migrations | Goose — per-module migration directories |
| PG driver | pgx/v5 — driver + connection pooling |
| Cache | Valkey — permission caching + cart session storage |
| Identity provider | Keycloak — JWT issuance, user registration |
| JWT validation | go-oidc/v3 — stateless OIDC discovery + JWKS validation |
| Containers | Docker / Docker Compose |

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
│   ├── api.go                         ← public API interface (sync — kept for cases that need an inline result)
│   └── integrationevents/
│       └── integration_events.go      ← cross-module event contracts (async bus) — the only cross-module dependency an architecture test allows
├── module.go                          ← wiring + DI (only exported entry point)
└── internal/
    ├── domain/                        ← entities, value objects, domain events, business rules
    │
    ├── ports/
    │   ├── inbound/                   ← service interfaces (called by driving adapters)
    │   └── outbound/                  ← repository interfaces + identity provider
    │
    ├── app/
    │   ├── commands/                  ← write side: one package per use case
    │   ├── queries/                   ← read side: one package per use case
    │   ├── consumers/                 ← integration event consumers (bus subscribers)
    │   └── event_handlers/            ← domain event handlers (run later, off the outbox — see Reliable messaging)
    │
    └── adapters/
        ├── driving/
        │   └── http/                  ← HTTP handler (calls inbound service interfaces)
        └── driven/
            ├── postgres/              ← SQLC queries, Goose migrations, repo implementations
            │   ├── migrations/
            │   ├── sqlc.yaml
            │   ├── query.sql
            │   └── generated/
            └── keycloak/              ← Keycloak Admin API client (raw HTTP)
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
                  adapters/driven/keycloak
```

### Inter-module communication

Modules communicate through two mechanisms, both defined in `modules/<module>/api/`:

**Synchronous (in-process):** A module exposes an interface in `api/api.go`. Other modules depend on the interface, not the concrete implementation. Used when the caller needs a result inline (e.g. validating a resource exists before writing). In practice, no module calls another module's sync API today — cross-module reads (`UsersAPI.GetUser`, `EventsAPI.GetTicketType`) were only ever self-implemented and never invoked externally, a leftover from before the project moved to the event bus. The interface is kept for the day a real inline cross-module call is needed, but the architecture tests forbid any *other* module from depending on it.

**Asynchronous (event bus):** A module publishes Integration Events to a shared in-memory `EventBus`. Other modules subscribe consumers at startup. Integration event types live in their own `api/integrationevents/` package — kept separate from `api/api.go` specifically so the architecture tests can allow-list "depend on the async contract" while still forbidding "depend on the sync interface" at the package level.

Events publishes six integration events today: `EventCreatedIntegrationEvent`, `EventRescheduledIntegrationEvent`, `EventCanceledIntegrationEvent`, `EventCancellationStartedIntegrationEvent`, `EventCancellationCompletedIntegrationEvent`, `TicketTypeCreatedIntegrationEvent`, and `TicketTypePriceChangedIntegrationEvent`. Ticketing consumes the creation/reschedule/price ones to keep its own `Event`/`TicketType` replica tables in sync — see [Database](#database) — and the cancellation ones as part of the [cancel-event saga](#saga-pattern-cancel-event-orchestration). Users publishes `UserRegisteredIntegrationEvent` and `UserProfileUpdatedIntegrationEvent`, consumed by Ticketing to keep its `Customer` replica in sync.

**Enforced by tests:** `TestModuleIsolation_NoModuleDependsOnAnotherModule` (in `test/architecture/`) fails the build if any module imports another module's `internal/` packages or its `api` package directly — the only cross-module import allowed is `api/integrationevents`.

The full path from a domain event to a cross-module side effect is asynchronous end to end — nothing in this chain runs inside the original HTTP request:

```
UserRepository.Insert(user)                              (Users, HTTP request)
        │  same transaction
        ▼
users.outbox_messages row (UserRegisteredDomainEvent)
        │
        ▼  outbox.Worker ticks, later, out of band
Idempotent("UserRegisteredHandler") → UserRegisteredDomainEventHandler
        │  publishes
        ▼
EventBus.Publish(UserRegisteredIntegrationEvent)   ← defined in users/api/
        │
        ▼
Idempotent("UserRegisteredConsumer") → UserRegisteredConsumer (ticketing/internal/app/consumers/)
        │  calls
        ▼
CreateCustomerCommand → ticketing.customers table
```

The `EventBus` lives in `internal/shared/eventbus/`. It is in-memory today — swappable for a real broker (NATS, Redis Streams) without touching domain logic. Every hop shown as `Idempotent(...)` is described in [Reliable messaging](#reliable-messaging-outbox-idempotent-consumer--inbox).

### Key design decisions

**CQRS** — commands (write) and queries (read) are strictly separated. Commands go through domain logic and repositories. Queries project directly from the database into read-model DTOs, bypassing the domain entirely.

**Hexagonal ports** — `ports/inbound/` defines what each driving adapter can call. `ports/outbound/` defines what the application needs from the outside world. Neither the domain nor the application layer knows about HTTP or PostgreSQL.

**No ORM** — SQLC generates type-safe Go code from `.sql` files. The generated code lives alongside its migrations inside `adapters/driven/postgres/`.

**Domain events vs. integration events** — Domain events are internal to a module (raised by aggregates, persisted to the outbox in the same transaction as the aggregate, dispatched later by a background worker via `events.Dispatcher`). Integration events are the module's public async contract, published from inside a domain event handler. These two are intentionally separate types. See [Reliable messaging](#reliable-messaging-outbox-idempotent-consumer--inbox) below for why dispatch is asynchronous and how retries stay safe.

**Encapsulated entities** — Every aggregate (`Event`, `Order`, `User`, ...) has unexported fields, exposed only through getter methods. There is no way to construct one outside its own package except through an exported `New<Type>` (enforces invariants) or `Rehydrate<Type>` (used only by the repository to reconstruct persisted state, no invariant checks, no domain events raised). This is Go's answer to C#'s "entities may only have a private constructor" rule — Go has no constructors, so the equivalent guarantee comes from field visibility instead.

**Prices in minor units** — All monetary values are stored as `int64` (cents). `5000` = $50.00.

---

## Testing

Four tiers, from fastest/most isolated to slowest/most end-to-end. `task lint`/`task test`/`task test:integration` (see [Taskfile.yml](Taskfile.yml)) are the source of truth for how each tier actually runs — CI calls the exact same tasks, so "works on my machine" and "works in CI" can't drift apart.

| Tier | Location | Needs Docker? | Command |
|---|---|---|---|
| Unit | `modules/*/internal/domain/*_test.go` | No | `go test ./modules/... -race -cover` |
| Architecture | `test/architecture/` | No | `go test ./test/architecture/...` |
| Integration | `test/integration/` | Yes (Postgres, Valkey, Keycloak) | `go test ./test/integration/... -timeout 5m` |
| System integration | `test/integration/system_flows_test.go` | Yes (same harness) | included in the command above |

### Unit tests

White-box, colocated with the code they test (`package domain`, not `domain_test`) — Go convention, and it lets tests reach unexported fields when useful. Every aggregate (`Event`, `Category`, `TicketType`, `Order`, `Payment`, `Ticket`, `Customer`, `User`, ...) has a `<type>_test.go` covering its invariants and the domain events it raises, using a small generic helper (`assertDomainEventPublished[T]` / `assertNoDomainEventPublished[T]`, one per module's `domain_test_helpers_test.go`) instead of hand-rolling event assertions per test. `testify/assert` + `testify/require` for assertions; table-driven tests for pure validation logic. No Bogus-style fake-data library — the domain types here don't need one.

```bash
go test ./modules/... -v          # verbose, see every test name
go test ./modules/... -cover      # coverage per package
```

### Architecture tests

`test/architecture/` is the Go port of the C# course's `Evently.ArchitectureTests` + `Evently.IntegrationTests`' `ModuleTests.cs` combined (this project has no MSBuild-style per-module test assemblies, so there's no need to split them the way C# does). Go has no runtime IL to reflect over, so the suite is built on `go/packages` + `go/types` instead — it parses and type-checks every package in the module and inspects the resulting import graph and type declarations statically.

```bash
go test ./test/architecture/...
```

What it enforces, one file per concern:

| File | Enforces |
|---|---|
| `module_isolation_test.go` | A module may depend on another module's `api/integrationevents` package only — never its `internal/` packages, never its synchronous `api` package. Checked in both directions for Events ↔ Ticketing, the one bidirectional dependency in this codebase (see [Saga Pattern](#saga-pattern-cancel-event-orchestration)). |
| `layers_test.go` | Hexagonal dependency direction within a module: `domain` has zero internal dependencies; `app` depends on `domain`/`ports` but never `adapters`; `adapters/driving` and `adapters/driven` never depend on each other. A second check denies importing concrete infra packages (`pgx`, `pgxpool`, `database/sql`, `valkey-go`) straight from `app`/`domain`, since a local-path check alone can't see a leak coming through a third-party driver. |
| `domain_test.go` | Every entity has zero exported fields plus an exported `New<Type>`/`Rehydrate<Type>` pair; every type implementing `events.DomainEvent` is named `*DomainEvent`. |
| `application_test.go` | Every `commands/*`/`queries/*` package exposes `Command`/`Query` + `Handler` + `NewHandler`; `Handler` has no exported fields; any `Validate` method is exactly `func() error`. |
| `presentation_test.go` | Any type that subscribes to the event bus (has a `Handle` method, lives in an `app/consumers` package) is named `*Consumer`. |

Some C# NetArchTest rules have no Go equivalent and are intentionally not ported: "sealed" doesn't exist because Go structs cannot be subclassed at all, and Go's own `internal/` visibility rule already makes cross-module internal imports a compile error — the test suite only needs to guard what the compiler *can't* see (the public `api` surface, and layer leaks through third-party packages).

### Integration tests

`test/integration/` boots the **real** application — real Postgres, real Valkey, real Keycloak — behind an `httptest.Server`, and drives it purely over HTTP (no direct handler access, unlike C#'s MediatR `ISender`: this module's services aren't exposed outside their own package, so HTTP is the only seam available, which also means these tests exercise routing and auth middleware for free). `TestMain` starts all three containers once and shares them across every test in the package.

There is no official `testcontainers-go` module for Keycloak (unlike Postgres/Redis, which do have one) — it's assembled by hand in `keycloak.go` from a generic image, using `testdata/evently-realm-export.json` as a minimal realm fixture (two clients: a public one for password-grant tokens, a confidential one whose service account has the `realm-management` `manage-users` role, so it can call the Admin API the same way `modules/users/internal/adapters/driven/keycloak` does in production).

```bash
go test ./test/integration/... -v -timeout 5m
```

Requires Docker running locally. First run pulls three images (`postgres:18-alpine`, `valkey/valkey:9-alpine`, `quay.io/keycloak/keycloak:latest`) — expect ~25-40s; subsequent runs are faster once images are cached.

### System integration tests

`system_flows_test.go`, in the same package/harness as the integration tests above — Go equivalent of the C# course's top-level `Evently.IntegrationTests` (as opposed to the per-module `Users.IntegrationTests`): flows that cross module boundaries through the real outbox → event bus → inbox pipeline, not just one module in isolation.

`poller_test.go` is the Go port of C#'s `Poller.WaitAsync<T>`: cross-module propagation is eventually consistent (an outbox worker has to tick before the effect is visible elsewhere), so these tests retry an assertion every 500ms until it passes or a timeout is hit, instead of asserting once right after the triggering call.

`TestSystemFlow_UserCanAddItemToCartAfterRegistrationAndEventCreation` proves both propagation paths that exist today — Users → Ticketing (`Customer`) and Events → Ticketing (`Event`/`TicketType`) — through the only endpoint that observably depends on both: there's no dedicated "does this customer exist yet" read endpoint to poll individually, so a successful `PUT /carts/add` is the signal that both landed.

---

## CI/CD

`.github/workflows/ci.yml` runs on every push to `main` and every pull request. Four jobs, all calling the same `task` targets described in [Testing](#testing) above — nothing CI-only, nothing that can pass locally and fail in CI (or the reverse) because the commands differ:

```
lint ──┐
       ├──► integration ──► build
unit ──┘
```

| Job | Runs | Needs |
|---|---|---|
| `lint` | `task lint` (`go vet` + `gofmt -l`) | — |
| `unit` | `task test` (unit + architecture tests) | — |
| `integration` | `task test:integration` (real Postgres/Valkey/Keycloak via testcontainers) | `lint`, `unit` — no point paying for three containers if a cheap check already failed |
| `build` | `task build` (binary) + `docker build --target builder` (image) | `unit`, `integration` |

`lint` and `unit` run in parallel — both are fast and don't depend on each other. `ubuntu-latest` GitHub-hosted runners ship a running Docker daemon, so `testcontainers-go` works with no `services:` block or Docker-in-Docker setup, exactly like running the integration suite locally.

Deliberately out of scope: no `golangci-lint` (no existing config in this repo, and introducing one with default rules would produce a wall of first-run warnings unrelated to any actual change — `go vet` + `gofmt` were already clean), and no deploy/registry-push step (this pipeline proves the code is correct and buildable; it doesn't publish anywhere).

---

## Reliable messaging: Outbox, Idempotent Consumer & Inbox

Every module writes to Postgres and dispatches events. Those two things are not naturally atomic: if the process crashed between "row committed" and "event dispatched", the event — and whatever cross-module side effect depended on it — would be lost forever, silently. Three cooperating patterns close that gap. All three live in `internal/shared/` and are shared code, not copy-pasted per module.

### Outbox — durable, asynchronous domain event dispatch

A repository never dispatches a domain event directly. Instead, when an aggregate is written, its raised domain events are serialized into a `{schema}.outbox_messages` row **in the same database transaction** as the aggregate — both commit together or neither does:

```go
return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
    q := r.queries.WithTx(tx)
    // ... insert/update the aggregate via q ...
    domainEvents := aggregate.DomainEvents()
    aggregate.ClearDomainEvents()
    return outbox.InsertMessages(ctx, tx, schema, domainEvents)
})
```

A background `outbox.Worker` — one per module, started as a goroutine in `cmd/api/main.go` and stopped on `SIGINT`/`SIGTERM` — polls its schema's `outbox_messages` on a ticker, claims a batch with `SELECT ... FOR UPDATE SKIP LOCKED`, decodes each row back into a concrete Go struct via an `outbox.TypeRegistry` (a `type name → reflect.Type` map populated once at startup — the Go equivalent of Newtonsoft's `TypeNameHandling`, without embedding .NET-style metadata in the payload), and dispatches it through the module's `events.Dispatcher`. This is the **only** place a domain event handler runs — it never runs inline with the HTTP request that raised it.

### Idempotent Consumer — safe retries when a handler fails

A single domain event can fan out to more than one handler (e.g. `EventCancelledDomainEvent` triggers both archiving tickets and refunding payments). If one handler fails, the worker marks the whole outbox message as errored and retries it on the next tick — which would re-run every handler for that event, including the ones that already succeeded, unless each handler tracks its own completion.

`outbox.Idempotent` decorates a handler so it runs at most once per `(outbox_message_id, handler_name)` pair, backed by `{schema}.outbox_message_consumers`:

```go
sharedevents.Register(app.Dispatcher, outbox.Idempotent(
    "ArchiveTicketsHandler", app.DB, schema,
    eventhandlers.NewArchiveTicketsHandler(eventRepo).Handle,
))
```

Every handler registration in every `module.go` goes through this decorator — there is no "trusted" handler that skips it. The message id it keys on is threaded through `context.Context` (`events.WithMessageID`/`events.MessageIDFromContext`), set once by the worker and read by every decorator downstream, rather than being added as a field on every domain event type.

### Inbox — the same protection on the receiving side of the event bus

`EventBus.Publish` is a synchronous in-process call, not a broker — there's no message redelivery to guard against the way there would be with Kafka or RabbitMQ. But Users and Ticketing still commit to **independent Postgres transactions**: if the process crashes between Ticketing committing its side effect and Users committing its own `outbox_message_consumers` row, the next retry re-publishes from scratch, and Ticketing's consumer — with no protection of its own — would run again.

`internal/shared/inbox.Idempotent` closes that gap on the consuming module's side, mechanically identical to `outbox.Idempotent` but keyed against `{schema}.inbox_message_consumers`, reusing the same message id that already arrived via `ctx`:

```go
eventbus.Subscribe[usersintegrationevents.UserRegisteredIntegrationEvent](app.EventBus, inbox.Idempotent(
    "UserRegisteredConsumer", app.DB, schema,
    consumers.NewUserRegisteredConsumer(createCustomerHandler).Handle,
))
```

Unlike the C# reference (MassTransit consumer → `inbox_messages` payload row → background `ProcessInboxJob` → handler), there is no separate payload table or background job on this side: dispatch is already synchronous, so the integration event is in hand the moment the consumer runs — the inbox only needs to remember "did I already do this," not "what was I asked to do."

### Configuration

Each module's outbox worker is tuned independently in `configs/modules.<module>.yaml`:

```yaml
users:
  outbox:
    interval_seconds: 10
    batch_size: 20
```

### Database tables

`outbox_messages` / `outbox_message_consumers` exist in the `events`, `users`, and `ticketing` schemas. `inbox_message_consumers` exists in `events` and `ticketing` — every module that subscribes to another module's integration events gets one. See [Database](#database) for the full column layout.

---

## Materialized View: Event Statistics

Answering "how many tickets were sold, how many attendees checked in, how many check-in attempts failed" for an event would otherwise mean joining `tickets` against `orders` and scanning check-in state at read time. Instead, `ticketing.event_statistics` is a denormalized read model kept in sync by projection handlers reacting to the same ticket domain events that already flow through the outbox — no new dispatch mechanism, just handlers that write to a different table than the one that raised the event.

```
Ticket domain events                             (already raised by NewTicket / Ticket.CheckIn)
        │
        ▼  outbox.Worker ticks — same mechanism as any other handler
TicketCreatedStatisticsHandler          → EnsureRow + IncrementTicketsSold
TicketCheckedInStatisticsHandler        → IncrementAttendeesCheckedIn
TicketCheckInDuplicateStatisticsHandler → AppendDuplicateCheckIn(ticket code)
TicketCheckInInvalidStatisticsHandler   → AppendInvalidCheckIn(ticket code)
        │
        ▼
ticketing.event_statistics (event_id, tickets_sold, attendees_checked_in,
                             duplicate_check_in_tickets[], invalid_check_in_tickets[])
        │
        ▼  read side bypasses the domain and the aggregate entirely
GET /events/{id}/statistics → GetEventStatisticsByEventID (sqlc, no joins)
```

`Ticket.CheckIn(customerID)` (`modules/ticketing/internal/domain/ticket.go`) is the one domain method worth calling out here: unlike `Archive()`, it raises a domain event on **every** outcome, including the two failure paths (wrong customer, already used) — the whole point of this view is counting failed attempts too, not just successes. `event_statistics` deliberately stores counts only, not the event's title/location: even though `ticketing.events` is now kept in sync (see [Inter-module communication](#inter-module-communication)), denormalizing those fields into every statistics row would mean re-syncing them on every rename, which isn't worth it for a read model that's only ever queried by `event_id`.

---

## Saga Pattern: Cancel-Event Orchestration

Cancelling an event triggers two independent side effects in Ticketing — refund every payment, archive every ticket — and neither one knows about the other. Something has to wait for **both** to finish, in whatever order they happen to complete, before the cancellation can be considered done. That coordinator is a saga: process state keyed by `event_id`, living outside any single request or message, that survives a crash mid-flight.

The C# reference implements this with MassTransit's `MassTransitStateMachine`, persisted in Redis. This port keeps the same observable behavior — start, wait for both branches, complete — as a much smaller mechanism: a single Postgres table plus an atomic bitmask join, since a full state-machine DSL would be overkill for one saga with two parallel steps.

```
events.Event.Cancel()                                     (Events, HTTP request)
        │  same transaction
        ▼
events.outbox_messages row (EventCancelledDomainEvent)
        │
        ▼  outbox.Worker ticks
Idempotent("EventCancelledHandler") → publishes EventCanceledIntegrationEvent   ← events/api/
        │
        ▼  EventBus.Publish is synchronous — the rest of this chain runs inline
Idempotent("CancelEventSagaStartedHandler")
        │  INSERT events.cancel_event_saga_state (event_id)
        │  publishes
        ▼
EventBus.Publish(EventCancellationStartedIntegrationEvent)
        │
        ▼
Idempotent("EventCancellationStartedConsumer")             (ticketing/internal/app/consumers/)
        │  calls CancelEventCommand → ticketing's own Event.Cancel()
        ▼
ticketing.outbox_messages row (EventCancelledDomainEvent, ticketing's own copy)
        │
        ▼  outbox.Worker ticks (ticketing)
ArchiveTicketsHandler + RefundPaymentsHandler run, each raising its own completion event
        │
        ▼
ticketing.outbox_messages rows (EventTicketsArchivedDomainEvent, EventPaymentsRefundedDomainEvent)
        │
        ▼  outbox.Worker ticks again
TicketsArchivedIntegrationEventPublisher / PaymentsRefundedIntegrationEventPublisher
        │  publish
        ▼
EventBus.Publish(EventTicketsArchivedIntegrationEvent / EventPaymentsRefundedIntegrationEvent)  ← ticketing/api/
        │
        ▼
Idempotent("CancelEventSagaTicketsArchivedHandler" / "...PaymentsRefundedHandler")
        │  UPDATE events.cancel_event_saga_state SET completed_steps = completed_steps | $step
        │  RETURNING completed_steps        (atomic — safe under redelivery or either arrival order)
        ▼
completed_steps == AllSteps? → publish EventCancellationCompletedIntegrationEvent, delete the row
```

The saga lives in `modules/events/internal/app/sagas/event_cancellation/` — three handlers (`StartedHandler`, `PaymentsRefundedHandler`, `TicketsArchivedHandler`), each a plain consumer wired the same way as any other `eventbus.Subscribe` + `inbox.Idempotent` pair. There's nothing framework-specific about it: `CancelEventSagaRepository` is just another `ports/outbound` interface, and the "state machine" is two bits in a `SMALLINT` column. Because Events subscribes to Ticketing's integration events and Ticketing subscribes to Events' — the first bidirectional cross-module dependency in this codebase — `TestModuleIsolation_NoModuleDependsOnAnotherModule` was checked explicitly to confirm both directions are still only through `api/integrationevents`, which the architecture test allows regardless of direction.

---

## Authentication & Authorization

### Authentication (JWT — stateless)

All endpoints require a valid Bearer token issued by Keycloak, except:

| Path | Public |
|---|---|
| `POST /users/register` | ✅ No token required |
| `GET /health` | ✅ No token required |
| Everything else | 🔒 `Authorization: Bearer <token>` required |

Tokens are validated **stateless** at the middleware layer using OIDC discovery:

1. `oidc.NewProvider` fetches the Keycloak discovery document once at startup and caches the JWKS.
2. Every request calls `verifier.Verify(token)` — no Keycloak roundtrip.
3. After verification, the middleware resolves the user's permissions (see below) and stores everything in the request context as `auth.Claims`.

```go
type Claims struct {
    Sub         string      // Keycloak subject (identity_id)
    Email       string
    UserID      uuid.UUID   // internal user UUID from DB
    Permissions []string    // e.g. ["events:read", "carts:add", ...]
}
```

### Authorization (RBAC)

Roles and permissions are stored in the `users` schema. Two roles exist out of the box:

| Role | Assigned on |
|---|---|
| `Member` | Automatically on user registration |
| `Administrator` | Manual assignment |

Permissions are 17 granular codes, one per operation:

```
users:read          users:update
events:read         events:search       events:update
ticket-types:read   ticket-types:update
categories:read     categories:update
carts:read          carts:add           carts:remove
orders:read         orders:create
tickets:read        tickets:check-in
event-statistics:read
```

#### Permission caching (Valkey)

To avoid a DB query on every request, permissions are cached in Valkey:

- **Key:** `permissions:{identity_id}` (where `identity_id` = Keycloak `sub` claim)
- **Value:** `{ "user_id": "uuid", "permissions": ["...", "..."] }` (JSON)
- **TTL:** 5 minutes

Flow: JWT validated → check Valkey → (miss) query DB → store in Valkey → populate `Claims`.

#### Protecting a route

```go
mux.Handle("GET /events", middleware.RequirePermission(domain.PermEventsRead)(handler))
```

`RequirePermission` checks `claims.Permissions` in the context and returns `403 Forbidden` if the permission is absent.

> **Known gap — enforcement is not wired yet.** `middleware.RequirePermission` exists and is unit-testable in isolation, but no route in any module is actually wrapped with it today — `rg "RequirePermission" -g "*.go"` outside the middleware package itself returns nothing. In practice, **every** endpoint (except `POST /users/register` and `GET /health`) only checks for a valid Bearer token; a `Member` can call any route a permission table below marks as needing a higher permission, and will get a normal `2xx` response instead of `403`. The "Permission" column in [Endpoints](#endpoints) documents the *intended* gate, computed and available in `claims.Permissions` on every request — it's just not enforced at the route level yet. Wire it the same way the snippet above shows, one route at a time.

---

## Database

Each module owns its own PostgreSQL schema. Cross-schema queries are prohibited.

### Schema: `events`

| Table | Description |
|---|---|
| `events.categories` | Event categories |
| `events.events` | Events (draft → published → cancelled) |
| `events.ticket_types` | Ticket types belonging to an event |
| `events.outbox_messages` | Pending/processed domain events — see [Reliable messaging](#reliable-messaging-outbox-idempotent-consumer--inbox) |
| `events.outbox_message_consumers` | Per-handler idempotency tracking for the rows above |
| `events.inbox_message_consumers` | Per-consumer idempotency tracking for integration events received from other modules |
| `events.cancel_event_saga_state` | Cancel-event saga correlation state — one row per in-flight cancellation, bitmask of which parallel steps (refund, archive) have completed — see [Saga Pattern](#saga-pattern-cancel-event-orchestration) |

### Schema: `users`

| Table | Description |
|---|---|
| `users.users` | Registered users (`identity_id` links to Keycloak subject) |
| `users.roles` | Available roles (`Member`, `Administrator`) |
| `users.permissions` | Permission codes (17 entries) |
| `users.role_permissions` | Role → permission mapping |
| `users.user_roles` | User → role assignment |
| `users.outbox_messages` | Pending/processed domain events — see [Reliable messaging](#reliable-messaging-outbox-idempotent-consumer--inbox) |
| `users.outbox_message_consumers` | Per-handler idempotency tracking for the rows above |

### Schema: `ticketing`

| Table | Description |
|---|---|
| `ticketing.customers` | Customer replica (synced from Users via integration event) |
| `ticketing.events` | Event replica |
| `ticketing.ticket_types` | TicketType replica with available quantity tracking |
| `ticketing.orders` | Purchase orders |
| `ticketing.order_items` | Line items per order |
| `ticketing.tickets` | Issued tickets (one per order item unit); check-in state tracked via `used_at_utc` |
| `ticketing.payments` | Payments linked to orders |
| `ticketing.event_statistics` | Materialized view of per-event ticket/check-in counts — see [Materialized View](#materialized-view-event-statistics) |
| `ticketing.outbox_messages` | Pending/processed domain events — see [Reliable messaging](#reliable-messaging-outbox-idempotent-consumer--inbox) |
| `ticketing.outbox_message_consumers` | Per-handler idempotency tracking for the rows above |
| `ticketing.inbox_message_consumers` | Per-consumer idempotency tracking for integration events received from other modules |

---

## Getting started

```bash
# Start all infrastructure (Postgres + Valkey + Keycloak)
docker compose up evently.database evently.cache evently.identity -d

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
APP_ENV=development go run ./cmd/api
```

Or run everything with Docker:

```bash
docker compose up --build
```

The API will be available at `http://localhost:8080`.

### Keycloak

Keycloak runs at `http://localhost:18080`. The realm `evently` is pre-configured with two clients:

| Client | Purpose |
|---|---|
| `evently-public-client` | Frontend / token issuance |
| `evently-confidential-client` | Backend admin API (user registration) |

To get a token for testing:

```bash
curl -s -X POST http://localhost:18080/realms/evently/protocol/openid-connect/token \
  -d "grant_type=password" \
  -d "client_id=evently-public-client" \
  -d "username=<email>" \
  -d "password=<password>" | jq .access_token
```

---

## Endpoints

> **All endpoints except `POST /users/register` and `GET /health` require `Authorization: Bearer <token>`.**
> The "Permission" column below is the *intended* gate — see the [known gap](#authorization-rbac): it's not enforced by any route yet, so any authenticated user can currently call any endpoint.

### Health

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/health` | Public | Liveness + readiness check (Postgres + Valkey) |

---

### Users

| Method | Path | Auth | Permission | Description |
|--------|------|------|------------|-------------|
| `POST` | `/users/register` | Public | — | Register a new user |
| `GET` | `/users/me/profile` | 🔒 | `users:read` | Get the authenticated user's profile |
| `PUT` | `/users/me/profile` | 🔒 | `users:update` | Update the authenticated user's profile |

#### `POST /users/register`
```json
{
  "email": "string",
  "first_name": "string",
  "last_name": "string",
  "password": "string"
}
```
Response `201`: `{ "id": "uuid" }`

Registers the user in Keycloak and in the local DB. The new user is assigned the `Member` role automatically.

#### `GET /users/me/profile`
Response `200`:
```json
{
  "id": "uuid",
  "email": "string",
  "first_name": "string",
  "last_name": "string"
}
```

#### `PUT /users/me/profile`
```json
{
  "first_name": "string",
  "last_name": "string"
}
```
Response `204`

---

### Events

| Method | Path | Auth | Permission | Description |
|--------|------|------|------------|-------------|
| `POST` | `/events` | 🔒 | `events:update` | Create a new event |
| `GET` | `/events` | 🔒 | `events:read` | List all events |
| `GET` | `/events/search` | 🔒 | `events:search` | Search events |
| `GET` | `/events/{id}` | 🔒 | `events:read` | Get event by ID (includes ticket types) |
| `POST` | `/events/{id}/publish` | 🔒 | `events:update` | Publish an event |
| `POST` | `/events/{id}/cancel` | 🔒 | `events:update` | Cancel an event — triggers the [cancel-event saga](#saga-pattern-cancel-event-orchestration) |
| `PUT` | `/events/{id}/reschedule` | 🔒 | `events:update` | Reschedule an event |

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

| Method | Path | Auth | Permission | Description |
|--------|------|------|------------|-------------|
| `POST` | `/categories` | 🔒 | `categories:update` | Create a category |
| `GET` | `/categories` | 🔒 | `categories:read` | List all categories |
| `GET` | `/categories/{id}` | 🔒 | `categories:read` | Get a category by ID |
| `POST` | `/categories/{id}/archive` | 🔒 | `categories:update` | Archive a category |
| `PUT` | `/categories/{id}/name` | 🔒 | `categories:update` | Rename a category |

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

| Method | Path | Auth | Permission | Description |
|--------|------|------|------------|-------------|
| `POST` | `/ticket-types` | 🔒 | `ticket-types:update` | Create a ticket type for an event |
| `GET` | `/ticket-types` | 🔒 | `ticket-types:read` | List ticket types (`?event_id=uuid`) |
| `GET` | `/ticket-types/{id}` | 🔒 | `ticket-types:read` | Get a ticket type by ID |
| `PUT` | `/ticket-types/{id}/price` | 🔒 | `ticket-types:update` | Update price |

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

### Carts

| Method | Path | Auth | Permission | Description |
|--------|------|------|------------|-------------|
| `PUT` | `/carts/add` | 🔒 | `carts:add` | Add a ticket type to a customer's cart |

#### `PUT /carts/add`
```json
{
  "customer_id": "uuid",
  "ticket_type_id": "uuid",
  "quantity": 2
}
```
Response `200`. The cart is stored in Valkey keyed by `customer_id`.

> **Note:** `customer_id` is the same UUID as the user's `id`. A customer record is created automatically when a user registers, via the `UserRegisteredIntegrationEvent` flow.

---

### Orders

| Method | Path | Auth | Permission | Description |
|--------|------|------|------------|-------------|
| `POST` | `/orders` | 🔒 | `orders:create` | Place an order |

#### `POST /orders`
```json
{
  "customer_id": "uuid",
  "ticket_types": [
    { "ticket_type_id": "uuid", "quantity": 2 }
  ]
}
```
Response `201`: `{ "id": "uuid" }`

Availability is validated and decremented atomically under a row lock (`SELECT ... FOR UPDATE`) as part of order creation — a ticket type selling out here raises `TicketTypeSoldOutDomainEvent`. Ticket issuance itself is **not** part of this request: `OrderCreatedDomainEvent` is recorded in the outbox in the same transaction as the order, and a background worker later creates one `Ticket` per unit of quantity ordered and marks the order as fulfilled. Poll `GET /orders/{id}` (not yet implemented) or the eventual order-confirmation flow to observe completion.

> Cancelling an event (`POST /events/{id}/cancel`) archives its tickets and refunds its payments in Ticketing automatically, end to end, coordinated by the [cancel-event saga](#saga-pattern-cancel-event-orchestration) — no direct call between the two modules. Payment refunds call a `PaymentGateway` port backed by an in-memory fake — no real payment processor is integrated.

---

### Tickets

| Method | Path | Auth | Permission | Description |
|--------|------|------|------------|-------------|
| `POST` | `/tickets/{id}/check-in` | 🔒 | `tickets:check-in` | Check a ticket in |

#### `POST /tickets/{id}/check-in`
```json
{ "customer_id": "uuid" }
```
Response `204` on success.

`customer_id` must match the ticket's owner. `Ticket.CheckIn` (see [Materialized View](#materialized-view-event-statistics)) raises a domain event on every outcome, including failures, so both of these surface as ordinary Problem Details responses and are still counted in `event_statistics`:
- `409 ticket.already_checked_in` — the ticket was already used.
- `400 ticket.check_in_invalid` — `customer_id` does not own this ticket.

---

### Event Statistics

| Method | Path | Auth | Permission | Description |
|--------|------|------|------------|-------------|
| `GET` | `/events/{id}/statistics` | 🔒 | `event-statistics:read` | Read an event's materialized ticket/check-in counters |

#### `GET /events/{id}/statistics`
Response `200`:
```json
{
  "event_id": "uuid",
  "tickets_sold": 42,
  "attendees_checked_in": 30,
  "duplicate_check_in_tickets": ["tc_..."],
  "invalid_check_in_tickets": ["tc_..."]
}
```
Response `404` if no ticket has ever been sold for the event (the row is created lazily on first sale — see [Materialized View](#materialized-view-event-statistics)).

---

## Error responses

All errors follow [RFC 9457 Problem Details](https://www.rfc-editor.org/rfc/rfc9457):

```json
{
  "title": "string",
  "status": 400,
  "detail": "string"
}
```

| HTTP Status | Meaning |
|-------------|---------|
| `400` | Validation error |
| `401` | Missing or invalid Bearer token |
| `403` | Authenticated but missing required permission |
| `404` | Resource not found |
| `409` | Business rule conflict (e.g. publishing an event with no tickets, duplicate email) |
| `500` | Internal server error |

---

## Credit

Original course and C# reference implementation by
[Milan Jovanovic](https://www.milanjovanovic.tech/modular-monolith-architecture).
