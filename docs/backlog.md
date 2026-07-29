# Backlog

Ideas worth doing **later**. Nothing here is implemented, and nothing here may be implemented
without first being promoted onto [the roadmap](roadmap.md) and approved.

The point of this file is to stop good ideas from derailing the current milestone.

## Format

Each entry states the idea, why it might be worth it, and why it is not now.

---

## Deferred

### A `seed` command for demo data
A re-runnable command that fills the database with realistic demo rows, so the pages are worth
looking at during development. It must be a **command-line action**, not a SQL file applied by hand.

Without Cobra in the stack, the simplest form is a flag on the existing binary — `go run . -seed`,
wrapped as `make seed` — which needs no new dependency and no second executable.

Deferred until there is more than one table worth seeding; a single `customers` table is quicker to
insert by hand than to build a seeder for.

### Docker Compose for local PostgreSQL
Would make the database reproducible with one command. Deferred because the machine already has
`psql` 18.4 installed, so a local server may be enough — an environment decision for Milestone 2.

### Embed the Vite bundle with `go:embed`
The server serves `public/build` from the filesystem, so it must be started from the repository root
and deployment means shipping the directory alongside the binary. `//go:embed` compiles the bundle
into the executable: one file to deploy, no working-directory assumption, no chance of a missing
stylesheet in production.

Deferred because development *wants* the filesystem version — embedded assets are frozen at compile
time, so `make watch` would stop having any effect. Worth doing when deployment is real.

### Graceful shutdown
`Run` currently calls `ListenAndServe` and nothing else, so Ctrl-C or a SIGTERM ends the process at
once and any request in flight is dropped. The production version listens for those signals, stops
accepting new connections, and gives in-flight requests a few seconds to finish.

It was written that way at step 0.4 and then deliberately removed: the mechanics — a goroutine, a
buffered error channel, `signal.NotifyContext`, and a second context for the shutdown deadline —
are a lot of concurrency to carry while the rest of the code is still being learned, and none of it
matters until real users can lose a request. Milestone 9 is the honest place for it, and by then
`http.Server.Shutdown` will be easier to read.

### Structured logging (`log/slog`) with request IDs
Real applications need correlated logs. Deferred to Milestone 9 so early code stays readable
with plain `log` while we are still learning the shape of the project.

### Database transactions across store calls
Needed once a single business action writes to several tables (a service job consuming parts, for
example). Deferred until Milestone 6 actually creates that need, so we do not invent an
abstraction before we can see its shape.

### Pagination and search on list pages
Every list page will eventually need it. Deferred until a list is long enough to hurt; adding it
to the customers list on day one would obscure the basic CRUD pattern we are trying to learn.

### Soft deletes / audit trail
A workshop probably should not hard-delete a customer with service history. Worth revisiting when
Milestone 6 gives customers a history worth preserving.

### Re-evaluate Gin against `net/http` + `ServeMux`
[ADR-004](decisions/ADR-004-gin.md) chose Gin over the standard library's Go 1.22+ routing, which
closed most of the historical gap. The honest test is real code: at Milestone 9, count what Gin
actually does for us across six modules and decide whether the dependency still earns its place.
Deferred because the answer is unknowable before the handlers exist.

### CI pipeline (build, vet, test, lint)
Valuable, but it is a DevOps/repository-configuration task and the test suite barely exists yet.
Revisit around Milestone 9.

### File uploads (vehicle photos, invoice PDFs)
A plausible real-world feature, out of scope for the core domain.
