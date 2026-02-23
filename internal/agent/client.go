package agent

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"google.golang.org/genai"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
)

const defaultGeminiModel = "gemini-2.5-flash"

var runIDCounter uint64

func NewGeminiModel(ctx context.Context, apiKey string, modelName string) (model.LLM, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("google api key is required")
	}
	if strings.TrimSpace(modelName) == "" {
		modelName = defaultGeminiModel
	}
	return gemini.NewModel(ctx, modelName, &genai.ClientConfig{APIKey: apiKey})
}

func runAgentOnce(ctx context.Context, appName string, agnt agent.Agent, input string) (string, error) {
	sessionService := session.InMemoryService()
	r, err := runner.New(runner.Config{
		AppName:        appName,
		Agent:          agnt,
		SessionService: sessionService,
	})
	if err != nil {
		return "", fmt.Errorf("create adk runner: %w", err)
	}

	runID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), atomic.AddUint64(&runIDCounter, 1))
	userID := "safebites-agent-" + runID
	sessionID := "safebites-session-" + runID
	_, err = sessionService.Create(ctx, &session.CreateRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		return "", fmt.Errorf("create adk session: %w", err)
	}

	var out string
	for event, runErr := range r.Run(ctx, userID, sessionID, genai.NewContentFromText(input, genai.RoleUser), agent.RunConfig{}) {
		if runErr != nil {
			return "", runErr
		}
		if event == nil || event.LLMResponse.Content == nil {
			continue
		}
		for _, part := range event.LLMResponse.Content.Parts {
			if strings.TrimSpace(part.Text) != "" {
				out = part.Text
			}
		}
	}

	if strings.TrimSpace(out) == "" {
		return "", fmt.Errorf("agent returned empty text")
	}

	return strings.TrimSpace(out), nil
}

func stripJSONCodeFences(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "```") {
		return trimmed
	}

	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)

	if idx := strings.IndexByte(trimmed, '\n'); idx >= 0 {
		firstLine := strings.TrimSpace(trimmed[:idx])
		if firstLine == "json" || firstLine == "JSON" {
			trimmed = strings.TrimSpace(trimmed[idx+1:])
		}
	}

	if before, ok := strings.CutSuffix(trimmed, "```"); ok {
		trimmed = before
	}

	return strings.TrimSpace(trimmed)
}
