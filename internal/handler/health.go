package handler

import "net/http"

func Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "SafeBites Go Backend is running!",
		"status":  "healthy",
		"version": "1.0.0",
	})
}
