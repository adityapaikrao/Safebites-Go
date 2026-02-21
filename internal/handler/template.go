package handler

import (
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/safebites/backend-go/internal/model"
	"github.com/safebites/backend-go/internal/repository"
)

type TemplateHandler struct {
	Users repository.UserRepository
}

func (h *TemplateHandler) List(w http.ResponseWriter, _ *http.Request) {
	templates := make([]model.DietaryTemplate, 0, len(model.DietaryTemplates))
	for _, tpl := range model.DietaryTemplates {
		templates = append(templates, tpl)
	}
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Key < templates[j].Key
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{"templates": templates})
}

func (h *TemplateHandler) Apply(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	templateKey := chi.URLParam(r, "template_key")

	template, ok := model.DietaryTemplates[templateKey]
	if !ok {
		writeError(w, http.StatusNotFound, "template not found")
		return
	}

	updated, err := h.Users.UpdatePreferences(r.Context(), userID, model.UserPreferences{
		Allergies:        template.Allergies,
		DietGoals:        template.DietGoals,
		AvoidIngredients: template.AvoidIngredients,
	})
	if err != nil {
		if err == repository.ErrNotFound {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to apply template")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":     updated,
		"template": template,
	})
}
