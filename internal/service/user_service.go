package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/safebites/backend-go/internal/model"
	"github.com/safebites/backend-go/internal/repository"
)

type userService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) UserService {
	return &userService{users: users}
}

func (s *userService) GetByID(ctx context.Context, userID string) (*model.User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user id is required")
	}
	return s.users.GetByID(ctx, userID)
}

func (s *userService) Upsert(ctx context.Context, user *model.User) (*model.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user is required")
	}
	if strings.TrimSpace(user.ID) == "" {
		return nil, fmt.Errorf("user id is required")
	}
	if strings.TrimSpace(user.Email) == "" {
		return nil, fmt.Errorf("user email is required")
	}
	return s.users.Upsert(ctx, user)
}

func (s *userService) UpdatePreferences(ctx context.Context, userID string, preferences model.UserPreferences) (*model.User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user id is required")
	}
	return s.users.UpdatePreferences(ctx, userID, preferences)
}

func (s *userService) ApplyTemplate(ctx context.Context, userID string, templateKey string) (*model.User, model.DietaryTemplate, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, model.DietaryTemplate{}, fmt.Errorf("user id is required")
	}

	template, ok := model.DietaryTemplates[templateKey]
	if !ok {
		return nil, model.DietaryTemplate{}, repository.ErrNotFound
	}

	updated, err := s.users.UpdatePreferences(ctx, userID, model.UserPreferences{
		Allergies:        template.Allergies,
		DietGoals:        template.DietGoals,
		AvoidIngredients: template.AvoidIngredients,
	})
	if err != nil {
		return nil, model.DietaryTemplate{}, err
	}

	return updated, template, nil
}
