package agent

import (
	"context"
	"fmt"
	"iter"
	"sync"

	"google.golang.org/genai"

	adkmodel "google.golang.org/adk/model"
)

type fakeLLM struct {
	mu        sync.Mutex
	responses []string
	requests  []*adkmodel.LLMRequest
}

func newFakeLLM(responses ...string) *fakeLLM {
	return &fakeLLM{responses: responses}
}

func (f *fakeLLM) Name() string { return "fake-llm" }

func (f *fakeLLM) GenerateContent(_ context.Context, req *adkmodel.LLMRequest, _ bool) iter.Seq2[*adkmodel.LLMResponse, error] {
	f.mu.Lock()
	f.requests = append(f.requests, req)
	if len(f.responses) == 0 {
		f.mu.Unlock()
		return func(yield func(*adkmodel.LLMResponse, error) bool) {
			yield(nil, fmt.Errorf("no fake responses left"))
		}
	}
	resp := f.responses[0]
	f.responses = f.responses[1:]
	f.mu.Unlock()

	return func(yield func(*adkmodel.LLMResponse, error) bool) {
		yield(&adkmodel.LLMResponse{Content: genai.NewContentFromText(resp, genai.RoleModel)}, nil)
	}
}

type fakeVisionClient struct {
	text string
	err  error
}

func (f *fakeVisionClient) GenerateContent(context.Context, string, []*genai.Content, *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &genai.GenerateContentResponse{Candidates: []*genai.Candidate{{Content: genai.NewContentFromText(f.text, genai.RoleModel)}}}, nil
}
