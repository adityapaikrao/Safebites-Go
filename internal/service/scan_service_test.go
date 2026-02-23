package service

import (
	"context"
	"testing"

	"github.com/safebites/backend-go/internal/model"
	"github.com/stretchr/testify/require"
)

type mockServiceScanRepo struct {
	listByUser func(ctx context.Context, userID string, limit int) ([]model.Scan, error)
	create     func(ctx context.Context, scan *model.Scan) (*model.Scan, error)
	getStats   func(ctx context.Context, userID string) (*model.UserStats, error)
}

func (m *mockServiceScanRepo) ListByUser(ctx context.Context, userID string, limit int) ([]model.Scan, error) {
	return m.listByUser(ctx, userID, limit)
}

func (m *mockServiceScanRepo) Create(ctx context.Context, scan *model.Scan) (*model.Scan, error) {
	return m.create(ctx, scan)
}

func (m *mockServiceScanRepo) GetStats(ctx context.Context, userID string) (*model.UserStats, error) {
	return m.getStats(ctx, userID)
}

func TestScanServiceListByUser(t *testing.T) {
	svc := NewScanService(&mockServiceScanRepo{
		listByUser: func(_ context.Context, userID string, limit int) ([]model.Scan, error) {
			return []model.Scan{{ID: "scan-1", UserID: userID, ProductName: "Granola"}}, nil
		},
		create:   nil,
		getStats: nil,
	})

	scans, err := svc.ListByUser(context.Background(), "user-1", 10)
	require.NoError(t, err)
	require.Len(t, scans, 1)
	require.Equal(t, "user-1", scans[0].UserID)
}

func TestScanServiceListByUserValidation(t *testing.T) {
	svc := NewScanService(&mockServiceScanRepo{
		listByUser: func(_ context.Context, _ string, _ int) ([]model.Scan, error) {
			t.Fatal("repo should not be called")
			return nil, nil
		},
		create:   nil,
		getStats: nil,
	})

	_, err := svc.ListByUser(context.Background(), "   ", 10)
	require.Error(t, err)
	require.ErrorContains(t, err, "user id is required")
}

func TestScanServiceCreate(t *testing.T) {
	svc := NewScanService(&mockServiceScanRepo{
		listByUser: nil,
		create: func(_ context.Context, scan *model.Scan) (*model.Scan, error) {
			return scan, nil
		},
		getStats: nil,
	})

	created, err := svc.Create(context.Background(), &model.Scan{UserID: "user-1", ProductName: "Granola", SafetyScore: 80})
	require.NoError(t, err)
	require.Equal(t, "Granola", created.ProductName)
}

func TestScanServiceCreateValidation(t *testing.T) {
	svc := NewScanService(&mockServiceScanRepo{
		listByUser: nil,
		create: func(_ context.Context, _ *model.Scan) (*model.Scan, error) {
			t.Fatal("repo should not be called")
			return nil, nil
		},
		getStats: nil,
	})

	_, err := svc.Create(context.Background(), nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "scan is required")

	_, err = svc.Create(context.Background(), &model.Scan{ProductName: "Granola", SafetyScore: 80})
	require.Error(t, err)
	require.ErrorContains(t, err, "user id is required")

	_, err = svc.Create(context.Background(), &model.Scan{UserID: "user-1", SafetyScore: 80})
	require.Error(t, err)
	require.ErrorContains(t, err, "product name is required")

	_, err = svc.Create(context.Background(), &model.Scan{UserID: "user-1", ProductName: "Granola", SafetyScore: 120})
	require.Error(t, err)
	require.ErrorContains(t, err, "between 0 and 100")
}

func TestScanServiceGetStats(t *testing.T) {
	svc := NewScanService(&mockServiceScanRepo{
		listByUser: nil,
		create:     nil,
		getStats: func(_ context.Context, userID string) (*model.UserStats, error) {
			return &model.UserStats{TotalScans: 3, AverageScore: 75}, nil
		},
	})

	stats, err := svc.GetStats(context.Background(), "user-1")
	require.NoError(t, err)
	require.Equal(t, 3, stats.TotalScans)
}

func TestScanServiceGetStatsValidation(t *testing.T) {
	svc := NewScanService(&mockServiceScanRepo{
		listByUser: nil,
		create:     nil,
		getStats: func(_ context.Context, _ string) (*model.UserStats, error) {
			t.Fatal("repo should not be called")
			return nil, nil
		},
	})

	_, err := svc.GetStats(context.Background(), "")
	require.Error(t, err)
	require.ErrorContains(t, err, "user id is required")
}
