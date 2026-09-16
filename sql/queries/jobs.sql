-- name: ListJobs :many
-- The list page shows which car and whose it is, so the join happens here rather
-- than as one extra query per row on the page.
SELECT
    j.id,
    j.status,
    j.description,
    j.labour_millimes,
    j.opened_at,
    v.id   AS vehicle_id,
    v.plate,
    v.make,
    v.model,
    c.id   AS customer_id,
    c.name AS customer_name
FROM service_jobs j
    JOIN vehicles v ON v.id = j.vehicle_id
    JOIN customers c ON c.id = v.customer_id
ORDER BY j.opened_at DESC, j.id DESC;

-- name: GetJob :one
-- No join: the handler already loads the vehicle and its owner through the store
-- methods that exist, and three lookups by primary key are cheaper to read than a
-- fourth generated row type.
SELECT * FROM service_jobs
WHERE id = $1;

-- name: ListJobsByCustomer :many
-- A customer's jobs across all of their vehicles, for their page. The plate comes
-- along so each row says which car it was.
SELECT
    j.id,
    j.status,
    j.description,
    j.labour_millimes,
    j.opened_at,
    v.id AS vehicle_id,
    v.plate
FROM service_jobs j
    JOIN vehicles v ON v.id = j.vehicle_id
WHERE v.customer_id = $1
ORDER BY j.opened_at DESC, j.id DESC;

-- name: CreateJob :one
INSERT INTO service_jobs (vehicle_id, description, status, labour_millimes)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: DeleteJob :exec
-- No route reaches this: a job that is not happening is cancelled, not erased.
-- It exists so tests and the seeder can leave the database as they found it.
DELETE FROM service_jobs
WHERE id = $1;
