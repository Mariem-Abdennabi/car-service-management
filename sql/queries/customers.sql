-- name: ListCustomers :many
-- An empty search returns everything, so one query serves both the full list and a
-- filtered one. ILIKE is case-insensitive.
SELECT * FROM customers
WHERE sqlc.arg(search)::text = '' OR name ILIKE '%' || sqlc.arg(search)::text || '%'
ORDER BY created_at DESC, id DESC;

-- name: GetCustomer :one
SELECT * FROM customers
WHERE id = $1;

-- name: CreateCustomer :one
INSERT INTO customers (name, phone, city)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateCustomer :one
UPDATE customers
SET name = $2, phone = $3, city = $4
WHERE id = $1
RETURNING *;

-- name: DeleteCustomer :exec
DELETE FROM customers
WHERE id = $1;
