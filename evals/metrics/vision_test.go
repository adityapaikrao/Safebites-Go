package metrics_test

import (
	"testing"

	"github.com/safebites/backend-go/evals/metrics"
)

func TestVisionExactMatch(t *testing.T) {
	if !metrics.VisionExactMatch("Coca-Cola Zero", "coca-cola zero") {
		t.Error("case-insensitive exact match failed")
	}
	if !metrics.VisionExactMatch("  Pepsi  ", "Pepsi") {
		t.Error("whitespace-trimmed exact match failed")
	}
	if metrics.VisionExactMatch("A", "B") {
		t.Error("different strings should not match")
	}
}

func TestVisionFuzzyMatch(t *testing.T) {
	if !metrics.VisionFuzzyMatch("Coca-Cola Zero", "Coca Cola Zero Sugar", 0.65) {
		t.Error("close strings within threshold should match")
	}
	if metrics.VisionFuzzyMatch("apple", "carburetor", 0.85) {
		t.Error("very different strings should not match")
	}
}

func TestVisionEmpty(t *testing.T) {
	if !metrics.VisionEmpty("") {
		t.Error("empty string should be empty")
	}
	if !metrics.VisionEmpty("   ") {
		t.Error("whitespace-only should be empty")
	}
	if !metrics.VisionEmpty("I cannot identify the product") {
		t.Error("refusal phrase should be considered empty")
	}
	if metrics.VisionEmpty("Pepsi") {
		t.Error("real name should not be empty")
	}
}
