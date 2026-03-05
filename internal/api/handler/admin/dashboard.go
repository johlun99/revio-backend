package admin

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/johlun99/revio/internal/repository"
)

type DashboardHandler struct {
	queries *repository.Queries
}

func NewDashboardHandler(pool *pgxpool.Pool) *DashboardHandler {
	return &DashboardHandler{queries: repository.New(pool)}
}

func (h *DashboardHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.queries.GetDashboardStats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load stats")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

func (h *DashboardHandler) Trends(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.ReviewTrends(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load trends")
		return
	}
	if rows == nil {
		rows = []repository.ReviewTrendsRow{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rows)
}
