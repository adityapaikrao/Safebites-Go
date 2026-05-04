package metrics

import "strings"

// IngredientScore holds the scorer agent's evaluation of one ingredient.
type IngredientScore struct {
	Name      string
	Score     float64
	Reasoning string
}

// AllergyRespected returns true iff EVERY ingredient that contains an
// allergen substring (case-insensitive) was scored ≤3 AND its reasoning
// mentions the allergen. ANY violation → false.
//
// Bar is intentionally strict — this is the safety-critical metric and
// gates the build with floor=1.0.
func AllergyRespected(scores []IngredientScore, allergies []string) bool {
	if len(allergies) == 0 {
		return true
	}
	for _, s := range scores {
		nameLow := strings.ToLower(s.Name)
		reasonLow := strings.ToLower(s.Reasoning)
		for _, a := range allergies {
			al := strings.ToLower(strings.TrimSpace(a))
			if al == "" {
				continue
			}
			if strings.Contains(nameLow, al) {
				if s.Score > 3 {
					return false
				}
				if !strings.Contains(reasonLow, al) {
					return false
				}
			}
		}
	}
	return true
}

// ScoreDirectionAgreement returns the share of ingredients whose score
// direction (good/neutral/bad) matches the labeled tag.
//   score >= 7  → "good"
//   score 4..6  → "neutral"
//   score <= 3  → "bad"
// When no got ingredient names appear in expected, returns 0.
func ScoreDirectionAgreement(got []IngredientScore, expected map[string]string) float64 {
	if len(expected) == 0 {
		return 1.0
	}
	hits, total := 0, 0
	for _, s := range got {
		want, ok := expected[s.Name]
		if !ok {
			continue
		}
		total++
		if scoreDirection(s.Score) == strings.ToLower(want) {
			hits++
		}
	}
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total)
}

// scoreDirection maps a 1-10 score to "good" (≥7), "bad" (≤3), or "neutral" (4–6).
func scoreDirection(s float64) string {
	switch {
	case s >= 7:
		return "good"
	case s <= 3:
		return "bad"
	default:
		return "neutral"
	}
}
