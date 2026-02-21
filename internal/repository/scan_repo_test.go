package repository

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/safebites/backend-go/internal/model"
	"github.com/stretchr/testify/require"
)

func TestScanRepoListByUserSuccess(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	now := time.Now().UTC()
	rows := pgxmock.NewRows([]string{"id", "user_id", "product_name", "brand", "image", "safety_score", "is_safe", "ingredients", "timestamp"}).
		AddRow("scan-1", "user-1", "Granola Bar", "Brand A", "", 78, true, []byte(`[{"name":"oats"}]`), now)

	mock.ExpectQuery("SELECT id, user_id").WithArgs("user-1", 10).WillReturnRows(rows)

	repo := &scanRepo{q: mock}
	scans, err := repo.ListByUser(context.Background(), "user-1", 10)
	require.NoError(t, err)
	require.Len(t, scans, 1)
	require.Equal(t, "scan-1", scans[0].ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScanRepoCreateSuccess(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	now := time.Now().UTC()
	rows := pgxmock.NewRows([]string{"id", "user_id", "product_name", "brand", "image", "safety_score", "is_safe", "ingredients", "timestamp"}).
		AddRow("scan-1", "user-1", "Granola Bar", "Brand A", "", 78, true, []byte(`[{"name":"oats"}]`), now)

	mock.ExpectQuery("INSERT INTO scans").WithArgs(
		"scan-1",
		"user-1",
		"Granola Bar",
		"Brand A",
		"",
		78,
		true,
		pgxmock.AnyArg(),
	).WillReturnRows(rows)

	repo := &scanRepo{q: mock}
	created, err := repo.Create(context.Background(), &model.Scan{
		ID:          "scan-1",
		UserID:      "user-1",
		ProductName: "Granola Bar",
		Brand:       "Brand A",
		Image:       "",
		SafetyScore: 78,
		IsSafe:      true,
		Ingredients: []map[string]interface{}{{"name": "oats"}},
	})
	require.NoError(t, err)
	require.Equal(t, "scan-1", created.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScanRepoGetStatsSuccess(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	rows := pgxmock.NewRows([]string{"total_scans", "today_scans", "safe_today", "risky_today", "average_score"}).
		AddRow(10, 3, 2, 1, 81.2)

	mock.ExpectQuery("SELECT").WithArgs("user-1").WillReturnRows(rows)

	repo := &scanRepo{q: mock}
	stats, err := repo.GetStats(context.Background(), "user-1")
	require.NoError(t, err)
	require.Equal(t, 10, stats.TotalScans)
	require.Equal(t, 81.2, stats.AverageScore)
	require.NoError(t, mock.ExpectationsWereMet())
}
