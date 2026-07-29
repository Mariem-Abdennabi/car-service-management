# Roadmap

Every milestone and step in this project: what is done, what is next, and what was decided along the
way. This is the single file to read to know where things stand.

The ordering follows one rule: **build the thinnest thing that runs, then grow it** — and from
Milestone 2 onward, **one new tool per step**, so each tool's contribution stays legible.

---

## Where we are

| | |
| --- | --- |
| **Done** | Milestone 0 foundation · Milestone 1 frontend pipeline · Milestone 2 database |
| **Next** | Milestone 3 — Customers, the first full feature · awaiting approval |
| **Blocked on you** | `cd web && bun install` then `make assets` — needed before any page renders |

### The one outstanding action

`web/package.json` pins **Vite 8**, per the stack. Vite 8's Rolldown native binding could not be
downloaded in the environment these steps were written in: `bun install` failed to extract the
tarball, `npm` failed with an SSL cipher error. Both are network failures, not project failures.

So `web/bun.lock` and `web/node_modules/` do not exist yet. Run:

```sh
cd web && bun install && cd ..
make assets
make run
```

The Vite config was verified end to end on Vite 7, which takes identical configuration, producing the
hashed bundle and manifest exactly as expected.

### The PostgreSQL server

Running since 2026-07-29: **PostgreSQL 16.4** in a container, database `car_service`, reached with the
`DATABASE_URL` in `.env`. Verify with:

```sh
pg_isready -h localhost -p 5432
psql "$DATABASE_URL" -c "select version();"
```

Pages before the database was considered on 2026-07-29 and **rejected** — the database comes first, so
every screen is built against real data instead of a fixture that has to be thrown away.

---

## Milestone 0 — Project foundation ✅

**Goal:** `make run` starts an HTTP server that answers a health check. No database, no HTML.

| Step | What | Result |
| ---- | ---- | ------ |
| 0.1 | Repository scaffolding | `CLAUDE.md`, `docs/`, ADRs 001–004, `.gitignore`, `README.md` |
| 0.2 | Go module and entry point | `go.mod`, first successful build, zero dependencies |
| 0.3 | Configuration | `internal/config`: typed `Config`, defaults, validation, tests |
| 0.3a | *Revision* | `.env` support added via `godotenv`, at your request |
| 0.4 | Gin server | `internal/server`: `Server` struct, `GET /healthz`, `Recovery` always on and `Logger` in development only |
| 0.4a | *Revision* | Graceful shutdown removed — too much concurrency for this stage; deferred to Milestone 9 |
| 0.5 | Day-to-day commands | `Makefile`, `docs/development-workflow.md` |

**Why first:** it establishes where code goes, how configuration reaches it, and how the process
starts. Everything later plugs into this.

## Milestone 1 — Frontend pipeline ✅ *(with one caveat)*

**Goal:** the server renders a real HTML page, styled, from a Templ template.

| Step | What | Result |
| ---- | ---- | ------ |
| 1.1 | Templ | `views/`: `Layout` with a children slot, `Home`, `render` helper, `GET /` |
| 1.2 | Tailwind CSS v4 | first via the Tailwind CLI — later replaced |
| 1.3 | htmx + Alpine.js | first vendored into the repo, then CDN tags, then npm dependencies |
| 1.4 | Live reload | `air` added — later removed, not part of the stack |
| 1.5 | Restructure | `main.go` at the root, `views/` as a flat package, all frontend tooling inside `web/`, `assets/` reading Vite's manifest, output to `public/build` |

**Built:** Tailwind v4 through `@tailwindcss/vite`, with htmx and Alpine as npm dependencies compiled
from `web/src/app.js` into one hashed JavaScript file. Filenames carry a content hash, so `assets/`
reads Vite's manifest at startup and the layout links whatever it reports. A missing manifest is a
startup error — a server that started anyway would serve every page unstyled, which looks like a CSS
bug rather than a missing build.

**The caveat:** htmx and Alpine are bundled and loaded, but **nothing on the page uses either**. They
are installed, not exercised. Their first real use is Milestone 3.

**Removed during 1.5:** the vendored JS directory, a throwaway htmx demo, `air`, and an ADR that had
argued against Vite. Steps 1.2–1.4 churned — three approaches to CSS, three to JavaScript delivery —
which is what the one-tool-per-step rule exists to prevent.

---

## Milestone 2 — Database ✅

**Goal:** the application reads and writes PostgreSQL, with migrations and type-safe queries.

Split into three committed steps so each tool arrives on its own, and so the boilerplate sqlc removes
is felt before sqlc removes it:

| Step | New tool | What | |
| ---- | -------- | ---- | - |
| 2a | pgx/v5 | `internal/store`: pool, `customers` table, **hand-written SQL** with `Query`/`Scan`. `DATABASE_URL` now required. `/healthz` pings the database and answers `503` when it is unreachable. Schema applied by hand with `make schema`. | ✅ |
| 2b | golang-migrate | `sql/migrations/` embedded with `go:embed` and applied on startup. `sql/schema.sql` and `make schema` deleted; `make migrate-up` / `make migrate-down` added for deliberate control. | ✅ |
| 2c | sqlc | `sqlc.yaml`, `sql/queries/customers.sql`, generated `internal/db/`, with `internal/store` as a thin wrapper. Every hand-written `rows.Scan` gone; no caller or test changed. | ✅ |

**Why 2a before 2c:** writing `rows.Scan(&a, &b, &c)` and misaligning two columns of the same type is
what makes sqlc's guarantee obvious. Going straight to the generated version means having the
conclusion without the experience.

