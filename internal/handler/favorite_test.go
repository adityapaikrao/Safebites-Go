package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/safebites/backend-go/internal/model"
	"github.com/safebites/backend-go/internal/repository"
	"github.com/stretchr/testify/require"
)

type mockFavoriteRepo struct {
	listByUser func(ctx context.Context, userID string) ([]model.Favorite, error)
	create     func(ctx context.Context, favorite *model.Favorite) (*model.Favorite, error)
	delete     func(ctx context.Context, userID string, favoriteID int) error
	exists     func(ctx context.Context, userID, productName string) (bool, error)
}

func (m *mockFavoriteRepo) ListByUser(ctx context.Context, userID string) ([]model.Favorite, error) {
	return m.listByUser(ctx, userID)
}

func (m *mockFavoriteRepo) Create(ctx context.Context, favorite *model.Favorite) (*model.Favorite, error) {
	return m.create(ctx, favorite)
}

func (m *mockFavoriteRepo) Delete(ctx context.Context, userID string, favoriteID int) error {
	return m.delete(ctx, userID, favoriteID)
}

func (m *mockFavoriteRepo) Exists(ctx context.Context, userID, productName string) (bool, error) {
	return m.exists(ctx, userID, productName)
}

func TestFavoriteHandlerCheck(t *testing.T) {
	h := &FavoriteHandler{Favorites: &mockFavoriteRepo{
		listByUser: nil,
		create:     nil,
		delete:     nil,
		exists: func(_ context.Context, userID, productName string) (bool, error) {
			require.Equal(t, "user-1", userID)
			require.Equal(t, "Granola", productName)
			return true, nil
		},
	}}

	req := httptest.NewRequest(http.MethodGet, "/api/users/user-1/favorites/check/Granola", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("user_id", "user-1")
	rctx.URLParams.Add("product_name", "Granola")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Check(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestFavoriteHandlerDeleteNotFound(t *testing.T) {
	h := &FavoriteHandler{Favorites: &mockFavoriteRepo{
		listByUser: nil,
		create:     nil,
		delete: func(_ context.Context, _ string, _ int) error {
			return repository.ErrNotFound
		},
		exists: nil,
	}}

	req := httptest.NewRequest(http.MethodDelete, "/api/users/user-1/favorites/999", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("user_id", "user-1")
	rctx.URLParams.Add("favorite_id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Delete(rr, req)
	require.Equal(t, http.StatusNotFound, rr.Code)
}
