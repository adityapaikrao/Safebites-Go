package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/safebites/backend-go/internal/model"
)

type recommendationRunner interface {
	Recommend(ctx context.Context, productName string, score float64) (*model.RecommenderResult, error)
}

type recommendService struct {
	recommender recommendationRunner
}

func NewRecommendService(recommender recommendationRunner) RecommendService {
	return &recommendService{recommender: recommender}
}

func (s *recommendService) Recommend(ctx context.Context, productName string, score float64) (*model.RecommenderResult, error) {
	if s.recommender == nil {
		return nil, fmt.Errorf("recommender dependency is required")
	}
	if strings.TrimSpace(productName) == "" {
		return nil, fmt.Errorf("product name is required")
	}
	if score < 0 {
		return nil, fmt.Errorf("score must be non-negative")
	}

	result, err := s.recommender.Recommend(ctx, strings.TrimSpace(productName), score)
	if err != nil {
		return nil, fmt.Errorf("run recommender workflow: %w", err)
	}
	if result == nil {
		return nil, fmt.Errorf("recommender workflow returned empty result")
	}

	return result, nil
}
