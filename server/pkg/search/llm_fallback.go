package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"recipes/pkg/ingredients"
	"recipes/pkg/storage/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	fallbackTopMatchThreshold = 0.45
	fallbackMinResultCount    = 5
	defaultPromptVersion      = "v1"
)

type llmRecipeResponse struct {
	Recipes []llmRecipe `json:"recipes"`
}

type llmRecipe struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Ingredients []llmRecipeIngredient `json:"ingredients"`
	Steps       []string              `json:"steps"`
	PrepMinutes int                   `json:"prepMinutes"`
	CookMinutes int                   `json:"cookMinutes"`
	Difficulty  string                `json:"difficulty"`
	Cuisine     string                `json:"cuisine"`
	DietaryTags []string              `json:"dietaryTags"`
	Servings    int                   `json:"servings"`
	SafetyNotes []string              `json:"safetyNotes"`
}

type llmRecipeIngredient struct {
	Name     string `json:"name"`
	Amount   string `json:"amount"`
	Optional bool   `json:"optional"`
}

func shouldTriggerLLMFallback(dbOnly bool, dbResults []SearchResultItem) bool {
	if dbOnly {
		return false
	}

	if len(dbResults) == 0 {
		return true
	}

	top := dbResults[0].MatchPercent
	return top < fallbackTopMatchThreshold && len(dbResults) < fallbackMinResultCount
}

