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

type TenantsHandler struct {
	queries *repository.Queries
}

func NewTenantsHandler(pool *pgxpool.Pool) *TenantsHandler {
	return &TenantsHandler{queries: repository.New(pool)}
}

func (h *TenantsHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.ListTenants(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tenants")
		return
	}
	if rows == nil {
		rows = []repository.ListTenantsRow{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rows)
}

func (h *TenantsHandler) Get(w http.ResponseWriter, r *http.Request) {
	var uid pgtype.UUID
	if err := uid.Scan(chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}
	row, err := h.queries.GetTenant(r.Context(), uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get tenant")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(row)
}

type createTenantRequest struct {
	Name string `json:"name"`
}

func (h *TenantsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	row, err := h.queries.CreateTenant(r.Context(), req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create tenant")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(row)
}

func (h *TenantsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var uid pgtype.UUID
	if err := uid.Scan(chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	row, err := h.queries.UpdateTenantName(r.Context(), repository.UpdateTenantNameParams{
		ID:   uid,
		Name: req.Name,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update tenant")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(row)
}

func (h *TenantsHandler) SetWebhook(w http.ResponseWriter, r *http.Request) {
	var uid pgtype.UUID
	if err := uid.Scan(chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}
	var req struct {
		WebhookURL *string `json:"webhook_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// Empty string is treated as clearing the webhook
	if req.WebhookURL != nil && *req.WebhookURL == "" {
		req.WebhookURL = nil
	}
	row, err := h.queries.UpdateTenantWebhook(r.Context(), repository.UpdateTenantWebhookParams{
		ID:         uid,
		WebhookUrl: req.WebhookURL,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update webhook")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(row)
}

func (h *TenantsHandler) RotateKey(w http.ResponseWriter, r *http.Request) {
	var uid pgtype.UUID
	if err := uid.Scan(chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}
	row, err := h.queries.RotateTenantAPIKey(r.Context(), uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to rotate api key")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(row)
}
