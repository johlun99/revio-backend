-- name: ListTenants :many
SELECT id, name, api_key, created_at FROM tenants ORDER BY created_at DESC;

-- name: GetTenant :one
SELECT id, name, api_key, created_at FROM tenants WHERE id = $1;

-- name: CreateTenant :one
INSERT INTO tenants (name) VALUES ($1) RETURNING id, name, api_key, created_at;

-- name: UpdateTenantName :one
UPDATE tenants SET name = @name, updated_at = now()
WHERE id = @id
RETURNING id, name, api_key, created_at;

-- name: RotateTenantAPIKey :one
UPDATE tenants SET api_key = gen_random_uuid()::text, updated_at = now()
WHERE id = @id
RETURNING id, name, api_key, created_at;
