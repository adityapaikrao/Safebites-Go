package main

import (
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/safebites/backend-go/internal/config"
	"github.com/safebites/backend-go/internal/handler"
	"github.com/safebites/backend-go/internal/middleware"
	"github.com/safebites/backend-go/internal/repository"
)

// buildRouter wires all middleware, routes, and handlers together.
// Dependencies are injected here; each handler package receives only what it needs.
//
// Phase progression:
//   - Phase 1: health check only (DB connectivity verified)
//   - Phase 2: repository layer (models + SQL)
//   - Phase 3: all REST handlers + auth middleware
//   - Phase 4: agent layer (Gemini)
//   - Phase 5: analyze + recommend endpoints wired end-to-end
func buildRouter(cfg *config.Config, db *repository.DB) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logging)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.OptionalAuth(cfg))

	userHandler := &handler.UserHandler{Users: repository.NewUserRepository(db)}
	templateHandler := &handler.TemplateHandler{Users: repository.NewUserRepository(db)}
	scanHandler := &handler.ScanHandler{
		Scans: repository.NewScanRepository(db),
		Users: repository.NewUserRepository(db),
	}
	favoriteHandler := &handler.FavoriteHandler{Favorites: repository.NewFavoriteRepository(db)}

	r.Get("/", handler.Health)

	r.Route("/api", func(api chi.Router) {
		api.With(middleware.RequireAuth(cfg)).Get("/users/me", userHandler.GetMe)

		api.Get("/users/{user_id}", userHandler.GetByID)
		api.Post("/users", userHandler.Upsert)
		api.Post("/users/{user_id}/preferences", userHandler.UpdatePreferences)

		api.Get("/dietary-templates", templateHandler.List)
		api.Post("/users/{user_id}/apply-template/{template_key}", templateHandler.Apply)

		api.Get("/users/{user_id}/scans", scanHandler.ListByUser)
		api.Post("/users/{user_id}/scans", scanHandler.Create)
		api.Get("/users/{user_id}/stats", scanHandler.Stats)

		api.Get("/users/{user_id}/favorites", favoriteHandler.ListByUser)
		api.Post("/users/{user_id}/favorites", favoriteHandler.Create)
		api.Delete("/users/{user_id}/favorites/{favorite_id}", favoriteHandler.Delete)
		api.Get("/users/{user_id}/favorites/check/{product_name}", favoriteHandler.Check)
	})

	return r
}
