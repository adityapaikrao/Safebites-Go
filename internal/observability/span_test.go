package observability_test

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/safebites/backend-go/internal/observability"
)

func TestStartAgentSpan_SetsGenAIAttributes(t *testing.T) {
	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	observability.SetTracerProvider(tp)
	t.Cleanup(func() { observability.SetTracerProvider(nil) })

	ctx, span := observability.StartAgentSpan(context.Background(), "SearchAgent")
	span.SetGenAIInput("hello")
	span.SetGenAIOutput("world")
	span.SetTokens(10, 20)
	span.End()

	spans := rec.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	s := spans[0]
	if s.Name() != "SearchAgent" {
		t.Errorf("span name = %q, want SearchAgent", s.Name())
	}

	attrs := map[string]string{}
	intAttrs := map[string]int64{}
	for _, kv := range s.Attributes() {
		switch kv.Value.Type() {
		case 4: // STRING
			attrs[string(kv.Key)] = kv.Value.AsString()
		case 2: // INT64
			intAttrs[string(kv.Key)] = kv.Value.AsInt64()
		}
	}
	if attrs["gen_ai.prompt"] != "hello" {
		t.Errorf("gen_ai.prompt = %q, want hello", attrs["gen_ai.prompt"])
	}
	if attrs["gen_ai.completion"] != "world" {
		t.Errorf("gen_ai.completion = %q, want world", attrs["gen_ai.completion"])
	}
	if intAttrs["gen_ai.usage.input_tokens"] != 10 {
		t.Errorf("input_tokens = %d, want 10", intAttrs["gen_ai.usage.input_tokens"])
	}
	if intAttrs["gen_ai.usage.output_tokens"] != 20 {
		t.Errorf("output_tokens = %d, want 20", intAttrs["gen_ai.usage.output_tokens"])
	}

	_ = ctx
}

func TestStartPipelineSpan_SetsTraceName(t *testing.T) {
	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	observability.SetTracerProvider(tp)
	t.Cleanup(func() { observability.SetTracerProvider(nil) })

	_, span := observability.StartPipelineSpan(context.Background(), "analyze")
	span.End()

	spans := rec.Ended()
	if len(spans) != 1 {
		t.Fatalf("want 1 span, got %d", len(spans))
	}
	if spans[0].Name() != "analyze" {
		t.Errorf("span name = %q, want analyze", spans[0].Name())
	}
}
