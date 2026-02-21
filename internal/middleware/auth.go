package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/safebites/backend-go/internal/config"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDContextKey).(string)
	if !ok || strings.TrimSpace(v) == "" {
		return "", false
	}
	return v, true
}

func OptionalAuth(cfg *config.Config) func(http.Handler) http.Handler {
	_ = cfg
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r.Header.Get("Authorization"))
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := userIDFromToken(token)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r.Header.Get("Authorization"))
			if token == "" {
				if cfg.DevModeAuth() {
					ctx := context.WithValue(r.Context(), userIDContextKey, "dev-test-user")
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				writeAuthError(w, "missing authorization token")
				return
			}

			userID, ok := userIDFromToken(token)
			if !ok {
				writeAuthError(w, "invalid authorization token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(authorizationHeader string) string {
	parts := strings.SplitN(strings.TrimSpace(authorizationHeader), " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func userIDFromToken(rawToken string) (string, bool) {
	// TODO: Production hardening: replace ParseUnverified with full JWT verification
	// against Auth0 JWKS, issuer, audience, and signing algorithm constraints.
	claims := jwt.MapClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(rawToken, claims)
	if err != nil {
		return "", false
	}

	sub, ok := claims["sub"].(string)
	if !ok || strings.TrimSpace(sub) == "" {
		return "", false
	}

	return sub, true
}

func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
