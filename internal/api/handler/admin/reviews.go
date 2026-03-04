package admin

// ReviewsHandler will handle admin review moderation endpoints.
// Phase 2: list, approve, reject, flag reviews.
//
// Routes (to be wired in router.go):
//   GET    /api/v1/admin/reviews          - list with filters (status, product, tenant)
//   GET    /api/v1/admin/reviews/{id}     - get single review
//   PATCH  /api/v1/admin/reviews/{id}/status - update status
