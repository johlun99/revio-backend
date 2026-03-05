-- name: ListAdminUsers :many
SELECT id, email, role, created_at, updated_at
FROM admin_users
ORDER BY created_at ASC;

-- name: CreateAdminUser :one
INSERT INTO admin_users (email, password_hash, role)
VALUES (@email, @password_hash, @role)
RETURNING id, email, role, created_at, updated_at;

-- name: DeleteAdminUser :exec
DELETE FROM admin_users WHERE id = @id;
