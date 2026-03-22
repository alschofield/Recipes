package models

import "time"

// User represents a registered account.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Ingredient is a canonical cooking ingredient with optional aliases.
type Ingredient struct {
	ID                   string     `json:"id"`
	CanonicalName        string     `json:"canonicalName"`
	Category             *string    `json:"category,omitempty"`
	NaturalSource        *string    `json:"naturalSource,omitempty"`
	FlavourMoleculeCount *int       `json:"flavourMoleculeCount,omitempty"`
	SourceCoverage       int        `json:"sourceCoverage"`
	QualityScore         float64    `json:"qualityScore"`
	AnalysisStatus       string     `json:"analysisStatus"`
	AnalysisNotes        *string    `json:"analysisNotes,omitempty"`
	LastAnalyzedAt       *time.Time `json:"lastAnalyzedAt,omitempty"`
	Aliases              []string   `json:"aliases,omitempty"`
}

// Recipe is a cooking instruction set with metadata and provenance.
type Recipe struct {
	ID                  string             `json:"id"`
	Name                string             `json:"name"`
	Description         string             `json:"description,omitempty"`
	Steps               []string           `json:"steps"`
	Ingredients         []RecipeIngredient `json:"ingredients,omitempty"`
	Source              string             `json:"source"`
	GenerationModel     *string            `json:"generationModel,omitempty"`
	GenerationTimestamp *time.Time         `json:"generationTimestamp,omitempty"`
	PromptVersion       *string            `json:"promptVersion,omitempty"`
	Reviewable          bool               `json:"reviewable"`
	QualityScore        float64            `json:"qualityScore"`
	PrepMinutes         *int               `json:"prepMinutes,omitempty"`
	CookMinutes         *int               `json:"cookMinutes,omitempty"`
	Difficulty          string             `json:"difficulty"`
	Cuisine             *string            `json:"cuisine,omitempty"`
	Servings            int                `json:"servings"`
	DietaryTags         []string           `json:"dietaryTags,omitempty"`
	SafetyNotes         []string           `json:"safetyNotes,omitempty"`
	CreatedAt           time.Time          `json:"createdAt"`
	UpdatedAt           time.Time          `json:"updatedAt"`
}

// RecipeIngredient links a recipe to one of its ingredients.
type RecipeIngredient struct {
	ID            string `json:"id"`
	RecipeID      string `json:"recipeId"`
	IngredientID  string `json:"ingredientID"`
	CanonicalName string `json:"canonicalName,omitempty"`
	Amount        string `json:"amount,omitempty"`
	Unit          string `json:"unit,omitempty"`
	Optional      bool   `json:"optional"`
}

// Favorite is a user's saved recipe.
type Favorite struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	RecipeID  string    `json:"recipeId"`
	CreatedAt time.Time `json:"createdAt"`
}

// SearchResult wraps a search response with metadata.
type SearchResult struct {
	Mode       string               `json:"mode"`
	Query      SearchQuery          `json:"query"`
	Pagination PaginationResult     `json:"pagination"`
	Results    []SearchResultRecipe `json:"results"`
}

// SearchQuery is the normalized query that was executed.
type SearchQuery struct {
	Ingredients []string `json:"ingredients"`
}

// PaginationResult is the pagination metadata for a search response.
type PaginationResult struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

// SearchResultRecipe is a single recipe result with match info.
type SearchResultRecipe struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Source             string   `json:"source"`
	MatchPercent       float64  `json:"matchPercent"`
	MatchedIngredients []string `json:"matchedIngredients"`
	MissingIngredients []string `json:"missingIngredients"`
	PrepMinutes        *int     `json:"prepMinutes,omitempty"`
	CookMinutes        *int     `json:"cookMinutes,omitempty"`
	Difficulty         string   `json:"difficulty"`
	Cuisine            *string  `json:"cuisine,omitempty"`
}

// OptionalSubstitution maps a missing ingredient to alternatives.
type OptionalSubstitution struct {
	Missing     string   `json:"missing"`
	Substitutes []string `json:"substitutes"`
}

// RecipeQualityAnalysis stores post-generation quality component scores.
type RecipeQualityAnalysis struct {
	ID                      string     `json:"id"`
	RecipeID                string     `json:"recipeId"`
	IngredientCoverageScore float64    `json:"ingredientCoverageScore"`
	NutritionBalanceScore   float64    `json:"nutritionBalanceScore"`
	FlavourAlignmentScore   float64    `json:"flavourAlignmentScore"`
	NoveltyScore            float64    `json:"noveltyScore"`
	OverallScore            float64    `json:"overallScore"`
	Status                  string     `json:"status"`
	Notes                   *string    `json:"notes,omitempty"`
	ComputedAt              *time.Time `json:"computedAt,omitempty"`
	CreatedAt               time.Time  `json:"createdAt"`
	UpdatedAt               time.Time  `json:"updatedAt"`
}
