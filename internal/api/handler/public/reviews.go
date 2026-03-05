package public

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/netip"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apimw "github.com/johlun99/revio/internal/api/middleware"
	"github.com/johlun99/revio/internal/repository"
	"github.com/johlun99/revio/internal/webhook"
)

type ReviewsHandler struct {
	queries *repository.Queries
}

func NewReviewsHandler(pool *pgxpool.Pool) *ReviewsHandler {
	return &ReviewsHandler{queries: repository.New(pool)}
}

type submitRequest struct {
	ExternalProductID string  `json:"product_id"`
	AuthorName        string  `json:"author_name"`
	AuthorEmail       *string `json:"author_email,omitempty"`
	Rating            int16   `json:"rating"`
	Title             *string `json:"title,omitempty"`
	Body              string  `json:"body"`
	VerifiedPurchase  bool    `json:"verified_purchase"`
}

func (h *ReviewsHandler) Submit(w http.ResponseWriter, r *http.Request) {
	tenant := apimw.TenantFromContext(r.Context())

	var req submitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.AuthorName == "" || req.Body == "" {
		writeError(w, http.StatusBadRequest, "author_name and body are required")
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		writeError(w, http.StatusBadRequest, "rating must be between 1 and 5")
		return
	}

	productID, err := h.queries.GetProductByExternalID(r.Context(), repository.GetProductByExternalIDParams{
		TenantID:   tenant.ID,
		ExternalID: req.ExternalProductID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	params := repository.CreateReviewParams{
		ProductID:        productID,
		TenantID:         tenant.ID,
		AuthorName:       req.AuthorName,
		Rating:           req.Rating,
		Body:             req.Body,
		VerifiedPurchase: req.VerifiedPurchase,
		AuthorEmail:      req.AuthorEmail,
		Title:            req.Title,
	}

	// Capture IP — strip port if present
	host := r.RemoteAddr
	if h, _, err := splitHostPort(host); err == nil {
		host = h
	}
	if addr, err := netip.ParseAddr(host); err == nil {
		params.IpAddress = &addr
	}

	ua := r.UserAgent()
	if ua != "" {
		params.UserAgent = &ua
	}

	review, err := h.queries.CreateReview(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Fire webhook asynchronously — does not block the response.
	if tenant.WebhookURL != nil {
		webhook.Dispatch(*tenant.WebhookURL, tenant.APIKey, webhook.ReviewDetail{
			ID:               webhook.UUIDString(review.ID),
			ProductID:        webhook.UUIDString(review.ProductID),
			TenantID:         webhook.UUIDString(review.TenantID),
			AuthorName:       review.AuthorName,
			Rating:           review.Rating,
			Title:            review.Title,
			Body:             review.Body,
			Status:           string(review.Status),
			VerifiedPurchase: review.VerifiedPurchase,
			CreatedAt:        review.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"id":     review.ID.String(),
		"status": string(review.Status),
	})
}

func (h *ReviewsHandler) List(w http.ResponseWriter, r *http.Request) {
	externalProductID := r.URL.Query().Get("product_id")
	if externalProductID == "" {
		writeError(w, http.StatusBadRequest, "product_id is required")
		return
	}

	tenant := apimw.TenantFromContext(r.Context())

	productID, err := h.queries.GetProductByExternalID(r.Context(), repository.GetProductByExternalIDParams{
		TenantID:   tenant.ID,
		ExternalID: externalProductID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	limit := int32(clamp(queryInt(r, "limit", 10), 1, 50))
	offset := int32(queryInt(r, "offset", 0))

	rows, err := h.queries.ListApprovedReviews(r.Context(), repository.ListApprovedReviewsParams{
		ProductID: productID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	stats, err := h.queries.CountApprovedReviews(r.Context(), productID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if rows == nil {
		rows = []repository.ListApprovedReviewsRow{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data":       rows,
		"total":      stats.Count,
		"avg_rating": stats.AvgRating,
		"limit":      limit,
		"offset":     offset,
	})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func queryInt(r *http.Request, key string, fallback int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func splitHostPort(addr string) (host, port string, err error) {
	i := strings.LastIndex(addr, ":")
	if i < 0 {
		return addr, "", nil
	}
	return addr[:i], addr[i+1:], nil
}
