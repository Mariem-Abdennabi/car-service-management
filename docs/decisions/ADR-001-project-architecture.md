# ADR-001 — Layered server-rendered monolith

- **Status:** Accepted
- **Date:** 2026-07-28

## Context

We are building a Car Service & Spare Parts Management System: one workshop's staff managing
customers, vehicles, parts, repair jobs, and invoices. Expected load is a handful of concurrent
users. The project also has to teach Go and backend development, so the structure must be
readable by someone learning the language.

The shape of the application had to be decided before any code, because it determines package
layout, and package layout is expensive to change once imports exist.

## Decision

A **single Go binary**, a **single Go module**, serving **server-rendered HTML**, organised in
**layers** — handler → service → store → database — with dependencies pointing only
downward.

Application packages live under `internal/`, so the compiler prevents anything outside the module
from importing them. Dependencies are wired by hand in `main.go` into a `Server` struct; there is
no dependency-injection framework. Layers are created when a feature needs them, not upfront.

## Consequences

**Good**

- One process to build, run, and debug. A stack trace covers the whole request.
- No network boundary inside the application, so no serialization, no retries, no partial
  failures between components.
- The layer boundaries make the codebase teachable: each file answers one kind of question.
- Testable without infrastructure — services take plain Go types, so they need no HTTP server.
- Server-rendered HTML means one language for logic and one place where state lives. No API
  contract to keep in sync, no client-side store to reconcile.

**Bad, and accepted**

- The whole application scales as one unit. Irrelevant at this size; if a piece ever needs to
  scale independently, the service layer is the natural seam to split on.
- Layers add indirection. A three-column CRUD form touches four files. The payoff appears once
  business rules exist — and this domain has real ones (stock levels, job statuses, invoice
  totals).
- Rich client interactivity is harder than in a JavaScript framework. HTMX covers the interactions
  this domain actually needs; Alpine.js is available for the rest.

## Alternatives considered

**Flat structure — everything in `main` or one `app` package.** Genuinely simpler for the first
week, and tempting given the "do not over-engineer" instruction. Rejected because this project
spans nine milestones and six domain modules; by Milestone 6 a flat package would be a single
directory of dozens of unrelated files with no rule about what may call what. The layers are the
one piece of structure that pays for itself at that size.

**Clean/hexagonal architecture with interfaces at every boundary.** Ports, adapters, and a domain
layer that depends on nothing. Rejected as over-engineering for one workshop's application: it
would mean writing an interface and a mapping struct for every entity before the first page
renders, and it's the opposite of the brief's "self-explanatory code".

**Go JSON API + separate SPA (React/Vue).** The mainstream choice, and the right one if a mobile
client or a public API were planned. Neither is. It would mean two build systems, two languages,
duplicated validation, and an API contract to version — all to render forms and tables that HTML
already renders. The chosen stack (Templ + HTMX) exists precisely to avoid that cost.

**Microservices.** No. One team member, one deployment, no independent scaling needs.
