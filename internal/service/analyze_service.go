package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/safebites/backend-go/internal/agent"
	"github.com/safebites/backend-go/internal/model"
)

type productNameExtractor interface {
	ExtractProductName(ctx context.Context, imageBytes []byte, mimeType string) (string, error)
}

type analyzeWorkflow interface {
	AnalyzeAndImprove(ctx context.Context, productName string, prefs *model.UserPreferences) (*agent.WorkflowResult, error)
}

type analyzeService struct {
	vision       productNameExtractor
	orchestrator analyzeWorkflow
}

func NewAnalyzeService(vision productNameExtractor, orchestrator analyzeWorkflow) AnalyzeService {
	return &analyzeService{
		vision:       vision,
		orchestrator: orchestrator,
	}
}

func (s *analyzeService) Analyze(ctx context.Context, imageBytes []byte, mimeType string, prefs *model.UserPreferences) (string, *agent.WorkflowResult, error) {
	if s.vision == nil {
		return "", nil, fmt.Errorf("vision dependency is required")
	}
	if s.orchestrator == nil {
		return "", nil, fmt.Errorf("orchestrator dependency is required")
	}
	if len(imageBytes) == 0 {
		return "", nil, fmt.Errorf("image bytes are required")
	}
	if strings.TrimSpace(mimeType) == "" {
		mimeType = "image/jpeg"
	}

	productName, err := s.vision.ExtractProductName(ctx, imageBytes, mimeType)
	if err != nil {
		return "", nil, fmt.Errorf("extract product name: %w", err)
	}
	if strings.TrimSpace(productName) == "" {
		return "", nil, fmt.Errorf("product name extraction returned empty value")
	}

	result, err := s.orchestrator.AnalyzeAndImprove(ctx, strings.TrimSpace(productName), prefs)
	if err != nil {
		return "", nil, fmt.Errorf("run analyze workflow: %w", err)
	}
	if result == nil {
		return "", nil, fmt.Errorf("analyze workflow returned empty result")
	}

	return strings.TrimSpace(productName), result, nil
}
