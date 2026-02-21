package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/safebites/backend-go/internal/middleware"
	"github.com/safebites/backend-go/internal/model"
	"github.com/safebites/backend-go/internal/repository"
)

type UserHandler struct {
	Users repository.UserRepository
}

type createUserRequest struct {
	ID               string   `json:"id"`
	Email            string   `json:"email"`
	Name             string   `json:"name"`
	Picture          string   `json:"picture"`
	Allergies        []string `json:"allergies"`
	DietGoals        []string `json:"dietGoals"`
	AvoidIngredients []string `json:"avoidIngredients"`
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	user, err := h.Users.GetByID(r.Context(), userID)
	if err != nil {
		if err == repository.ErrNotFound {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing user_id")
		return
	}

	user, err := h.Users.GetByID(r.Context(), userID)
	if err != nil {
		if err == repository.ErrNotFound {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

func (h *UserHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.ID == "" || req.Email == "" {
		writeError(w, http.StatusBadRequest, "id and email are required")
		return
	}

	created, err := h.Users.Upsert(r.Context(), &model.User{
		ID:               req.ID,
		Email:            req.Email,
		Name:             req.Name,
		Picture:          req.Picture,
		Allergies:        req.Allergies,
		DietGoals:        req.DietGoals,
		AvoidIngredients: req.AvoidIngredients,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to upsert user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"user": created})
}

func (h *UserHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing user_id")
		return
	}

	var preferences model.UserPreferences
	if err := json.NewDecoder(r.Body).Decode(&preferences); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	updated, err := h.Users.UpdatePreferences(r.Context(), userID, preferences)
	if err != nil {
		if err == repository.ErrNotFound {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update preferences")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"user": updated})
}
