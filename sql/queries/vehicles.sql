-- name: ListVehiclesByCustomer :many
SELECT * FROM vehicles
WHERE customer_id = $1
ORDER BY created_at DESC, id DESC;

-- name: GetVehicle :one
SELECT * FROM vehicles
WHERE id = $1;

-- name: UpdateVehicle :one
UPDATE vehicles
SET plate = $2, make = $3, model = $4, year = $5
WHERE id = $1
RETURNING *;

-- name: CreateVehicle :one
INSERT INTO vehicles (customer_id, plate, make, model, year)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteVehicle :exec
DELETE FROM vehicles
WHERE id = $1;
