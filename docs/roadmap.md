# Roadmap

Every milestone and step in this project: what is done, what is next, and what was decided along the
way. This is the single file to read to know where things stand.

The ordering follows one rule: **build the thinnest thing that runs, then grow it** — and from
Milestone 2 onward, **one new tool per step**, so each tool's contribution stays legible.

---

## Where we are

| | |
| --- | --- |
| **Done** | Milestone 0 foundation · 1 frontend · 2 database · 3 customers · 4 vehicles |
| **Done** | Milestone 5 — Spare parts (full CRUD) |
| **Current** | Design pass and the dashboard · done. Next: `make seed`, awaiting approval |

The application runs end to end: `make run`, then `http://localhost:8080`. Customers can be listed,
searched, created, edited, and deleted, against PostgreSQL, with the Vite bundle served from
`public/build`.

### Frontend build

`web/package.json` pins **Vite 8** with Rolldown, per the stack. `make assets` calls
`web/node_modules/.bin/vite` — the binary `bun install` puts there, so it is the version pinned in
`bun.lock`. `bun run build` reaches the same place through `package.json`, and works if your `bun run`
is healthy; it is not in the environment these steps were written in.

### The PostgreSQL server

Running since 2026-07-29: **PostgreSQL 16.4** in a container, database `car_service`, reached with the
`DATABASE_URL` in `.env`. Verify with:

```sh
pg_isready -h localhost -p 5432
psql "$DATABASE_URL" -c "select version();"
```

Pages before the database was considered on 2026-07-29 and **rejected** — the database comes first, so
every screen is built against real data instead of a fixture that has to be thrown away.

### Design and validation pass — 2026-07-30

A visual pass over the whole application, not a new module:

- **Design tokens in `web/src/app.css`.** A `brand` colour defined in `@theme`, so changing it later is
  one line rather than a find-and-replace through every template.
- **`views/ui.templ`** holds the pieces every page reuses: `PageHeader`, `BackLink`, `EmptyState`,
  `Card`. Shared markup written once.
- **The header marks the current section**, using `aria-current="page"` as well as colour — colour alone
  says it only to people who can see it.
- **The home page is a dashboard**: counts, the five most recent customers, and the parts that need
  restocking. One query for the counts rather than four round trips.
- **Field rules** live in `internal/server/validate.go` and are shared by all three forms: length limits
  counted in *runes* rather than bytes, and a phone check that requires enough digits to be a real
  number without dictating a format — `+216 20 145 872`, `20145872`, and `(216) 20-145-872` are all the
  same number, and a strict pattern mostly rejects valid input.

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

**The caveat, now closed:** htmx and Alpine were bundled but unused until Milestone 3 — installed, not
exercised. Search (3d) and the delete confirm (3e) fixed that.

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

## Milestone 3 — Customers ✅

**Goal:** one complete feature, end to end: list, create, view, edit, delete.

| Step | What | |
| ---- | ---- | - |
| 3a | List and detail pages, `views/customers.templ`, a 404 page used for unmatched routes too | ✅ |
| 3b | Create: `GET /customers/new`, `POST /customers`, required-field validation that redraws the form with what was typed | ✅ |
| 3c | Edit and delete: one form component shared by new and edit, `customerFromPath` shared by every handler that takes an `:id` | ✅ |
| 3d | htmx: `?q=` search on the list. One handler answers both a page and a fragment, switching on the `HX-Request` header. **Closes half the Milestone 1 caveat** — htmx now does real work. | ✅ |
| 3e | Alpine: confirm before deleting, as a progressive enhancement — the button stays a real submit, so delete still works with JavaScript off. **Closes the Milestone 1 caveat.** | ✅ |

**The pattern every later module copies:** query in `sql/queries` → `make sqlc` → a method on `store`
→ a handler on `Server` → a component in `views` → a test that runs against the real database.

Three things settled here that are worth not re-deciding:

- `store.ErrNotFound` — the store translates the driver's `pgx.ErrNoRows` into its own sentinel, so
  handlers answer 404 without importing pgx.
