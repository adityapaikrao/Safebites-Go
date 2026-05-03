package metrics

import "strings"

func VisionExactMatch(got, want string) bool {
	return strings.EqualFold(strings.TrimSpace(got), strings.TrimSpace(want))
}

func VisionFuzzyMatch(got, want string, threshold float64) bool {
	return LevenshteinRatio(strings.ToLower(strings.TrimSpace(got)), strings.ToLower(strings.TrimSpace(want))) >= threshold
}

// VisionEmpty reports refusal/empty outputs that should count against the
// empty_rate metric (lower is better, so a `true` here is "agent failed").
func VisionEmpty(got string) bool {
	t := strings.ToLower(strings.TrimSpace(got))
	if t == "" {
		return true
	}
	refusalPrefixes := []string{
		"i cannot",
		"i can't",
		"i am unable",
		"unable to identify",
		"no product",
		"cannot identify",
	}
	for _, p := range refusalPrefixes {
		if strings.HasPrefix(t, p) {
			return true
		}
	}
	return false
}
