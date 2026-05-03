package metrics

import (
	"math"
	"testing"
)

func TestIngredientRecallAndPrecision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		got       []string
		expected  []string
		wantRecall float64
		wantPrecision float64
	}{
		{
			name:           "basic recall and precision",
			got:            []string{"sugar", "salt", "Yeast"},
			expected:       []string{"Sugar", "Wheat Flour", "Salt"},
			wantRecall:     2.0 / 3.0,
			wantPrecision:  2.0 / 3.0,
		},
		{
			name:           "empty expected returns recall 1.0",
			got:            []string{"sugar", "salt"},
			expected:       []string{},
			wantRecall:     1.0,
			wantPrecision:  0.0,
		},
		{
			name:           "empty got returns recall 0 precision 0",
			got:            []string{},
			expected:       []string{"Sugar", "Salt"},
			wantRecall:     0.0,
			wantPrecision:  0.0,
		},
		{
			name:           "all match",
			got:            []string{"Sugar", "Salt"},
			expected:       []string{"sugar", "salt"},
			wantRecall:     1.0,
			wantPrecision:  1.0,
		},
		{
			name:           "no match",
			got:            []string{"butter", "milk"},
			expected:       []string{"sugar", "salt"},
			wantRecall:     0.0,
			wantPrecision:  0.0,
		},
		{
			name:           "substring match case-insensitive",
			got:            []string{"enriched wheat flour"},
			expected:       []string{"wheat flour"},
			wantRecall:     1.0,
			wantPrecision:  1.0,
		},
		{
			name:           "reverse substring match",
			got:            []string{"wheat"},
			expected:       []string{"enriched wheat flour"},
			wantRecall:     1.0,
			wantPrecision:  1.0,
		},
		{
			name:           "whitespace handling",
			got:            []string{"  sugar  ", "salt"},
			expected:       []string{"Sugar", "Salt"},
			wantRecall:     1.0,
			wantPrecision:  1.0,
		},
		{
			name:           "both empty",
			got:            []string{},
			expected:       []string{},
			wantRecall:     1.0,
			wantPrecision:  0.0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			recall := IngredientRecall(tt.got, tt.expected)
			precision := IngredientPrecision(tt.got, tt.expected)

			if math.Abs(recall-tt.wantRecall) > 1e-9 {
				t.Fatalf("IngredientRecall(%v, %v) = %v, want %v", tt.got, tt.expected, recall, tt.wantRecall)
			}
			if math.Abs(precision-tt.wantPrecision) > 1e-9 {
				t.Fatalf("IngredientPrecision(%v, %v) = %v, want %v", tt.got, tt.expected, precision, tt.wantPrecision)
			}
		})
	}
}

func TestIngredientRecall_EmptyExpected(t *testing.T) {
	t.Parallel()

	got := []string{"sugar", "salt"}
	expected := []string{}
	want := 1.0

	result := IngredientRecall(got, expected)
	if math.Abs(result-want) > 1e-9 {
		t.Fatalf("IngredientRecall(%v, %v) = %v, want %v", got, expected, result, want)
	}
}
