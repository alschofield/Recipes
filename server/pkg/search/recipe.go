package search

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"recipes/pkg/response"
	"recipes/pkg/storage/postgres"
)

// SearchRequest is the DTO for POST /recipes/search.
type SearchRequest struct {
	Ingredients  []string         `json:"ingredients"`
	Mode         string           `json:"mode"`
	DBOnly       bool             `json:"dbOnly,omitempty"`
	DebugNoCache bool             `json:"debugNoCache,omitempty"`
	Filters      *SearchFilters   `json:"filters,omitempty"`
	Pagination   *PaginationInput `json:"pagination,omitempty"`
}

// SearchFilters narrows search results.
type SearchFilters struct {
	MaxPrepMinutes  *int     `json:"maxPrepMinutes,omitempty"`
	MaxTotalMinutes *int     `json:"maxTotalMinutes,omitempty"`
	Servings        *int     `json:"servings,omitempty"`
	Cuisine         []string `json:"cuisine,omitempty"`
	Dietary         []string `json:"dietary,omitempty"`
	Difficulty      []string `json:"difficulty,omitempty"`
}

// PaginationInput controls result paging.
type PaginationInput struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// Substitution maps a missing ingredient to its aliases.
type Substitution struct {
	Missing     string   `json:"missing"`
	Substitutes []string `json:"substitutes"`
}

// SearchResultItem is a single recipe in the search response.
type SearchResultItem struct {
	ID                 string         `json:"id"`
	Name               string         `json:"name"`
	Source             string         `json:"source"`
	MatchPercent       float64        `json:"matchPercent"`
	MatchedIngredients []string       `json:"matchedIngredients"`
	MissingIngredients []string       `json:"missingIngredients"`
	Substitutions      []Substitution `json:"optionalSubstitutions,omitempty"`
	PrepMinutes        *int           `json:"prepMinutes,omitempty"`
	CookMinutes        *int           `json:"cookMinutes,omitempty"`
	TotalMinutes       int            `json:"totalMinutes"`
	Difficulty         string         `json:"difficulty"`
	Cuisine            *string        `json:"cuisine,omitempty"`
	Servings           int            `json:"servings"`
	DietaryTags        []string       `json:"dietaryTags,omitempty"`
}

// SearchResponse is the full response body.
type SearchResponse struct {
	Mode       string             `json:"mode"`
	Query      SearchQueryInfo    `json:"query"`
	Pagination PaginationInfo     `json:"pagination"`
	Results    []SearchResultItem `json:"results"`
}

// SearchQueryInfo is the normalized query that was executed.
type SearchQueryInfo struct {
	Ingredients []string `json:"ingredients"`
}

// PaginationInfo is paging metadata.
type PaginationInfo struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

// recipeRow holds DB rows before in-memory grouping.
type recipeRow struct {
	RecipeID        string
	RecipeName      string
	Source          string
	Steps           []string
	PrepMinutes     *int
	CookMinutes     *int
	Difficulty      string
	Cuisine         *string
	Servings        int
	QualityScore    float64
	DietaryTags     []string
	UpdatedAt       time.Time
	IngIngredientID string
	IngName         string
	IngOptional     bool
	IngPosition     int
	IngAmount       string
	IngUnit         string
}

// recipeGroup holds all ingredient rows for a recipe.
type recipeGroup struct {
	RecipeID     string
	RecipeName   string
	Source       string
	PrepMinutes  *int
	CookMinutes  *int
	Difficulty   string
	Cuisine      *string
	Servings     int
	QualityScore float64
	DietaryTags  []string
	UpdatedAt    time.Time
	Ingredients  []ingredientEntry
}

type ingredientEntry struct {
	IngredientID string
	Canonical    string
	Optional     bool
	Position     int
	Amount       string
	Unit         string
}

// aliasLookup maps a canonical ingredient name to its aliases.
type aliasLookup struct {
	canonicalToAliases map[string][]string
}

