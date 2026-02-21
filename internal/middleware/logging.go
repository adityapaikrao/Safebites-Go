package middleware

import (
	"net/http"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func Logging(next http.Handler) http.Handler {
	return chiMiddleware.Logger(next)
}
