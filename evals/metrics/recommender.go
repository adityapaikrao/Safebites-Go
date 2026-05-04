package metrics

import "strings"

// Recommendation holds the recommender agent's output for one product.
type Recommendation struct {
	Name  string
	Score float64
}

// DistinctFromInput returns true iff no recommendation's name equals the
// input product name (case-insensitive, whitespace-trimmed).
// Gated at floor=1.0 — recommending the input product back is a hard fail.
func DistinctFromInput(recs []Recommendation, input string) bool {
	inLow := strings.ToLower(strings.TrimSpace(input))
	for _, r := range recs {
		if strings.ToLower(strings.TrimSpace(r.Name)) == inLow {
			return false
		}
	}
	return true
}

// AllScoresImproved returns true iff every recommendation's score is at
// least inputScore+minDelta. Empty recs → false.
func AllScoresImproved(recs []Recommendation, inputScore, minDelta float64) bool {
	if len(recs) == 0 {
		return false
	}
	for _, r := range recs {
		if r.Score < inputScore+minDelta {
			return false
		}
	}
	return true
}

// CountEqualsThree returns true iff exactly three recommendations were returned.
func CountEqualsThree(recs []Recommendation) bool {
	return len(recs) == 3
}
