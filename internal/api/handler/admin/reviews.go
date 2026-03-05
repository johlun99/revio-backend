package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/johlun99/revio/internal/repository"
)

type ReviewsHandler struct {
	queries *repository.Queries
}

func NewReviewsHandler(pool *pgxpool.Pool) *ReviewsHandler {
	return &ReviewsHandler{queries: repository.New(pool)}
}

func (h *ReviewsHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := int32(clampInt(queryInt(r, "limit", 20), 1, 100))
	offset := int32(queryInt(r, "offset", 0))

	params := repository.ListReviewsParams{
		Limit:  limit,
		Offset: offset,
	}

	if s := r.URL.Query().Get("status"); s != "" {
		status := repository.ReviewStatus(s)
		if !status.Valid() {
			writeError(w, http.StatusBadRequest, "invalid status value")
			return
		}
		params.Status = repository.NullReviewStatus{ReviewStatus: status, Valid: true}
	}

	if tid := r.URL.Query().Get("tenant_id"); tid != "" {
		var uid pgtype.UUID
		if err := uid.Scan(tid); err != nil {
			writeError(w, http.StatusBadRequest, "invalid tenant_id")
			return
		}
		params.TenantID = uid
	}

	if pid := r.URL.Query().Get("product_id"); pid != "" {
		var uid pgtype.UUID
		if err := uid.Scan(pid); err != nil {
			writeError(w, http.StatusBadRequest, "invalid product_id")
			return
		}
		params.ProductID = uid
	}

	rows, err := h.queries.ListReviews(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list reviews")
		return
	}

	total, err := h.queries.CountReviews(r.Context(), repository.CountReviewsParams{
		Status:    params.Status,
		TenantID:  params.TenantID,
		ProductID: params.ProductID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to count reviews")
		return
	}

	if rows == nil {
		rows = []repository.ListReviewsRow{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data":   rows,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *ReviewsHandler) Get(w http.ResponseWriter, r *http.Request) {
	var uid pgtype.UUID
	if err := uid.Scan(chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusBadRequest, "invalid review id")
		return
	}

	row, err := h.queries.GetReview(r.Context(), uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "review not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get review")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(row)
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func (h *ReviewsHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var uid pgtype.UUID
	if err := uid.Scan(chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusBadRequest, "invalid review id")
		return
	}

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	status := repository.ReviewStatus(req.Status)
	if !status.Valid() {
		writeError(w, http.StatusBadRequest, "invalid status value")
		return
	}

	review, err := h.queries.UpdateReviewStatus(r.Context(), repository.UpdateReviewStatusParams{
		ID:     uid,
		Status: status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "review not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update review")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(review)
}

func queryInt(r *http.Request, key string, fallback int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
