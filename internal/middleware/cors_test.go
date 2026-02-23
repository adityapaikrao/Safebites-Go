package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/safebites/backend-go/internal/config"
	"github.com/stretchr/testify/require"
)

func TestCORSAllowedOrigins(t *testing.T) {
	cfg := &config.Config{CORSOrigins: []string{"http://localhost:3000", "https://app.example.com"}}

	h := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name            string
		origin          string
		expectedAllowed bool
	}{
		{name: "first configured origin allowed", origin: "http://localhost:3000", expectedAllowed: true},
		{name: "second configured origin allowed", origin: "https://app.example.com", expectedAllowed: true},
		{name: "non configured origin blocked", origin: "https://evil.example.com", expectedAllowed: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			req.Header.Set("Origin", tc.origin)
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if tc.expectedAllowed {
				require.Equal(t, tc.origin, rr.Header().Get("Access-Control-Allow-Origin"))
				return
			}

			require.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}
