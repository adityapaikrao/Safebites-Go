package metrics

import (
	"testing"
)

func TestDistinctFromInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		recs  []Recommendation
		input string
		want  bool
	}{
		{
			name: "no recs match input",
			recs: []Recommendation{
				{Name: "Brand A Granola", Score: 0},
				{Name: "Brand B", Score: 0},
			},
			input: "Original Cereal",
			want:  true,
		},
		{
			name: "rec equals input case-insensitive",
			recs: []Recommendation{
				{Name: "Original Cereal", Score: 0},
				{Name: "Other", Score: 0},
			},
			input: "original cereal",
			want:  false,
		},
		{
			name:  "empty recs",
			recs:  []Recommendation{},
			input: "anything",
			want:  true,
		},
		{
			name: "whitespace-insensitive",
			recs: []Recommendation{
				{Name: "  Original Cereal  ", Score: 0},
			},
			input: "Original Cereal",
			want:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := DistinctFromInput(tt.recs, tt.input)
			if got != tt.want {
				t.Errorf("DistinctFromInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllScoresImproved(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		recs       []Recommendation
		inputScore float64
		minDelta   float64
		want       bool
	}{
		{
			name: "all improve by delta",
			recs: []Recommendation{
				{Name: "A", Score: 8},
				{Name: "B", Score: 7.5},
				{Name: "C", Score: 9},
			},
			inputScore: 6.0,
			minDelta:   1.0,
			want:       true,
		},
		{
			name: "one fails delta",
			recs: []Recommendation{
				{Name: "A", Score: 7.4},
			},
			inputScore: 6.0,
			minDelta:   1.5,
			want:       false,
		},
		{
			name:       "empty recs",
			recs:       []Recommendation{},
			inputScore: 5.0,
			minDelta:   1.0,
			want:       false,
		},
		{
			name: "exact boundary passes",
			recs: []Recommendation{
				{Name: "A", Score: 7.0},
			},
			inputScore: 6.0,
			minDelta:   1.0,
			want:       true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := AllScoresImproved(tt.recs, tt.inputScore, tt.minDelta)
			if got != tt.want {
				t.Errorf("AllScoresImproved() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountEqualsThree(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		recs []Recommendation
		want bool
	}{
		{
			name: "exactly three",
			recs: []Recommendation{
				{Name: "A", Score: 8},
				{Name: "B", Score: 7},
				{Name: "C", Score: 9},
			},
			want: true,
		},
		{
			name: "two",
			recs: []Recommendation{
				{Name: "A", Score: 8},
				{Name: "B", Score: 7},
			},
			want: false,
		},
		{
			name: "five",
			recs: []Recommendation{
				{Name: "A", Score: 8},
				{Name: "B", Score: 7},
				{Name: "C", Score: 9},
				{Name: "D", Score: 8.5},
				{Name: "E", Score: 7.5},
			},
			want: false,
		},
		{
			name: "zero",
			recs: []Recommendation{},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CountEqualsThree(tt.recs)
			if got != tt.want {
				t.Errorf("CountEqualsThree() = %v, want %v", got, tt.want)
			}
		})
	}
}
