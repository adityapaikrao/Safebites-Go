package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/safebites/backend-go/internal/model"
)

type userQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type userRepo struct {
	q userQuerier
}

func NewUserRepository(db *DB) UserRepository {
	return &userRepo{q: db.Pool}
}

func (r *userRepo) GetByID(ctx context.Context, userID string) (*model.User, error) {
	const query = `
		SELECT id, email, COALESCE(name, ''), COALESCE(picture, ''), allergies, diet_goals, avoid_ingredients, created_at, updated_at
		FROM users
		WHERE id = $1`

	var user model.User
	var allergiesBytes []byte
	var dietGoalsBytes []byte
	var avoidIngredientsBytes []byte

	err := r.q.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Picture,
		&allergiesBytes,
		&dietGoalsBytes,
		&avoidIngredientsBytes,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query user by id: %w", err)
	}

	if err := unmarshalStringSlice(allergiesBytes, &user.Allergies); err != nil {
		return nil, fmt.Errorf("decode allergies: %w", err)
	}
	if err := unmarshalStringSlice(dietGoalsBytes, &user.DietGoals); err != nil {
		return nil, fmt.Errorf("decode diet goals: %w", err)
	}
	if err := unmarshalStringSlice(avoidIngredientsBytes, &user.AvoidIngredients); err != nil {
		return nil, fmt.Errorf("decode avoid ingredients: %w", err)
	}

	return &user, nil
}

func (r *userRepo) Upsert(ctx context.Context, user *model.User) (*model.User, error) {
	allergiesJSON, err := json.Marshal(user.Allergies)
	if err != nil {
		return nil, fmt.Errorf("marshal allergies: %w", err)
	}
	dietGoalsJSON, err := json.Marshal(user.DietGoals)
	if err != nil {
		return nil, fmt.Errorf("marshal diet goals: %w", err)
	}
	avoidIngredientsJSON, err := json.Marshal(user.AvoidIngredients)
	if err != nil {
		return nil, fmt.Errorf("marshal avoid ingredients: %w", err)
	}

	const query = `
		INSERT INTO users (id, email, name, picture, allergies, diet_goals, avoid_ingredients)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7::jsonb)
		ON CONFLICT (id)
		DO UPDATE SET
			email = EXCLUDED.email,
			name = EXCLUDED.name,
			picture = EXCLUDED.picture,
			updated_at = NOW()
		RETURNING id, email, COALESCE(name, ''), COALESCE(picture, ''), allergies, diet_goals, avoid_ingredients, created_at, updated_at`

	var created model.User
	var allergiesBytes []byte
	var dietGoalsBytes []byte
	var avoidIngredientsBytes []byte

	err = r.q.QueryRow(
		ctx,
		query,
		user.ID,
		user.Email,
		user.Name,
		user.Picture,
		allergiesJSON,
		dietGoalsJSON,
		avoidIngredientsJSON,
	).Scan(
		&created.ID,
		&created.Email,
		&created.Name,
		&created.Picture,
		&allergiesBytes,
		&dietGoalsBytes,
		&avoidIngredientsBytes,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}

	if err := unmarshalStringSlice(allergiesBytes, &created.Allergies); err != nil {
		return nil, fmt.Errorf("decode allergies: %w", err)
	}
	if err := unmarshalStringSlice(dietGoalsBytes, &created.DietGoals); err != nil {
		return nil, fmt.Errorf("decode diet goals: %w", err)
	}
	if err := unmarshalStringSlice(avoidIngredientsBytes, &created.AvoidIngredients); err != nil {
		return nil, fmt.Errorf("decode avoid ingredients: %w", err)
	}

	return &created, nil
}

func (r *userRepo) UpdatePreferences(ctx context.Context, userID string, preferences model.UserPreferences) (*model.User, error) {
	allergiesJSON, err := json.Marshal(preferences.Allergies)
	if err != nil {
		return nil, fmt.Errorf("marshal allergies: %w", err)
	}
	dietGoalsJSON, err := json.Marshal(preferences.DietGoals)
	if err != nil {
		return nil, fmt.Errorf("marshal diet goals: %w", err)
	}
	avoidIngredientsJSON, err := json.Marshal(preferences.AvoidIngredients)
	if err != nil {
		return nil, fmt.Errorf("marshal avoid ingredients: %w", err)
	}

	const query = `
		UPDATE users
		SET allergies = $2::jsonb,
			diet_goals = $3::jsonb,
			avoid_ingredients = $4::jsonb,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, COALESCE(name, ''), COALESCE(picture, ''), allergies, diet_goals, avoid_ingredients, created_at, updated_at`

	var user model.User
	var allergiesBytes []byte
	var dietGoalsBytes []byte
	var avoidIngredientsBytes []byte

	err = r.q.QueryRow(ctx, query, userID, allergiesJSON, dietGoalsJSON, avoidIngredientsJSON).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Picture,
		&allergiesBytes,
		&dietGoalsBytes,
		&avoidIngredientsBytes,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update user preferences: %w", err)
	}

	if err := unmarshalStringSlice(allergiesBytes, &user.Allergies); err != nil {
		return nil, fmt.Errorf("decode allergies: %w", err)
	}
	if err := unmarshalStringSlice(dietGoalsBytes, &user.DietGoals); err != nil {
		return nil, fmt.Errorf("decode diet goals: %w", err)
	}
	if err := unmarshalStringSlice(avoidIngredientsBytes, &user.AvoidIngredients); err != nil {
		return nil, fmt.Errorf("decode avoid ingredients: %w", err)
	}

	return &user, nil
}

func unmarshalStringSlice(in []byte, out *[]string) error {
	if len(in) == 0 {
		*out = []string{}
		return nil
	}
	if err := json.Unmarshal(in, out); err != nil {
		return err
	}
	if *out == nil {
		*out = []string{}
	}
	return nil
}
