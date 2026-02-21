package model

type Ingredient struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type WebSearchResult struct {
	ListOfIngredients []Ingredient `json:"List_of_ingredients"`
}

type IngredientScore struct {
	IngredientName string `json:"ingredient_name"`
	SafetyScore    string `json:"safety_score"`
	Reasoning      string `json:"reasoning"`
}

type ScorerResult struct {
	IngredientScores []IngredientScore `json:"ingredient_scores"`
	OverallScore     float64           `json:"overall_score"`
}

type Recommendation struct {
	ProductName string `json:"product_name"`
	HealthScore string `json:"health_score"`
	Reason      string `json:"reason"`
}

type RecommenderResult struct {
	Recommendations []Recommendation `json:"recommendations"`
}
