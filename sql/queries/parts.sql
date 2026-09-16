-- name: ListParts :many
SELECT * FROM parts
ORDER BY name, reference;

-- name: GetPart :one
SELECT * FROM parts
WHERE id = $1;

-- name: CreatePart :one
INSERT INTO parts (reference, name, price_millimes, quantity_on_hand)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdatePart :one
UPDATE parts
SET reference = $2, name = $3, price_millimes = $4, quantity_on_hand = $5
WHERE id = $1
RETURNING *;

-- name: DeletePart :exec
DELETE FROM parts
WHERE id = $1;