func generateAndStoreFallbackRecipes(
	ctx context.Context,
	req SearchRequest,
	normalizedIngredients []string,
	aliases *aliasLookup,
) ([]SearchResultItem, error) {
	raw, model, err := callLLMForRecipes(ctx, req, normalizedIngredients)
	if err != nil {
		return nil, err
	}

	parsed, err := parseAndValidateLLMResponse(raw)
	if err != nil {
		return nil, err
	}

	pool := storage.Pool()
	inputSet := make(map[string]struct{}, len(normalizedIngredients))
	for _, ing := range normalizedIngredients {
		inputSet[ing] = struct{}{}
	}

	items := make([]SearchResultItem, 0, len(parsed.Recipes))
	for _, generated := range parsed.Recipes {
		item, err := persistGeneratedRecipeAndBuildResult(ctx, pool, generated, req, model, inputSet, aliases)
		if err != nil {
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

func callLLMForRecipes(ctx context.Context, req SearchRequest, normalizedIngredients []string) (string, string, error) {
	apiKey := strings.TrimSpace(os.Getenv("LLM_API_KEY"))
	if apiKey == "" {
		return "", "", errors.New("LLM_API_KEY is not configured")
	}

	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("LLM_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	model := strings.TrimSpace(os.Getenv("LLM_MODEL"))
	if model == "" {
		model = "gpt-4o-mini"
	}

	prompt := buildLLMPrompt(req, normalizedIngredients)
	body := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You generate safe, practical cooking recipes. Respond with valid JSON only."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.2,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", "", err
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 25 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("LLM provider error: %d", resp.StatusCode)
	}

	var providerResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&providerResp); err != nil {
		return "", "", err
	}

	if len(providerResp.Choices) == 0 || strings.TrimSpace(providerResp.Choices[0].Message.Content) == "" {
		return "", "", errors.New("empty LLM response")
	}

	return providerResp.Choices[0].Message.Content, model, nil
}

func buildLLMPrompt(req SearchRequest, normalizedIngredients []string) string {
	filters := "none"
	if req.Filters != nil {
		b, _ := json.Marshal(req.Filters)
		filters = string(b)
	}

	return fmt.Sprintf(`Generate 1-3 recipes for these normalized ingredients: %s.
Search mode: %s.
Filters: %s.

Return JSON ONLY with this exact schema:
{
  "recipes": [
    {
      "name": "string",
      "description": "string",
      "ingredients": [
        { "name": "string", "amount": "string", "optional": false }
      ],
      "steps": ["string"],
      "prepMinutes": 30,
      "cookMinutes": 20,
      "difficulty": "easy|medium|hard",
      "cuisine": "string",
      "dietaryTags": ["string"],
      "servings": 2,
      "safetyNotes": ["string"]
    }
  ]
}

Do not include markdown, code fences, or any extra fields.`, strings.Join(normalizedIngredients, ", "), req.Mode, filters)
}

func parseAndValidateLLMResponse(raw string) (*llmRecipeResponse, error) {
	var parsed llmRecipeResponse
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&parsed); err != nil {
		return nil, err
	}
	if decoder.More() {
		return nil, errors.New("schema invalid: multiple JSON values")
	}

	if len(parsed.Recipes) == 0 {
		return nil, errors.New("schema invalid: recipes must contain at least one item")
	}

	for i, recipe := range parsed.Recipes {
		if strings.TrimSpace(recipe.Name) == "" {
			return nil, fmt.Errorf("schema invalid: recipes[%d].name is required", i)
		}
		if len(recipe.Ingredients) == 0 {
			return nil, fmt.Errorf("schema invalid: recipes[%d].ingredients must be non-empty", i)
		}
		if len(recipe.Steps) == 0 {
			return nil, fmt.Errorf("schema invalid: recipes[%d].steps must be non-empty", i)
		}
		if recipe.PrepMinutes < 0 || recipe.CookMinutes < 0 {
			return nil, fmt.Errorf("schema invalid: recipes[%d] prep/cook minutes must be >= 0", i)
		}
		difficulty := strings.ToLower(strings.TrimSpace(recipe.Difficulty))
		if difficulty != "easy" && difficulty != "medium" && difficulty != "hard" {
			return nil, fmt.Errorf("schema invalid: recipes[%d].difficulty invalid", i)
		}
		for j, ing := range recipe.Ingredients {
			if strings.TrimSpace(ing.Name) == "" {
				return nil, fmt.Errorf("schema invalid: recipes[%d].ingredients[%d].name is required", i, j)
			}
			if strings.TrimSpace(ing.Amount) == "" {
				return nil, fmt.Errorf("schema invalid: recipes[%d].ingredients[%d].amount is required", i, j)
			}
		}
	}

	return &parsed, nil
}

func persistGeneratedRecipeAndBuildResult(
	ctx context.Context,
	pool *pgxpool.Pool,
	generated llmRecipe,
	req SearchRequest,
	model string,
	inputSet map[string]struct{},
	aliases *aliasLookup,
) (SearchResultItem, error) {
	type normalizedIngredient struct {
		canonical string
		amount    string
		optional  bool
	}

	recipeName := strings.TrimSpace(generated.Name)
	description := strings.TrimSpace(generated.Description)
	difficulty := strings.ToLower(strings.TrimSpace(generated.Difficulty))
	cuisine := strings.TrimSpace(generated.Cuisine)
	if cuisine == "" {
		cuisine = "global"
	}

	steps := make([]string, 0, len(generated.Steps))
	for _, step := range generated.Steps {
		s := strings.TrimSpace(step)
		if s != "" {
			steps = append(steps, s)
		}
	}
	stepsJSON, err := json.Marshal(steps)
	if err != nil {
		return SearchResultItem{}, err
	}

	prep := generated.PrepMinutes
	cook := generated.CookMinutes
	servings := generated.Servings
	if servings <= 0 {
		servings = 2
	}

	normalizedIngredients := []normalizedIngredient{}
	matched := []string{}
	missing := []string{}
	seenIng := map[string]struct{}{}

	for _, ing := range generated.Ingredients {
		canonical := normalizeRawIngredient(ing.Name)
		if canonical == "" {
			continue
		}

		if _, seen := seenIng[canonical]; seen {
			continue
		}
		seenIng[canonical] = struct{}{}

		normalizedIngredients = append(normalizedIngredients, normalizedIngredient{
			canonical: canonical,
			amount:    strings.TrimSpace(ing.Amount),
			optional:  ing.Optional,
		})

		if ing.Optional {
			continue
		}
		if _, ok := inputSet[canonical]; ok {
			matched = append(matched, canonical)
		} else {
			missing = append(missing, canonical)
		}
	}

	totalRequired := len(matched) + len(missing)
	matchPercent := 0.0
	if totalRequired > 0 {
		matchPercent = float64(len(matched)) / float64(totalRequired)
	}

	if req.Mode == "strict" && len(missing) > 0 {
		return SearchResultItem{}, errors.New("generated recipe not strict compatible")
	}

	totalMinutes := prep + cook
	grp := &recipeGroup{
		PrepMinutes: prepPtr(prep),
		CookMinutes: prepPtr(cook),
		Difficulty:  difficulty,
		Cuisine:     strPtr(cuisine),
		Servings:    servings,
		DietaryTags: generated.DietaryTags,
	}
	if !matchesFilters(grp, req.Filters, totalMinutes) {
		return SearchResultItem{}, errors.New("generated recipe filtered out")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return SearchResultItem{}, err
	}
	defer tx.Rollback(ctx)

	promptVersion := strings.TrimSpace(os.Getenv("LLM_PROMPT_VERSION"))
	if promptVersion == "" {
		promptVersion = defaultPromptVersion
	}

	var recipeID string
	now := time.Now().UTC()
	err = tx.QueryRow(ctx, `
		INSERT INTO recipes (
			name, description, steps, source,
			generation_model, generation_timestamp, prompt_version,
			reviewable, quality_score,
			prep_minutes, cook_minutes, difficulty, cuisine,
			servings, dietary_tags, safety_notes
		) VALUES (
			$1, $2, $3, 'llm',
			$4, $5, $6,
			TRUE, 0.40,
			$7, $8, $9, $10,
			$11, $12, $13
		)
		RETURNING id`,
		recipeName,
		description,
		stepsJSON,
		model,
		now,
		promptVersion,
		prep,
		cook,
		difficulty,
		cuisine,
		servings,
		generated.DietaryTags,
		generated.SafetyNotes,
	).Scan(&recipeID)
	if err != nil {
		return SearchResultItem{}, err
	}

	for position, ing := range normalizedIngredients {
		resolved, err := ingredients.ResolveOrCreateForLLM(ctx, pool, ing.canonical)
		if err != nil || resolved.IngredientID == "" {
			continue
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO recipe_ingredients (recipe_id, ingredient_id, amount, unit, optional, position)
			VALUES ($1, $2, $3, '', $4, $5)`,
			recipeID,
			resolved.IngredientID,
			ing.amount,
			ing.optional,
			position,
		)
		if err != nil {
			return SearchResultItem{}, err
		}
	}

	if err := upsertRecipeQualityAnalysis(ctx, tx, recipeID); err != nil {
		return SearchResultItem{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return SearchResultItem{}, err
	}

	substitutions := []Substitution{}
	if aliases != nil {
		for _, m := range missing {
			if subs, ok := aliases.canonicalToAliases[m]; ok && len(subs) > 0 {
				substitutions = append(substitutions, Substitution{Missing: m, Substitutes: subs})
			}
		}
	}

	return SearchResultItem{
		ID:                 recipeID,
		Name:               recipeName,
		Source:             "llm",
		MatchPercent:       matchPercent,
		MatchedIngredients: matched,
		MissingIngredients: missing,
		Substitutions:      substitutions,
		PrepMinutes:        prepPtr(prep),
		CookMinutes:        prepPtr(cook),
		TotalMinutes:       totalMinutes,
		Difficulty:         difficulty,
		Cuisine:            strPtr(cuisine),
		Servings:           servings,
		DietaryTags:        generated.DietaryTags,
	}, nil
}

func prepPtr(v int) *int {
	return &v
}

func strPtr(v string) *string {
	return &v
}

func upsertRecipeQualityAnalysis(ctx context.Context, tx pgx.Tx, recipeID string) error {
	type quality struct {
		IngredientCoverage float64
		NutritionBalance   float64
		FlavourAlignment   float64
		Novelty            float64
		Overall            float64
	}

	var q quality
	err := tx.QueryRow(ctx, `
		WITH stats AS (
			SELECT
				COUNT(*)::numeric AS ingredient_count,
				COALESCE(AVG(LEAST(GREATEST(i.source_coverage, 0)::numeric / 5.0, 1.0)), 0) AS ingredient_coverage,
				COALESCE(AVG(LEAST(COALESCE(i.flavour_molecule_count, 0)::numeric / 200.0, 1.0)), 0) AS flavour_alignment,
				COALESCE(COUNT(DISTINCT NULLIF(i.category, ''))::numeric / NULLIF(COUNT(*), 0), 0) AS novelty
			FROM recipe_ingredients ri
			JOIN ingredients i ON i.id = ri.ingredient_id
			WHERE ri.recipe_id = $1
			  AND ri.optional = FALSE
		)
		SELECT
			ingredient_coverage,
			CASE
				WHEN ingredient_count >= 6 THEN 0.75
				WHEN ingredient_count >= 4 THEN 0.65
				WHEN ingredient_count >= 2 THEN 0.55
				ELSE 0.35
			END AS nutrition_balance,
			flavour_alignment,
			novelty,
			(
				ingredient_coverage * 0.40 +
				(CASE
					WHEN ingredient_count >= 6 THEN 0.75
					WHEN ingredient_count >= 4 THEN 0.65
					WHEN ingredient_count >= 2 THEN 0.55
					ELSE 0.35
				 END) * 0.25 +
				flavour_alignment * 0.20 +
				novelty * 0.15
			) AS overall
		FROM stats`, recipeID).Scan(
		&q.IngredientCoverage,
		&q.NutritionBalance,
		&q.FlavourAlignment,
		&q.Novelty,
		&q.Overall,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO recipe_quality_analysis (
			recipe_id,
			ingredient_coverage_score,
			nutrition_balance_score,
			flavour_alignment_score,
			novelty_score,
			overall_score,
			status,
			computed_at
		) VALUES ($1, $2, $3, $4, $5, $6, 'computed', NOW())
		ON CONFLICT (recipe_id)
		DO UPDATE SET
			ingredient_coverage_score = EXCLUDED.ingredient_coverage_score,
			nutrition_balance_score = EXCLUDED.nutrition_balance_score,
			flavour_alignment_score = EXCLUDED.flavour_alignment_score,
			novelty_score = EXCLUDED.novelty_score,
			overall_score = EXCLUDED.overall_score,
			status = 'computed',
			computed_at = NOW(),
			updated_at = NOW()`,
		recipeID,
		q.IngredientCoverage,
		q.NutritionBalance,
		q.FlavourAlignment,
		q.Novelty,
		q.Overall,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `UPDATE recipes SET quality_score = $2 WHERE id = $1`, recipeID, q.Overall)
	return err
}
