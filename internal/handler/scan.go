package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/safebites/backend-go/internal/model"
	"github.com/safebites/backend-go/internal/repository"
)

type ScanHandler struct {
	Scans repository.ScanRepository
	Users repository.UserRepository
}

type createScanRequest struct {
	ID          string                   `json:"id"`
	ProductName string                   `json:"productName"`
	Brand       string                   `json:"brand"`
	Image       string                   `json:"image"`
	SafetyScore int                      `json:"safetyScore"`
	IsSafe      bool                     `json:"isSafe"`
	Ingredients []map[string]interface{} `json:"ingredients"`
}

func (h *ScanHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	limit := 20
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}

	scans, err := h.Scans.ListByUser(r.Context(), userID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch scans")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"scans": scans})
}

func (h *ScanHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")

	var req createScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.ProductName == "" {
		writeError(w, http.StatusBadRequest, "productName is required")
		return
	}

	scanID := req.ID
	if scanID == "" {
		scanID = uuid.NewString()
	}

	created, err := h.Scans.Create(r.Context(), &model.Scan{
		ID:          scanID,
		UserID:      userID,
		ProductName: req.ProductName,
		Brand:       req.Brand,
		Image:       req.Image,
		SafetyScore: req.SafetyScore,
		IsSafe:      req.IsSafe,
		Ingredients: req.Ingredients,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create scan")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"scan":   created,
		"status": "created",
	})
}

func (h *ScanHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")

	_, err := h.Users.GetByID(r.Context(), userID)
	if err != nil {
		if err == repository.ErrNotFound {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	stats, err := h.Scans.GetStats(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch stats")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"stats": stats})
}
