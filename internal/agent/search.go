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

type SearchAgent struct {
	agent agent.Agent
}

func NewSearchAgent(llm adkmodel.LLM) (*SearchAgent, error) {
	a, err := llmagent.New(llmagent.Config{
		Name:        "search_agent",
		Model:       llm,
		Description: "Finds product ingredients using grounded web search.",
		Instruction: webSearchAgentInstructions,
		Tools: []tool.Tool{
			geminitool.GoogleSearch{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create search agent: %w", err)
	}
	return &SearchAgent{agent: a}, nil
}

func (a *SearchAgent) Search(ctx context.Context, productName string) (*sbmodel.WebSearchResult, error) {
	if strings.TrimSpace(productName) == "" {
		return nil, fmt.Errorf("product name is required")
	}

	raw, err := runAgentOnce(ctx, "safebites-search", a.agent, productName)
	if err != nil {
		return nil, err
	}

	var out sbmodel.WebSearchResult
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("parse search result: %w", err)
	}

	return &out, nil
}
