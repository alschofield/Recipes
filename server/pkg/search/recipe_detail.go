package search

import (
	"encoding/json"
	"net/http"

	"recipes/pkg/response"
	"recipes/pkg/storage/postgres"
)

type RecipeDetailResponse struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	Description  string               `json:"description,omitempty"`
	Difficulty   string               `json:"difficulty,omitempty"`
	Cuisine      *string              `json:"cuisine,omitempty"`
	PrepMinutes  *int                 `json:"prepMinutes,omitempty"`
	CookMinutes  *int                 `json:"cookMinutes,omitempty"`
	TotalMinutes int                  `json:"totalMinutes"`
	Servings     int                  `json:"servings,omitempty"`
	Steps        []string             `json:"steps,omitempty"`
	Ingredients  []string             `json:"ingredients,omitempty"`
	Analysis     *RecipeQualityDetail `json:"analysis,omitempty"`
}

type RecipeQualityDetail struct {
	IngredientCoverageScore float64 `json:"ingredientCoverageScore"`
	NutritionBalanceScore   float64 `json:"nutritionBalanceScore"`
	FlavourAlignmentScore   float64 `json:"flavourAlignmentScore"`
	NoveltyScore            float64 `json:"noveltyScore"`
	OverallScore            float64 `json:"overallScore"`
	Status                  string  `json:"status"`
	Notes                   string  `json:"notes,omitempty"`
}

func GetRecipeDetail(w http.ResponseWriter, r *http.Request) {
	recipeID := r.PathValue("recipeid")
	if recipeID == "" {
		response.BadRequest(w, "Recipe ID is required")
		return
	}

	pool := storage.Pool()
	var out RecipeDetailResponse
	var stepsJSON []byte
	err := pool.QueryRow(r.Context(), `
		SELECT id, name, COALESCE(description, ''), difficulty, cuisine,
		       prep_minutes, cook_minutes, servings, steps
		FROM recipes WHERE id = $1`, recipeID,
	).Scan(&out.ID, &out.Name, &out.Description, &out.Difficulty, &out.Cuisine, &out.PrepMinutes, &out.CookMinutes, &out.Servings, &stepsJSON)
	if err != nil {
		response.NotFound(w, "Recipe not found")
		return
	}

	_ = json.Unmarshal(stepsJSON, &out.Steps)
	if out.PrepMinutes != nil {
		out.TotalMinutes += *out.PrepMinutes
	}
	if out.CookMinutes != nil {
		out.TotalMinutes += *out.CookMinutes
	}

	rows, err := pool.Query(r.Context(), `
		SELECT i.canonical_name
		FROM recipe_ingredients ri
		JOIN ingredients i ON i.id = ri.ingredient_id
		WHERE ri.recipe_id = $1
		ORDER BY ri.position ASC`, recipeID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ing string
			if rows.Scan(&ing) == nil {
				out.Ingredients = append(out.Ingredients, ing)
			}
		}
	}

	var analysis RecipeQualityDetail
	if err := pool.QueryRow(r.Context(), `
		SELECT ingredient_coverage_score::float8,
		       nutrition_balance_score::float8,
		       flavour_alignment_score::float8,
		       novelty_score::float8,
		       overall_score::float8,
		       status,
		       COALESCE(notes, '')
		FROM recipe_quality_analysis
		WHERE recipe_id = $1`, recipeID,
	).Scan(
		&analysis.IngredientCoverageScore,
		&analysis.NutritionBalanceScore,
		&analysis.FlavourAlignmentScore,
		&analysis.NoveltyScore,
		&analysis.OverallScore,
		&analysis.Status,
		&analysis.Notes,
	); err == nil {
		out.Analysis = &analysis
	}

	response.WriteJSON(w, http.StatusOK, out)
}
