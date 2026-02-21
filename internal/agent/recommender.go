package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	adkmodel "google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/geminitool"

	sbmodel "github.com/safebites/backend-go/internal/model"
)

type RecommenderAgent struct {
	agent agent.Agent
}

func NewRecommenderAgent(llm adkmodel.LLM) (*RecommenderAgent, error) {
	a, err := llmagent.New(llmagent.Config{
		Name:        "recommender_agent",
		Model:       llm,
		Description: "Finds healthier alternatives for a product.",
		Instruction: recommenderAgentInstructions,
		Tools: []tool.Tool{
			geminitool.GoogleSearch{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create recommender agent: %w", err)
	}
	return &RecommenderAgent{agent: a}, nil
}

func (a *RecommenderAgent) Recommend(ctx context.Context, productName string, score float64) (*sbmodel.RecommenderResult, error) {
	if strings.TrimSpace(productName) == "" {
		return nil, fmt.Errorf("product name is required")
	}

	input := map[string]interface{}{
		"product_name":  productName,
		"overall_score": score,
	}
	buf, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal recommender payload: %w", err)
	}

	raw, err := runAgentOnce(ctx, "safebites-recommender", a.agent, string(buf))
	if err != nil {
		return nil, err
	}

	var out sbmodel.RecommenderResult
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("parse recommender result: %w", err)
	}

	return &out, nil
}
