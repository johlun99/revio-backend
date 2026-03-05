-- name: GetDashboardStats :one
SELECT
    COUNT(*) FILTER (WHERE status = 'pending')  AS pending,
    COUNT(*) FILTER (WHERE status = 'approved') AS approved,
    COUNT(*) FILTER (WHERE status = 'rejected') AS rejected,
    COUNT(*) FILTER (WHERE status = 'flagged')  AS flagged,
    COUNT(*)                                     AS total
FROM reviews;

-- name: ReviewTrends :many
SELECT
    TO_CHAR(DATE(created_at AT TIME ZONE 'UTC'), 'YYYY-MM-DD') AS day,
    COUNT(*)::int AS count
FROM reviews
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY day
ORDER BY day ASC;
