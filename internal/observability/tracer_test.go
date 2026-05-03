package observability_test

import (
	"context"
	"testing"
	"time"

	"github.com/safebites/backend-go/internal/config"
	"github.com/safebites/backend-go/internal/observability"
)

func TestInitTracer_NoopWhenDisabled(t *testing.T) {
	ctx := context.Background()
	cfg := config.LangfuseConfig{} // empty → disabled

	shutdown := observability.InitTracer(ctx, cfg)
	if shutdown == nil {
		t.Fatal("shutdown func is nil")
	}
	if err := shutdown(ctx); err != nil {
		t.Errorf("shutdown returned err: %v", err)
	}
}

func TestInitTracer_NeverPanicsOnBadHost(t *testing.T) {
	// Use a short timeout so a DNS lookup on the invalid host doesn't block CI.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	t.Cleanup(func() { observability.SetTracerProvider(nil) })

	cfg := config.LangfuseConfig{
		PublicKey: "pk", SecretKey: "sk",
		Host: "http://invalid.invalid.invalid",
	}

	// Must not panic; must always return a usable cleanup func.
	shutdown := observability.InitTracer(ctx, cfg)
	if shutdown == nil {
		t.Fatal("shutdown is nil — must always return a usable cleanup func")
	}
	_ = shutdown(ctx)
}
