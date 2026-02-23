package model

import (
	"encoding/json"
	"fmt"
)

// FlexibleString accepts both JSON strings and numbers, converting numbers to
// their string representation. This is useful when an LLM may return either
// form despite being instructed to return a string.
type FlexibleString string

func (f *FlexibleString) UnmarshalJSON(data []byte) error {
	// Try string first.
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexibleString(s)
		return nil
	}
	// Fall back to number.
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexibleString(n.String())
		return nil
	}
	return fmt.Errorf("flexibleString: cannot unmarshal %s", string(data))
}

type Ingredient struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type WebSearchResult struct {
	ListOfIngredients []Ingredient `json:"List_of_ingredients"`
}

type IngredientScore struct {
	IngredientName string         `json:"ingredient_name"`
	SafetyScore    FlexibleString `json:"safety_score"`
	Reasoning      string         `json:"reasoning"`
}

type ScorerResult struct {
	IngredientScores []IngredientScore `json:"ingredient_scores"`
	OverallScore     float64           `json:"overall_score"`
}

type Recommendation struct {
	ProductName string         `json:"product_name"`
	HealthScore FlexibleString `json:"health_score"`
	Reason      string         `json:"reason"`
}

type RecommenderResult struct {
	Recommendations []Recommendation `json:"recommendations"`
}
