-- name: CreateOrder :one
INSERT INTO orders (
  user_id,
  status,
  total_cents
) VALUES (
  $1,
  $2,
  $3
) RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders WHERE id = $1;

-- name: GetOrdersByUserID :many
SELECT * FROM orders WHERE user_id = $1 ORDER BY created_at DESC;

-- name: UpdateOrderStatus :one
UPDATE orders SET status = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, quantity, price_cents)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOrderItems :many
SELECT * FROM order_items WHERE order_id = $1;