// HandleSearch handles POST /recipes/search.
func HandleSearch(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON body")
		return
	}

	if len(req.Ingredients) == 0 {
		response.BadRequest(w, "At least one ingredient is required")
		return
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "strict"
	}
	if mode != "strict" && mode != "inclusive" {
		response.WriteError(w, http.StatusUnprocessableEntity, "mode must be 'strict' or 'inclusive'", "INVALID_MODE")
		return
	}
	req.Mode = mode

	page := 1
	pageSize := 20
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.PageSize > 0 && req.Pagination.PageSize <= 100 {
			pageSize = req.Pagination.PageSize
		}
	}

	ctx := r.Context()
	pool := storage.Pool()

	// Normalize input ingredients to canonical names.
	normalized, err := normalizeIngredients(ctx, req.Ingredients)
	if err != nil {
		response.InternalError(w, "Failed to normalize ingredients")
		return
	}

	if len(normalized) == 0 {
		resp := SearchResponse{
			Mode:       mode,
			Query:      SearchQueryInfo{Ingredients: normalized},
			Pagination: PaginationInfo{Page: page, PageSize: pageSize, Total: 0},
			Results:    []SearchResultItem{},
		}
		response.WriteJSON(w, http.StatusOK, resp)
		return
	}

	cacheBypass := shouldBypassSearchCache(req.DebugNoCache)
	cacheKey := buildSearchCacheKey(normalized, req, page, pageSize)
	if !cacheBypass {
		if cached, ok := getSearchCache(cacheKey); ok {
			response.WriteJSON(w, http.StatusOK, cached)
			return
		}
	}

	// Query recipes that share at least one ingredient with the input set.
	rows, err := pool.Query(ctx, `
		SELECT
			r.id, r.name, r.source, r.steps, r.prep_minutes, r.cook_minutes,
			r.difficulty, r.cuisine, r.servings, r.quality_score::float8, r.dietary_tags,
			r.updated_at,
			ri.ingredient_id, i.canonical_name, ri.optional, ri.position,
			ri.amount, ri.unit
		FROM recipes r
		JOIN recipe_ingredients ri ON ri.recipe_id = r.id
		JOIN ingredients i ON i.id = ri.ingredient_id
		ORDER BY r.id, ri.position`,
	)
	if err != nil {
		response.InternalError(w, "Search query failed")
		return
	}
	defer rows.Close()

	aliases, err := loadAliases(ctx)
	if err != nil {
		response.InternalError(w, "Failed to load ingredient aliases")
		return
	}

	// Group results by recipe.
	grouped := map[string]*recipeGroup{}
	for rows.Next() {
		var rr recipeRow
		var stepsJSON []byte
		var amount sql.NullString
		var unit sql.NullString

		err := rows.Scan(
			&rr.RecipeID, &rr.RecipeName, &rr.Source, &stepsJSON,
			&rr.PrepMinutes, &rr.CookMinutes, &rr.Difficulty, &rr.Cuisine,
			&rr.Servings, &rr.QualityScore, &rr.DietaryTags, &rr.UpdatedAt,
			&rr.IngIngredientID, &rr.IngName, &rr.IngOptional, &rr.IngPosition,
			&amount, &unit,
		)
		if err != nil {
			response.InternalError(w, "Failed to parse search rows")
			return
		}

		rr.IngAmount = amount.String
		rr.IngUnit = unit.String

		grp, ok := grouped[rr.RecipeID]
		if !ok {
			grp = &recipeGroup{
				RecipeID:     rr.RecipeID,
				RecipeName:   rr.RecipeName,
				Source:       rr.Source,
				PrepMinutes:  rr.PrepMinutes,
				CookMinutes:  rr.CookMinutes,
				Difficulty:   rr.Difficulty,
				Cuisine:      rr.Cuisine,
				Servings:     rr.Servings,
				QualityScore: rr.QualityScore,
				UpdatedAt:    rr.UpdatedAt,
				DietaryTags:  rr.DietaryTags,
			}
			grouped[rr.RecipeID] = grp
		}

		grp.Ingredients = append(grp.Ingredients, ingredientEntry{
			IngredientID: rr.IngIngredientID,
			Canonical:    rr.IngName,
			Optional:     rr.IngOptional,
			Position:     rr.IngPosition,
			Amount:       rr.IngAmount,
			Unit:         rr.IngUnit,
		})
	}

	// Build normalized input set for matching.
	inputSet := make(map[string]struct{}, len(normalized))
	for _, n := range normalized {
		inputSet[n] = struct{}{}
	}

	// Compute match info and build candidate list.
	candidates := make([]SearchResultItem, 0, len(grouped))
	for _, grp := range grouped {
		var matched, missing []string

		for _, ing := range grp.Ingredients {
			if ing.Optional {
				continue
			}

			if _, ok := inputSet[ing.Canonical]; ok {
				matched = append(matched, ing.Canonical)
			} else {
				missing = append(missing, ing.Canonical)
			}
		}

		if !passesMode(mode, len(missing)) {
			continue
		}

		totalRequired := len(matched) + len(missing)
		matchPercent := 0.0
		if totalRequired > 0 {
			matchPercent = math.Round(float64(len(matched))/float64(totalRequired)*100) / 100
		}

		// Build substitutions for missing ingredients.
		var substitutions []Substitution
		for _, m := range missing {
			if subs, ok := aliases.canonicalToAliases[m]; ok && len(subs) > 0 {
				substitutions = append(substitutions, Substitution{
					Missing:     m,
					Substitutes: subs,
				})
			}
		}

		totalMinutes := 0
		if grp.PrepMinutes != nil {
			totalMinutes += *grp.PrepMinutes
		}
		if grp.CookMinutes != nil {
			totalMinutes += *grp.CookMinutes
		}

		if !matchesFilters(grp, req.Filters, totalMinutes) {
			continue
		}

		item := SearchResultItem{
			ID:                 grp.RecipeID,
			Name:               grp.RecipeName,
			Source:             grp.Source,
			MatchPercent:       matchPercent,
			MatchedIngredients: matched,
			MissingIngredients: missing,
			Substitutions:      substitutions,
			PrepMinutes:        grp.PrepMinutes,
			CookMinutes:        grp.CookMinutes,
			TotalMinutes:       totalMinutes,
			Difficulty:         grp.Difficulty,
			Cuisine:            grp.Cuisine,
			Servings:           grp.Servings,
			DietaryTags:        grp.DietaryTags,
		}

		candidates = append(candidates, item)
	}

	// Sort by ranking rules.
	sort.Slice(candidates, func(i, j int) bool {
		return compareSearchResults(candidates[i], candidates[j], grouped)
	})

	results := candidates
	if shouldTriggerLLMFallback(req.DBOnly, candidates) {
		generated, err := generateAndStoreFallbackRecipes(ctx, req, normalized, aliases)
		if err == nil && len(generated) > 0 {
			results = append(results, generated...)
		}
	}

	total := len(results)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paged := results[start:end]

	if paged == nil {
		paged = []SearchResultItem{}
	}

	resp := SearchResponse{
		Mode:  mode,
		Query: SearchQueryInfo{Ingredients: normalized},
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
		Results: paged,
	}

	if !cacheBypass {
		setSearchCache(cacheKey, resp)
	}

	response.WriteJSON(w, http.StatusOK, resp)
}

