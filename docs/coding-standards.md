# Coding standards

Conventions this codebase follows. Naming rules live in
[project-structure.md](project-structure.md); layer responsibilities live in
[architecture.md](architecture.md). This file covers how the code inside them is written.

The guiding rule, from `task.md`: **code must be self-explanatory. Do not over-engineer.** Simple
first, with good bones, refactored toward production-ready later.

## Go

### Errors are returned, wrapped, and handled once

```go
if err != nil {
    return fmt.Errorf("create customer: %w", err)
}
```

Wrap with `%w` and a short phrase saying what was being attempted. By the time an error reaches a
log line it describes its own path: `create customer: insert row: duplicate key`.

- **Never panic** for an expected failure. Panics are for programmer errors — a nil map, an
  impossible switch case.
- **Only the handler turns an error into a status code**, because it is the only layer that knows
  what HTTP is.
- **Never discard an error silently.** If ignoring one is genuinely correct, say so with `_ =` and
  a comment explaining why — see `render` in `internal/server/render.go`.

### Accept interfaces, return structs

A function takes the narrowest type it needs and returns a concrete one. But **do not add an
interface until a second implementation or a test double needs it.** Go's convention is to
introduce abstraction on demand, not in advance.

### Dependencies arrive through a struct, never a global

Handlers are methods on `Server`, so they reach config and the database pool through the receiver.
Package-level `var db *pgxpool.Pool` is what makes a codebase untestable — tests then have to set up
and tear down global state and can never run in parallel.

### Constructors are `New`, and they validate

`New` or `NewThing`, returning the value (or a pointer plus an error if it can fail). Validate
inputs there rather than trusting every call site.

### Keep `main` small

`main.go` reads config, loads the asset manifest, builds the server, runs it, and reports a fatal
error. It should read as a summary of how the application is assembled.

### Comments explain why, not what

The code says what it does. A comment earns its place by explaining a decision, a constraint, or a
trap:

```go
// A fresh context, because ctx is already cancelled — that cancellation is
// what got us here, and reusing it would abort the shutdown immediately.
```

Every exported identifier gets a doc comment starting with its name. Packages get a `// Package x
...` comment on one file.

### Formatting is not a matter of opinion

`make check` runs `go fmt` and rewrites files. Do not argue with it. `go vet` runs in the same
command and catches what the compiler allows — a `Printf` whose arguments do not match its format
string, a struct tag typo, a discarded result.

## Tests

- **Table-driven by default.** Cases are data in a slice; the assertion logic is written once. See
  `internal/config/config_test.go`.
- **Name the behaviour, not the function.** `"rejects a port outside the valid range"` beats
  `"TestLoad2"`. Subtest names appear in failure output and become the bug report.
- **Set all the state a test depends on**, including the variables it does not care about. A test
  that reads whatever happens to be in your shell passes or fails depending on whose machine it runs
  on.
- **Assert on landmarks, not whole documents.** An HTML test that pins every byte fails on any
  styling change, which trains you to update it without reading it. At that point it tests that the
  file did not change, not that the page works.
- **`t.Helper()` in helpers**, so failures point at the calling test rather than the helper.
- **Say what is deliberately untested and why**, in a comment where the test would have been.

## Templ

- **One component per concern.** A page, a layout, a row, a form field. Components are Go functions;
  compose them.
- **`Layout` owns everything shared** — `<head>`, stylesheet, scripts, header. A page never repeats
  them.
- **Pass typed parameters**, not `map[string]any`. Type safety is the reason for choosing Templ over
  `html/template`; a map throws it away.
- **Generated `*_templ.go` files are committed and never hand-edited.** `make templ` overwrites
  them, and it runs automatically before `run`, `build`, and `test`.
- **A fragment is a component rendered without a layout around it.** That is the whole trick behind
  HTMX here; nothing special is required.
- **All HTML goes through `render`** in `internal/server/render.go`, so `Content-Type` is set in one
  place.

## HTMX and Alpine.js

The order to reach for things:

1. **A plain link or form.** If a normal `GET` or `POST` does the job, use it. It works without
   JavaScript and needs no explanation.
2. **HTMX**, when only part of the page should change — a list row, a validation message, a search
   result. State stays on the server, which is the reason this stack was chosen.
3. **Alpine.js**, only for state that has no business on the server: a dropdown being open, a modal
   being visible, a confirm toggle.

If something needs more than Alpine, that is a signal to reconsider the design of that one page —
not to add a framework.

Use `x-on:click` rather than Alpine's `@click` shorthand: `@` is templ's own syntax for rendering a
component.

Both libraries are npm dependencies bundled by Vite from `web/src/app.js`, not CDN script tags, so
the browser loads one JavaScript file from our own server.

## SQL

Conventions are fixed at Milestone 2 in `docs/database.md`. Two are already settled:

- Migrations in `sql/migrations/` are **append-only forever**. Once one has run anywhere, you write a
  new migration rather than editing it. Every `.up.sql` gets a `.down.sql` written at the same time.
- Hand-written SQL in `sql/queries/`, with sqlc generating the Go. Never hand-edit `internal/db/`.

## URLs

Settled in [api.md](api.md): plural lowercase resource paths, nested paths reflecting ownership,
and `POST` rather than `PUT`/`PATCH` for updates — because HTML forms only support `GET` and `POST`,
and staying inside that keeps the application working without JavaScript.
