package config

import "testing"

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
