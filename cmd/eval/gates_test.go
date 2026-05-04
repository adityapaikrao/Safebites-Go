package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeBaseline writes content to a temp file and returns its path.
func writeBaseline(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "baseline.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeBaseline: %v", err)
	}
	return path
}

func TestEvaluateGates(t *testing.T) {
	t.Run("missing baseline - floor violation fires", func(t *testing.T) {
		// allergy_respect_rate floor = 1.0, value = 0.8 → should fail
		report := &Report{
			Agents: map[string]map[string]float64{
				"scorer": {"allergy_respect_rate": 0.8},
			},
		}
		nonExistent := filepath.Join(t.TempDir(), "no_such_file.json")
		failures := EvaluateGates(report, nonExistent)
		if len(failures) == 0 {
			t.Fatal("expected floor failure for allergy_respect_rate=0.8, got none")
		}
		found := false
		for _, f := range failures {
			if strings.Contains(f, "allergy_respect_rate") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected failure mentioning allergy_respect_rate, got: %v", failures)
		}
	})

	t.Run("missing baseline - tripwire-only metric passes", func(t *testing.T) {
		// exact_match has no floor, only tripwire — with no baseline, tripwire is skipped
		report := &Report{
			Agents: map[string]map[string]float64{
				"vision": {"exact_match": 0.5},
			},
		}
		nonExistent := filepath.Join(t.TempDir(), "no_such_file.json")
		failures := EvaluateGates(report, nonExistent)
		if len(failures) != 0 {
			t.Errorf("expected no failures for tripwire-only metric with missing baseline, got: %v", failures)
		}
	})

	t.Run("unparseable baseline - single failure returned immediately", func(t *testing.T) {
		report := &Report{
			Agents: map[string]map[string]float64{
				"scorer": {"allergy_respect_rate": 0.8},
			},
		}
		dir := t.TempDir()
		path := writeBaseline(t, dir, `{not valid json}`)
		failures := EvaluateGates(report, path)
		if len(failures) != 1 {
			t.Fatalf("expected exactly 1 failure for bad baseline, got %d: %v", len(failures), failures)
		}
		if !strings.Contains(failures[0], "could not be parsed") {
			t.Errorf("expected parse error message, got: %s", failures[0])
		}
	})

	t.Run("floor violation - json_valid below 0.9", func(t *testing.T) {
		report := &Report{
			Agents: map[string]map[string]float64{
				"search": {"json_valid": 0.85},
			},
		}
		nonExistent := filepath.Join(t.TempDir(), "no_such_file.json")
		failures := EvaluateGates(report, nonExistent)
		if len(failures) == 0 {
			t.Fatal("expected floor failure for json_valid=0.85, got none")
		}
		found := false
		for _, f := range failures {
			if strings.Contains(f, "json_valid") && strings.Contains(f, "floor") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected floor failure for json_valid, got: %v", failures)
		}
	})

	t.Run("floor pass - json_valid above 0.9", func(t *testing.T) {
		report := &Report{
			Agents: map[string]map[string]float64{
				"search": {"json_valid": 0.95},
			},
		}
		nonExistent := filepath.Join(t.TempDir(), "no_such_file.json")
		failures := EvaluateGates(report, nonExistent)
		if len(failures) != 0 {
			t.Errorf("expected no failures for json_valid=0.95, got: %v", failures)
		}
	})

	t.Run("tripwire violation - exact_match drops > 10pp", func(t *testing.T) {
		// baseline: 0.80, new: 0.65 → delta = -0.15 > 0.10pp
		report := &Report{
			Agents: map[string]map[string]float64{
				"vision": {"exact_match": 0.65},
			},
		}
		dir := t.TempDir()
		baseline := `{"agents": {"vision": {"exact_match": 0.80}}}`
		path := writeBaseline(t, dir, baseline)
		failures := EvaluateGates(report, path)
		if len(failures) == 0 {
			t.Fatal("expected tripwire failure for exact_match drop, got none")
		}
		found := false
		for _, f := range failures {
			if strings.Contains(f, "exact_match") && strings.Contains(f, "dropped") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected tripwire drop message for exact_match, got: %v", failures)
		}
	})

	t.Run("tripwire pass - exact_match drops < 10pp", func(t *testing.T) {
		// baseline: 0.80, new: 0.75 → delta = -0.05 < 0.10pp
		report := &Report{
			Agents: map[string]map[string]float64{
				"vision": {"exact_match": 0.75},
			},
		}
		dir := t.TempDir()
		baseline := `{"agents": {"vision": {"exact_match": 0.80}}}`
		path := writeBaseline(t, dir, baseline)
		failures := EvaluateGates(report, path)
		if len(failures) != 0 {
			t.Errorf("expected no failures for small exact_match drop, got: %v", failures)
		}
	})

	t.Run("tripwire violation HigherIsWorse - empty_rate rises > 10pp", func(t *testing.T) {
		// baseline: 0.10, new: 0.25 → delta = +0.15 > 0.10pp
		report := &Report{
			Agents: map[string]map[string]float64{
				"vision": {"empty_rate": 0.25},
			},
		}
		dir := t.TempDir()
		baseline := `{"agents": {"vision": {"empty_rate": 0.10}}}`
		path := writeBaseline(t, dir, baseline)
		failures := EvaluateGates(report, path)
		if len(failures) == 0 {
			t.Fatal("expected tripwire failure for empty_rate rise, got none")
		}
		found := false
		for _, f := range failures {
			if strings.Contains(f, "empty_rate") && strings.Contains(f, "rose") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected tripwire rise message for empty_rate, got: %v", failures)
		}
	})

	t.Run("tripwire pass HigherIsWorse - empty_rate rises < 10pp", func(t *testing.T) {
		// baseline: 0.10, new: 0.15 → delta = +0.05 < 0.10pp
		report := &Report{
			Agents: map[string]map[string]float64{
				"vision": {"empty_rate": 0.15},
			},
		}
		dir := t.TempDir()
		baseline := `{"agents": {"vision": {"empty_rate": 0.10}}}`
		path := writeBaseline(t, dir, baseline)
		failures := EvaluateGates(report, path)
		if len(failures) != 0 {
			t.Errorf("expected no failures for small empty_rate rise, got: %v", failures)
		}
	})

	t.Run("agent not in report - skipped without error", func(t *testing.T) {
		// Report has no "vision" key → vision rules should be skipped
		report := &Report{
			Agents: map[string]map[string]float64{
				// vision deliberately absent
			},
		}
		nonExistent := filepath.Join(t.TempDir(), "no_such_file.json")
		failures := EvaluateGates(report, nonExistent)
		if len(failures) != 0 {
			t.Errorf("expected no failures when agent is absent from report, got: %v", failures)
		}
	})

	t.Run("all gates pass - empty failure slice", func(t *testing.T) {
		report := &Report{
			Agents: map[string]map[string]float64{
				"vision": {
					"exact_match": 0.90,
					"fuzzy_match": 0.92,
					"empty_rate":  0.05,
				},
				"search": {
					"ingredient_recall":    0.85,
					"ingredient_precision": 0.88,
					"json_valid":           0.95,
				},
				"scorer": {
					"allergy_respect_rate":      1.0,
					"score_direction_agreement": 0.80,
					"overall_score_mae":         0.10,
					"json_valid":                0.95,
				},
				"recommender": {
					"distinct_from_input":    1.0,
					"score_improvement_rate": 0.85,
					"count_equals_three":     0.90,
					"json_valid":             0.95,
				},
			},
		}
		dir := t.TempDir()
		// Baseline values slightly below current — no regressions
		baseline := `{
			"agents": {
				"vision": {
					"exact_match": 0.88,
					"fuzzy_match": 0.90,
					"empty_rate":  0.06
				},
				"search": {
					"ingredient_recall":    0.83,
					"ingredient_precision": 0.86
				},
				"scorer": {
					"score_direction_agreement": 0.78,
					"overall_score_mae": 0.11
				},
				"recommender": {
					"score_improvement_rate": 0.83,
					"count_equals_three":     0.88
				}
			}
		}`
		path := writeBaseline(t, dir, baseline)
		failures := EvaluateGates(report, path)
		if len(failures) != 0 {
			t.Errorf("expected all gates to pass, got failures: %v", failures)
		}
	})
}
