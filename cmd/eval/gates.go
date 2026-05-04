package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
)

type baselineFile struct {
	Agents map[string]map[string]float64 `json:"agents"`
}

// gateRule describes how to evaluate a single (agent, metric).
type gateRule struct {
	Floor          *float64 // hard floor — fail if < floor
	HigherIsWorse  bool     // for metrics like empty_rate / score_mae
	TripwireDeltaP float64  // tripwire threshold in absolute points (0.10 = 10pp)
}

// gateRules captures the locked decisions from the spec (B3-light).
var gateRules = map[string]map[string]gateRule{
	"vision": {
		"exact_match": {TripwireDeltaP: 0.10},
		"fuzzy_match": {TripwireDeltaP: 0.10},
		"empty_rate":  {TripwireDeltaP: 0.10, HigherIsWorse: true},
	},
	"search": {
		"ingredient_recall":    {TripwireDeltaP: 0.10},
		"ingredient_precision": {TripwireDeltaP: 0.10},
		"json_valid":           {Floor: ptr(0.9)},
	},
	"scorer": {
		"allergy_respect_rate":      {Floor: ptr(1.0)},
		"score_direction_agreement": {TripwireDeltaP: 0.10},
		"overall_score_mae":         {TripwireDeltaP: 0.10, HigherIsWorse: true},
		"json_valid":                {Floor: ptr(0.9)},
	},
	"recommender": {
		"distinct_from_input":    {Floor: ptr(1.0)},
		"score_improvement_rate": {TripwireDeltaP: 0.10},
		"count_equals_three":     {TripwireDeltaP: 0.10},
		"json_valid":             {Floor: ptr(0.9)},
	},
}

func ptr(f float64) *float64 { return &f }

// EvaluateGates returns the list of human-readable failure messages.
// An empty slice means all gates passed.
func EvaluateGates(report *Report, baselinePath string) []string {
	var failures []string

	base := baselineFile{Agents: map[string]map[string]float64{}}
	if b, err := os.ReadFile(baselinePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			failures = append(failures, fmt.Sprintf("baseline %s could not be read: %v", baselinePath, err))
			return failures
		}
		// File not found → no tripwire gating, only floor gating.
	} else if err := json.Unmarshal(b, &base); err != nil {
		failures = append(failures, fmt.Sprintf("baseline %s could not be parsed: %v", baselinePath, err))
		return failures
	}

	for agent, rules := range gateRules {
		observed := report.Agents[agent]
		if observed == nil {
			continue // agent not run this time — skip
		}
		for metric, rule := range rules {
			v, ok := observed[metric]
			if !ok {
				continue
			}
			if rule.Floor != nil {
				if rule.HigherIsWorse {
					if v > *rule.Floor {
						failures = append(failures, fmt.Sprintf("%s.%s = %.4f > floor %.4f (higher is worse)", agent, metric, v, *rule.Floor))
					}
				} else {
					if v < *rule.Floor {
						failures = append(failures, fmt.Sprintf("%s.%s = %.4f < floor %.4f", agent, metric, v, *rule.Floor))
					}
				}
			}
			if rule.TripwireDeltaP > 0 {
				baseVal, hasBase := base.Agents[agent][metric]
				if !hasBase {
					continue // empty baseline → tripwire passes
				}
				delta := v - baseVal
				if rule.HigherIsWorse {
					if delta > rule.TripwireDeltaP {
						failures = append(failures, fmt.Sprintf("%s.%s rose by %.4f (>%.0fpp tripwire); base=%.4f new=%.4f", agent, metric, delta, rule.TripwireDeltaP*100, baseVal, v))
					}
				} else {
					if -delta > rule.TripwireDeltaP {
						failures = append(failures, fmt.Sprintf("%s.%s dropped by %.4f (>%.0fpp tripwire); base=%.4f new=%.4f", agent, metric, math.Abs(delta), rule.TripwireDeltaP*100, baseVal, v))
					}
				}
			}
		}
	}
	return failures
}
