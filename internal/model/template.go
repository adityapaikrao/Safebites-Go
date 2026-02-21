package model

type DietaryTemplate struct {
	Key              string   `json:"key"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Allergies        []string `json:"allergies"`
	DietGoals        []string `json:"dietGoals"`
	AvoidIngredients []string `json:"avoidIngredients"`
}

var DietaryTemplates = map[string]DietaryTemplate{
	"vegan": {
		Key:              "vegan",
		Name:             "Vegan",
		Description:      "No animal products",
		Allergies:        []string{},
		DietGoals:        []string{"vegan"},
		AvoidIngredients: []string{"meat", "fish", "dairy", "eggs", "honey", "gelatin", "whey", "casein", "lactose", "shellac"},
	},
	"vegetarian": {
		Key:              "vegetarian",
		Name:             "Vegetarian",
		Description:      "No meat or fish",
		Allergies:        []string{},
		DietGoals:        []string{"vegetarian"},
		AvoidIngredients: []string{"meat", "fish", "gelatin", "shellac"},
	},
	"gluten_free": {
		Key:              "gluten_free",
		Name:             "Gluten-Free",
		Description:      "No gluten-containing grains",
		Allergies:        []string{"gluten"},
		DietGoals:        []string{"gluten-free"},
		AvoidIngredients: []string{"wheat", "barley", "rye", "malt", "brewer's yeast"},
	},
	"keto": {
		Key:              "keto",
		Name:             "Keto",
		Description:      "Low carb, high fat diet",
		Allergies:        []string{},
		DietGoals:        []string{"keto", "low-carb"},
		AvoidIngredients: []string{"sugar", "corn syrup", "high fructose corn syrup", "maltodextrin", "dextrose", "sucrose"},
	},
	"dairy_free": {
		Key:              "dairy_free",
		Name:             "Dairy-Free",
		Description:      "No dairy products",
		Allergies:        []string{"dairy"},
		DietGoals:        []string{"dairy-free"},
		AvoidIngredients: []string{"milk", "cream", "butter", "cheese", "yogurt", "whey", "casein", "lactose"},
	},
	"nut_free": {
		Key:              "nut_free",
		Name:             "Nut-Free",
		Description:      "No tree nuts or peanuts",
		Allergies:        []string{"tree nuts", "peanuts"},
		DietGoals:        []string{"nut-free"},
		AvoidIngredients: []string{"almonds", "cashews", "walnuts", "pecans", "pistachios", "hazelnuts", "macadamia", "brazil nuts", "peanuts"},
	},
	"paleo": {
		Key:              "paleo",
		Name:             "Paleo",
		Description:      "Whole foods, no processed ingredients",
		Allergies:        []string{},
		DietGoals:        []string{"paleo"},
		AvoidIngredients: []string{"sugar", "corn syrup", "soy", "legumes", "dairy", "grains", "processed oils", "artificial sweeteners"},
	},
}
