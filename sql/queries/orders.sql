-- name: CreateOrder :one
INSERT INTO orders (customer_id)
VALUES ($1
) RETURNING *;


-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, quantity, price_in_cents)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOrders :many
SELECT * FROM orders
ORDER BY created_at DESC;

-- name: GetOrderItemsByOrderID :many
SELECT oi.*, p.name AS product_name 
FROM order_items oi
JOIN products p ON oi.product_id = p.id
WHERE oi.order_id = $1;

-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id = $1;

-- name: UpdateProductStock :exec
UPDATE products
SET quantity = $2, updated_at = NOW()
WHERE id = $1;  

