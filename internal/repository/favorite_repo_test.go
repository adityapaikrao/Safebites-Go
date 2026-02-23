package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/safebites/backend-go/internal/model"
	"github.com/stretchr/testify/require"
)

func TestFavoriteRepoListByUserSuccess(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	now := time.Now().UTC()
	rows := pgxmock.NewRows([]string{"id", "user_id", "product_name", "brand", "safety_score", "image", "added_at"}).
		AddRow(1, "user-1", "Granola Bar", "Brand A", 85, "", now)

	mock.ExpectQuery("SELECT id, user_id").WithArgs("user-1").WillReturnRows(rows)

	repo := &favoriteRepo{q: mock}
	favorites, err := repo.ListByUser(context.Background(), "user-1")
	require.NoError(t, err)
	require.Len(t, favorites, 1)
	require.Equal(t, "Granola Bar", favorites[0].ProductName)
	require.NotNil(t, favorites[0].SafetyScore)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFavoriteRepoCreateSuccess(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	now := time.Now().UTC()
	rows := pgxmock.NewRows([]string{"id", "user_id", "product_name", "brand", "safety_score", "image", "added_at"}).
		AddRow(1, "user-1", "Granola Bar", "Brand A", 85, "", now)

	score := 85
	mock.ExpectQuery("INSERT INTO favorites").WithArgs(
		"user-1",
		"Granola Bar",
		"Brand A",
		&score,
		"",
	).WillReturnRows(rows)

	repo := &favoriteRepo{q: mock}
	created, err := repo.Create(context.Background(), &model.Favorite{
		UserID:      "user-1",
		ProductName: "Granola Bar",
		Brand:       "Brand A",
		SafetyScore: &score,
		Image:       "",
	})
	require.NoError(t, err)
	require.Equal(t, 1, created.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFavoriteRepoDeleteNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectExec("DELETE FROM favorites").WithArgs("user-1", 99).WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repo := &favoriteRepo{q: mock}
	err = repo.Delete(context.Background(), "user-1", 99)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNotFound))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFavoriteRepoExistsTrue(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	rows := pgxmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("user-1", "Granola Bar").WillReturnRows(rows)

	repo := &favoriteRepo{q: mock}
	exists, err := repo.Exists(context.Background(), "user-1", "Granola Bar")
	require.NoError(t, err)
	require.True(t, exists)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFavoriteRepoListByUserNullSafetyScore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	now := time.Now().UTC()
	rows := pgxmock.NewRows([]string{"id", "user_id", "product_name", "brand", "safety_score", "image", "added_at"}).
		AddRow(1, "user-1", "Granola Bar", "Brand A", nil, "", now)

	mock.ExpectQuery("SELECT id, user_id").WithArgs("user-1").WillReturnRows(rows)

	repo := &favoriteRepo{q: mock}
	favorites, err := repo.ListByUser(context.Background(), "user-1")
	require.NoError(t, err)
	require.Len(t, favorites, 1)
	require.Nil(t, favorites[0].SafetyScore)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFavoriteRepoDeleteSuccess(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectExec("DELETE FROM favorites").WithArgs("user-1", 1).WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := &favoriteRepo{q: mock}
	err = repo.Delete(context.Background(), "user-1", 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFavoriteRepoExistsError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT EXISTS").WithArgs("user-1", "Granola Bar").WillReturnError(errors.New("db failure"))

	repo := &favoriteRepo{q: mock}
	_, err = repo.Exists(context.Background(), "user-1", "Granola Bar")
	require.Error(t, err)
	require.ErrorContains(t, err, "check favorite exists")
	require.NoError(t, mock.ExpectationsWereMet())
}
