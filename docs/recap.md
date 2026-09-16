# Recap — what we built, with what, and why

A single document explaining the stack, every file type in the repository, and the order we did
things in. Written to be read start to finish once, then dipped into.

For *where things stand* and what is next, see [roadmap.md](roadmap.md).

---

## 1. What the application does

A workshop management system. Staff record **customers**, the **vehicles** those customers own, and
the **spare parts** on the shelf. Everything is a server-rendered web page — the browser receives
HTML, never JSON.

Working today: full create/read/update/delete for all three, a searchable customer list, a dashboard,
and demo data on demand.

---

## 2. The stack, with versions

Every version below was read from the project, not from memory.

### Backend

| Tool | Version | What it does here |
| ---- | ------- | ----------------- |
| **Go** | 1.26.2 | The language. One compiled binary, no runtime to install on the server. |
| **Gin** | v1.12.0 | HTTP routing and middleware. Matches `/customers/:id` to a handler, recovers from panics, logs requests in development. |
| **PostgreSQL** | 16.4 | The database. Chosen for real constraints — foreign keys, `CHECK`, unique indexes — that hold even when the application has a bug. |
| **pgx/v5** | v5.10.0 | The PostgreSQL driver and connection pool. Speaks PostgreSQL's own protocol rather than going through `database/sql`. |
| **sqlc** | v1.31.1 (CLI) | Reads our `.sql` files and generates Go. We write SQL; it writes the structs and the `Scan` calls. |
| **golang-migrate** | v4.19.1 | Applies schema migrations, embedded in the binary and run at startup. |
| **godotenv** | v1.5.1 | Loads `.env` in development so settings need not be typed on every run. |
| **pgerrcode** | (pinned) | Names PostgreSQL's error codes, so the code says `pgerrcode.UniqueViolation` rather than `"23505"`. |

### Frontend

| Tool | Version | What it does here |
| ---- | ------- | ----------------- |
| **templ** | v0.3.1020 (CLI + library) | HTML templates that compile to Go functions. Type-checked at build time. |
| **Tailwind CSS** | v4.3.3 (via `@tailwindcss/vite`) | Styling as utility classes written directly in the markup. |
| **htmx** | 2.0.10 | Sends a request and swaps the returned HTML into the page. Used for the customer search. |
| **Alpine.js** | 3.15.12 | Small pieces of browser-only state. Used for confirm-before-delete. |
| **Vite** | 8.1.5 (Rolldown) | Bundles the CSS and JavaScript into two hashed files. |
| **Bun** | 1.3.14 | Installs the frontend dependencies and runs Vite. |
| **Node** | 20.19.4 | Present because Vite runs on it. Not a runtime dependency of the app. |

**Nothing loads from a CDN.** The browser fetches one CSS file and one JS file, both from our own
server.

---

## 3. Why these tools, and why together

The stack is one coherent idea: **the server owns the HTML, and the browser gets pages rather than
data**. Each tool exists because that idea needs it.

### Go + Gin — the server

Go compiles to a single binary with no interpreter or runtime to install. Gin adds what the standard
library leaves you to write by hand: route parameters (`c.Param("id")`), form reading
(`c.PostForm("name")`), route grouping, and panic recovery.

*Example from this project:* `s.router.GET("/customers/:id", s.handleCustomer)` — one line, and the
`:id` is available as a string. The trade-off is recorded in
[ADR-004](decisions/ADR-004-gin.md): Go 1.22+ can route without a framework, and we may revisit.

### templ — HTML that the compiler checks

`html/template` in the standard library resolves everything at runtime, so a renamed struct field
becomes a blank space on a live page. templ compiles each template into a Go function, so the same
mistake stops the build.

*Example:* `views.CustomerDetail(assets, customer, vehicles, notice)` is a function call. Pass the
wrong type and it does not compile.

**How it complements Gin:** a handler renders a component to the response writer. And because a
component is just a function, rendering *part* of a page is nothing special — which is exactly what
htmx needs.

### htmx — interactivity without a JavaScript application

htmx turns HTML attributes into requests. The server answers with HTML, and htmx puts it in the page.

*Example — the customer search:*

