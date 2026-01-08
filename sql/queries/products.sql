-- name: GetProducts :many
SELECT id, name, price, stock
FROM products
WHERE is_active = true
ORDER BY created_at DESC;

-- name: GetProductByID :one
SELECT id, name, price, stock
FROM products
WHERE is_active = true AND id = $1;

