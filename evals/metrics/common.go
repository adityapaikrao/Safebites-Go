package metrics

import (
	"encoding/json"
	"io"
	"math"
	"strings"
)

// Pair stores a prediction and reference value for error metrics.
type Pair struct {
	Got  float64
	Want float64
}

// CaseResult captures evaluation output for one test case.
type CaseResult struct {
	ID        string
	Pass      bool
	Err       string
	Metrics   map[string]float64
	LatencyMs int64
}

// JSONValid reports whether s contains exactly one valid JSON value.
func JSONValid(s string) bool {
	dec := json.NewDecoder(strings.NewReader(s))
	var v any
	if err := dec.Decode(&v); err != nil {
		return false
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return false
	}

	return true
}

// LevenshteinRatio computes 1 - distance/max(len(a), len(b)).
// For two empty strings, it returns 1.
func LevenshteinRatio(a, b string) float64 {
	ar := []rune(a)
	br := []rune(b)
	maxLen := maxInt(len(ar), len(br))
	if maxLen == 0 {
		return 1
	}

	distance := levenshteinDistance(ar, br)
	return 1 - float64(distance)/float64(maxLen)
}

// MeanAbsoluteError returns the mean absolute difference between each pair.
func MeanAbsoluteError(pairs []Pair) float64 {
	if len(pairs) == 0 {
		return 0
	}

	sum := 0.0
	for _, p := range pairs {
		sum += math.Abs(p.Got - p.Want)
	}
	return sum / float64(len(pairs))
}

// AggregateRate returns the arithmetic mean of vals, or 0 for empty input.
func AggregateRate(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func levenshteinDistance(a, b []rune) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}

			deletion := prev[j] + 1
			insertion := curr[j-1] + 1
			substitution := prev[j-1] + cost
			curr[j] = minInt3(deletion, insertion, substitution)
		}
		prev, curr = curr, prev
	}

	return prev[len(b)]
}

func minInt3(a, b, c int) int {
	if a > b {
		a = b
	}
	if a > c {
		return c
	}
	return a
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
