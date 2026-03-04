-- name: GetDashboardStats :one
SELECT
    COUNT(*) FILTER (WHERE status = 'pending')  AS pending,
    COUNT(*) FILTER (WHERE status = 'approved') AS approved,
    COUNT(*) FILTER (WHERE status = 'rejected') AS rejected,
    COUNT(*) FILTER (WHERE status = 'flagged')  AS flagged,
    COUNT(*)                                     AS total
FROM reviews;
