# ADR-004 — Gin as the HTTP framework

- **Status:** Accepted
- **Date:** 2026-07-28

## Context

The application needs to route HTTP requests, read path and form parameters, group routes under
shared middleware, and write responses. Go's standard library can do all of this, and since Go
1.22 `net/http`'s `ServeMux` also supports method-based patterns and path wildcards
(`GET /customers/{id}`) — which removes the historical reason most projects reached for a router
at all.

So the question was a real one, not a formality.

## Decision

Use **[Gin](https://gin-gonic.com)**.

Handlers are methods on our `Server` struct with the signature `func(c *gin.Context)`, so they can
reach configuration and the database pool without package-level globals. Routes are registered in
one file under `internal/server/`. Gin's `Recovery` and `Logger` middleware are enabled from the
start.

`*gin.Context` stays in the handler layer. Services take plain Go types — passing a context object
downward would weld the business rules to the web framework and break the dependency rule in
[ADR-001](ADR-001-project-architecture.md).

## Consequences

**Good**

- **Route grouping and middleware are built in.** `r.Group("/customers")` with shared middleware
  is one line. This matters at Milestone 8, when authentication has to wrap most routes but not
  the login page.
- **Parameter handling without ceremony.** `c.Param("id")`, `c.Query("q")`, `c.PostForm("name")`,
  and form binding into a struct are all provided. With the standard library these are several
  lines each, repeated per handler.
- **`Recovery` middleware** means a panic in one handler returns 500 for that request instead of
  killing the process — worth having from day one while the code is new.
- **Enormous ecosystem and documentation.** For someone learning, the odds that a given question
  already has a clear answer are as high as for any Go library. That is a genuine feature, not a
  popularity argument.
- **Fits Templ cleanly.** We write to `c.Writer` and call a Templ component's `Render` — no
  fighting Gin's own HTML rendering.

**Bad, and accepted**

- **A dependency where the standard library would nearly do.** Honest trade-off. Accepted because
  the middleware and parameter ergonomics save real code on every one of the six domain modules.
- **`*gin.Context` is a framework type in our handler signatures.** Migrating away later would mean
  touching every handler. Contained by keeping it strictly at the handler layer: services and
  repositories never see it, so at most one layer would change.
- **Gin's `c.JSON`, binding tags, and validator** are features we mostly won't use, since we render
  HTML. Unused surface area to ignore.
- **Its `Context` doubles as a value bag** (`c.Set`/`c.Get`), which is untyped and easy to abuse.
  We will use it only where middleware genuinely must pass something down, such as the current
  user at Milestone 8.

## Alternatives considered

**`net/http` with Go 1.22+ `ServeMux`.** Zero dependencies, and the routing gap that once made a
framework mandatory is closed. The strongest alternative by a distance. Rejected because you still
hand-roll middleware chaining, form parsing, and error-response helpers — and a beginner rolling
their own middleware chain learns about closures instead of about the domain. Worth revisiting at
Milestone 9 with real code to measure the dependency against; noted in the backlog rather than
pretended away.

**Chi.** A thin router that composes with `net/http` idiomatically — handlers stay
`http.HandlerFunc`, so nothing framework-specific leaks into signatures. Arguably the most
tasteful choice, and the one that would age best. Rejected narrowly: it provides less out of the
box than Gin (no form binding, fewer response helpers), and Gin's larger body of tutorials is
worth more to someone learning than Chi's cleaner interface signature.

**Echo.** Very close to Gin in features and design, with its own context type. No decisive
advantage either way; Gin's larger community broke the tie.

**Fiber.** Fast, Express-like, but built on `fasthttp` rather than `net/http`, which means it is
incompatible with the standard `http.Handler` ecosystem. Rejected: stepping outside `net/http`
costs access to the wider Go ecosystem, and the performance argument is irrelevant for one
workshop's internal application.
