-- name: Counts :one
-- One round trip for the home page rather than four.
SELECT
    (SELECT COUNT(*) FROM customers)                          AS customers,
    (SELECT COUNT(*) FROM vehicles)                           AS vehicles,
    (SELECT COUNT(*) FROM parts)                              AS parts,
    (SELECT COUNT(*) FROM parts WHERE quantity_on_hand = 0)   AS out_of_stock;

-- name: ListPartsOutOfStock :many
SELECT * FROM parts
WHERE quantity_on_hand = 0
ORDER BY name
LIMIT 5;

-- name: ListRecentCustomers :many
SELECT * FROM customers
ORDER BY created_at DESC, id DESC
LIMIT 5;
