package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/safebites/backend-go/internal/model"
	"github.com/safebites/backend-go/internal/repository"
	"github.com/stretchr/testify/require"
)

type mockUserRepo struct {
	getByID           func(ctx context.Context, userID string) (*model.User, error)
	upsert            func(ctx context.Context, user *model.User) (*model.User, error)
	updatePreferences func(ctx context.Context, userID string, preferences model.UserPreferences) (*model.User, error)
}

func (m *mockUserRepo) GetByID(ctx context.Context, userID string) (*model.User, error) {
	return m.getByID(ctx, userID)
}

func (m *mockUserRepo) Upsert(ctx context.Context, user *model.User) (*model.User, error) {
	return m.upsert(ctx, user)
}

func (m *mockUserRepo) UpdatePreferences(ctx context.Context, userID string, preferences model.UserPreferences) (*model.User, error) {
	return m.updatePreferences(ctx, userID, preferences)
}

func TestUserHandlerGetByID(t *testing.T) {
	h := &UserHandler{Users: &mockUserRepo{
		getByID: func(_ context.Context, userID string) (*model.User, error) {
			return &model.User{ID: userID, Email: "user@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
		},
		upsert:            nil,
		updatePreferences: nil,
	}}

	req := httptest.NewRequest(http.MethodGet, "/api/users/user-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("user_id", "user-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.GetByID(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserHandlerUpsert(t *testing.T) {
	h := &UserHandler{Users: &mockUserRepo{
		getByID: nil,
		upsert: func(_ context.Context, user *model.User) (*model.User, error) {
			return user, nil
		},
		updatePreferences: nil,
	}}

	body, _ := json.Marshal(map[string]interface{}{
		"id":               "user-1",
		"email":            "user@example.com",
		"allergies":        []string{"peanut"},
		"dietGoals":        []string{"vegan"},
		"avoidIngredients": []string{"gelatin"},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	h.Upsert(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserHandlerUpdatePreferencesNotFound(t *testing.T) {
	h := &UserHandler{Users: &mockUserRepo{
		getByID: nil,
		upsert:  nil,
		updatePreferences: func(_ context.Context, _ string, _ model.UserPreferences) (*model.User, error) {
			return nil, repository.ErrNotFound
		},
	}}

	body, _ := json.Marshal(map[string]interface{}{"dietGoals": []string{"keto"}})
	req := httptest.NewRequest(http.MethodPost, "/api/users/user-1/preferences", bytes.NewBuffer(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("user_id", "user-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.UpdatePreferences(rr, req)
	require.Equal(t, http.StatusNotFound, rr.Code)
}
