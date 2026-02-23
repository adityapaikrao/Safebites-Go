package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
)

const maxJSONBodyBytes int64 = 10 << 20

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	log.Printf("handler client error status=%d message=%q", status, message)
	writeJSON(w, status, map[string]string{"error": message})
}

func writeRequestError(w http.ResponseWriter, r *http.Request, status int, message string) {
	log.Printf(
		"handler client error method=%s path=%s status=%d content_length=%d message=%q",
		r.Method,
		r.URL.Path,
		status,
		r.ContentLength,
		message,
	)
	writeError(w, status, message)
}

func writeInternalError(w http.ResponseWriter, r *http.Request, message string, err error) {
	log.Printf("handler error method=%s path=%s status=%d message=%q err=%v", r.Method, r.URL.Path, http.StatusInternalServerError, message, err)
	writeError(w, http.StatusInternalServerError, message)
}

func readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	if ct := strings.TrimSpace(r.Header.Get("Content-Type")); ct != "" {
		mediaType, _, err := mime.ParseMediaType(ct)
		if err != nil || mediaType != "application/json" {
			writeRequestError(w, r, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
			return false
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		var maxBytesErr *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxErr):
			writeRequestError(w, r, http.StatusBadRequest, "request body contains malformed JSON")
		case errors.Is(err, io.EOF):
			writeRequestError(w, r, http.StatusBadRequest, "request body must not be empty")
		case errors.As(err, &typeErr):
			writeRequestError(w, r, http.StatusBadRequest, "request body contains invalid value type")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			writeRequestError(w, r, http.StatusBadRequest, strings.TrimPrefix(err.Error(), "json: "))
		case errors.As(err, &maxBytesErr):
			writeRequestError(w, r, http.StatusRequestEntityTooLarge, "request body too large")
		default:
			writeRequestError(w, r, http.StatusBadRequest, "invalid JSON body")
		}
		return false
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		writeRequestError(w, r, http.StatusBadRequest, "request body must contain only one JSON object")
		return false
	}

	return true
}
