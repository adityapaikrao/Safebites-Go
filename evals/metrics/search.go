package metrics

import "strings"

// IngredientRecall = |expected ∩ got| / |expected|.
// Matching is case-insensitive substring match in both directions
// (so "wheat flour" matches "enriched wheat flour").
func IngredientRecall(got, expected []string) float64 {
	if len(expected) == 0 {
		return 1.0
	}
	hits := 0
	for _, e := range expected {
		if containsAnyMatch(e, got) {
			hits++
		}
	}
	return float64(hits) / float64(len(expected))
}

// IngredientPrecision = |expected ∩ got| / |got|.
func IngredientPrecision(got, expected []string) float64 {
	if len(got) == 0 {
		return 0.0
	}
	hits := 0
	for _, g := range got {
		if containsAnyMatch(g, expected) {
			hits++
		}
	}
	return float64(hits) / float64(len(got))
}

func containsAnyMatch(needle string, haystack []string) bool {
	n := strings.ToLower(strings.TrimSpace(needle))
	if n == "" {
		return false
	}
	for _, h := range haystack {
		h = strings.ToLower(strings.TrimSpace(h))
		if h == "" {
			continue
		}
		if strings.Contains(h, n) || strings.Contains(n, h) {
			return true
		}
	}
	return false
}
