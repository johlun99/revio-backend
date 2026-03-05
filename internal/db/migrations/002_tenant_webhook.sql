-- +goose Up
ALTER TABLE tenants ADD COLUMN webhook_url TEXT;

-- +goose Down
ALTER TABLE tenants DROP COLUMN webhook_url;
