package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/johlun99/revio/internal/api/handler"
	"github.com/johlun99/revio/internal/api/handler/admin"
	"github.com/johlun99/revio/internal/api/handler/public"
	authmw "github.com/johlun99/revio/internal/api/middleware"
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

	authHandler := admin.NewAuthHandler(pool, cfg.JWTSecret)

	r.Route("/api/v1", func(r chi.Router) {
		// Public widget API (API key auth + rate limiting)
		publicReviews := public.NewReviewsHandler(pool)
		r.Group(func(r chi.Router) {
			r.Use(authmw.RequireAPIKey(pool))
			r.Use(authmw.RateLimit(10, 20, func(r *http.Request) string {
				return r.Header.Get("X-API-Key")
			}))
			r.Get("/reviews", publicReviews.List)
			r.Post("/reviews", publicReviews.Submit)
		})

		// Public auth routes
		r.Post("/admin/auth/login", authHandler.Login)
		r.Post("/admin/auth/logout", authHandler.Logout)

		// Protected admin routes
		r.Group(func(r chi.Router) {
			r.Use(authmw.RequireAuth(cfg.JWTSecret))

			r.Get("/admin/auth/me", authHandler.Me)

			dashboardHandler := admin.NewDashboardHandler(pool)
			r.Get("/admin/stats", dashboardHandler.Stats)

			reviewsHandler := admin.NewReviewsHandler(pool)
			r.Get("/admin/reviews", reviewsHandler.List)
			r.Get("/admin/reviews/{id}", reviewsHandler.Get)
			r.Patch("/admin/reviews/{id}/status", reviewsHandler.UpdateStatus)

			tenantsHandler := admin.NewTenantsHandler(pool)
			r.Get("/admin/tenants", tenantsHandler.List)
			r.Post("/admin/tenants", tenantsHandler.Create)
			r.Get("/admin/tenants/{id}", tenantsHandler.Get)

			productsHandler := admin.NewProductsHandler(pool)
			r.Get("/admin/products", productsHandler.List)
			r.Post("/admin/products", productsHandler.Create)
			r.Get("/admin/products/{id}", productsHandler.Get)
			r.Delete("/admin/products/{id}", productsHandler.Delete)
		})
	})

	return r
}
