package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/safebites/backend-go/internal/model"
	"github.com/stretchr/testify/require"
)

func TestUserRepoGetByIDSuccess(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	now := time.Now().UTC()
	rows := pgxmock.NewRows([]string{"id", "email", "name", "picture", "allergies", "diet_goals", "avoid_ingredients", "created_at", "updated_at"}).
		AddRow("user-1", "user@example.com", "Test User", "", []byte(`["peanut"]`), []byte(`["vegan"]`), []byte(`["gelatin"]`), now, now)

	mock.ExpectQuery("SELECT id, email").WithArgs("user-1").WillReturnRows(rows)

	repo := &userRepo{q: mock}
	user, err := repo.GetByID(context.Background(), "user-1")
	require.NoError(t, err)
	require.Equal(t, "user-1", user.ID)
	require.Equal(t, []string{"peanut"}, user.Allergies)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepoGetByIDNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT id, email").WithArgs("missing").WillReturnError(pgx.ErrNoRows)

	repo := &userRepo{q: mock}
	_, err = repo.GetByID(context.Background(), "missing")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNotFound))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepoUpsertSuccess(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	now := time.Now().UTC()
	rows := pgxmock.NewRows([]string{"id", "email", "name", "picture", "allergies", "diet_goals", "avoid_ingredients", "created_at", "updated_at"}).
		AddRow("user-1", "user@example.com", "Test User", "", []byte(`["peanut"]`), []byte(`["vegan"]`), []byte(`["gelatin"]`), now, now)

	mock.ExpectQuery("INSERT INTO users").WithArgs(
		"user-1",
		"user@example.com",
		"Test User",
		"",
		pgxmock.AnyArg(),
		pgxmock.AnyArg(),
		pgxmock.AnyArg(),
	).WillReturnRows(rows)

	repo := &userRepo{q: mock}
	created, err := repo.Upsert(context.Background(), &model.User{
		ID:               "user-1",
		Email:            "user@example.com",
		Name:             "Test User",
		Allergies:        []string{"peanut"},
		DietGoals:        []string{"vegan"},
		AvoidIngredients: []string{"gelatin"},
	})
	require.NoError(t, err)
	require.Equal(t, "user-1", created.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepoUpdatePreferencesSuccess(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	now := time.Now().UTC()
	rows := pgxmock.NewRows([]string{"id", "email", "name", "picture", "allergies", "diet_goals", "avoid_ingredients", "created_at", "updated_at"}).
		AddRow("user-1", "user@example.com", "Test User", "", []byte(`["peanut"]`), []byte(`["keto"]`), []byte(`["sugar"]`), now, now)

	mock.ExpectQuery("UPDATE users").WithArgs(
		"user-1",
		pgxmock.AnyArg(),
		pgxmock.AnyArg(),
		pgxmock.AnyArg(),
	).WillReturnRows(rows)

	repo := &userRepo{q: mock}
	updated, err := repo.UpdatePreferences(context.Background(), "user-1", model.UserPreferences{
		Allergies:        []string{"peanut"},
		DietGoals:        []string{"keto"},
		AvoidIngredients: []string{"sugar"},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"keto"}, updated.DietGoals)
	require.NoError(t, mock.ExpectationsWereMet())
}
