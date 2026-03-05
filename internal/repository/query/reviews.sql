-- name: ListReviews :many
SELECT
    r.id,
    r.product_id,
    r.tenant_id,
    r.author_name,
    r.author_email,
    r.rating,
    r.title,
    r.body,
    r.status,
    r.verified_purchase,
    r.ip_address,
    r.user_agent,
    r.created_at,
    r.updated_at,
    p.name AS product_name,
    t.name AS tenant_name
FROM reviews r
JOIN products p ON p.id = r.product_id
JOIN tenants t  ON t.id = r.tenant_id
WHERE
    (sqlc.narg('status')::review_status IS NULL OR r.status = sqlc.narg('status'))
    AND (sqlc.narg('tenant_id')::uuid IS NULL OR r.tenant_id = sqlc.narg('tenant_id'))
    AND (sqlc.narg('product_id')::uuid IS NULL OR r.product_id = sqlc.narg('product_id'))
ORDER BY r.created_at DESC
LIMIT  sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: CountReviews :one
SELECT COUNT(*) FROM reviews
WHERE
    (sqlc.narg('status')::review_status IS NULL OR status = sqlc.narg('status'))
    AND (sqlc.narg('tenant_id')::uuid IS NULL OR tenant_id = sqlc.narg('tenant_id'))
    AND (sqlc.narg('product_id')::uuid IS NULL OR product_id = sqlc.narg('product_id'));

-- name: GetReview :one
SELECT
    r.id,
    r.product_id,
    r.tenant_id,
    r.author_name,
    r.author_email,
    r.rating,
    r.title,
    r.body,
    r.status,
    r.verified_purchase,
    r.ip_address,
    r.user_agent,
    r.created_at,
    r.updated_at,
    p.name AS product_name,
    t.name AS tenant_name
FROM reviews r
JOIN products p ON p.id = r.product_id
JOIN tenants t  ON t.id = r.tenant_id
WHERE r.id = $1;

-- name: UpdateReviewStatus :one
UPDATE reviews
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;
