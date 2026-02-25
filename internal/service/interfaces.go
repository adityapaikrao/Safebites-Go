package service

import (
	"context"

	"github.com/safebites/backend-go/internal/model"
)

type AnalyzeService interface {
	Analyze(ctx context.Context, imageBytes []byte, mimeType string, prefs *model.UserPreferences) (string, *model.ScorerResult, error)
}

type RecommendService interface {
	Recommend(ctx context.Context, productName string, score float64) (*model.RecommenderResult, error)
}

type UserService interface {
	GetByID(ctx context.Context, userID string) (*model.User, error)
	Upsert(ctx context.Context, user *model.User) (*model.User, error)
	UpdatePreferences(ctx context.Context, userID string, preferences model.UserPreferences) (*model.User, error)
	ApplyTemplate(ctx context.Context, userID string, templateKey string) (*model.User, model.DietaryTemplate, error)
}

type ScanService interface {
	ListByUser(ctx context.Context, userID string, limit int) ([]model.Scan, error)
	Create(ctx context.Context, scan *model.Scan) (*model.Scan, error)
	GetStats(ctx context.Context, userID string) (*model.UserStats, error)
}
