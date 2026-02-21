package agent

import (
	"context"
	"testing"

	"github.com/safebites/backend-go/internal/model"
	"github.com/stretchr/testify/require"
)

func TestScorerScoreIngredients(t *testing.T) {
	fake := newFakeLLM(`{"ingredient_scores":[{"ingredient_name":"Sugar","safety_score":"LOW","reasoning":"High added sugar"}],"overall_score":3.5}`)
	a, err := NewScorerAgent(fake)
	require.NoError(t, err)

	prefs := &model.UserPreferences{Allergies: []string{"nuts"}, DietGoals: []string{"keto"}}
	out, err := a.ScoreIngredients(context.Background(), []model.Ingredient{{Name: "Sugar", Description: "Sweetener"}}, prefs)
	require.NoError(t, err)
	require.Equal(t, 3.5, out.OverallScore)
	require.Len(t, fake.requests, 1)
}

func TestScorerScoreRecommendations(t *testing.T) {
	fake := newFakeLLM(`{"ingredient_scores":[{"ingredient_name":"Alt Product","safety_score":"HIGH","reasoning":"Cleaner profile"}],"overall_score":8.9}`)
	a, err := NewScorerAgent(fake)
	require.NoError(t, err)

	recs := []model.Recommendation{{ProductName: "Alt Product", HealthScore: "HIGH", Reason: "Less additives"}}
	out, err := a.ScoreRecommendations(context.Background(), recs, nil)
	require.NoError(t, err)
	require.Equal(t, 8.9, out.OverallScore)
}