**External dependency:** a running PostgreSQL server — see "The PostgreSQL server" above. The database
is `car_service`, reached with the `DATABASE_URL` already documented in `.env.example`.

**What 2c proved.** `store`'s four methods kept their signatures, so `server`, `main`, and every test
were untouched — the sign that 2a put the data in the right place. And a wrong column is now a
generation error rather than a runtime one:

```
sql/queries/customers.sql:7:7: column "customer_id" does not exist
```

Demo data still has no seeder — recorded in [backlog.md](backlog.md) as a command-line action, since a
single table is quicker to insert by hand than to build a seeder for.

**Migrations are append-only.** Once `000001` has run anywhere, it is never edited; a change means a
new numbered pair. Editing an applied migration means the schema in the database and the schema in the
files disagree, with nothing to detect it.

## Milestone 3 — Customers ⬜

**Goal:** one complete feature, end to end: list, create, view, edit, delete.

**Also where htmx and Alpine finally do real work** — an htmx-swapped list row, an Alpine
delete-confirm — closing the caveat left by Milestone 1.

**Why customers first:** the simplest entity in the domain, with no dependencies on any other table.
The honest place to establish the pattern every later module copies: SQL → sqlc → store → handler →
Templ page → htmx → test → docs.

## Milestone 4 — Vehicles ⬜

Vehicles belonging to customers. The first foreign key, the first nested URL, and the first "list of
children on a parent's page".

## Milestone 5 — Spare parts and stock ⬜

A parts catalogue with quantity on hand, so parts can later be consumed by a repair job.

## Milestone 6 — Service jobs ⬜

The heart of the domain. A job is opened against a vehicle, moves through statuses, consumes parts
from stock, and records labour.

**Why late:** it depends on customers, vehicles, and parts all existing. Building it earlier would
mean inventing fake versions of all three.

## Milestone 7 — Invoicing ⬜

Turn a completed job into an invoice with line items and totals.

## Milestone 8 — Users, authentication, and roles ⬜

Sessions, login, and access control (mechanic vs. manager).

**Why last among the features:** authentication is a cross-cutting concern, far easier to add once
the routes it must protect exist.

## Milestone 9 — Hardening ⬜

Broader test coverage, structured logging, request IDs, graceful shutdown, error handling review,
input validation review, and a deliberate refactor pass over the simple early code.

---

## Decisions

| Decision | Where |
| -------- | ----- |
| Layered server-rendered monolith, single binary | [ADR-001](decisions/ADR-001-project-architecture.md) |
| sqlc + pgx/v5 rather than an ORM | [ADR-002](decisions/ADR-002-sqlc.md) |
| Templ rather than `html/template` | [ADR-003](decisions/ADR-003-templ.md) |
| Gin rather than `net/http` or Chi | [ADR-004](decisions/ADR-004-gin.md) |
| Module path `github.com/Mariem-Abdennabi/car-service-management` | 0.2 |
| `.env` via godotenv; real environment variables win | 0.3a |
| No graceful shutdown yet | 0.4a, deferred to Milestone 9 |
| htmx and Alpine bundled by Vite, not CDN or vendored | 1.5 |
| Hashed asset filenames, resolved from Vite's manifest | 1.5 |
| Customers is the first feature | approved 2026-07-28 |

Ideas deliberately **not** being built now live in [backlog.md](backlog.md), each with the reason it
is deferred.

## Dependencies

Go — five direct:

| Dependency | Why |
| ---------- | --- |
| `github.com/gin-gonic/gin` | HTTP routing and middleware |
| `github.com/a-h/templ` | runtime for the generated templates. Pinned to the installed CLI version (v0.3.1020) — the two drifting apart causes confusing codegen errors. |
| `github.com/joho/godotenv` | reads `.env` in development |
| `github.com/jackc/pgx/v5` | PostgreSQL driver and connection pool — see [ADR-002](decisions/ADR-002-sqlc.md) |
| `github.com/golang-migrate/migrate/v4` | applies the embedded migrations at startup |

sqlc and templ are command-line tools, not imports: they generate committed code, so a build needs
only the Go toolchain.

Gin alone pulls in around 30 transitive modules, including a JSON codec, a validator, and a QUIC
implementation we never call. That is the concrete cost [ADR-004](decisions/ADR-004-gin.md) weighed
against the standard library, and the number to revisit at Milestone 9.

Frontend — in `web/package.json`:

| Dependency | Version | Role |
| ---------- | ------- | ---- |
| `vite` | ^8 | bundler, run through Bun |
| `@tailwindcss/vite` | ^4.3.3 | Tailwind CSS v4 as a Vite plugin |
| `htmx.org` | ^2.0.10 | partial page updates |
| `alpinejs` | ^3.15.12 | small pieces of UI state |

Nothing loads from a CDN: the browser fetches one JS file and one CSS file from our own server.

## Environment

Checked 2026-07-28, recorded so we stop re-checking:

| Tool | Version |
| ---- | ------- |
| Go | 1.26.2 (darwin/arm64) |
| Bun | 1.3.14 |
| Node | 20.19.4 |
| templ | 0.3.1020 |
| sqlc | 1.31.1 |
| golang-migrate | installed (`~/go/bin/migrate`) |
| psql | 18.4 |
| Docker | CLI installed, **daemon not running** |

Milestone 2 needs a reachable PostgreSQL server; whether that is the local install or a container is
decided then.

## How this roadmap is kept

- One commit per step. **Each completed milestone is pushed before the next starts.**
- A step is done when `make check` passes, the behaviour has been exercised, the affected
  documentation is updated in the same change, and this file records it.
- Reversals are logged as their own step (`0.3a`, `0.4a`) rather than folded in silently, so the
  history shows decisions being made and changed.
