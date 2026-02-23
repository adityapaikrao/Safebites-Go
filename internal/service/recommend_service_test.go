package service

import (
	"context"
	"errors"
	"testing"

	"github.com/safebites/backend-go/internal/model"
	"github.com/stretchr/testify/require"
)

type mockRecommendationRunner struct {
	recommend func(ctx context.Context, productName string, score float64) (*model.RecommenderResult, error)
}

func (m *mockRecommendationRunner) Recommend(ctx context.Context, productName string, score float64) (*model.RecommenderResult, error) {
	return m.recommend(ctx, productName, score)
}

func TestRecommendServiceRecommendSuccess(t *testing.T) {
	svc := NewRecommendService(&mockRecommendationRunner{
		recommend: func(_ context.Context, productName string, score float64) (*model.RecommenderResult, error) {
			require.Equal(t, "Product A", productName)
			require.Equal(t, 4.5, score)
			return &model.RecommenderResult{
				Recommendations: []model.Recommendation{{ProductName: "Better Product", HealthScore: "HIGH", Reason: "Less sugar"}},
			}, nil
		},
	})

	result, err := svc.Recommend(context.Background(), "Product A", 4.5)
	require.NoError(t, err)
	require.Len(t, result.Recommendations, 1)
	require.Equal(t, "Better Product", result.Recommendations[0].ProductName)
}

func TestRecommendServiceRecommendValidation(t *testing.T) {
	svc := NewRecommendService(nil)
	_, err := svc.Recommend(context.Background(), "Product A", 4.5)
	require.Error(t, err)
	require.ErrorContains(t, err, "recommender dependency is required")

	svc = NewRecommendService(&mockRecommendationRunner{
		recommend: func(_ context.Context, _ string, _ float64) (*model.RecommenderResult, error) {
			t.Fatal("recommender should not be called")
			return nil, nil
		},
	})

	_, err = svc.Recommend(context.Background(), "   ", 4.5)
	require.Error(t, err)
	require.ErrorContains(t, err, "product name is required")

	_, err = svc.Recommend(context.Background(), "Product A", -1)
	require.Error(t, err)
	require.ErrorContains(t, err, "score must be non-negative")
}

func TestRecommendServiceRecommendRunnerError(t *testing.T) {
	runnerErr := errors.New("runner failed")
	svc := NewRecommendService(&mockRecommendationRunner{
		recommend: func(_ context.Context, _ string, _ float64) (*model.RecommenderResult, error) {
			return nil, runnerErr
		},
	})

	_, err := svc.Recommend(context.Background(), "Product A", 3.2)
	require.Error(t, err)
	require.ErrorContains(t, err, "run recommender workflow")
	require.ErrorIs(t, err, runnerErr)
}

func TestRecommendServiceRecommendNilResult(t *testing.T) {
	svc := NewRecommendService(&mockRecommendationRunner{
		recommend: func(_ context.Context, _ string, _ float64) (*model.RecommenderResult, error) {
			return nil, nil
		},
	})

	_, err := svc.Recommend(context.Background(), "Product A", 3.2)
	require.Error(t, err)
	require.ErrorContains(t, err, "recommender workflow returned empty result")
}
