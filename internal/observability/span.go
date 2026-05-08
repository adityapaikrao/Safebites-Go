package observability

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// GenAI semantic-convention attribute keys.
// OTel-Go does not yet expose these as stable constants, so we hard-code them.
// Reference: https://opentelemetry.io/docs/specs/semconv/gen-ai/
const (
	AttrGenAIModel        = "gen_ai.request.model"
	AttrGenAIPrompt       = "gen_ai.prompt"
	AttrGenAICompletion   = "gen_ai.completion"
	AttrGenAIInputTokens  = "gen_ai.usage.input_tokens"
	AttrGenAIOutputTokens = "gen_ai.usage.output_tokens"

	// Langfuse-specific OTel attributes.
	AttrLangfuseObservationName = "langfuse.observation.name"
	AttrLangfuseTraceName       = "langfuse.trace.name"
	AttrLangfuseUserID          = "langfuse.user.id"
)

// AgentSpan wraps trace.Span with typed setters that enforce attribute names.
type AgentSpan struct{ trace.Span }

func (s AgentSpan) SetModel(name string) {
	s.SetAttributes(attribute.String(AttrGenAIModel, name))
}

func (s AgentSpan) SetGenAIInput(text string) {
	s.SetAttributes(attribute.String(AttrGenAIPrompt, text))
}

func (s AgentSpan) SetGenAIOutput(text string) {
	s.SetAttributes(attribute.String(AttrGenAICompletion, text))
}

func (s AgentSpan) SetTokens(input, output int64) {
	s.SetAttributes(
		attribute.Int64(AttrGenAIInputTokens, input),
		attribute.Int64(AttrGenAIOutputTokens, output),
	)
}

func tracer() trace.Tracer {
	providerMu.RLock()
	tp := currentProvider
	providerMu.RUnlock()
	if tp != nil {
		return tp.Tracer(tracerName)
	}
	return otel.Tracer(tracerName)
}

// StartAgentSpan opens a child span named after the agent.
func StartAgentSpan(ctx context.Context, agentName string) (context.Context, AgentSpan) {
	ctx, sp := tracer().Start(ctx, agentName,
		trace.WithAttributes(attribute.String(AttrLangfuseObservationName, agentName)),
	)
	return ctx, AgentSpan{sp}
}

// StartPipelineSpan opens the root span for an end-to-end request.
// Child agent spans nest under it.
func StartPipelineSpan(ctx context.Context, pipelineName string) (context.Context, AgentSpan) {
	ctx, sp := tracer().Start(ctx, pipelineName,
		trace.WithAttributes(attribute.String(AttrLangfuseTraceName, pipelineName)),
	)
	return ctx, AgentSpan{sp}
}

// HashUserID returns a SHA-256 hex digest of the user ID for safe inclusion
// in traces without leaking Auth0 subs.
func HashUserID(userID string) string {
	if userID == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(userID))
	return hex.EncodeToString(sum[:])
}