```html
hx-get="/customers" hx-target="#customer-list" hx-swap="outerHTML"
hx-trigger="input changed delay:300ms"
```

The **same handler** answers both a full page and a fragment, choosing by the `HX-Request` header
that htmx sets. No second route, no JSON, no duplicated markup — the fragment is 719 bytes against
2 KB for the page.

**How it complements templ:** without components-as-functions, returning a fragment would mean a
second template that duplicates the first and drifts from it.

### Alpine.js — the state that never leaves the browser

Some state has no business on the server. "Have I clicked delete once?" is not a fact about the
workshop.

*Example — confirm before deleting:*

```html
x-data="{ confirming: false }"
x-on:submit="if (!confirming) { $event.preventDefault(); confirming = true }"
```

**The dividing line we use:** htmx when the answer lives in the database, Alpine when it does not.

### Tailwind CSS — styling in the markup

Utility classes mean the styling lives beside the element, so a component is genuinely
self-contained — no separate stylesheet to keep in step, and no class names to invent.

### Vite + Bun — one build for both

Vite compiles Tailwind, bundles htmx and Alpine into a single JavaScript file, and writes both with a
content hash in the filename (`app-CRIVrv26.css`), so browsers can cache forever and still pick up a
new version instantly. Bun installs the packages and runs Vite.

### PostgreSQL + pgx + sqlc + golang-migrate — the data

These four are one workflow:

1. **golang-migrate** changes the schema, in numbered files.
2. **sqlc** reads those same files to learn the schema, then generates Go from our queries.
3. **pgx** executes the generated code.
4. **PostgreSQL** enforces the rules regardless of what the application believes.

**Why they complement each other:** sqlc knows every column's real type *because* the migrations
describe them. Rename a column in a migration and `sqlc generate` fails with a line number — before
the code runs.

```
sql/queries/customers.sql:7:7: column "customer_id" does not exist
```

---

## 4. The path we took

One new tool per step, so each tool's contribution stays visible.

| Milestone | Step | What arrived | Why then |
| --------- | ---- | ------------ | -------- |
| **0** Foundation | 0.1 | Docs, ADRs, `.gitignore` | Decide the shape before code exists |
| | 0.2 | `go.mod`, `main.go` | A build that passes, so later failures have one cause |
| | 0.3 | `internal/config` | Settings read once, validated at startup |
| | 0.4 | Gin, `/healthz` | The server answers a request |
| | 0.5 | `Makefile` | The commands we type daily |
| **1** Frontend | 1.1 | templ | HTML from a component |
| | 1.2–1.4 | Tailwind, htmx, Alpine | Styling and the two JS libraries |
| | 1.5 | Vite, `assets/` | One bundle, hashed, resolved through a manifest |
| **2** Database | 2a | pgx — **hand-written SQL** | Write `rows.Scan` once, by hand |
| | 2b | golang-migrate | The schema stops being applied by hand |
| | 2c | sqlc | Delete every `Scan`; keep the SQL |
| **3** Customers | 3a–3c | List, detail, create, edit, delete | The first complete feature |
| | 3d | htmx search | htmx does real work |
| | 3e | Alpine confirm | Alpine does real work |
| **4** Vehicles | 4a–4c | The first foreign key, nested routes | A relationship between two tables |
| **5** Parts | 5a–5c | Catalogue, money, stock | Money and quantities |
| — | — | Dashboard, design pass, seed | Something worth opening |

**Why 2a before 2c** is the clearest example of the method. We wrote this by hand first:

```go
rows.Scan(&c.ID, &c.Name, &c.Phone, &c.City, &c.CreatedAt)
```

Nothing checks that the order matches the `SELECT`. Swap two `TEXT` columns and it compiles, runs,
and shows the wrong data. Having written that, sqlc's guarantee is a felt relief rather than a claim.

---

## 5. Every file type, and why it exists

### `.go` — hand-written

Compiled Go. 18 files of ours, in four packages: `config`, `server`, `store`, and `assets`.

### `*_templ.go` — generated by `templ generate`, **committed**

Each `.templ` compiles to one of these. They are committed on purpose, so `go build ./...` works for
anyone with only the Go toolchain — no templ CLI needed to build. The cost is noisier diffs.

Never edited by hand: the next `make templ` overwrites them.

