package middleware

import (
	"log"
	"net/http"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		requestID := chiMiddleware.GetReqID(r.Context())
		if requestID == "" {
			requestID = "-"
		}

		log.Printf(
			"request_id=%s method=%s path=%s query=%q status=%d bytes=%d duration=%s remote=%s ua=%q content_length=%d",
			requestID,
			r.Method,
			r.URL.Path,
			r.URL.RawQuery,
			ww.Status(),
			ww.BytesWritten(),
			time.Since(start),
			r.RemoteAddr,
			r.UserAgent(),
			r.ContentLength,
		)
	})
}
