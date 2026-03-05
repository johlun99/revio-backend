-- name: GetTenantByAPIKey :one
SELECT id, name, api_key, webhook_url FROM tenants WHERE api_key = $1;

-- name: GetProductByExternalID :one
SELECT id FROM products WHERE tenant_id = $1 AND external_id = $2;

-- name: CreateReview :one
INSERT INTO reviews (
    product_id, tenant_id, author_name, author_email,
    rating, title, body, ip_address, user_agent, verified_purchase
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: ListApprovedReviews :many
SELECT
    id, author_name, rating, title, body, verified_purchase, created_at
FROM reviews
WHERE product_id = $1 AND status = 'approved'
ORDER BY created_at DESC
LIMIT  sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: CountApprovedReviews :one
SELECT COUNT(*), COALESCE(AVG(rating), 0)::FLOAT8 AS avg_rating
FROM reviews
WHERE product_id = $1 AND status = 'approved';