- **Writes are always `POST`**, even deletes, because that is what an HTML form can send. One route
  that works without JavaScript beats two routes where one needs it.
- **htmx for server state, Alpine for client state.** Searching asks the server, because the answer
  lives in the database. Confirming a delete does not, because "am I sure?" never leaves the browser.

**Why customers first:** the simplest entity in the domain, with no dependencies on any other table.
The honest place to establish the pattern every later module copies: SQL → sqlc → store → handler →
Templ page → htmx → test → docs.

## Milestone 4 — Vehicles ✅

Vehicles belonging to customers. The first foreign key, the first nested URL, and the first "list of
children on a parent's page".

| Step | What | |
| ---- | ---- | - |
| 4a | `vehicles` table with a foreign key, `VehiclesByCustomer`, vehicles listed on the customer page, and a refused delete explained rather than crashed | ✅ |
| 4b | Add a vehicle: nested form, first non-string field (`year`) with its own parse-and-range validation | ✅ |
| 4c | Edit and delete a vehicle, on unnested `/vehicles/:id` routes, with `vehicleFromPath` loading the owner from the vehicle | ✅ |

**`ON DELETE RESTRICT`, not `CASCADE`.** Deleting a customer who still has vehicles is refused, so a
workshop's records cannot vanish because someone clicked Delete. Cascade would have been less code and
the wrong default.

PostgreSQL reports the refusal as SQLSTATE 23503. The store translates it to `store.ErrInUse` and the
handler redraws the page with a message and **409 Conflict** — without that translation the visitor
gets a 500 for a perfectly reasonable click.

The foreign key is indexed explicitly: PostgreSQL indexes a primary key automatically but **not** a
foreign key, and every vehicle lookup goes through its customer.

**Nested for creating, unnested for the rest.** `POST /customers/:id/vehicles` needs the customer,
because that is where the owner comes from. `/vehicles/:id` does not: the id is unique, and putting the
customer back in the path would allow `/customers/1/vehicles/2` where the vehicle belongs to someone
else. `vehicleFromPath` reads the owner from the vehicle, so a page cannot show one customer's name
above another's car.

**Form values are always strings in the view model.** `VehicleForm.Year` is a `string`, not an `int`,
because a rejected form has to redraw exactly what was typed and "not a year" is not an `int`. Parsing
is part of validation, and `validateVehicle` returns the parsed year so the caller never parses it
twice.

## Milestone 5 — Spare parts and stock ✅

A parts catalogue with quantity on hand, so parts can later be consumed by a repair job.

| Step | What | |
| ---- | ---- | - |
| 5a | `parts` table, catalogue list with prices and an out-of-stock marker | ✅ |
| 5b | Create a part, with the price typed in dinars and stored exactly as millimes, and a unique `reference` | ✅ |
| 5c | Edit and delete | ✅ |

**Money is a whole number of millimes, never a float.** `0.1 + 0.2` is not `0.3` in binary floating
point, and prices get added up — a rounding error that is invisible on one part becomes a wrong invoice
total. `price_millimes` is an `INTEGER`; `views.Millimes` puts the decimal point back for display, and
is the only place it exists.

**Two `CHECK` constraints** — price and quantity cannot go negative. The form will refuse it too, but
the database is the only place that still holds when two requests adjust the same row at once, which is
exactly what happens once repair jobs consume parts.

**Parsing a price never goes through a float.** `strconv.ParseFloat("8.29") * 1000` gives 8289, not
8290, which is the exact error the integer column exists to avoid. `parseMillimes` splits on the
decimal point and parses two integers. More than three decimals is rejected rather than rounded — a
price the visitor did not type is worse than an error.

**A duplicate reference is a field error, not a 500.** The unique index reports SQLSTATE 23505; the
store translates it to `store.ErrDuplicate` and the form points at the reference box.

**Quantity is a plain number for now.** Real inventory records movements and derives the total.
Deferred to [backlog.md](backlog.md) with the trigger: Milestone 6, when jobs start consuming stock.

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
