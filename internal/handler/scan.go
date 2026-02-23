package handler

import (
	"net/http"
	"strconv"
	"strings"

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
	if strings.TrimSpace(userID) == "" {
		writeError(w, http.StatusBadRequest, "missing user_id")
		return
	}

	limit := 20
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed <= 0 {
			writeError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		if parsed > 100 {
			parsed = 100
		}
		limit = parsed
	}

	scans, err := h.Scans.ListByUser(r.Context(), userID, limit)
	if err != nil {
		writeInternalError(w, r, "failed to fetch scans", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"scans": scans})
}

func (h *ScanHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if strings.TrimSpace(userID) == "" {
		writeError(w, http.StatusBadRequest, "missing user_id")
		return
	}

	var req createScanRequest
	if ok := readJSON(w, r, &req); !ok {
		return
	}
	if strings.TrimSpace(req.ProductName) == "" {
		writeError(w, http.StatusBadRequest, "productName is required")
		return
	}
	if req.SafetyScore < 0 || req.SafetyScore > 100 {
		writeError(w, http.StatusBadRequest, "safetyScore must be between 0 and 100")
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
		writeInternalError(w, r, "failed to create scan", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"scan":   created,
		"status": "created",
	})
}

func (h *ScanHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if strings.TrimSpace(userID) == "" {
		writeError(w, http.StatusBadRequest, "missing user_id")
		return
	}

	_, err := h.Users.GetByID(r.Context(), userID)
	if err != nil {
		if err == repository.ErrNotFound {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeInternalError(w, r, "failed to fetch user", err)
		return
	}

	stats, err := h.Scans.GetStats(r.Context(), userID)
	if err != nil {
		writeInternalError(w, r, "failed to fetch stats", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"stats": stats})
}
