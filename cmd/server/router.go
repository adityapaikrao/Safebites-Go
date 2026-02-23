package main

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	sbagent "github.com/safebites/backend-go/internal/agent"
	"github.com/safebites/backend-go/internal/config"
	"github.com/safebites/backend-go/internal/handler"
	"github.com/safebites/backend-go/internal/middleware"
	"github.com/safebites/backend-go/internal/repository"
	"github.com/safebites/backend-go/internal/service"
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
func buildRouter(cfg *config.Config, db *repository.DB) (*chi.Mux, error) {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.Logging)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.OptionalAuth(cfg))

	userRepo := repository.NewUserRepository(db)
	scanRepo := repository.NewScanRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)

	llm, err := sbagent.NewGeminiModel(context.Background(), cfg.GoogleAPIKey, "")
	if err != nil {
		return nil, fmt.Errorf("initialize gemini model: %w", err)
	}

	visionOCR, err := sbagent.NewVisionOCRFromAPIKey(cfg.GoogleAPIKey)
	if err != nil {
		return nil, fmt.Errorf("initialize vision ocr: %w", err)
	}

	orchestrator, err := sbagent.NewOrchestratorFromModel(llm, sbagent.WorkflowConfig{})
	if err != nil {
		return nil, fmt.Errorf("initialize analysis orchestrator: %w", err)
	}

	recommenderAgent, err := sbagent.NewRecommenderAgent(llm)
	if err != nil {
		return nil, fmt.Errorf("initialize recommender agent: %w", err)
	}

	userService := service.NewUserService(userRepo)
	analyzeService := service.NewAnalyzeService(visionOCR, orchestrator)
	recommendService := service.NewRecommendService(recommenderAgent)

	userHandler := &handler.UserHandler{Users: userRepo}
	templateHandler := &handler.TemplateHandler{Users: userRepo}
	scanHandler := &handler.ScanHandler{
		Scans: scanRepo,
		Users: userRepo,
	}
	favoriteHandler := &handler.FavoriteHandler{Favorites: favoriteRepo}
	analyzeHandler := &handler.AnalyzeHandler{Analyze: analyzeService, Users: userService}
	recommendHandler := &handler.RecommendHandler{Recommend: recommendService}

	r.Get("/", handler.Health)

	// Swagger / OpenAPI docs
	r.Get("/docs", handler.SwaggerUI)
	r.Get("/docs/openapi.json", handler.OpenAPISpec)

	r.Route("/api", func(api chi.Router) {
		api.Post("/analyze", analyzeHandler.AnalyzeImage)
		api.Get("/reccomendations/{product_name}/{overall_score}", recommendHandler.RecommendProducts)

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

	return r, nil
}
