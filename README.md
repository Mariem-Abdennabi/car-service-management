# Car Service & Spare Parts Management System

A management application for a car workshop: customers, their vehicles, spare-parts stock, repair
jobs, and invoices.

Built as a Go full-stack application, server-rendered, one milestone at a time.

> **Status: early development.** The server runs, renders a styled page from a Templ template, serves a
> Vite bundle, and talks to PostgreSQL through pgx with migrations and sqlc-generated queries. No
> feature screens yet — see
> [docs/roadmap.md](docs/roadmap.md) for exactly where things stand.

## Stack

**Backend** — Go · [Gin](https://gin-gonic.com) · PostgreSQL ·
[pgx/v5](https://github.com/jackc/pgx) · [sqlc](https://sqlc.dev) ·
[golang-migrate](https://github.com/golang-migrate/migrate)

**Frontend** — [Templ](https://templ.guide) · [htmx](https://htmx.org) ·
[Alpine.js](https://alpinejs.dev) · [Tailwind CSS v4](https://tailwindcss.com) ·
[Vite](https://vite.dev) · [Bun](https://bun.sh)

Pages are rendered as HTML on the server. HTMX swaps fragments of that HTML for interactivity, so
there is no separate JavaScript application and no JSON API to keep in sync. The reasoning is in
[ADR-001](docs/decisions/ADR-001-project-architecture.md).

## Domain

Six modules, built in dependency order:

| Module | Holds | Depends on |
| ------ | ----- | ---------- |
| Customers | the people who bring cars in | — |
| Vehicles | plate, make, model, VIN, mileage | Customers |
| Spare parts | catalogue and quantity on hand | — |
| Service jobs | a repair: status, labour, parts consumed | Vehicles, Spare parts |
| Invoices | line items and totals for a finished job | Service jobs |
| Users | login and roles (mechanic, manager) | — |

## Documentation

| Document | What it covers |
| -------- | -------------- |
| [recap.md](docs/recap.md) | **Start here** — the stack, every file type, and why each choice was made |
| [architecture.md](docs/architecture.md) | Layers, dependency direction, configuration, request lifecycle |
| [api.md](docs/api.md) | Every URL the application answers |
| [development-workflow.md](docs/development-workflow.md) | Daily loop, commands, definition of done, commit convention |
| [coding-standards.md](docs/coding-standards.md) | How the code is written: errors, tests, Templ, HTMX vs Alpine |
| [project-structure.md](docs/project-structure.md) | Every directory, why it exists, naming conventions |
| [roadmap.md](docs/roadmap.md) | **Every milestone and step — done, next, and why** |
| [backlog.md](docs/backlog.md) | Deferred ideas, with the reason each is deferred |
| [decisions/](docs/decisions/) | ADRs: architecture, sqlc, Templ, Gin |

`database.md` arrives with Milestone 2, which gives it content.

## Requirements

Go 1.26+, PostgreSQL 15+, Bun, and the `sqlc`, `templ`, and `migrate` command-line tools.

## Running it

```sh
cp .env.example .env             # local settings; git-ignored
cd web && bun install && cd ..   # frontend dependencies
make assets                      # build the Vite bundle
make run                         # http://localhost:8080
```

A PostgreSQL server must be reachable at the `DATABASE_URL` in `.env`. The database can be empty —
migrations run on startup.

`make` on its own lists every available command. `make check` — format, vet, test — is the one to run
before committing.

Anything exported in your shell overrides `.env`. Ctrl-C stops the server.

See [development-workflow.md](docs/development-workflow.md) for the full loop.

## How this project is built

Incrementally, and deliberately slowly. One milestone at a time, one micro-step at a time, with
the reasoning written down as it happens — the point is a working application *and* an
understanding of why it is built this way. The working agreement is in
[CLAUDE.md](CLAUDE.md); the original brief is in [task.md](task.md).
