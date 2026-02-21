package repository

import (
	"context"
	"errors"

	"github.com/safebites/backend-go/internal/model"
)

var ErrNotFound = errors.New("not found")

type UserRepository interface {
	GetByID(ctx context.Context, userID string) (*model.User, error)
	Upsert(ctx context.Context, user *model.User) (*model.User, error)
	UpdatePreferences(ctx context.Context, userID string, preferences model.UserPreferences) (*model.User, error)
}

type ScanRepository interface {
	ListByUser(ctx context.Context, userID string, limit int) ([]model.Scan, error)
	Create(ctx context.Context, scan *model.Scan) (*model.Scan, error)
	GetStats(ctx context.Context, userID string) (*model.UserStats, error)
}

type FavoriteRepository interface {
	ListByUser(ctx context.Context, userID string) ([]model.Favorite, error)
	Create(ctx context.Context, favorite *model.Favorite) (*model.Favorite, error)
	Delete(ctx context.Context, userID string, favoriteID int) error
	Exists(ctx context.Context, userID, productName string) (bool, error)
}
