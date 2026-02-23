package agent

import (
	"context"
	"testing"

	"github.com/safebites/backend-go/internal/model"
	"github.com/stretchr/testify/require"
)

func TestRecommenderRecommend(t *testing.T) {
	fake := newFakeLLM(`{"recommendations":[{"product_name":"Better Cereal","health_score":"HIGH","reason":"Lower sugar"}]}`)
	a, err := NewRecommenderAgent(fake)
	require.NoError(t, err)

	out, err := a.Recommend(context.Background(), "Sugary Cereal", 3.0)
	require.NoError(t, err)
	require.Len(t, out.Recommendations, 1)
	require.Len(t, fake.requests, 1)
}

func TestRecommenderEmptyProductName(t *testing.T) {
	fake := newFakeLLM(`{"recommendations":[]}`)
	a, err := NewRecommenderAgent(fake)
	require.NoError(t, err)

	_, err = a.Recommend(context.Background(), "   ", 3.0)
	require.Error(t, err)
	require.ErrorContains(t, err, "product name is required")
}

func TestRecommenderMalformedJSON(t *testing.T) {
	fake := newFakeLLM(`not-json`)
	a, err := NewRecommenderAgent(fake)
	require.NoError(t, err)

	_, err = a.Recommend(context.Background(), "Sugary Cereal", 3.0)
	require.Error(t, err)
	require.ErrorContains(t, err, "parse recommender result")
}

func TestRecommenderCodeFenceJSON(t *testing.T) {
	fake := newFakeLLM("```json\n{\"recommendations\":[{\"product_name\":\"Better Cereal\",\"health_score\":\"HIGH\",\"reason\":\"Lower sugar\"}]}\n```")
	a, err := NewRecommenderAgent(fake)
	require.NoError(t, err)

	out, err := a.Recommend(context.Background(), "Sugary Cereal", 3.0)
	require.NoError(t, err)
	require.Len(t, out.Recommendations, 1)
}

func TestOrchestratorStopsWhenScoreImproves(t *testing.T) {
	fake := newFakeLLM(
		`{"List_of_ingredients":[{"name":"Sugar","description":"Sweetener"}]}`,
		`{"ingredient_scores":[{"ingredient_name":"Sugar","safety_score":"LOW","reasoning":"High sugar"}],"overall_score":2.1}`,
		`{"recommendations":[{"product_name":"Unsweetened Oats","health_score":"HIGH","reason":"No added sugar"}]}`,
		`{"ingredient_scores":[{"ingredient_name":"Unsweetened Oats","safety_score":"HIGH","reasoning":"Minimal processing"}],"overall_score":8.3}`,
	)

	searcher, err := NewSearchAgent(fake)
	require.NoError(t, err)
	scorer, err := NewScorerAgent(fake)
	require.NoError(t, err)
	recommender, err := NewRecommenderAgent(fake)
	require.NoError(t, err)
	orch := NewOrchestrator(searcher, scorer, recommender, WorkflowConfig{MinAcceptableScore: 7.0, MaxRecommendationTx: 2})

	res, err := orch.AnalyzeAndImprove(context.Background(), "Sugary Oatmeal", &model.UserPreferences{DietGoals: []string{"low-sugar"}})
	require.NoError(t, err)
	require.Equal(t, 2.1, res.InitialScore.OverallScore)
	require.Equal(t, 8.3, res.FinalScore.OverallScore)
	require.Len(t, res.Turns, 1)
}

func TestOrchestratorMaxTwoRecommendationTurns(t *testing.T) {
	fake := newFakeLLM(
		`{"List_of_ingredients":[{"name":"Additive","description":"Unknown blend"}]}`,
		`{"ingredient_scores":[{"ingredient_name":"Additive","safety_score":"LOW","reasoning":"Unclear"}],"overall_score":2.0}`,
		`{"recommendations":[{"product_name":"Alt1","health_score":"MEDIUM","reason":"Slightly better"}]}`,
		`{"ingredient_scores":[{"ingredient_name":"Alt1","safety_score":"MEDIUM","reasoning":"Some concerns"}],"overall_score":4.0}`,
		`{"recommendations":[{"product_name":"Alt2","health_score":"MEDIUM","reason":"Better profile"}]}`,
		`{"ingredient_scores":[{"ingredient_name":"Alt2","safety_score":"MEDIUM","reasoning":"Still not ideal"}],"overall_score":5.0}`,
	)

	searcher, err := NewSearchAgent(fake)
	require.NoError(t, err)
	scorer, err := NewScorerAgent(fake)
	require.NoError(t, err)
	recommender, err := NewRecommenderAgent(fake)
	require.NoError(t, err)
	orch := NewOrchestrator(searcher, scorer, recommender, WorkflowConfig{MinAcceptableScore: 7.0, MaxRecommendationTx: 2})

	res, err := orch.AnalyzeAndImprove(context.Background(), "Unknown Snack", nil)
	require.NoError(t, err)
	require.Len(t, res.Turns, 2)
	require.Equal(t, 5.0, res.FinalScore.OverallScore)
}

func TestOrchestratorSkipsLoopWhenInitialScoreIsHigh(t *testing.T) {
	fake := newFakeLLM(
		`{"List_of_ingredients":[{"name":"Oats","description":"Whole grain"}]}`,
		`{"ingredient_scores":[{"ingredient_name":"Oats","safety_score":"HIGH","reasoning":"Minimal processing"}],"overall_score":8.0}`,
	)

	searcher, err := NewSearchAgent(fake)
	require.NoError(t, err)
	scorer, err := NewScorerAgent(fake)
	require.NoError(t, err)
	recommender, err := NewRecommenderAgent(fake)
	require.NoError(t, err)
	orch := NewOrchestrator(searcher, scorer, recommender, WorkflowConfig{MinAcceptableScore: 7.0, MaxRecommendationTx: 2})

	res, err := orch.AnalyzeAndImprove(context.Background(), "Plain Oats", nil)
	require.NoError(t, err)
	require.Equal(t, 8.0, res.InitialScore.OverallScore)
	require.Equal(t, 8.0, res.FinalScore.OverallScore)
	require.Len(t, res.Turns, 0)
	require.Len(t, fake.requests, 2)
}
