package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/safebites/backend-go/internal/service"
)

type RecommendHandler struct {
	Recommend service.RecommendService
}

func (h *RecommendHandler) RecommendProducts(w http.ResponseWriter, r *http.Request) {
	if h.Recommend == nil {
		writeError(w, http.StatusInternalServerError, "recommend service is not configured")
		return
	}

	productName := strings.TrimSpace(chi.URLParam(r, "product_name"))
	if productName == "" {
		writeError(w, http.StatusBadRequest, "missing product_name")
		return
	}

	rawScore := strings.TrimSpace(chi.URLParam(r, "overall_score"))
	if rawScore == "" {
		writeError(w, http.StatusBadRequest, "missing overall_score")
		return
	}

	overallScore, err := strconv.ParseFloat(rawScore, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "overall_score must be a number")
		return
	}

	result, err := h.Recommend.Recommend(r.Context(), productName, overallScore)
	if err != nil {
		writeInternalError(w, r, "failed to generate recommendations", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":           "success",
		"reccomender_data": result,
	})
}
