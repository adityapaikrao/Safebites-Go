package metrics_test

import (
	"testing"

	"github.com/safebites/backend-go/evals/metrics"
)

func TestAllergyRespect_AllAllergiesRespected(t *testing.T) {
	scores := []metrics.IngredientScore{
		{Name: "Peanut Oil", Score: 1, Reasoning: "user is allergic to peanut"},
		{Name: "Sugar", Score: 6, Reasoning: ""},
	}
	if !metrics.AllergyRespected(scores, []string{"peanut"}) {
		t.Error("low score + reasoning should pass allergy check")
	}
}

func TestAllergyRespect_MissedAllergy(t *testing.T) {
	scores := []metrics.IngredientScore{
		{Name: "Peanut Oil", Score: 8, Reasoning: "natural oil"}, // missed
	}
	if metrics.AllergyRespected(scores, []string{"peanut"}) {
		t.Error("high score on allergen ingredient must fail")
	}
}

func TestAllergyRespect_NoAllergyPresent(t *testing.T) {
	scores := []metrics.IngredientScore{
		{Name: "Sugar", Score: 5, Reasoning: ""},
	}
	if !metrics.AllergyRespected(scores, []string{"peanut"}) {
		t.Error("absence of allergen should pass")
	}
}

func TestAllergyRespect_AllergenLowScoreButReasoningMisses(t *testing.T) {
	scores := []metrics.IngredientScore{
		{Name: "Peanut Oil", Score: 2, Reasoning: "natural oil"},
	}
	if metrics.AllergyRespected(scores, []string{"peanut"}) {
		t.Error("low score on allergen must still fail if reasoning omits allergen")
	}
}

func TestAllergyRespect_EmptyAllergies(t *testing.T) {
	scores := []metrics.IngredientScore{
		{Name: "Peanut Oil", Score: 9, Reasoning: ""},
	}
	if !metrics.AllergyRespected(scores, []string{}) {
		t.Error("empty allergies list should always pass")
	}

	scores2 := []metrics.IngredientScore{
		{Name: "Sugar", Score: 5, Reasoning: ""},
	}
	if !metrics.AllergyRespected(scores2, nil) {
		t.Error("nil allergies list should always pass")
	}
}

func TestScoreDirectionAgreement(t *testing.T) {
	got := []metrics.IngredientScore{
		{Name: "Sugar", Score: 3},
		{Name: "Quinoa", Score: 9},
	}
	expected := map[string]string{
		"Sugar":  "bad",
		"Quinoa": "good",
	}
	rate := metrics.ScoreDirectionAgreement(got, expected)
	if rate != 1.0 {
		t.Errorf("agreement = %v, want 1.0", rate)
	}
}

func TestScoreDirectionAgreement_PartialMismatch(t *testing.T) {
	got := []metrics.IngredientScore{
		{Name: "Sugar", Score: 3},
		{Name: "Quinoa", Score: 9},
	}
	expected := map[string]string{
		"Sugar":  "bad",
		"Quinoa": "bad", // mismatch
	}
	rate := metrics.ScoreDirectionAgreement(got, expected)
	if rate != 0.5 {
		t.Errorf("agreement = %v, want 0.5", rate)
	}
}

func TestScoreDirectionAgreement_NeutralBand(t *testing.T) {
	got := []metrics.IngredientScore{
		{Name: "Salt", Score: 4},
		{Name: "Pepper", Score: 5},
		{Name: "Cumin", Score: 6},
	}
	expected := map[string]string{
		"Salt":   "neutral",
		"Pepper": "neutral",
		"Cumin":  "neutral",
	}
	rate := metrics.ScoreDirectionAgreement(got, expected)
	if rate != 1.0 {
		t.Errorf("agreement = %v, want 1.0 for neutral band 4-6", rate)
	}
}

func TestScoreDirectionAgreement_EmptyExpected(t *testing.T) {
	got := []metrics.IngredientScore{
		{Name: "Sugar", Score: 5},
	}
	rate := metrics.ScoreDirectionAgreement(got, map[string]string{})
	if rate != 1.0 {
		t.Errorf("agreement = %v, want 1.0 for empty expected", rate)
	}
}
