package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/johlun99/revio/internal/repository"
)

type tenantKey string

const TenantKey tenantKey = "tenant"

type Tenant struct {
	ID   pgtype.UUID
	Name string
}

func RequireAPIKey(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	queries := repository.New(pool)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				writeAPIError(w, http.StatusUnauthorized, "missing X-API-Key header")
				return
			}

			row, err := queries.GetTenantByAPIKey(r.Context(), key)
			if err != nil {
				writeAPIError(w, http.StatusUnauthorized, "invalid API key")
				return
			}

			ctx := context.WithValue(r.Context(), TenantKey, Tenant{ID: row.ID, Name: row.Name})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TenantFromContext(ctx context.Context) *Tenant {
	t, _ := ctx.Value(TenantKey).(Tenant)
	return &t
}

func writeAPIError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
