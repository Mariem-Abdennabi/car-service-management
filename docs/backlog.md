# Backlog

Ideas worth doing **later**. Nothing here is implemented, and nothing here may be implemented
without first being promoted onto [the roadmap](roadmap.md) and approved.

The point of this file is to stop good ideas from derailing the current milestone.

## Format

Each entry states the idea, why it might be worth it, and why it is not now.

---

## Deferred

### Stock movements instead of a stored quantity
`parts.quantity_on_hand` is a single number edited directly. Real inventory records *movements* —
received 20, consumed 2 on job 41 — and derives the total, which gives you a history, an audit trail,
and correct behaviour when two people adjust stock at the same moment.

Deferred because nothing consumes parts yet. Revisit at Milestone 6, when a repair job takes parts out
of stock: that is the point where "why is this number wrong?" becomes a real question. The `CHECK
(quantity_on_hand >= 0)` constraint is already in place, so the database will refuse to go negative in
the meantime.

**Re-examined at the Milestone 6 design and deferred again.** Jobs do consume stock now, so the
trigger fired — but step 6d already introduces the project's first transaction, and a movements table
would mean introducing a derived quantity in the same breath. Two new concepts in one step teaches
neither. Revisit at Milestone 9, or sooner if a stock number is ever visibly wrong.

### Itemised labour on a job
`service_jobs.labour_millimes` is a single charge for the whole job. A real invoice usually prints
labour as lines — "diagnostics 1h, brake replacement 2h" — each with hours and a rate.

Deferred to Milestone 7, where printing an invoice makes the difference visible. Until something reads
the breakdown, a second table stores detail nothing displays.

### Prices that change over time
A part has one price. Invoicing a job from six months ago should use the price *then*, not now, which
means either price history or copying the price onto the invoice line at the time. Deferred to
Milestone 7, where invoices make the question concrete — copying onto the line is likely the simpler
correct answer.

The Milestone 6 design deliberately leaves a price off `job_parts` for this reason: the frozen price
belongs wherever the invoice is written, and a column nothing reads is a column that drifts.

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

**Promoted.** Milestone 6 created exactly that need, and step 6d is where it lands — one transaction
around "insert the line, lower the stock", written where it is used rather than as a general
abstraction. Moves to *Promoted and built* once 6d is done.

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

### A separate database for the tests
The tests run against the same database as `make run` and `make seed`, so a test fixture and a demo
row can collide, and a row left behind by a failed test is still there on the next run. Each test
cleans up after itself, which is enough while that stays true — the honest fix is a database the tests
own, created and dropped around the run.

Deferred because it is environment work: another database to create, another URL to configure, and a
decision about whether it lives in the same container. Revisit at Milestone 9, alongside CI, which
needs the same thing.

### CI pipeline (build, vet, test, lint)
Valuable, but it is a DevOps/repository-configuration task and the test suite barely exists yet.
Revisit around Milestone 9.

### File uploads (vehicle photos, invoice PDFs)
A plausible real-world feature, out of scope for the core domain.

---

## Promoted and built

Kept so the file does not re-suggest something that already exists.

### A `seed` command for demo data
Built at step 5d as `go run . -seed`, wrapped as `make seed`. See
[roadmap.md](roadmap.md#milestone-5--spare-parts-and-stock-).
