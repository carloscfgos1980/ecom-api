-- name: GetProducts :many
SELECT *
FROM products;

-- name: GetProductByID :one
SELECT *
FROM products
WHERE id = $1;
