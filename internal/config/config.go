package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port             string
	Env              string
	DatabaseURL      string
	MigrationsPath   string
	GoogleAPIKey     string
	Auth0Domain      string
	Auth0APIAudience string
	CORSOrigins      []string
}

// Load reads configuration from environment variables, loading .env if present.
func Load() *Config {
	// Load .env file in dev — ignore error if file doesn't exist
	_ = godotenv.Load()

	cfg := &Config{
		Port:             getEnv("PORT", "8080"),
		Env:              getEnv("ENV", "development"),
		DatabaseURL:      requireEnv("DATABASE_URL"),
		MigrationsPath:   getEnv("MIGRATIONS_PATH", "migrations"),
		GoogleAPIKey:     requireEnv("GOOGLE_API_KEY"),
		Auth0Domain:      getEnv("AUTH0_DOMAIN", ""),
		Auth0APIAudience: getEnv("AUTH0_API_AUDIENCE", ""),
		CORSOrigins:      parseCORSOrigins(getEnv("CORS_ORIGINS", "http://localhost:3000")),
	}

	return cfg
}

// IsDev returns true when running in development mode.
func (c *Config) IsDev() bool {
	return c.Env == "development"
}

// DevModeAuth returns true when Auth0 is not configured (dev bypass).
func (c *Config) DevModeAuth() bool {
	return c.Auth0Domain == "" || c.Auth0APIAudience == ""
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %q is not set", key)
	}
	return v
}

func parseCORSOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
