package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/johlun99/revio/internal/api/handler"
	"github.com/johlun99/revio/internal/config"
)

func NewRouter(pool *pgxpool.Pool) http.Handler {
	cfg := config.Load()
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   strings.Split(cfg.CORSAllowedOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	healthHandler := handler.NewHealthHandler(pool)
	r.Get("/health", healthHandler.Handle)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/admin", func(r chi.Router) {
			// TODO: r.Use(jwtMiddleware)
			r.Mount("/reviews", adminReviewsRouter())
		})
	})

	return r
}

func adminReviewsRouter() http.Handler {
	r := chi.NewRouter()
	// Stub — Phase 2 will add: GET /, GET /{id}, PATCH /{id}/status
	return r
}
