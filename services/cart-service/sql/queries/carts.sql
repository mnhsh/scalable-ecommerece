-- name: GetCartByUserID :one
SELECT id, user_id, created_at, updated_at
FROM carts
WHERE user_id = $1;

-- name: CreateCart :one
INSERT INTO carts (
  user_id,
  created_at,
  updated_at
) VALUES (
    $1,
    now(),
    now()
) RETURNING *;

-- name: GetCartItems :many
SELECT id, cart_id, product_id, quantity, price_cents, created_at, updated_at
FROM cart_items
WHERE cart_id = $1;

-- name: GetCartItemByID :one
SELECT id, cart_id, product_id, quantity, price_cents, created_at, updated_at
FROM cart_items
WHERE id = $1;

-- name: AddCartItem :one
INSERT INTO cart_items (
  cart_id,
  product_id,
  quantity,
  price_cents,
  created_at,
  updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    now(),
    now()
) RETURNING *;

-- name: UpdateCartItemQuantity :one
UPDATE cart_items
SET
  quantity = $2,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteCartItem :exec
DELETE FROM cart_items
WHERE id = $1;

-- name: ClearCart :exec
DELETE FROM cart_items
WHERE cart_id = $1;

-- name: DeleteCart :exec
DELETE FROM carts
WHERE id = $1;

-- name: GetCartItemByProductID :one
SELECT id, cart_id, product_id, quantity, price_cents, created_at, updated_at
FROM cart_items
WHERE cart_id = $1 AND product_id = $2;

