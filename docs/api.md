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

## Customers

| Method | Path | Purpose | |
| ------ | ---- | ------- | - |
| `GET` | `/customers` | list | ✅ |
| `GET` | `/customers/:id` | detail | ✅ |
| `GET` | `/customers/new` | the create form | ✅ |
| `POST` | `/customers` | create | ✅ |
| `GET` | `/customers/:id/edit` | the edit form | ✅ |
| `POST` | `/customers/:id` | update | ✅ |
| `POST` | `/customers/:id/delete` | delete | ✅ |

Every write is a `POST`, because HTML forms only support `GET` and `POST`. Deleting could have been
`DELETE /customers/:id` with htmx issuing it, but that route would only work with JavaScript enabled —
a second route for the same action. One `POST` that a plain form can reach is simpler, and htmx can
still enhance it.

### `GET /customers`

The list, newest first. Optional `?q=` filters by name, case-insensitively and on any part of the
name. An empty `q` returns everyone.

**Two response shapes from one handler:**

| Request | Response |
| ------- | -------- |
| normal | the full page |
| `HX-Request: true` | just the `#customer-list` element |

htmx sets that header, so searching needs no second route. The fragment is the same component the page
renders, called without a layout around it — about 700 bytes against 2 KB for the page.

The fragment carries the `id="customer-list"` that htmx targets. It has to: the swap is `outerHTML`, so
a response without the id would replace the element with something the next search cannot find.

### `GET /customers/:id`

One customer, with their vehicles listed underneath. **404** when no row has that id, and also when the id is not a number — `/customers/abc`
is simply not a page, so there is nothing to report as a bad request.

### `GET /customers/new`

The create form.

### `POST /customers`

Creates a customer from `name`, `phone`, and `city`. All three are required, and values are trimmed, so
a field of spaces counts as empty.

- **303 See Other** to `/customers/:id` on success. A redirect rather than a rendered page, so a
  refresh cannot submit the form twice.
- **422 Unprocessable Content** on a validation failure, re-rendering the form with the submitted
  values and one message per bad field. Not 400: the request was well formed, its contents were not.

### `GET /customers/:id/edit`

The edit form, pre-filled with the customer's current details. **404** for an unknown id.

### `POST /customers/:id`

Updates a customer. Same validation and the same 303/422 behaviour as creating one, and a rejected
submission leaves the stored row untouched. **404** for an unknown id.

### `POST /customers/:id/delete`

Deletes a customer, then **303** to `/customers`. **404** for an unknown id.

**409 Conflict** when the customer still has vehicles: the page is redrawn with a message explaining
why, and nothing is deleted. Their vehicles have to go first.

The button on the detail page asks for confirmation first, held in Alpine. It is a real submit inside a
real form, so with JavaScript off the first click deletes — the page keeps working, it just loses the
extra question.

## Vehicles

Vehicles are always reached through their customer, because a vehicle without one does not exist in
this domain — the foreign key is `NOT NULL`.

| Method | Path | Purpose | |
| ------ | ---- | ------- | - |
| `GET` | `/customers/:id/vehicles/new` | the create form | ✅ |
| `POST` | `/customers/:id/vehicles` | create | ✅ |
| `GET` | `/vehicles/:id/edit` | the edit form | ✅ |
| `POST` | `/vehicles/:id` | update | ✅ |
| `POST` | `/vehicles/:id/delete` | delete | ✅ |

Creating is nested under the customer, since that is where the vehicle's owner comes from. Editing and
deleting are not: a vehicle id is already unique, so repeating the customer in the path would allow
`/customers/1/vehicles/2` where vehicle 2 belongs to customer 3 — a mismatch to validate rather than a
URL to support. The owner is read from the vehicle instead.

### `POST /customers/:id/vehicles`

Creates a vehicle from `plate`, `make`, `model`, and `year`. All required. **303** to the customer's
page on success, **422** with the form redrawn on a validation failure, **404** if the customer does
not exist.

`year` must be a number between 1900 and next year — next year because new models are sold before the
year they are named for. The input is `type="number"`, but the server validates regardless: nothing
stops a request arriving without a browser.

### `GET /vehicles/:id/edit`

The edit form, pre-filled. The customer's name is shown and links back, read from the vehicle rather
than the URL. **404** for an unknown id.

### `POST /vehicles/:id`

Updates a vehicle. Same validation as creating one. **303** to the owner's page, **422** with the form
redrawn, **404** for an unknown id.

### `POST /vehicles/:id/delete`

Deletes a vehicle, then **303** to the owner's page. Asks for confirmation first, in Alpine, the same
way deleting a customer does.

## Spare parts

| Method | Path | Purpose | |
| ------ | ---- | ------- | - |
| `GET` | `/parts` | the catalogue | ✅ |
| `GET` | `/parts/new` | the create form | ✅ |
| `POST` | `/parts` | create | ✅ |
| `GET` | `/parts/:id/edit` | the edit form | ✅ |
| `POST` | `/parts/:id` | update | ✅ |
| `POST` | `/parts/:id/delete` | delete | ✅ |

### `GET /parts`

The catalogue, by name. Prices are stored as millimes and shown as dinars — `42500` renders as
`42.500 TND`. A part with nothing in stock is marked rather than showing a bare `0`, since that is the
difference between "we can fit this today" and "we cannot".

### `POST /parts` and `POST /parts/:id`

Price is typed in dinars (`42.500`) and stored as millimes. More than three decimal places is
rejected rather than rounded. `reference` is unique: a clash comes back as **422** with a message on
that field, not a 500.

## Errors

| Response | When |
| -------- | ---- |
| `404` HTML page | unknown URL, unknown record, or a malformed id |
| `409` HTML page | the action conflicts with the current state — deleting a customer who has vehicles |
| `500` plain text | anything unexpected; the error is logged, never sent to the browser |

The 404 page is rendered for unmatched routes too, so a mistyped URL and a missing record look the
same to a visitor. The 500 is deliberately plain text rather than a rendered page: that path means
something is already broken, and rendering a template could fail again.

## Conventions

- **Unknown path** → `404`, handled by Gin's default.
- **Resource paths are plural and lowercase** — `/customers`, `/customers/:id`.
- **Nested resources** reflect ownership — `/customers/:id/vehicles`.
