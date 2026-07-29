# Architecture

## What this application is

A single Go binary that serves HTML over HTTP and stores its data in PostgreSQL. Pages are
rendered on the server; the browser gets HTML, not JSON. Interactivity comes from HTMX swapping
fragments of that HTML, with Alpine.js only where something genuinely needs client-side state.

This is a **server-rendered monolith**, and that is a deliberate choice rather than a starting
compromise. See [ADR-001](decisions/ADR-001-project-architecture.md).

## The layers

Four layers, each with one job:

```
   HTTP request
        │
        ▼
┌───────────────┐
│    handler    │  parse the request, call a service, render HTML or a fragment
└───────┬───────┘
        │
        ▼
┌───────────────┐
│    service    │  business rules, validation, orchestration
└───────┬───────┘
        │
        ▼
┌───────────────┐
│     store     │  database access, wrapping sqlc-generated queries
└───────┬───────┘
        │
        ▼
┌───────────────┐
│  PostgreSQL   │
└───────────────┘
```

### The one rule: dependencies point down

`handler` imports `service`. `service` imports `store`. **Nothing imports upward.** The store never
knows a web request exists; a service never knows what a status code is.

Go enforces part of this for free — import cycles are a compile error, so an accidental upward
dependency won't build. But the *spirit* of the rule needs discipline: passing a
`*gin.Context` into a service compiles fine and quietly welds your business rules to your web
framework. It's the mistake to watch for.

Two things follow from the rule, and they're the whole payoff:

- A service can be tested without starting an HTTP server.
- A business rule can be read without reading any SQL.

### What goes where, concretely

Take "delete a customer":

| Layer | Responsibility | Example |
| ----- | -------------- | ------- |
| handler | translate HTTP ↔ Go | read `:id` from the URL, return 404 vs. 204, re-render the list |
| service | decide whether it's allowed | refuse if the customer has an open service job |
| store | execute it | `DELETE FROM customers WHERE id = $1` |

The recurring question is *"is this a rule of the business, or a detail of the web/database?"*
Business rules go in the middle; details go at the edges.

## Why the layers arrive late

`internal/` currently contains only `config` and `server`, and handlers live in `server` as methods.
`store` and the sqlc-generated `db` package arrive at Milestone 2, and business logic separates out
at Milestone 3 if and when a rule needs somewhere to live.

A health-check endpoint routed through four layers would demonstrate the layers and nothing else.
The brief says not to over-engineer, and empty abstractions are the most expensive kind: they must
be read and understood by everyone who follows, and they pay nothing back.

## Configuration strategy

Configuration is read **once, at startup, from the environment**, into a typed struct:

```go
type Config struct {
    Env         string // "development" | "production"
    Port        int    // 1-65535, validated at startup
    DatabaseURL string
}
```

Three properties matter:

1. **Environment variables, not files committed to git.** Credentials never enter the repository.
   `.env.example` documents the required names with fake values; the real `.env` is git-ignored.
2. **Typed, not stringly.** Callers get `cfg.Port` as an `int`, which the compiler checks, rather
   than `os.Getenv("PORT")`, which a typo silently turns into `""`.
3. **Validated at startup.** A missing database URL kills the process immediately with a clear
   message, rather than surfacing as a confusing error on the first request that touches the
   database.

A local `.env` file, loaded by [godotenv](https://github.com/joho/godotenv), makes development
convenient. It is a *source* of environment variables, not a competing configuration system:
`config.LoadFile(".env")` fills in variables that are not already set, then `config.Load()` reads
and validates the environment as usual. An exported shell variable therefore always beats the
file, and in production — where no `.env` exists — the missing file is silently fine.

The config struct is constructed in `main.go` and handed to the server. Nothing deeper in the
application reaches for the environment on its own.

## Request lifecycle

Once Milestone 3 lands, a typical page request looks like:

1. Gin matches the route and runs middleware (recovery, logging).
2. The handler method — a method on `Server`, so it has access to config, the asset URLs, and the
   database pool — parses path and form values.
3. The handler calls a service method with plain Go types.
4. The service applies rules and calls the store.
5. The store runs a sqlc-generated query against the pgx pool.
6. Results travel back up as Go structs.
7. The handler renders a Templ component: a full page for a normal request, a fragment for an
   HTMX request.

Step 7 is the notable one. **The same handler can answer both**, choosing a full page or a
fragment based on the `HX-Request` header. That is what keeps HTMX from doubling the number of
routes.

## Error handling

Kept simple on purpose, and revisited at Milestone 9:

- Errors are returned, never panicked, and wrapped with context as they travel up:
  `fmt.Errorf("create customer: %w", err)`. By the time an error reaches a log line it describes
  its own path.
- The handler is the only layer that turns an error into a status code, because it is the only
  layer that knows what HTTP is.
- Gin's recovery middleware catches genuine panics so one bad request can't take down the
  process.

## What is deliberately absent

Named so that their absence reads as a decision rather than an oversight:

- **No dependency-injection framework.** A struct with fields, constructed in `main.go`.
- **No live reload.** Restart with `make run`; `make watch` handles the frontend bundle.
- **No ORM.** Hand-written SQL through sqlc — see [ADR-002](decisions/ADR-002-sqlc.md).
- **No JSON API.** Nothing consumes one yet. If a mobile client ever appears, the service layer
  is already the right place to add one; it would only need new handlers.
- **No caching, no message queue, no microservices.** No measured problem calls for them.
- **No interfaces around the store — yet.** Go's convention is to introduce an interface when a
  second implementation or a test double genuinely needs one, not in advance. Milestone 3 will show
  whether that point has arrived.

Related: [project-structure.md](project-structure.md) · [roadmap.md](roadmap.md) ·
[decisions/](decisions/)
