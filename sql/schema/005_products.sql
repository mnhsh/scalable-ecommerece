-- +goose Up
ALTER TABLE products
RENAME COLUMN price to price_cents;

-- +goose Down
ALTER TABLE products
RENAME COLUMN price_cents TO price;
