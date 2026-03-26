-- name: CreateOrder :one
INSERT INTO orders (customer_id)
VALUES ($1
) RETURNING *;


-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, quantity, price_in_cents)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOrders :many
SELECT o.*, oi.id AS order_item_id, oi.product_id, oi.quantity, oi.price_in_cents, oi.subtotal_in_cents, p.name AS product_name
FROM orders o
LEFT JOIN order_items oi ON o.id = oi.order_id
LEFT JOIN products p ON oi.product_id = p.id
ORDER BY o.created_at DESC;

-- name: GetOrderItemsByOrderID :many
SELECT oi.*, p.name AS product_name 
FROM order_items oi
JOIN products p ON oi.product_id = p.id
WHERE oi.order_id = $1;

-- name: GetOrderByID :one
SELECT o.*, oi.id AS order_item_id, oi.product_id, oi.quantity, oi.price_in_cents, oi.subtotal_in_cents, p.name AS product_name
FROM orders o
LEFT JOIN order_items oi ON o.id = oi.order_id
LEFT JOIN products p ON oi.product_id = p.id
WHERE o.id = $1;

-- name: UpdateProductStock :exec
UPDATE products
SET quantity = $2, updated_at = NOW()
WHERE id = $1;  

