-- name: CreateOrder :one
INSERT INTO orders (customer_id)
VALUES ($1
) RETURNING *;


-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, quantity, price_in_cents)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOrders :many
SELECT o.id, o.customer_id, o.created_at,
       json_agg(json_build_object(
           'id', oi.id,
           'product_id', oi.product_id,
           'quantity', oi.quantity,
           'price_in_cents', oi.price_in_cents
       )) AS items
FROM orders o
JOIN order_items oi ON o.id = oi.order_id
GROUP BY o.id;

-- name: GetOrderByID :one
SELECT o.id, o.customer_id, o.created_at,
       json_agg(json_build_object(
           'id', oi.id,
           'product_id', oi.product_id,
           'quantity', oi.quantity,
           'price_in_cents', oi.price_in_cents
       )) AS items
FROM orders o
JOIN order_items oi ON o.id = oi.order_id
WHERE o.id = $1
GROUP BY o.id;

-- name: UpdateProductStock :exec
UPDATE products
SET quantity = $2
WHERE id = $1;  