### `.templ` — the templates

Go with HTML in it. `templ Customers(a assets.Assets, customers []store.Customer, search string)`
becomes a Go function taking exactly those arguments.

### `internal/db/*.sql.go` — generated by sqlc, **committed**

The generated queries and models. Same reasoning as `*_templ.go`: committed so a build needs only Go.
Marked `// Code generated by sqlc. DO NOT EDIT.`

### `sql/migrations/000002_create_vehicles.up.sql`

Applies a schema change. Numbered, and **append-only forever** — once a migration has run anywhere,
you write a new one rather than editing it. The database records which versions it has applied in a
`schema_migrations` table; editing an applied file makes the database and the repository disagree
with nothing to detect it.

### `sql/migrations/000002_create_vehicles.down.sql`

Undoes it. Written at the same time as the `up`, while you still remember what it did. It is what
makes a mistaken migration recoverable — `make migrate-down` rolls back one step.

Why *two files* rather than one: a rollback must be a separate statement the tool can run on its own.

### `sql/queries/*.sql` — sqlc's input

The queries we write, each with an annotation naming the Go method:

```sql
-- name: GetCustomer :one
SELECT * FROM customers WHERE id = $1;
```

`:one`, `:many`, `:exec` tell sqlc what shape to return. Separate from migrations because they change
at completely different rates — queries freely, migrations never.

### `sqlc.yaml`

Tells sqlc where the schema is, where the queries are, where to write Go, and that we use pgx. Also
holds one override mapping `timestamptz` to `time.Time` instead of `pgtype.Timestamptz`, keeping a pgx
type out of the templates.

### `web/package.json`

The frontend dependencies with version *ranges* (`^8.1.5`) and the two scripts.

### `web/bun.lock` — **committed**

The exact resolved version of every package, including transitive ones. `package.json` says "Vite 8 or
later minor"; the lockfile says "exactly 8.1.5, with these 200 sub-dependencies". Committing it is
what makes a build on another machine produce the same bytes.

### `go.mod` / `go.sum`

Go's equivalents. `go.mod` lists direct dependencies; `go.sum` holds a cryptographic hash of every
module, so a tampered dependency fails the build rather than running.

### `web/src/app.css` and `app.js`

The two entry points Vite compiles. `app.js` imports the CSS, htmx, and Alpine — which is why one
bundle contains all three.

### `.env` (ignored) and `.env.example` (committed)

`.env` holds real settings including the database password; it is git-ignored. `.env.example`
documents the variable names with fake values, so a new clone knows what to set.

### `public/build/` — generated, ignored

Vite's output plus `.vite/manifest.json`, which maps `src/app.js` to the hashed filename actually
produced. The Go package `assets/` reads that manifest at startup, since no Go code can know a
content hash in advance.

### `Makefile`

The project's commands in one place: `make` alone lists them. It also encodes *ordering* — `run`,
`build`, and `test` all depend on `templ`, because an edited template that was not regenerated is
silently ignored, which is a confusing ten minutes the first time.

---

## 6. Running it

### Before anything: PostgreSQL must be up

Everything else fails without it, and the error will not obviously say so.

```sh
pg_isready -h localhost -p 5432        # want: accepting connections
```

If the database runs in Docker, `docker start car-service-db` first.

### First time on a machine

```sh
cp .env.example .env         # 1. settings; .env is git-ignored
cd web && bun install        # 2. frontend packages
cd .. && make assets         # 3. build the CSS and JS bundle
make seed                    # 4. optional: demo data to look at
make run                     # 5. start
```

Then open **http://localhost:8080**.

**Step by step, and why each one:**

1. **`cp .env.example .env`** — copies the documented variable names with fake values into the file
   the app actually reads. `.env` is git-ignored because it holds the real database password;
   `.env.example` is committed so you know what to fill in. Without it the app exits at once:
   `DATABASE_URL is required`.
2. **`bun install`** — reads `web/package.json` and `web/bun.lock` and downloads the exact pinned
   versions into `web/node_modules`. Run in `web/` because that is where the frontend project lives.
   Only needed once, and again when a dependency changes.
