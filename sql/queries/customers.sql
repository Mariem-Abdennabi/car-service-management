-- name: ListCustomers :many
SELECT * FROM customers
ORDER BY created_at DESC, id DESC;

-- name: GetCustomer :one
SELECT * FROM customers
WHERE id = $1;

-- name: CreateCustomer :one
INSERT INTO customers (name, phone, city)
VALUES ($1, $2, $3)
RETURNING *;
