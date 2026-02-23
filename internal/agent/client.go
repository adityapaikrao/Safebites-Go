package agent

import (
	"context"
	"fmt"
	"log"
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
	start := time.Now()
	log.Printf("agent run start app=%s input_len=%d input_preview=%q", appName, len(input), previewText(input, 160))

	sessionService := session.InMemoryService()
	r, err := runner.New(runner.Config{
		AppName:        appName,
		Agent:          agnt,
		SessionService: sessionService,
	})
	if err != nil {
		log.Printf("agent run failed app=%s stage=runner_init err=%v", appName, err)
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
		log.Printf("agent run failed app=%s stage=session_create err=%v", appName, err)
		return "", fmt.Errorf("create adk session: %w", err)
	}

	var out string
	eventCount := 0
	partsCount := 0
	for event, runErr := range r.Run(ctx, userID, sessionID, genai.NewContentFromText(input, genai.RoleUser), agent.RunConfig{}) {
		if runErr != nil {
			log.Printf("agent run failed app=%s stage=run_stream event_count=%d err=%v", appName, eventCount, runErr)
			return "", runErr
		}
		if event == nil || event.LLMResponse.Content == nil {
			continue
		}
		eventCount++
		for _, part := range event.LLMResponse.Content.Parts {
			partsCount++
			if strings.TrimSpace(part.Text) != "" {
				out = part.Text
			}
		}
	}

	if strings.TrimSpace(out) == "" {
		log.Printf("agent run failed app=%s stage=empty_output duration=%s events=%d parts=%d", appName, time.Since(start), eventCount, partsCount)
		return "", fmt.Errorf("agent returned empty text")
	}

	out = strings.TrimSpace(out)
	log.Printf("agent run complete app=%s duration=%s events=%d parts=%d output_len=%d output_preview=%q", appName, time.Since(start), eventCount, partsCount, len(out), previewText(out, 160))
	return out, nil
}

func previewText(raw string, max int) string {
	text := strings.TrimSpace(raw)
	if max <= 0 || len(text) <= max {
		return text
	}
	return text[:max] + "..."
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
