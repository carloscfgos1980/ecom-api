-- name: CreateCustomer :one
INSERT INTO users (id, created_at, updated_at, email, password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3
) 
RETURNING *;

-- name: GetCustomerByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetCustomerByID :one
SELECT * FROM users
WHERE id = $1;