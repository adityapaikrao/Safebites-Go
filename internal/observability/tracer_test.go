package observability_test

import (
	"context"
	"testing"

	"github.com/safebites/backend-go/internal/config"
	"github.com/safebites/backend-go/internal/observability"
)

func TestInitTracer_NoopWhenDisabled(t *testing.T) {
	ctx := context.Background()
	cfg := config.LangfuseConfig{} // empty → disabled

	shutdown, err := observability.InitTracer(ctx, cfg)
	if err != nil {
		t.Fatalf("InitTracer returned err: %v", err)
	}
	if shutdown == nil {
		t.Fatal("shutdown func is nil")
	}
	if err := shutdown(ctx); err != nil {
		t.Errorf("shutdown returned err: %v", err)
	}
}

func TestInitTracer_NeverPanicsOnBadHost(t *testing.T) {
	ctx := context.Background()
	cfg := config.LangfuseConfig{
		PublicKey: "pk", SecretKey: "sk",
		Host: "http://invalid.invalid.invalid",
	}

	// Must not panic; caller is allowed to ignore the error.
	shutdown, _ := observability.InitTracer(ctx, cfg)
	if shutdown == nil {
		t.Fatal("shutdown is nil — must always return a usable cleanup func")
	}
	_ = shutdown(ctx)
}