func compareSearchResults(a, b SearchResultItem, grouped map[string]*recipeGroup) bool {
	// 1. Higher match percent first.
	if a.MatchPercent != b.MatchPercent {
		return a.MatchPercent > b.MatchPercent
	}

	// 2. Fewer missing ingredients first.
	if len(a.MissingIngredients) != len(b.MissingIngredients) {
		return len(a.MissingIngredients) < len(b.MissingIngredients)
	}

	// 3. Higher quality score first.
	qa := qualityOf(grouped[a.ID])
	qb := qualityOf(grouped[b.ID])
	if qa != qb {
		return qa > qb
	}

	// 4. Lower prep time first (known prep beats unknown prep).
	if a.PrepMinutes == nil && b.PrepMinutes != nil {
		return false
	}
	if a.PrepMinutes != nil && b.PrepMinutes == nil {
		return true
	}
	if a.PrepMinutes != nil && b.PrepMinutes != nil && *a.PrepMinutes != *b.PrepMinutes {
		return *a.PrepMinutes < *b.PrepMinutes
	}

	// 5. Newer updated first.
	ua := updatedAtOf(grouped[a.ID])
	ub := updatedAtOf(grouped[b.ID])
	if !ua.Equal(ub) {
		return ua.After(ub)
	}

	// 6. Deterministic final tie-breaker.
	if a.ID != b.ID {
		return a.ID < b.ID
	}

	return a.Name < b.Name
}

func passesMode(mode string, missingCount int) bool {
	if mode == "strict" {
		return missingCount == 0
	}
	return true
}

func qualityOf(grp *recipeGroup) float64 {
	if grp != nil {
		return grp.QualityScore
	}
	return 0.0
}

func updatedAtOf(grp *recipeGroup) time.Time {
	if grp != nil {
		return grp.UpdatedAt
	}
	return time.Time{}
}

