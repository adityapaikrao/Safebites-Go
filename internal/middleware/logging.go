package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		if shouldSkipRequestLog(r) {
			return
		}

		requestID := chiMiddleware.GetReqID(r.Context())
		if requestID == "" {
			requestID = "-"
		}

		log.Printf("%s", formatRequestLog(requestID, r.Method, r.URL.Path, ww.Status(), time.Since(start), shouldUseColorLogs()))
	})
}

func shouldSkipRequestLog(r *http.Request) bool {
	return r.Method == http.MethodOptions
}

func formatRequestLog(requestID, method, path string, status int, duration time.Duration, useColor bool) string {
	if !useColor {
		return fmt.Sprintf("%d %s %s %s req=%s", status, method, path, duration.Round(time.Millisecond), requestID)
	}

	statusColor := colorForStatus(status)
	methodColor := colorForMethod(method)

	return fmt.Sprintf(
		"\x1b[%sm%d\x1b[0m \x1b[%sm%s\x1b[0m %s \x1b[90m%s\x1b[0m \x1b[2mreq=%s\x1b[0m",
		statusColor,
		status,
		methodColor,
		method,
		path,
		duration.Round(time.Millisecond),
		requestID,
	)
}

func shouldUseColorLogs() bool {
	if strings.EqualFold(os.Getenv("ENV"), "production") {
		return false
	}

	if strings.EqualFold(os.Getenv("NO_COLOR"), "1") {
		return false
	}

	return !strings.EqualFold(os.Getenv("TERM"), "dumb")
}

func colorForStatus(status int) string {
	switch {
	case status >= 500:
		return "31"
	case status >= 400:
		return "33"
	case status >= 300:
		return "36"
	default:
		return "32"
	}
}

func colorForMethod(method string) string {
	switch method {
	case http.MethodGet:
		return "34"
	case http.MethodPost:
		return "35"
	case http.MethodPut, http.MethodPatch:
		return "36"
	case http.MethodDelete:
		return "31"
	default:
		return "37"
	}
}
