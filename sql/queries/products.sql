-- name: GetProducts :many
SELECT id, name, price_cents, stock
FROM products
WHERE is_active = true
ORDER BY created_at DESC;

-- name: GetProductByID :one
SELECT id, name, price_cents, stock
FROM products
WHERE is_active = true AND id = $1;

-- name: CreateProduct :one
INSERT INTO products (
  id,
  created_at,
  updated_at,
  name,
  description,
  price_cents,
  stock,
  is_active
) VALUES (
    gen_random_uuid(),
    now(),
    now(),
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING *;


