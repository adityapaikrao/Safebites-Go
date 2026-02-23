package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoggingSkipsOptionsAndRoot(t *testing.T) {
	h := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("options request is skipped", func(t *testing.T) {
		var buf bytes.Buffer
		oldWriter := log.Writer()
		log.SetOutput(&buf)
		t.Cleanup(func() { log.SetOutput(oldWriter) })

		req := httptest.NewRequest(http.MethodOptions, "/api/users/u/scans", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		require.Empty(t, buf.String())
	})

	t.Run("root health request is logged", func(t *testing.T) {
		var buf bytes.Buffer
		oldWriter := log.Writer()
		log.SetOutput(&buf)
		t.Cleanup(func() { log.SetOutput(oldWriter) })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		require.Contains(t, buf.String(), "GET")
		require.Contains(t, buf.String(), "/")
	})

	t.Run("api request is still logged", func(t *testing.T) {
		var buf bytes.Buffer
		oldWriter := log.Writer()
		log.SetOutput(&buf)
		t.Cleanup(func() { log.SetOutput(oldWriter) })

		req := httptest.NewRequest(http.MethodGet, "/api/users/u/scans", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		require.Contains(t, buf.String(), "GET")
		require.Contains(t, buf.String(), "/api/users/u/scans")
	})
}
