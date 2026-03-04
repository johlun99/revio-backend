-- name: ListProducts :many
SELECT p.id, p.tenant_id, p.external_id, p.name, p.created_at, t.name AS tenant_name
FROM products p
JOIN tenants t ON t.id = p.tenant_id
WHERE (sqlc.narg('tenant_id')::uuid IS NULL OR p.tenant_id = sqlc.narg('tenant_id'))
ORDER BY p.created_at DESC;

-- name: GetProduct :one
SELECT p.id, p.tenant_id, p.external_id, p.name, p.created_at, t.name AS tenant_name
FROM products p
JOIN tenants t ON t.id = p.tenant_id
WHERE p.id = $1;

-- name: CreateProduct :one
INSERT INTO products (tenant_id, external_id, name)
VALUES ($1, $2, $3)
RETURNING id, tenant_id, external_id, name, created_at;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1;
