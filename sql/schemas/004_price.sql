-- +goose Up
ALTER TABLE products RENAME COLUMN price_in_cents TO price;

ALTER TABLE products
  ALTER COLUMN price TYPE double precision
  USING price / 100.00;

-- +goose Down
ALTER TABLE products
  ALTER COLUMN price TYPE INTEGER
  USING (price * 100)::integer;