3. **`make assets`** — compiles the frontend. See below for what it actually runs. Without it the
   server **refuses to start**: `read asset manifest (run 'make assets')`. Failing loudly is
   deliberate — a server that started anyway would serve every page unstyled, which looks like a CSS
   bug rather than a missing build.
4. **`make seed`** — fills the database with eight customers, nine vehicles, and twelve parts. Skip it
   if you would rather type your own data in; every page has a designed empty state.
5. **`make run`** — starts the server. Ctrl-C stops it.

There is **no database setup step**: migrations run automatically at startup, so an empty database is
enough.

### Every command, what it runs, and why

| Command | Actually runs | Why it exists |
| ------- | ------------- | ------------- |
| `make` | prints the list | A bare `make` should tell you things, not do something surprising |
| `make run` | `templ generate` then `go run .` | Compile and start in one step, no binary left behind |
| `make build` | `templ generate`, `vite build`, `go build -o bin/server .` | Produces the deployable binary |
| `make templ` | `templ generate` | Turns every `.templ` into a `*_templ.go` |
| `make sqlc` | `sqlc generate` | Regenerates `internal/db` from `sql/queries` |
| `make assets` | `cd web && ./node_modules/.bin/vite build` | Bundles CSS + JS into `public/build`, with hashed names |
| `make watch` | the same, with `--watch` | Rebuilds the bundle on every save while styling |
| `make seed` | `go run . -seed` | Replaces all data with demo data; refuses outside development |
| `make migrate-up` | `migrate ... up` | Applies pending migrations deliberately |
| `make migrate-down` | `migrate ... down 1` | Rolls back **one** migration |
| `make test` | `templ generate` then `go test ./...` with `.env` loaded | Runs every test, including the ones needing a real database |
| `make vet` | `go vet ./...` | Reports mistakes the compiler allows |
| `make fmt` | `go fmt ./...` | Formats in place |
| `make check` | `fmt`, then `vet`, then `test` | The one command to run before committing |
| `make tidy` | `go mod tidy` | Syncs `go.mod`/`go.sum` with the imports |
| `make clean` | `rm -rf bin/ public/build` | Deletes generated output |

**Five of these are worth understanding rather than memorising:**

**`run` and `build` regenerate templates first.** `run: templ` is a *prerequisite*: Make runs `templ`
before `go run`. Without it, editing a `.templ` and reloading shows the old page with no error
explaining why — a genuinely confusing ten minutes the first time.

**`build` also depends on `assets`, but `run` does not.** A binary without its stylesheet is not
deployable, so `build` always rebuilds it. Starting the server, though, should not require Bun — the
bundle only has to be built once, or watched with `make watch` while styling.

**`sqlc` is deliberately *not* a prerequisite of anything.** Templates change constantly while
styling, so a stale one is a real trap. Queries change rarely, and making every test run require the
sqlc CLI would mean nobody could run the suite without installing it. The generated code is
committed, so a build needs only Go.

**`make check` modifies your files**, because `fmt` rewrites them. That is on purpose: Go formatting
is not a matter of opinion, so having it fixed rather than reported is one less thing to argue with.

**`make test` loads `.env`.** The tests that need a real database skip themselves when `DATABASE_URL`
is unset — convenient on a machine without PostgreSQL, but it would also let a broken query pass
unnoticed, so the everyday command supplies the variable.

**`migrate-down` rolls back exactly one step** (`down 1`), not everything. Unqualified `down` would
drop the entire schema on a typo.

### The everyday loop

```sh
make run        # start; Ctrl-C to stop
# edit code
make check      # format, vet, test
git commit
```

Changing styles or templates? Run the bundler in a second terminal:

```sh
make watch      # rebuilds on save — then refresh the browser
```

There is **no live reload for Go code**: after changing a `.go` or `.templ` file, stop the server and
`make run` again. `air` is not part of the stack.

### What to click once it is running

| Try | What it demonstrates |
| --- | -------------------- |
| Type in the customer search box | **htmx** — the table filters with no page reload, and the URL gains `?q=` |
| Click **Delete** on a customer | **Alpine** — it becomes "Yes, delete" before anything happens |
| Delete a customer who has a vehicle | The foreign key refusing, shown as a message rather than a crash |
| Enter a price like `42.5000` | Validation rejecting rather than silently rounding |
| Enter a phone like `call me` | Field-level validation, with what you typed still in the form |

