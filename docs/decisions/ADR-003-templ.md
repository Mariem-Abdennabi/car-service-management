# ADR-003 — Templ for HTML, with HTMX and Alpine.js

- **Status:** Accepted
- **Date:** 2026-07-28

## Context

The application renders HTML on the server (see [ADR-001](ADR-001-project-architecture.md)). It
needs a way to produce that HTML: a page layout, reusable pieces (a customer row, a form field, a
status badge), and — because HTMX swaps parts of a page — the ability to render a *fragment* on its
own, not just a whole document.

Go ships with `html/template` in the standard library, which would do the job. The question was
whether a code-generating alternative earns its extra step.

## Decision

Use **[Templ](https://templ.guide)** for HTML. Templates are `.templ` files that compile to Go
functions; `templ generate` produces a `*_templ.go` file beside each template.

```
templ CustomerRow(c Customer) {
    <tr><td>{ c.Name }</td><td>{ c.Phone }</td></tr>
}
```

becomes an ordinary Go function you call as `CustomerRow(c).Render(ctx, w)`.

Add **HTMX** for interactivity: HTML attributes that issue requests and swap the returned HTML
into the page, so a search-as-you-type box or an inline delete needs no JavaScript we write.

Add **Alpine.js** only for genuinely client-side state — a dropdown that opens, a modal that
closes, a confirm toggle — where a server round-trip would be absurd. It stays unused until a
feature needs it.

Generated `*_templ.go` files are committed, so `go build ./...` works with only the Go toolchain.

## Consequences

**Good**

- **Type-safe templates.** A template takes typed parameters. Passing the wrong struct, or reading
  a field that doesn't exist, is a compile error. With `html/template` the same mistake is a
  runtime error on the page — and often a silently empty value, which is worse.
- **Editor support that actually works.** Autocomplete and go-to-definition inside markup, because
  by the time the compiler sees it, it *is* Go.
- **Components are just functions.** Composition, parameters, and defaults use language features
  already being learned instead of a separate template mini-language (`{{define}}`,
  `{{template}}`, `{{block}}`).
- **Fragments are free.** Rendering one row for HTMX means calling one function. This is the
  property that makes the HTMX approach pleasant rather than fiddly.
- **Contextual escaping.** Templ escapes by default and understands HTML context, so
  interpolated values don't become an XSS hole.
- **HTMX keeps state on the server.** One source of truth, no client-side store to reconcile —
  which is the entire reason this stack was chosen over an SPA.

**Bad, and accepted**

- **Another generation step**, alongside sqlc's. Edit a `.templ`, run `templ generate`. Both go
  behind one `make` target, and `templ generate --watch` handles the inner loop during
  development.
- **Committed generated code inflates diffs.** A one-line template change shows up as a larger
  generated-file change. Accepted for the guarantee that the repository always compiles.
- **Smaller ecosystem** than the standard library. Fewer examples, and it's a younger project
  (v0.3.x) that has made breaking changes before. Mitigated by pinning the version and by the fact
  that a `.templ` file is close enough to HTML that migrating away would be mechanical.
- **HTMX has limits.** Interactions that need optimistic updates or heavy local state fight it.
  Alpine.js is the designated escape hatch; if something ever needs more than Alpine, that's a
  signal to reconsider for that page only, not for the application.

## Alternatives considered

**`html/template` (standard library).** Zero dependencies, zero build steps, and it can render
fragments via named templates. The strongest alternative, and it would be the right call for a
handful of pages. Rejected because runtime-only checking is a poor fit for a project meant to
teach how Go's type system helps you: a renamed struct field should break the build, not the
customer list page. Its `{{...}}` language is also a second thing to learn that transfers nowhere.

**A JavaScript framework (React/Vue/Svelte) against a JSON API.** Rejected in ADR-001 — it would
mean a second language, a second build system, duplicated validation, and an API contract to keep
in sync, in exchange for interactivity this domain doesn't need.

**Go + htmx with plain string concatenation.** Some small projects build HTML with `fmt.Sprintf`.
Fast to start, and an XSS vulnerability waiting to happen. Not seriously considered.

**gomponents.** HTML as pure Go function calls, no codegen, no new file type. Type-safe and
appealingly small. Rejected because the markup stops looking like markup — nested `Div(Class(...),
Table(...))` calls are harder to read against a designer's HTML, and Tailwind class strings are
easier to manage in something that resembles HTML.
