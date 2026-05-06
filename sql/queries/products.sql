-- name: GetProducts :many
SELECT *
FROM products
ORDER BY id ASC;

-- name: GetProductByID :one
SELECT *
FROM products
WHERE id = $1;
