-- +goose Up
ALTER TABLE order_items
ADD COLUMN subtotal_in_cents INTEGER GENERATED ALWAYS AS (quantity * price_in_cents) STORED;

-- +goose Down
ALTER TABLE order_items
DROP COLUMN subtotal_in_cents;