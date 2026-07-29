# API

Every URL the application answers. The authoritative list is
`internal/server/routes.go`; this file explains the intent behind each entry.

Most endpoints return **HTML**, not JSON — pages for a normal request, fragments for an HTMX
request. Anything returning JSON is noted as such.

HTML responses are written by `internal/server/render.go`, which is the single place the
`Content-Type` is set.

## Assets

### `GET /build/*`

Serves `public/build/` — the Vite bundle: one hashed JavaScript file (Tailwind-free application
code plus htmx and Alpine) and one hashed CSS file.

Filenames contain a content hash, so the templates cannot hard-code them. `assets.Load` reads
Vite's manifest at startup and the layout links whatever it reports.

The path is relative to the process's working directory, so the server must run from the repository
root. Embedding the bundle in the binary is in [backlog.md](backlog.md).

## Pages

### `GET /`

The landing page. Renders `views.Home` inside `views.Layout`.

**Response** `200 OK`, `text/html; charset=utf-8`.

It becomes a real dashboard once there is data to summarise.

## Operational

### `GET /healthz`

Reports that the process is alive and able to serve requests.

**Response** `200 OK` when the database is reachable, `503 Service Unavailable` when it is not:

```json
{"status":"ok"}
{"status":"unavailable"}
```

The error itself is logged, never returned — this endpoint is public enough that it should not leak
connection strings or hostnames.

This is the only endpoint whose audience is machines rather than people: a container runtime
deciding whether to restart the instance, a load balancer deciding whether to send it traffic, an
uptime check deciding whether to page somebody.

The distinction between the two failures matters to whoever reads it: no response at all means the
process is gone, while `503` means the process is up but not ready for traffic — a different
instruction to a load balancer.

It pings the database rather than running a query, because a server that cannot reach its database is
not usefully alive, and a health check should not be able to fail for reasons of its own.

Why `/healthz` and not `/health`: the trailing `z` is a Kubernetes-era convention that keeps
operational endpoints from colliding with a real application route. If this system ever gains a page
about vehicle health, `/health` is still free.

## Features

None yet. Customers arrive at Milestone 3, and this section grows with them.

Each feature will follow the same URL shape:

| Method | Path | Purpose |
| ------ | ---- | ------- |
| `GET` | `/customers` | list |
| `GET` | `/customers/new` | the create form |
| `POST` | `/customers` | create |
| `GET` | `/customers/:id` | detail |
| `GET` | `/customers/:id/edit` | the edit form |
| `POST` | `/customers/:id` | update |
| `DELETE` | `/customers/:id` | delete |

`POST` rather than `PUT`/`PATCH` for updates because HTML forms only support `GET` and `POST`, and
HTMX is the only thing that would issue the others. Sticking to what a plain form can do keeps the
application working without JavaScript.

## Conventions

- **Unknown path** → `404`, handled by Gin's default.
- **Resource paths are plural and lowercase** — `/customers`, `/customers/:id`.
- **Nested resources** reflect ownership — `/customers/:id/vehicles`.