### When something goes wrong

| Symptom | Cause and fix |
| ------- | ------------- |
| `read asset manifest (run 'make assets')` | The bundle has never been built. Run `make assets`. |
| `DATABASE_URL is required` | No `.env`. Run `cp .env.example .env`. |
| `ping database: failed to connect` | PostgreSQL is not running. |
| Pages look unstyled | The bundle is stale or missing — `make assets`. |
| `address already in use` | An older server is still running: `lsof -ti:8080 \| xargs kill -9`. This one is worth knowing: it looks exactly like your code being broken, because the old binary answers with its old routes. |
| A template edit changes nothing | You ran `go run .` directly instead of `make run`, so `templ generate` never ran. |

## 7. Why the directories are shaped this way

```
main.go              composition root: read config, migrate, connect, serve
internal/
  config/            settings, read once at startup
  server/            routes, handlers, middleware
  store/             database access
  db/                sqlc output — never hand-edited
views/               templ components
assets/              reads Vite's manifest
web/                 the Vite project (package.json, node_modules, src/)
public/build/        Vite output — generated
sql/
  migrations/        schema history
  queries/           sqlc input
docs/
```

**`internal/`** is special to Go: nothing outside this module can import it. The compiler enforces it,
so these packages can be renamed and refactored freely.

**`store/` between handlers and `db/`** is what let sqlc arrive without touching anything else. At
step 2c every hand-written query was replaced by generated code and **no handler or test changed** —
the sign the seam was in the right place.

**`views/` is one flat package** so any component can render any other without an import cycle.

**`web/` holds the whole frontend toolchain** — its own `package.json`, lockfile, and `node_modules` —
so the repository root stays Go.

**`assets/` is separate** because it is the bridge between two worlds: Vite writes hashed filenames,
Go needs to link them, and the manifest is how they meet.

---

## 8. A request, end to end

`GET /customers/3`:

1. **Gin** matches `/customers/:id`, runs `Recovery`, calls `handleCustomer`.
2. The **handler** parses `3`, calls `store.Customer(ctx, 3)`.
3. The **store** calls sqlc's generated `GetCustomer`, which runs the SQL through the **pgx** pool.
4. A missing row comes back as `pgx.ErrNoRows`; the store translates it to `store.ErrNotFound` so the
   handler never imports pgx.
5. The handler calls `store.VehiclesByCustomer` for the vehicles.
6. It renders `views.CustomerDetail(...)` — a **templ** function — to the response.
7. The page links `/build/assets/app-CRIVrv26.css`, whose name came from the **manifest** at startup.
8. The browser loads the bundle: **Tailwind**'s styles, **htmx**, and **Alpine**.

---

## 9. Decisions worth remembering

| Decision | Why |
| -------- | --- |
| Money as integer millimes | `0.1 + 0.2 ≠ 0.3` in floating point, and prices get added up. `42.500 TND` is stored as `42500`. Parsing splits on the decimal point rather than using `ParseFloat`, which would turn `8.29` into `8289`. |
| Writes are always `POST` | HTML forms only support `GET` and `POST`. One route that works without JavaScript beats two where one needs it. |
| `303 See Other` after a successful form | The browser follows with `GET`, so a refresh cannot submit twice. |
| `422` and redraw on a rejected form | The values stay in the boxes. `400` would mean "I could not parse this", which is not what happened. |
| `ON DELETE RESTRICT` on vehicles | Deleting a customer must not silently destroy their records. The refusal becomes a `409` with an explanation, not a `500`. |
| Driver errors translated at the store | `ErrNotFound`, `ErrInUse`, `ErrDuplicate`. Handlers never see a SQLSTATE. |
| Generated code committed | A build needs only the Go toolchain. |
| Length limits counted in runes | `len("أميرة")` is 10 bytes but 5 characters. |

---

## 10. Deliberately not built

Service jobs, invoicing, authentication, reports, stock movements, live reload. All are on the
[roadmap](roadmap.md) or in the [backlog](backlog.md), each with the condition that would trigger
building it. Nothing here is an oversight; the list exists so absence reads as a decision.
