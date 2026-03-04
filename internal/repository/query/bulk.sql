-- name: BulkUpdateReviewStatus :execrows
UPDATE reviews
SET status = @status, updated_at = now()
WHERE id = ANY(@ids::uuid[]);
