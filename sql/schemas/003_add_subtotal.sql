-- +goose Up
ALTER TABLE order_items
ADD subtotal_in_cents AS (quantity * price_in_cents) PERSISTED;

-- +goose Down
ALTER TABLE order_items
DROP COLUMN subtotal_in_cents;