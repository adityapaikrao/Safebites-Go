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
	"github.com/stretchr/testify/require"
)

type mockScanRepo struct {
	listByUser func(ctx context.Context, userID string, limit int) ([]model.Scan, error)
	create     func(ctx context.Context, scan *model.Scan) (*model.Scan, error)
	getStats   func(ctx context.Context, userID string) (*model.UserStats, error)
}

func (m *mockScanRepo) ListByUser(ctx context.Context, userID string, limit int) ([]model.Scan, error) {
	return m.listByUser(ctx, userID, limit)
}

func (m *mockScanRepo) Create(ctx context.Context, scan *model.Scan) (*model.Scan, error) {
	return m.create(ctx, scan)
}

func (m *mockScanRepo) GetStats(ctx context.Context, userID string) (*model.UserStats, error) {
	return m.getStats(ctx, userID)
}

func TestScanHandlerListByUser(t *testing.T) {
	h := &ScanHandler{Scans: &mockScanRepo{
		listByUser: func(_ context.Context, userID string, limit int) ([]model.Scan, error) {
			require.Equal(t, "user-1", userID)
			require.Equal(t, 5, limit)
			return []model.Scan{{ID: "scan-1", UserID: userID}}, nil
		},
		create:   nil,
		getStats: nil,
	}, Users: &mockUserRepo{}}

	req := httptest.NewRequest(http.MethodGet, "/api/users/user-1/scans?limit=5", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("user_id", "user-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.ListByUser(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestScanHandlerCreate(t *testing.T) {
	h := &ScanHandler{Scans: &mockScanRepo{
		listByUser: nil,
		create: func(_ context.Context, scan *model.Scan) (*model.Scan, error) {
			scan.Timestamp = time.Now()
			return scan, nil
		},
		getStats: nil,
	}, Users: &mockUserRepo{}}

	body, _ := json.Marshal(map[string]interface{}{
		"productName": "Granola Bar",
		"brand":       "Brand A",
		"safetyScore": 80,
		"isSafe":      true,
		"ingredients": []map[string]interface{}{{"name": "oats"}},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/users/user-1/scans", bytes.NewBuffer(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("user_id", "user-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Create(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}
