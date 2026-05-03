-- +goose Up
ALTER TABLE order_items DROP COLUMN subtotal_in_cents;

ALTER TABLE order_items RENAME COLUMN price_in_cents TO price;
ALTER TABLE order_items
  ALTER COLUMN price TYPE double precision
  USING price / 100.00;

ALTER TABLE order_items
  ADD COLUMN subtotal double precision GENERATED ALWAYS AS (quantity * price) STORED;

-- +goose Down
ALTER TABLE order_items DROP COLUMN subtotal;

ALTER TABLE order_items
  ALTER COLUMN price TYPE INTEGER
  USING (price * 100)::integer;
ALTER TABLE order_items RENAME COLUMN price TO price_in_cents;

ALTER TABLE order_items
  ADD COLUMN subtotal_in_cents integer GENERATED ALWAYS AS (quantity * price_in_cents) STORED;