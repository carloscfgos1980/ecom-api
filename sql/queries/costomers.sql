-- name: CreateCustomer :one
INSERT INTO customers (id, created_at, updated_at, email, password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
) 
RETURNING *;

-- name: GetCustomerByEmail :one
SELECT * FROM customers
WHERE email = $1;

-- name: GetCustomerByID :one
SELECT * FROM customers
WHERE id = $1;