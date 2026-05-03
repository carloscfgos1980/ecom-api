-- +goose Up
ALTER TABLE products
ADD description TEXT;

-- +goose Down
ALTER TABLE products
DROP COLUMN description;