package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/safebites/backend-go/internal/config"
	"github.com/stretchr/testify/require"
)

func TestRequireAuthDevBypass(t *testing.T) {
	cfg := &config.Config{Env: "development"}

	h := RequireAuth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		require.True(t, ok)
		require.Equal(t, "dev-test-user", userID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireAuthToken(t *testing.T) {
	cfg := &config.Config{Env: "production", Auth0Domain: "example.auth0.com"}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "auth0|abc"})
	rawToken, err := token.SignedString([]byte("secret"))
	require.NoError(t, err)

	h := RequireAuth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		require.True(t, ok)
		require.Equal(t, "auth0|abc", userID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+rawToken)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}
