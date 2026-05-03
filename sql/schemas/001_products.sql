-- +goose Up
CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price_in_cents INTEGER NOT NULL CHECK (price_in_cents >= 0),
    quantity INTEGER NOT NULL CHECK (quantity >= 0) DEFAULT 0,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE products;