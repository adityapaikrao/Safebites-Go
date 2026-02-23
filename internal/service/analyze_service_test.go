package service

import (
	"context"
	"errors"
	"testing"

	"github.com/safebites/backend-go/internal/agent"
	"github.com/safebites/backend-go/internal/model"
	"github.com/stretchr/testify/require"
)

type mockVisionExtractor struct {
	extractProductName func(ctx context.Context, imageBytes []byte, mimeType string) (string, error)
}

func (m *mockVisionExtractor) ExtractProductName(ctx context.Context, imageBytes []byte, mimeType string) (string, error) {
	return m.extractProductName(ctx, imageBytes, mimeType)
}

type mockAnalyzeWorkflow struct {
	analyzeAndImprove func(ctx context.Context, productName string, prefs *model.UserPreferences) (*agent.WorkflowResult, error)
}

func (m *mockAnalyzeWorkflow) AnalyzeAndImprove(ctx context.Context, productName string, prefs *model.UserPreferences) (*agent.WorkflowResult, error) {
	return m.analyzeAndImprove(ctx, productName, prefs)
}

func TestAnalyzeServiceAnalyzeSuccess(t *testing.T) {
	svc := NewAnalyzeService(
		&mockVisionExtractor{
			extractProductName: func(_ context.Context, imageBytes []byte, mimeType string) (string, error) {
				require.Equal(t, []byte("img"), imageBytes)
				require.Equal(t, "image/png", mimeType)
				return "Product A", nil
			},
		},
		&mockAnalyzeWorkflow{
			analyzeAndImprove: func(_ context.Context, productName string, prefs *model.UserPreferences) (*agent.WorkflowResult, error) {
				require.Equal(t, "Product A", productName)
				require.Equal(t, []string{"vegan"}, prefs.DietGoals)
				return &agent.WorkflowResult{FinalScore: model.ScorerResult{OverallScore: 8.2}}, nil
			},
		},
	)

	name, result, err := svc.Analyze(context.Background(), []byte("img"), "image/png", &model.UserPreferences{DietGoals: []string{"vegan"}})
	require.NoError(t, err)
	require.Equal(t, "Product A", name)
	require.NotNil(t, result)
	require.Equal(t, 8.2, result.FinalScore.OverallScore)
}

func TestAnalyzeServiceAnalyzeDefaultsMimeType(t *testing.T) {
	svc := NewAnalyzeService(
		&mockVisionExtractor{
			extractProductName: func(_ context.Context, _ []byte, mimeType string) (string, error) {
				require.Equal(t, "image/jpeg", mimeType)
				return "Product A", nil
			},
		},
		&mockAnalyzeWorkflow{
			analyzeAndImprove: func(_ context.Context, _ string, _ *model.UserPreferences) (*agent.WorkflowResult, error) {
				return &agent.WorkflowResult{FinalScore: model.ScorerResult{OverallScore: 5.0}}, nil
			},
		},
	)

	_, _, err := svc.Analyze(context.Background(), []byte("img"), "   ", nil)
	require.NoError(t, err)
}

func TestAnalyzeServiceAnalyzeVisionError(t *testing.T) {
	visionErr := errors.New("vision failed")
	svc := NewAnalyzeService(
		&mockVisionExtractor{
			extractProductName: func(_ context.Context, _ []byte, _ string) (string, error) {
				return "", visionErr
			},
		},
		&mockAnalyzeWorkflow{
			analyzeAndImprove: func(_ context.Context, _ string, _ *model.UserPreferences) (*agent.WorkflowResult, error) {
				t.Fatal("workflow should not be called")
				return nil, nil
			},
		},
	)

	_, _, err := svc.Analyze(context.Background(), []byte("img"), "image/jpeg", nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "extract product name")
	require.ErrorIs(t, err, visionErr)
}

func TestAnalyzeServiceAnalyzeWorkflowError(t *testing.T) {
	workflowErr := errors.New("workflow failed")
	svc := NewAnalyzeService(
		&mockVisionExtractor{
			extractProductName: func(_ context.Context, _ []byte, _ string) (string, error) {
				return "Product A", nil
			},
		},
		&mockAnalyzeWorkflow{
			analyzeAndImprove: func(_ context.Context, _ string, _ *model.UserPreferences) (*agent.WorkflowResult, error) {
				return nil, workflowErr
			},
		},
	)

	_, _, err := svc.Analyze(context.Background(), []byte("img"), "image/jpeg", nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "run analyze workflow")
	require.ErrorIs(t, err, workflowErr)
}

func TestAnalyzeServiceAnalyzeValidation(t *testing.T) {
	svc := NewAnalyzeService(nil, nil)

	_, _, err := svc.Analyze(context.Background(), []byte("img"), "image/jpeg", nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "vision dependency is required")

	svc = NewAnalyzeService(
		&mockVisionExtractor{
			extractProductName: func(_ context.Context, _ []byte, _ string) (string, error) {
				return "Product A", nil
			},
		},
		nil,
	)
	_, _, err = svc.Analyze(context.Background(), []byte("img"), "image/jpeg", nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "orchestrator dependency is required")

	svc = NewAnalyzeService(
		&mockVisionExtractor{
			extractProductName: func(_ context.Context, _ []byte, _ string) (string, error) {
				t.Fatal("vision should not be called")
				return "", nil
			},
		},
		&mockAnalyzeWorkflow{
			analyzeAndImprove: func(_ context.Context, _ string, _ *model.UserPreferences) (*agent.WorkflowResult, error) {
				t.Fatal("workflow should not be called")
				return nil, nil
			},
		},
	)
	_, _, err = svc.Analyze(context.Background(), nil, "image/jpeg", nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "image bytes are required")
}
