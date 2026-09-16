# Development workflow

How work actually gets done in this repository, day to day.

## First-time setup

```sh
cp .env.example .env      # local settings; .env is git-ignored
cd web && bun install     # frontend dependencies
cd .. && make assets      # build the Vite bundle into public/build
make check                # confirm the toolchain works
```

A PostgreSQL server must be reachable at the `DATABASE_URL` in `.env` before the app will start —
`DATABASE_URL` is required, and `store.New` pings the database as it opens the pool, so a wrong
password fails at startup rather than on the first request.

There is no schema step: the application applies its own migrations on startup, so an empty database
is enough.

`make assets` is required before the first `make run`: the server refuses to start without Vite's
manifest, because a server that started anyway would serve every page with no styles and no
JavaScript — a confusing failure that looks like a CSS bug.

## The loop

```sh
make run          # start the server, Ctrl-C to stop
# edit code
make check        # format, vet, test
git commit
```

When changing templates or styles, run the bundler in a second terminal so the bundle rebuilds on
save:

```sh
make watch        # then refresh the browser
```

There is no live reload for Go code: after changing a `.go` or `.templ` file, stop the server and
`make run` again.

## The commands

| Command | What it does |
| ------- | ------------ |
| `make run` | `go run .` — starts on the port from `.env` |
| `make build` | compiles to `bin/server` (regenerates templates and assets first) |
| `make templ` | regenerates Go from `.templ` files |
| `make sqlc` | regenerates `internal/db` from `sql/queries` |
| `make seed` | replaces all data with demo data — development only, destructive |
| `make migrate-up` | applies pending migrations — the app also does this on startup |
| `make migrate-down` | rolls back the most recent migration |
| `make assets` | builds the Vite bundle into `public/build` |
| `make watch` | the same, rebuilding on every save |
| `make test` | `go test ./...`, with `.env` loaded so the database tests run |
| `make vet` | `go vet ./...` |
| `make fmt` | `go fmt ./...` — rewrites files in place |
| `make check` | fmt, then vet, then test — the pre-commit command |
| `make tidy` | `go mod tidy` — run after adding or removing an import |
| `make clean` | deletes `bin/` and `public/build/` |

`make check` **modifies your files**, because `fmt` rewrites them. That is deliberate: formatting is
not a matter of opinion in Go, so having it fixed rather than reported is one less thing to argue
with.

**`vet` is the one people skip.** It catches mistakes the compiler permits: a `Printf` whose
arguments do not match its format string, a struct tag with a typo, a result that is silently
discarded. It costs nothing and it has never once been wrong.

**`templ` and `assets` are prerequisites, not chores.** `run`, `build`, and `test` all depend on
`templ`, because a `.templ` file compiles to Go and an edit you forget to regenerate is silently
ignored — the page just does not change, with no error saying why. `build` also depends on `assets`,
since a binary without its bundle is not deployable.

`make test` loads `.env` on purpose. The tests that need a real database skip themselves when
`DATABASE_URL` is unset — convenient on a machine without PostgreSQL, but it would also let a broken
query pass unnoticed, so the everyday command supplies the variable.

**Adding a migration.** Create the next numbered pair in `sql/migrations/` — `000002_<name>.up.sql`
and `.down.sql` — and restart the app, which applies it. Write the `.down.sql` at the same time, while
you still remember what the `up` did.

Never edit a migration that has already run. The database records which versions were applied, so an
edited file means the schema in PostgreSQL and the schema in the repository disagree with nothing to
detect it. A change is always a new migration.

**Changing a query.** Edit `sql/queries/*.sql`, run `make sqlc`, then use the new method. sqlc reads
`sql/migrations` for the schema, so a column that does not exist fails generation with a line number
rather than failing a request in production.

Unlike `templ`, `sqlc` is *not* a prerequisite of build or test. Templates change constantly while
styling; queries change rarely, and regenerating on every test run would mean everyone needs the sqlc
CLI to run the suite.

## Every feature follows the same path

The lifecycle, from `task.md`:

1. Requirement
2. Domain discussion
3. Database design
4. SQL
5. `sqlc generate`
6. Store
7. Service logic
8. HTTP handlers
9. Templ pages
10. HTMX interactions
11. Alpine.js — only if genuinely needed
12. Styling
13. Testing
14. Documentation
15. Refactoring

Steps 3–5 do not exist until Milestone 2. Until then a step is skipped rather than faked.

The order matters in one specific way: **the database comes before the Go code, and the Go code
before the HTML.** Designing the schema last means discovering at step 9 that the data cannot answer
the question the page asks.

## Definition of done

A micro-step is finished when all of these are true:

- `make check` passes.
- The new behaviour has been exercised — a test, or the running application, or both.
- Documentation affected by the change has been updated in the same step.
- `docs/roadmap.md` records what changed.
- There is a commit message written for it.

The third one is the easiest to skip and the most expensive later. Documentation that contradicts the
code is worse than none: it costs a reader time *before* misleading them. When a change makes a
document wrong, the document is part of the change.

## Commits

One short line:

```
type(scope): what changed
```

Types in use: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`.
Scope is the package or area — `config`, `server`, `customers`. Omit it when the change is
repository-wide.

```
feat(config): load settings from .env
feat(web): add Templ views and Vite bundle
refactor(server): simplify Run to ListenAndServe
docs: add architecture and ADRs
```

Use the imperative — *add*, not *added* — the convention Git's own tooling uses.

Add a body only when the change would otherwise be baffling later. Most commits do not need one.

One commit per micro-step, and **each completed milestone is pushed before the next one starts**, so
history reads as the sequence of decisions that built the project.

## Decisions

A choice with lasting consequences — a library, a schema shape, an architectural boundary — gets a
file in `docs/decisions/`, numbered in sequence, recording the context, the decision, the
consequences, and the alternatives that lost. An ADR without a rejected alternative is marketing.

Smaller choices are logged in `docs/roadmap.md`. A future improvement that is *not* being built now
goes in `docs/backlog.md`, with the reason it is deferred, so "not yet" does not decay into
"forgotten".
