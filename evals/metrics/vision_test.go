package metrics

import (
	"testing"
)

func TestVisionExactMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  string
		want string
		ok   bool
	}{
		{name: "case-insensitive exact match", got: "Coca-Cola Zero", want: "coca-cola zero", ok: true},
		{name: "whitespace-trimmed exact match", got: "  Pepsi  ", want: "Pepsi", ok: true},
		{name: "different strings should not match", got: "A", want: "B", ok: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := VisionExactMatch(tt.got, tt.want)
			if got != tt.ok {
				t.Fatalf("VisionExactMatch(%q, %q) = %v, want %v", tt.got, tt.want, got, tt.ok)
			}
		})
	}
}

func TestVisionFuzzyMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		got       string
		want      string
		threshold float64
		ok        bool
	}{
		{name: "close strings within threshold", got: "Coca-Cola Zero", want: "Coca Cola Zero Sugar", threshold: 0.60, ok: true},
		{name: "very different strings should not match", got: "apple", want: "carburetor", threshold: 0.85, ok: false},
		{name: "threshold 0.0 matches any non-empty", got: "foo", want: "bar", threshold: 0.0, ok: true},
		{name: "threshold 1.0 only matches identical", got: "hello", want: "hello", threshold: 1.0, ok: true},
		{name: "threshold 1.0 rejects similar", got: "hello", want: "hallo", threshold: 1.0, ok: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := VisionFuzzyMatch(tt.got, tt.want, tt.threshold)
			if got != tt.ok {
				t.Fatalf("VisionFuzzyMatch(%q, %q, %v) = %v, want %v", tt.got, tt.want, tt.threshold, got, tt.ok)
			}
		})
	}
}

func TestVisionEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  string
		want bool
	}{
		{name: "empty string", got: "", want: true},
		{name: "whitespace-only", got: "   ", want: true},
		{name: "refusal: i cannot", got: "I cannot identify the product", want: true},
		{name: "refusal: i can't", got: "I can't tell what product this is", want: true},
		{name: "refusal: i am unable", got: "I am unable to determine the brand", want: true},
		{name: "refusal: unable to identify", got: "Unfortunately unable to identify this item", want: true},
		{name: "refusal: no product", got: "There is no product visible here", want: true},
		{name: "refusal: cannot identify", got: "Sorry, I cannot identify what brand this is", want: true},
		{name: "real product name", got: "Pepsi", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := VisionEmpty(tt.got)
			if got != tt.want {
				t.Fatalf("VisionEmpty(%q) = %v, want %v", tt.got, got, tt.want)
			}
		})
	}
}
