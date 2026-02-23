package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/safebites/backend-go/internal/model"
	"github.com/stretchr/testify/require"
)

type mockRecommendService struct {
	recommend func(ctx context.Context, productName string, score float64) (*model.RecommenderResult, error)
}

func (m *mockRecommendService) Recommend(ctx context.Context, productName string, score float64) (*model.RecommenderResult, error) {
	return m.recommend(ctx, productName, score)
}

func makeRecommendRequest(productName string, overallScore string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/reccomendations/"+url.PathEscape(productName)+"/"+url.PathEscape(overallScore), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("product_name", productName)
	rctx.URLParams.Add("overall_score", overallScore)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestRecommendHandlerRecommendProductsSuccess(t *testing.T) {
	h := &RecommendHandler{
		Recommend: &mockRecommendService{
			recommend: func(_ context.Context, productName string, score float64) (*model.RecommenderResult, error) {
				require.Equal(t, "Granola", productName)
				require.Equal(t, 4.5, score)
				return &model.RecommenderResult{
					Recommendations: []model.Recommendation{{ProductName: "Oats", HealthScore: "HIGH", Reason: "Lower sugar"}},
				}, nil
			},
		},
	}

	req := makeRecommendRequest("Granola", "4.5")
	rr := httptest.NewRecorder()

	h.RecommendProducts(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), `"status":"success"`)
	require.Contains(t, rr.Body.String(), `"reccomender_data"`)
}

func TestRecommendHandlerRecommendProductsMissingProductName(t *testing.T) {
	h := &RecommendHandler{
		Recommend: &mockRecommendService{
			recommend: func(_ context.Context, _ string, _ float64) (*model.RecommenderResult, error) {
				t.Fatal("service should not be called")
				return nil, nil
			},
		},
	}

	req := makeRecommendRequest("", "4.5")
	rr := httptest.NewRecorder()

	h.RecommendProducts(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "missing product_name")
}

func TestRecommendHandlerRecommendProductsMissingOverallScore(t *testing.T) {
	h := &RecommendHandler{
		Recommend: &mockRecommendService{
			recommend: func(_ context.Context, _ string, _ float64) (*model.RecommenderResult, error) {
				t.Fatal("service should not be called")
				return nil, nil
			},
		},
	}

	req := makeRecommendRequest("Granola", "")
	rr := httptest.NewRecorder()

	h.RecommendProducts(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "missing overall_score")
}

func TestRecommendHandlerRecommendProductsInvalidScore(t *testing.T) {
	h := &RecommendHandler{
		Recommend: &mockRecommendService{
			recommend: func(_ context.Context, _ string, _ float64) (*model.RecommenderResult, error) {
				t.Fatal("service should not be called")
				return nil, nil
			},
		},
	}

	req := makeRecommendRequest("Granola", "not-a-number")
	rr := httptest.NewRecorder()

	h.RecommendProducts(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "overall_score must be a number")
}

func TestRecommendHandlerRecommendProductsServiceError(t *testing.T) {
	h := &RecommendHandler{
		Recommend: &mockRecommendService{
			recommend: func(_ context.Context, _ string, _ float64) (*model.RecommenderResult, error) {
				return nil, errors.New("service failed")
			},
		},
	}

	req := makeRecommendRequest("Granola", "4.5")
	rr := httptest.NewRecorder()

	h.RecommendProducts(rr, req)
	require.Equal(t, http.StatusInternalServerError, rr.Code)
}
