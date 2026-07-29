# ADR-002 — sqlc for database access, pgx/v5 as the driver

- **Status:** Accepted
- **Date:** 2026-07-28

## Context

The application needs to read and write PostgreSQL from Go. Go offers a wide spectrum here, from
`database/sql` with hand-written scanning at one end to full ORMs at the other. The choice affects
every feature we build, and it also affects what the project teaches: an ORM hides SQL, while raw
`database/sql` shows it at the cost of a lot of repetitive scanning code.

## Decision

**Write SQL by hand, and let [sqlc](https://sqlc.dev) generate the Go code from it.**
**Use [pgx/v5](https://github.com/jackc/pgx) as the PostgreSQL driver**, via its native interface
and connection pool rather than through `database/sql`.

The workflow: write a query in `db/queries/*.sql` with a comment naming it and its result shape,
run `sqlc generate`, and get a type-safe Go method in `internal/db/`.

```sql
-- name: GetCustomer :one
SELECT * FROM customers WHERE id = $1;
```

becomes a `GetCustomer(ctx, id) (Customer, error)` method, with the `Customer` struct derived from
the actual schema.

Generated code is committed to the repository and never hand-edited.

## Consequences

**Good**

- **SQL stays visible.** You learn PostgreSQL, not an ORM's query DSL. When something is slow you
  can read the query and run `EXPLAIN` on it.
- **Errors surface at compile time.** sqlc reads the migrations to learn the schema, so a renamed
  column or a wrong type breaks `sqlc generate` or the build — not a request in production. This
  is the single biggest argument for it over `sqlx` or an ORM.
- **No scanning boilerplate.** The repetitive, error-prone `rows.Scan(&a, &b, &c)` code is
  generated, in the right order, every time.
- **No hidden queries.** ORMs are prone to lazy-loading N+1 problems that are invisible in the Go
  code. Here, one query in the file is one query against the database.
- **pgx** supports PostgreSQL's real type system (arrays, `jsonb`, `numeric`, timestamps with
  proper time zone handling) instead of flattening everything through `database/sql`'s narrower
  interface, and its pool is well suited to a long-running server.

**Bad, and accepted**

- **A generation step.** Change the SQL, re-run `sqlc generate`, or the Go code is stale. This
  goes in the `Makefile` so it is one command, and it's a good habit to learn.
- **sqlc must be installed** to change queries. Already present (v1.31.1), and committing the
  generated code means anyone can still *build* the project with only Go.
- **Dynamic queries are awkward.** sqlc generates one function per static query, so filters that
  vary at runtime (optional search terms, sortable columns) don't fit neatly. When we hit that —
  probably on a list page — the answer is either a few explicit query variants or one hand-written
  query built with pgx directly. We'll decide it when it happens rather than guess now.
- **`SELECT *` in generated queries** couples the returned struct to column order and presence.
  Fine for single-table reads; we will name columns explicitly where it matters.

## Alternatives considered

**GORM.** The most popular Go ORM. Fastest for a first CRUD screen, and it handles migrations and
associations. Rejected on two grounds: it hides the SQL, which defeats a stated goal of learning
backend development properly; and its runtime-tag-driven API turns schema mistakes into runtime
errors, giving up exactly the compile-time safety that makes Go pleasant. Its automigration also
tends to diverge from a real migration history.

**Ent.** Schema-as-Go-code with strong generated types and graph traversal. Genuinely good, but a
large conceptual surface (its own schema DSL, codegen model, and query builder) and it hides SQL
much like an ORM. Too much to learn alongside Go itself.

**sqlx.** Thin helper over `database/sql`: struct scanning, named parameters, no codegen. Close
second, and simpler than sqlc. Rejected because its mapping is checked at runtime — a typo in a
`db:"..."` tag is an error on the request, not at build time — and it still leaves you writing the
struct definitions sqlc derives for free.

**Raw `database/sql` / pgx only.** Maximum control, maximum boilerplate. Every query means a
struct, a `Scan` call in the right order, and `rows.Err()` handling. Educational once, tedious
forever, and each hand-written `Scan` is a chance to swap two columns of the same type.
