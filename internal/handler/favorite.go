package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	favorites, err := h.Favorites.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch favorites")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"favorites": favorites})
}

func (h *FavoriteHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")

	var req createFavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.ProductName == "" {
		writeError(w, http.StatusBadRequest, "productName is required")
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
		writeError(w, http.StatusInternalServerError, "failed to add favorite")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"favorite": created,
		"status":   "created",
	})
}

func (h *FavoriteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	favoriteIDRaw := chi.URLParam(r, "favorite_id")
	favoriteID, err := strconv.Atoi(favoriteIDRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "favorite_id must be numeric")
		return
	}

	err = h.Favorites.Delete(r.Context(), userID, favoriteID)
	if err != nil {
		if err == repository.ErrNotFound {
			writeError(w, http.StatusNotFound, "favorite not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete favorite")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *FavoriteHandler) Check(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	productName := chi.URLParam(r, "product_name")

	exists, err := h.Favorites.Exists(r.Context(), userID, productName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check favorite")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"isFavorite": exists})
}
