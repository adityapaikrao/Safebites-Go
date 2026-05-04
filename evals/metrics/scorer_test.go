package metrics

import (
	"testing"
)

func TestAllergyRespected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		scores    []IngredientScore
		allergies []string
		want      bool
	}{
		{
			name: "all allergies respected: low score + reasoning mentions allergen",
			scores: []IngredientScore{
				{Name: "Peanut Oil", Score: 1, Reasoning: "user is allergic to peanut"},
				{Name: "Sugar", Score: 6, Reasoning: ""},
			},
			allergies: []string{"peanut"},
			want:      true,
		},
		{
			name: "missed allergy: high score on allergen ingredient",
			scores: []IngredientScore{
				{Name: "Peanut Oil", Score: 8, Reasoning: "natural oil"},
			},
			allergies: []string{"peanut"},
			want:      false,
		},
		{
			name: "no allergy present: non-allergen ingredient",
			scores: []IngredientScore{
				{Name: "Sugar", Score: 5, Reasoning: ""},
			},
			allergies: []string{"peanut"},
			want:      true,
		},
		{
			name: "allergen low score but reasoning misses allergen mention",
			scores: []IngredientScore{
				{Name: "Peanut Oil", Score: 2, Reasoning: "natural oil"},
			},
			allergies: []string{"peanut"},
			want:      false,
		},
		{
			name: "empty allergies list",
			scores: []IngredientScore{
				{Name: "Peanut Oil", Score: 9, Reasoning: ""},
			},
			allergies: []string{},
			want:      true,
		},
		{
			name: "nil allergies list",
			scores: []IngredientScore{
				{Name: "Sugar", Score: 5, Reasoning: ""},
			},
			allergies: nil,
			want:      true,
		},
		{
			name: "allergen score exactly 3.0 (boundary: must pass)",
			scores: []IngredientScore{
				{Name: "Peanut Oil", Score: 3.0, Reasoning: "contains peanut"},
			},
			allergies: []string{"peanut"},
			want:      true,
		},
		{
			name: "allergen score 3.5 (boundary: must fail)",
			scores: []IngredientScore{
				{Name: "Peanut Oil", Score: 3.5, Reasoning: "contains peanut"},
			},
			allergies: []string{"peanut"},
			want:      false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := AllergyRespected(tt.scores, tt.allergies)
			if got != tt.want {
				t.Errorf("AllergyRespected() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScoreDirectionAgreement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      []IngredientScore
		expected map[string]string
		want     float64
	}{
		{
			name: "full match: 1.0",
			got: []IngredientScore{
				{Name: "Sugar", Score: 3},
				{Name: "Quinoa", Score: 9},
			},
			expected: map[string]string{
				"Sugar":  "bad",
				"Quinoa": "good",
			},
			want: 1.0,
		},
		{
			name: "partial mismatch: 0.5",
			got: []IngredientScore{
				{Name: "Sugar", Score: 3},
				{Name: "Quinoa", Score: 9},
			},
			expected: map[string]string{
				"Sugar":  "bad",
				"Quinoa": "bad",
			},
			want: 0.5,
		},
		{
			name: "neutral band (4-6): 1.0",
			got: []IngredientScore{
				{Name: "Salt", Score: 4},
				{Name: "Pepper", Score: 5},
				{Name: "Cumin", Score: 6},
			},
			expected: map[string]string{
				"Salt":   "neutral",
				"Pepper": "neutral",
				"Cumin":  "neutral",
			},
			want: 1.0,
		},
		{
			name: "empty expected: 1.0",
			got: []IngredientScore{
				{Name: "Sugar", Score: 5},
			},
			expected: map[string]string{},
			want:     1.0,
		},
		{
			name: "no matching keys in got: returns 0",
			got: []IngredientScore{
				{Name: "Milk", Score: 5},
			},
			expected: map[string]string{
				"Sugar":  "neutral",
				"Quinoa": "good",
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ScoreDirectionAgreement(tt.got, tt.expected)
			if got != tt.want {
				t.Errorf("ScoreDirectionAgreement() = %v, want %v", got, tt.want)
			}
		})
	}
}
