package config

import (
	"testing"
)

func TestDevModeAuthRequiresBothAuth0Fields(t *testing.T) {
	cfg := &Config{Env: "development", Auth0Domain: "", Auth0APIAudience: ""}
	if !cfg.DevModeAuth() {
		t.Fatal("expected dev mode auth bypass when Auth0 config is missing")
	}

	cfg = &Config{Env: "development", Auth0Domain: "example.auth0.com", Auth0APIAudience: "api://aud"}
	if cfg.DevModeAuth() {
		t.Fatal("expected no auth bypass when Auth0 config is fully present")
	}
}

func TestParseCORSOriginsTrimsAndSkipsEmpty(t *testing.T) {
	origins := parseCORSOrigins(" http://localhost:3000, ,https://app.example.com  ")
	if len(origins) != 2 {
		t.Fatalf("expected 2 origins, got %d", len(origins))
	}
	if origins[0] != "http://localhost:3000" {
		t.Fatalf("unexpected first origin: %q", origins[0])
	}
	if origins[1] != "https://app.example.com" {
		t.Fatalf("unexpected second origin: %q", origins[1])
	}
}

func TestLoad_Langfuse(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("GOOGLE_API_KEY", "k")
	t.Setenv("LANGFUSE_PUBLIC_KEY", "pk-test")
	t.Setenv("LANGFUSE_SECRET_KEY", "sk-test")
	t.Setenv("LANGFUSE_BASE_URL", "https://example.langfuse.com")

	cfg := Load()

	if cfg.Langfuse.PublicKey != "pk-test" {
		t.Errorf("PublicKey = %q, want pk-test", cfg.Langfuse.PublicKey)
	}
	if cfg.Langfuse.SecretKey != "sk-test" {
		t.Errorf("SecretKey = %q, want sk-test", cfg.Langfuse.SecretKey)
	}
	if cfg.Langfuse.Host != "https://example.langfuse.com" {
		t.Errorf("Host = %q, want https://example.langfuse.com", cfg.Langfuse.Host)
	}
	if !cfg.Langfuse.Enabled() {
		t.Error("Enabled() = false, want true when both keys set")
	}
}

func TestLoad_LangfuseDisabledWhenKeysMissing(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("GOOGLE_API_KEY", "k")
	t.Setenv("LANGFUSE_PUBLIC_KEY", "")
	t.Setenv("LANGFUSE_SECRET_KEY", "")

	cfg := Load()

	if cfg.Langfuse.Enabled() {
		t.Error("Enabled() = true with no keys, want false")
	}
}