func matchesFilters(grp *recipeGroup, filters *SearchFilters, totalMinutes int) bool {
	if filters == nil {
		return true
	}

	if filters.MaxPrepMinutes != nil && grp.PrepMinutes != nil {
		if *grp.PrepMinutes > *filters.MaxPrepMinutes {
			return false
		}
	}

	if filters.MaxTotalMinutes != nil {
		if totalMinutes > *filters.MaxTotalMinutes {
			return false
		}
	}

	if filters.Servings != nil && grp.Servings > 0 {
		if grp.Servings < *filters.Servings {
			return false
		}
	}

	if len(filters.Difficulty) > 0 {
		found := false
		for _, d := range filters.Difficulty {
			if strings.EqualFold(d, grp.Difficulty) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(filters.Cuisine) > 0 && grp.Cuisine != nil {
		found := false
		for _, c := range filters.Cuisine {
			if strings.EqualFold(c, *grp.Cuisine) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(filters.Dietary) > 0 {
		for _, wanted := range filters.Dietary {
			found := false
			for _, tag := range grp.DietaryTags {
				if strings.EqualFold(wanted, tag) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

func normalizeIngredients(ctx context.Context, input []string) ([]string, error) {
	pool := storage.Pool()

	seen := map[string]struct{}{}
	var candidates []string

	for _, raw := range input {
		trimmed := normalizeRawIngredient(raw)
		if trimmed != "" {
			candidates = append(candidates, trimmed)
		}
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	// Map aliases to canonical names.
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT LOWER(ia.alias), i.canonical_name
		FROM ingredient_aliases ia
		JOIN ingredients i ON i.id = ia.ingredient_id
		WHERE LOWER(ia.alias) = ANY($1)`,
		candidates,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	aliasToCanonical := map[string]string{}
	for rows.Next() {
		var alias, canonical string
		if err := rows.Scan(&alias, &canonical); err != nil {
			continue
		}
		aliasToCanonical[alias] = canonical
	}

	var result []string
	for _, candidate := range candidates {
		canonical, ok := aliasToCanonical[candidate]
		if !ok {
			// Use input as-is if no alias found.
			canonical = candidate
		}
		if _, dup := seen[canonical]; dup {
			continue
		}
		seen[canonical] = struct{}{}
		result = append(result, canonical)
	}

	return result, nil
}

func loadAliases(ctx context.Context) (*aliasLookup, error) {
	pool := storage.Pool()

	rows, err := pool.Query(ctx, `
		SELECT i.canonical_name, ia.alias
		FROM ingredient_aliases ia
		JOIN ingredients i ON i.id = ia.ingredient_id
		ORDER BY i.canonical_name, ia.alias`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lookup := &aliasLookup{canonicalToAliases: map[string][]string{}}
	for rows.Next() {
		var canonical, alias string
		if err := rows.Scan(&canonical, &alias); err != nil {
			continue
		}
		lookup.canonicalToAliases[canonical] = append(
			lookup.canonicalToAliases[canonical], alias,
		)
	}

	return lookup, nil
}

func normalizeRawIngredient(raw string) string {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if trimmed == "" {
		return ""
	}

	tokens := strings.Fields(trimmed)
	for i := range tokens {
		tokens[i] = singularizeToken(tokens[i])
	}

	return strings.Join(tokens, " ")
}

func singularizeToken(token string) string {
	if token == "" {
		return token
	}

	irregular := map[string]string{
		"tomatoes": "tomato",
		"potatoes": "potato",
		"leaves":   "leaf",
		"knives":   "knife",
		"loaves":   "loaf",
	}
	if singular, ok := irregular[token]; ok {
		return singular
	}

	if strings.HasSuffix(token, "ies") && len(token) > 4 {
		return token[:len(token)-3] + "y"
	}

	if strings.HasSuffix(token, "ches") || strings.HasSuffix(token, "shes") || strings.HasSuffix(token, "xes") || strings.HasSuffix(token, "zes") {
		return token[:len(token)-2]
	}

	if strings.HasSuffix(token, "es") && (strings.HasSuffix(token, "oes") || strings.HasSuffix(token, "ses")) && len(token) > 3 {
		return token[:len(token)-2]
	}

	if strings.HasSuffix(token, "s") && !strings.HasSuffix(token, "ss") && !strings.HasSuffix(token, "us") && len(token) > 3 {
		return token[:len(token)-1]
	}

	return token
}
