package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/safebites/backend-go/internal/model"
	"github.com/safebites/backend-go/internal/repository"
)

type FavoriteHandler struct {
	Favorites repository.FavoriteRepository
}

type createFavoriteRequest struct {
	ProductName string `json:"productName"`
	Brand       string `json:"brand"`
	SafetyScore *int   `json:"safetyScore"`
	Image       string `json:"image"`
}

func (h *FavoriteHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if strings.TrimSpace(userID) == "" {
		writeError(w, http.StatusBadRequest, "missing user_id")
		return
	}

	favorites, err := h.Favorites.ListByUser(r.Context(), userID)
	if err != nil {
		writeInternalError(w, r, "failed to fetch favorites", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"favorites": favorites})
}

func (h *FavoriteHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if strings.TrimSpace(userID) == "" {
		writeError(w, http.StatusBadRequest, "missing user_id")
		return
	}

	var req createFavoriteRequest
	if ok := readJSON(w, r, &req); !ok {
		return
	}
	if strings.TrimSpace(req.ProductName) == "" {
		writeError(w, http.StatusBadRequest, "productName is required")
		return
	}
	if req.SafetyScore != nil && (*req.SafetyScore < 0 || *req.SafetyScore > 100) {
		writeError(w, http.StatusBadRequest, "safetyScore must be between 0 and 100")
		return
	}

	created, err := h.Favorites.Create(r.Context(), &model.Favorite{
		UserID:      userID,
		ProductName: req.ProductName,
		Brand:       req.Brand,
		SafetyScore: req.SafetyScore,
		Image:       req.Image,
	})
	if err != nil {
		writeInternalError(w, r, "failed to add favorite", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"favorite": created,
		"status":   "created",
	})
}

func (h *FavoriteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if strings.TrimSpace(userID) == "" {
		writeError(w, http.StatusBadRequest, "missing user_id")
		return
	}

	favoriteIDRaw := chi.URLParam(r, "favorite_id")
	favoriteID, err := strconv.Atoi(favoriteIDRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "favorite_id must be numeric")
		return
	}
	if favoriteID <= 0 {
		writeError(w, http.StatusBadRequest, "favorite_id must be positive")
		return
	}

	err = h.Favorites.Delete(r.Context(), userID, favoriteID)
	if err != nil {
		if err == repository.ErrNotFound {
			writeError(w, http.StatusNotFound, "favorite not found")
			return
		}
		writeInternalError(w, r, "failed to delete favorite", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *FavoriteHandler) Check(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	productName := chi.URLParam(r, "product_name")
	if strings.TrimSpace(userID) == "" {
		writeError(w, http.StatusBadRequest, "missing user_id")
		return
	}
	if strings.TrimSpace(productName) == "" {
		writeError(w, http.StatusBadRequest, "missing product_name")
		return
	}

	exists, err := h.Favorites.Exists(r.Context(), userID, productName)
	if err != nil {
		writeInternalError(w, r, "failed to check favorite", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"isFavorite": exists})
}
