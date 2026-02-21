package model

import "time"

type User struct {
	ID               string    `json:"id"`
	Email            string    `json:"email"`
	Name             string    `json:"name,omitempty"`
	Picture          string    `json:"picture,omitempty"`
	Allergies        []string  `json:"allergies"`
	DietGoals        []string  `json:"dietGoals"`
	AvoidIngredients []string  `json:"avoidIngredients"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type UserPreferences struct {
	Allergies        []string `json:"allergies"`
	DietGoals        []string `json:"dietGoals"`
	AvoidIngredients []string `json:"avoidIngredients"`
}

type UserStats struct {
	TotalScans   int     `json:"totalScans"`
	TodayScans   int     `json:"todayScans"`
	SafeToday    int     `json:"safeToday"`
	RiskyToday   int     `json:"riskyToday"`
	AverageScore float64 `json:"averageScore"`
}
