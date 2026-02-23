package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	adkmodel "google.golang.org/adk/model"

	sbmodel "github.com/safebites/backend-go/internal/model"
)

type ScorerAgent struct {
	ingredientAgent     agent.Agent
	recommendationAgent agent.Agent
}

func NewScorerAgent(llm adkmodel.LLM) (*ScorerAgent, error) {
	ingredientAgent, err := llmagent.New(llmagent.Config{
		Name:        "ingredient_scorer_agent",
		Model:       llm,
		Description: "Scores product ingredient safety with user preferences.",
		Instruction: scorerAgentInstructions,
	})
	if err != nil {
		return nil, fmt.Errorf("create ingredient scorer agent: %w", err)
	}

	recommendationAgent, err := llmagent.New(llmagent.Config{
		Name:        "recommendation_scorer_agent",
		Model:       llm,
		Description: "Scores recommended alternatives with user preferences.",
		Instruction: recommendationEvalSystemPrompt,
	})
	if err != nil {
		return nil, fmt.Errorf("create recommendation scorer agent: %w", err)
	}

	return &ScorerAgent{
		ingredientAgent:     ingredientAgent,
		recommendationAgent: recommendationAgent,
	}, nil
}

func (a *ScorerAgent) ScoreIngredients(ctx context.Context, ingredients []sbmodel.Ingredient, prefs *sbmodel.UserPreferences) (*sbmodel.ScorerResult, error) {
	payload := map[string]interface{}{"ingredients": ingredients}
	return a.scoreFromPayload(ctx, a.ingredientAgent, payload, prefs)
}

func (a *ScorerAgent) ScoreRecommendations(ctx context.Context, recommendations []sbmodel.Recommendation, prefs *sbmodel.UserPreferences) (*sbmodel.ScorerResult, error) {
	payload := map[string]interface{}{"recommendations": recommendations}
	return a.scoreFromPayload(ctx, a.recommendationAgent, payload, prefs)
}

func (a *ScorerAgent) scoreFromPayload(ctx context.Context, agnt agent.Agent, payload map[string]interface{}, prefs *sbmodel.UserPreferences) (*sbmodel.ScorerResult, error) {
	if prefs != nil {
		payload["user_preferences"] = prefs
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal scorer payload: %w", err)
	}

	raw, err := runAgentOnce(ctx, "safebites-scorer", agnt, string(buf))
	if err != nil {
		return nil, err
	}

	raw = stripJSONCodeFences(raw)

	var out sbmodel.ScorerResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &out); err != nil {
		return nil, fmt.Errorf("parse scorer result: %w", err)
	}

	return &out, nil
}
