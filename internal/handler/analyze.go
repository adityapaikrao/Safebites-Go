package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/safebites/backend-go/internal/middleware"
	"github.com/safebites/backend-go/internal/model"
	"github.com/safebites/backend-go/internal/repository"
	"github.com/safebites/backend-go/internal/service"
)

const maxAnalyzeImageBytes = 10 << 20

type AnalyzeHandler struct {
	Analyze service.AnalyzeService
	Users   service.UserService
}

func (h *AnalyzeHandler) AnalyzeImage(w http.ResponseWriter, r *http.Request) {
	if h.Analyze == nil {
		writeError(w, http.StatusInternalServerError, "analyze service is not configured")
		return
	}

	if err := r.ParseMultipartForm(maxAnalyzeImageBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form data")
		return
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "image file is required")
		return
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(io.LimitReader(file, maxAnalyzeImageBytes+1))
	if err != nil {
		writeInternalError(w, r, "failed to read image", err)
		return
	}
	if len(imageBytes) == 0 {
		writeError(w, http.StatusBadRequest, "image file must not be empty")
		return
	}
	if len(imageBytes) > maxAnalyzeImageBytes {
		writeError(w, http.StatusRequestEntityTooLarge, "image file too large")
		return
	}

	mimeType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
	if mimeType == "" {
		mimeType = http.DetectContentType(imageBytes)
	}

	var prefs *model.UserPreferences
	if userID, ok := middleware.UserIDFromContext(r.Context()); ok && h.Users != nil {
		user, userErr := h.Users.GetByID(r.Context(), userID)
		if userErr != nil {
			if userErr != repository.ErrNotFound {
				writeInternalError(w, r, "failed to fetch user preferences", userErr)
				return
			}
		} else {
			prefs = &model.UserPreferences{
				Allergies:        user.Allergies,
				DietGoals:        user.DietGoals,
				AvoidIngredients: user.AvoidIngredients,
			}
		}
	}

	productName, scorerResult, err := h.Analyze.Analyze(r.Context(), imageBytes, mimeType, prefs)
	if err != nil {
		writeInternalError(w, r, "failed to analyze product", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":               "success",
		"product_name":         productName,
		"ingredient_breakdown": scorerResult,
	})
}
