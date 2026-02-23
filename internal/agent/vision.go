package agent

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// VisionOCR is intentionally not modeled as an agent.
// It is a direct Gemini OCR call that extracts product name from image bytes.
type VisionOCR struct {
	client VisionClient
	model  string
}

type VisionClient interface {
	GenerateContent(context.Context, string, []*genai.Content, *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)
}

type geminiVisionClient struct {
	client *genai.Client
}

func (g *geminiVisionClient) GenerateContent(ctx context.Context, model string, contents []*genai.Content, cfg *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	return g.client.Models.GenerateContent(ctx, model, contents, cfg)
}

func NewVisionOCR(client VisionClient) *VisionOCR {
	return &VisionOCR{client: client, model: defaultGeminiModel}
}

func NewVisionOCRFromAPIKey(apiKey string) (*VisionOCR, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("google api key is required")
	}
	gClient, err := genai.NewClient(context.Background(), &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, fmt.Errorf("create gemini vision client: %w", err)
	}
	return NewVisionOCR(&geminiVisionClient{client: gClient}), nil
}

func (v *VisionOCR) ExtractProductName(ctx context.Context, imageBytes []byte, mimeType string) (string, error) {
	if len(imageBytes) == 0 {
		return "", fmt.Errorf("image bytes are required")
	}
	if v.client == nil {
		return "", fmt.Errorf("vision client is required")
	}
	if strings.TrimSpace(mimeType) == "" {
		mimeType = "image/jpeg"
	}

	if !isSupportedVisionMimeType(mimeType) {
		return "", fmt.Errorf("unsupported image mime type: %s", mimeType)
	}

	resp, err := v.client.GenerateContent(ctx, v.model, []*genai.Content{
		genai.NewContentFromParts([]*genai.Part{
			genai.NewPartFromBytes(imageBytes, mimeType),
			genai.NewPartFromText(visionOCRPrompt),
		}, genai.RoleUser),
	}, &genai.GenerateContentConfig{})
	if err != nil {
		return "", err
	}

	if resp == nil || strings.TrimSpace(resp.Text()) == "" {
		return "", fmt.Errorf("vision response is empty")
	}

	return strings.TrimSpace(resp.Text()), nil
}

func isSupportedVisionMimeType(mimeType string) bool {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/jpeg", "image/jpg", "image/png", "image/webp", "image/heic", "image/heif":
		return true
	default:
		return false
	}
}
