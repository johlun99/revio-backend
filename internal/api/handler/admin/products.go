package admin

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/johlun99/revio/internal/repository"
)

type ProductsHandler struct {
	queries *repository.Queries
}

func NewProductsHandler(pool *pgxpool.Pool) *ProductsHandler {
	return &ProductsHandler{queries: repository.New(pool)}
}

func (h *ProductsHandler) List(w http.ResponseWriter, r *http.Request) {
	var tenantUID pgtype.UUID // zero value = NULL = no filter
	if tid := r.URL.Query().Get("tenant_id"); tid != "" {
		if err := tenantUID.Scan(tid); err != nil {
			writeError(w, http.StatusBadRequest, "invalid tenant_id")
			return
		}
	}

	rows, err := h.queries.ListProducts(r.Context(), tenantUID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list products")
		return
	}
	if rows == nil {
		rows = []repository.ListProductsRow{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rows)
}

func (h *ProductsHandler) Get(w http.ResponseWriter, r *http.Request) {
	var uid pgtype.UUID
	if err := uid.Scan(chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}
	row, err := h.queries.GetProduct(r.Context(), uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get product")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(row)
}

type createProductRequest struct {
	TenantID   string `json:"tenant_id"`
	ExternalID string `json:"external_id"`
	Name       string `json:"name"`
}

func (h *ProductsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.TenantID == "" || req.ExternalID == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "tenant_id, external_id and name are required")
		return
	}

	var tenantUID pgtype.UUID
	if err := tenantUID.Scan(req.TenantID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant_id")
		return
	}

	row, err := h.queries.CreateProduct(r.Context(), repository.CreateProductParams{
		TenantID:   tenantUID,
		ExternalID: req.ExternalID,
		Name:       req.Name,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create product")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(row)
}

func (h *ProductsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	var uid pgtype.UUID
	if err := uid.Scan(chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}
	if err := h.queries.DeleteProduct(r.Context(), uid); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete product")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
