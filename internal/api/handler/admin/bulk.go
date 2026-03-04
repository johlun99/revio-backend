package admin

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/johlun99/revio/internal/repository"
)

type BulkHandler struct {
	queries *repository.Queries
}

func NewBulkHandler(pool *pgxpool.Pool) *BulkHandler {
	return &BulkHandler{queries: repository.New(pool)}
}

type bulkStatusRequest struct {
	IDs    []string `json:"ids"`
	Status string   `json:"status"`
}

func (h *BulkHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req bulkStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "ids must not be empty")
		return
	}

	status := repository.ReviewStatus(req.Status)
	if !status.Valid() {
		writeError(w, http.StatusBadRequest, "invalid status value")
		return
	}

	uuids := make([]pgtype.UUID, 0, len(req.IDs))
	for _, id := range req.IDs {
		var uid pgtype.UUID
		if err := uid.Scan(id); err != nil {
			writeError(w, http.StatusBadRequest, "invalid id: "+id)
			return
		}
		uuids = append(uuids, uid)
	}

	count, err := h.queries.BulkUpdateReviewStatus(r.Context(), repository.BulkUpdateReviewStatusParams{
		Status: status,
		Ids:    uuids,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update reviews")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int64{"updated": count})
}
