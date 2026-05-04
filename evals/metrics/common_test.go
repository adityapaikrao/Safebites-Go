package metrics

import (
	"math"
	"testing"
)

func TestLevenshteinRatio(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    string
		b    string
		want float64
	}{
		{name: "identical", a: "hello", b: "hello", want: 1.0},
		{name: "both empty", a: "", b: "", want: 1.0},
		{name: "non-empty vs empty", a: "abc", b: "", want: 0.0},
		{name: "kitten sitting", a: "kitten", b: "sitting", want: 1.0 - 3.0/7.0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := LevenshteinRatio(tt.a, tt.b)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Fatalf("LevenshteinRatio(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestJSONValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "valid object", in: `{"a":1}`, want: true},
		{name: "invalid object", in: `{"a":}`, want: false},
		{name: "prefixed junk", in: `prefix {"a":1}`, want: false},
		{name: "multiple values", in: `{"a":1} {"b":2}`, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := JSONValid(tt.in)
			if got != tt.want {
				t.Fatalf("JSONValid(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestMeanAbsoluteError(t *testing.T) {
	t.Parallel()

	t.Run("non-empty", func(t *testing.T) {
		t.Parallel()
		pairs := []Pair{{Got: 5, Want: 5}, {Got: 6, Want: 4}, {Got: 0, Want: 3}}
		want := (0.0 + 2.0 + 3.0) / 3.0
		got := MeanAbsoluteError(pairs)
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("MeanAbsoluteError() = %v, want %v", got, want)
		}
	})

	t.Run("empty returns zero", func(t *testing.T) {
		t.Parallel()
		got := MeanAbsoluteError(nil)
		if got != 0 {
			t.Fatalf("MeanAbsoluteError(nil) = %v, want 0", got)
		}
	})
}

func TestAggregateRate(t *testing.T) {
	t.Parallel()

	t.Run("non-empty sample mean", func(t *testing.T) {
		t.Parallel()
		vals := []float64{1, 2, 3, 4}
		want := 2.5
		got := AggregateRate(vals)
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("AggregateRate(%v) = %v, want %v", vals, got, want)
		}
	})

	t.Run("empty returns zero", func(t *testing.T) {
		t.Parallel()
		got := AggregateRate(nil)
		if got != 0 {
			t.Fatalf("AggregateRate(nil) = %v, want 0", got)
		}
	})
}
