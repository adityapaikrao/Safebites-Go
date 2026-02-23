package service

import (
	"context"
	"testing"
	"time"

	"github.com/safebites/backend-go/internal/model"
	"github.com/safebites/backend-go/internal/repository"
	"github.com/stretchr/testify/require"
)

type mockServiceUserRepo struct {
	getByID           func(ctx context.Context, userID string) (*model.User, error)
	upsert            func(ctx context.Context, user *model.User) (*model.User, error)
	updatePreferences func(ctx context.Context, userID string, preferences model.UserPreferences) (*model.User, error)
}

func (m *mockServiceUserRepo) GetByID(ctx context.Context, userID string) (*model.User, error) {
	return m.getByID(ctx, userID)
}

func (m *mockServiceUserRepo) Upsert(ctx context.Context, user *model.User) (*model.User, error) {
	return m.upsert(ctx, user)
}

func (m *mockServiceUserRepo) UpdatePreferences(ctx context.Context, userID string, preferences model.UserPreferences) (*model.User, error) {
	return m.updatePreferences(ctx, userID, preferences)
}

func TestUserServiceGetByID(t *testing.T) {
	now := time.Now()
	svc := NewUserService(&mockServiceUserRepo{
		getByID: func(_ context.Context, userID string) (*model.User, error) {
			return &model.User{ID: userID, Email: "user@example.com", CreatedAt: now, UpdatedAt: now}, nil
		},
		upsert:            nil,
		updatePreferences: nil,
	})

	user, err := svc.GetByID(context.Background(), "user-1")
	require.NoError(t, err)
	require.Equal(t, "user-1", user.ID)
}

func TestUserServiceGetByIDEmptyUserID(t *testing.T) {
	svc := NewUserService(&mockServiceUserRepo{
		getByID: func(_ context.Context, _ string) (*model.User, error) {
			t.Fatal("repo should not be called")
			return nil, nil
		},
		upsert:            nil,
		updatePreferences: nil,
	})

	_, err := svc.GetByID(context.Background(), "   ")
	require.Error(t, err)
	require.ErrorContains(t, err, "user id is required")
}

func TestUserServiceUpsert(t *testing.T) {
	svc := NewUserService(&mockServiceUserRepo{
		getByID: nil,
		upsert: func(_ context.Context, user *model.User) (*model.User, error) {
			return user, nil
		},
		updatePreferences: nil,
	})

	user, err := svc.Upsert(context.Background(), &model.User{ID: "user-1", Email: "user@example.com"})
	require.NoError(t, err)
	require.Equal(t, "user-1", user.ID)
}

func TestUserServiceUpsertValidation(t *testing.T) {
	svc := NewUserService(&mockServiceUserRepo{
		getByID: nil,
		upsert: func(_ context.Context, _ *model.User) (*model.User, error) {
			t.Fatal("repo should not be called")
			return nil, nil
		},
		updatePreferences: nil,
	})

	_, err := svc.Upsert(context.Background(), nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "user is required")

	_, err = svc.Upsert(context.Background(), &model.User{Email: "user@example.com"})
	require.Error(t, err)
	require.ErrorContains(t, err, "user id is required")

	_, err = svc.Upsert(context.Background(), &model.User{ID: "user-1"})
	require.Error(t, err)
	require.ErrorContains(t, err, "user email is required")
}

func TestUserServiceUpdatePreferences(t *testing.T) {
	svc := NewUserService(&mockServiceUserRepo{
		getByID: nil,
		upsert:  nil,
		updatePreferences: func(_ context.Context, userID string, preferences model.UserPreferences) (*model.User, error) {
			return &model.User{ID: userID, DietGoals: preferences.DietGoals}, nil
		},
	})

	updated, err := svc.UpdatePreferences(context.Background(), "user-1", model.UserPreferences{DietGoals: []string{"keto"}})
	require.NoError(t, err)
	require.Equal(t, []string{"keto"}, updated.DietGoals)
}

func TestUserServiceApplyTemplate(t *testing.T) {
	svc := NewUserService(&mockServiceUserRepo{
		getByID: nil,
		upsert:  nil,
		updatePreferences: func(_ context.Context, userID string, preferences model.UserPreferences) (*model.User, error) {
			return &model.User{
				ID:               userID,
				Allergies:        preferences.Allergies,
				DietGoals:        preferences.DietGoals,
				AvoidIngredients: preferences.AvoidIngredients,
			}, nil
		},
	})

	updated, template, err := svc.ApplyTemplate(context.Background(), "user-1", "vegan")
	require.NoError(t, err)
	require.Equal(t, "vegan", template.Key)
	require.Equal(t, template.DietGoals, updated.DietGoals)
}

func TestUserServiceApplyTemplateNotFound(t *testing.T) {
	svc := NewUserService(&mockServiceUserRepo{
		getByID: nil,
		upsert:  nil,
		updatePreferences: func(_ context.Context, _ string, _ model.UserPreferences) (*model.User, error) {
			t.Fatal("repo should not be called")
			return nil, nil
		},
	})

	_, _, err := svc.ApplyTemplate(context.Background(), "user-1", "does-not-exist")
	require.Error(t, err)
	require.ErrorIs(t, err, repository.ErrNotFound)
}
