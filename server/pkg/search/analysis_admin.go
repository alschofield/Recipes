package search

import (
	"net/http"
	"strconv"
	"strings"

	"recipes/pkg/response"
	"recipes/pkg/storage/postgres"
)

type AnalysisOverview struct {
	TotalRecipes              int     `json:"totalRecipes"`
	ComputedAnalyses          int     `json:"computedAnalyses"`
	PendingOrMissing          int     `json:"pendingOrMissing"`
	FailedAnalyses            int     `json:"failedAnalyses"`
	AverageOverallScore       float64 `json:"averageOverallScore"`
	AverageQualityScore       float64 `json:"averageQualityScore"`
	AverageIngredientCoverage float64 `json:"averageIngredientCoverage"`
}

type ScoreRow struct {
	RecipeID     string  `json:"recipeId"`
	Name         string  `json:"name"`
	Source       string  `json:"source"`
	Cuisine      string  `json:"cuisine"`
	OverallScore float64 `json:"overallScore"`
	QualityScore float64 `json:"qualityScore"`
	UpdatedAt    string  `json:"updatedAt"`
}

type GroupAverageRow struct {
	Label          string  `json:"label"`
	RecipeCount    int     `json:"recipeCount"`
	AverageOverall float64 `json:"averageOverallScore"`
	AverageQuality float64 `json:"averageQualityScore"`
}

type StaleRow struct {
	RecipeID        string `json:"recipeId"`
	Name            string `json:"name"`
	Source          string `json:"source"`
	Cuisine         string `json:"cuisine"`
	AnalysisStatus  string `json:"analysisStatus"`
	RecipeUpdatedAt string `json:"recipeUpdatedAt"`
	ComputedAt      string `json:"computedAt,omitempty"`
}

type RecipeCatalogRow struct {
	RecipeID        string  `json:"recipeId"`
	Name            string  `json:"name"`
	Source          string  `json:"source"`
	Cuisine         string  `json:"cuisine"`
	Difficulty      string  `json:"difficulty"`
	OverallScore    float64 `json:"overallScore"`
	QualityScore    float64 `json:"qualityScore"`
	AnalysisStatus  string  `json:"analysisStatus"`
	NeedsReview     bool    `json:"needsReview"`
	IngredientCount int     `json:"ingredientCount"`
	UpdatedAt       string  `json:"updatedAt"`
	ComputedAt      string  `json:"computedAt,omitempty"`
}

func GetAnalysisAdminDashboard(w http.ResponseWriter, r *http.Request) {
	pool := storage.Pool()

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	status := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("status")))
	if status == "" {
		status = "all"
	}
	source := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("source")))
	if source == "" {
		source = "all"
	}
	sortBy := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("sort")))
	if sortBy == "" {
		sortBy = "updated_desc"
	}

	minQuality := 0.0
	if raw := strings.TrimSpace(r.URL.Query().Get("minQuality")); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil && parsed >= 0 {
			minQuality = parsed
		}
	}
	maxQuality := 1.0
	if raw := strings.TrimSpace(r.URL.Query().Get("maxQuality")); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil && parsed >= 0 {
			maxQuality = parsed
		}
	}
	if maxQuality < minQuality {
		maxQuality = minQuality
	}

	needsReview := false
	if raw := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("needsReview"))); raw == "1" || raw == "true" || raw == "yes" {
		needsReview = true
	}

	page := 1
	if raw := strings.TrimSpace(r.URL.Query().Get("page")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			page = parsed
		}
	}
	pageSize := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("pageSize")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 200 {
			pageSize = parsed
		}
	}
	offset := (page - 1) * pageSize

	hasQATable := false
	_ = pool.QueryRow(r.Context(), `
		SELECT EXISTS(
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'recipe_quality_analysis'
		)`).Scan(&hasQATable)

	hasQAStatus := false
	hasQAOverall := false
	hasQACoverage := false
	hasQAComputed := false
	if hasQATable {
		_ = pool.QueryRow(r.Context(), `
			SELECT
				EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'recipe_quality_analysis' AND column_name = 'status'),
				EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'recipe_quality_analysis' AND column_name = 'overall_score'),
				EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'recipe_quality_analysis' AND column_name = 'ingredient_coverage_score'),
				EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'recipe_quality_analysis' AND column_name = 'computed_at')
		`).Scan(&hasQAStatus, &hasQAOverall, &hasQACoverage, &hasQAComputed)
	}

	joinQA := ""
	statusExpr := "'missing'"
	overallExpr := "COALESCE(r.quality_score::float8, 0)"
	coverageExpr := "0::float8"
	computedAtExpr := "''"
	needsReviewExpr := "FALSE"

	if hasQATable {
		joinQA = "LEFT JOIN recipe_quality_analysis qa ON qa.recipe_id = r.id"
		if hasQAStatus {
			statusExpr = "COALESCE(qa.status, 'missing')"
		}
		if hasQAOverall {
			overallExpr = "COALESCE(qa.overall_score::float8, r.quality_score::float8, 0)"
		}
		if hasQACoverage {
			coverageExpr = "COALESCE(qa.ingredient_coverage_score::float8, 0)"
		}
		if hasQAComputed {
			computedAtExpr = "COALESCE(qa.computed_at::text, '')"
			if hasQAStatus {
				needsReviewExpr = "(qa.recipe_id IS NULL OR COALESCE(qa.status, 'missing') <> 'computed' OR qa.computed_at < r.updated_at)"
			} else {
				needsReviewExpr = "(qa.recipe_id IS NULL OR qa.computed_at < r.updated_at)"
			}
		} else if hasQAStatus {
			needsReviewExpr = "(qa.recipe_id IS NULL OR COALESCE(qa.status, 'missing') <> 'computed')"
		} else {
			needsReviewExpr = "(qa.recipe_id IS NULL)"
		}
	}

	var overview AnalysisOverview
	overviewQuery := `
		SELECT
			COUNT(*)::int,
			COUNT(*) FILTER (WHERE ` + statusExpr + ` = 'computed')::int,
			COUNT(*) FILTER (WHERE ` + statusExpr + ` IN ('pending', 'missing'))::int,
			COUNT(*) FILTER (WHERE ` + statusExpr + ` = 'failed')::int,
			COALESCE(AVG(` + overallExpr + `), 0)::float8,
			COALESCE(AVG(COALESCE(r.quality_score::float8, 0)), 0)::float8,
			COALESCE(AVG(` + coverageExpr + `), 0)::float8
		FROM recipes r
		` + joinQA
	if err := pool.QueryRow(r.Context(), overviewQuery).Scan(
		&overview.TotalRecipes,
		&overview.ComputedAnalyses,
		&overview.PendingOrMissing,
		&overview.FailedAnalyses,
		&overview.AverageOverallScore,
		&overview.AverageQualityScore,
		&overview.AverageIngredientCoverage,
	); err != nil {
		response.InternalError(w, "Failed to compute analysis overview")
		return
	}

	whereSQL := `
		($1 = '' OR r.name ILIKE '%' || $1 || '%' OR COALESCE(r.cuisine, '') ILIKE '%' || $1 || '%')
		AND ($2 = 'all' OR ` + statusExpr + ` = $2)
		AND ($3 = 'all' OR LOWER(r.source) = $3)
		AND ` + overallExpr + ` >= $4
		AND ` + overallExpr + ` <= $5
		AND ($6 = FALSE OR ` + needsReviewExpr + `)`
	args := []any{query, status, source, minQuality, maxQuality, needsReview}

	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM recipes r
		` + joinQA + `
		WHERE ` + whereSQL
	if err := pool.QueryRow(r.Context(), countQuery, args...).Scan(&total); err != nil {
		response.InternalError(w, "Failed to count recipes")
		return
	}

	orderBy := "r.updated_at DESC"
	switch sortBy {
	case "quality_desc":
		orderBy = overallExpr + " DESC, r.updated_at DESC"
	case "quality_asc":
		orderBy = overallExpr + " ASC, r.updated_at DESC"
	case "name_asc":
		orderBy = "r.name ASC"
	case "name_desc":
		orderBy = "r.name DESC"
	case "updated_asc":
		orderBy = "r.updated_at ASC"
	}

	recipeRows := []RecipeCatalogRow{}
	listQuery := `
		SELECT
			r.id,
			r.name,
			r.source,
			COALESCE(r.cuisine, ''),
			r.difficulty,
			` + overallExpr + `,
			r.quality_score::float8,
			` + statusExpr + `,
			` + needsReviewExpr + ` AS needs_review,
			COALESCE(ic.ingredient_count, 0),
			r.updated_at::text,
			` + computedAtExpr + `
		FROM recipes r
		` + joinQA + `
		LEFT JOIN (
			SELECT recipe_id, COUNT(*) AS ingredient_count
			FROM recipe_ingredients
			GROUP BY recipe_id
		) ic ON ic.recipe_id = r.id
		WHERE ` + whereSQL + `
		ORDER BY ` + orderBy + `
		LIMIT $7 OFFSET $8`
	if rows, err := pool.Query(r.Context(), listQuery, append(args, pageSize, offset)...); err == nil {
		defer rows.Close()
		for rows.Next() {
			var row RecipeCatalogRow
			if rows.Scan(
				&row.RecipeID,
				&row.Name,
				&row.Source,
				&row.Cuisine,
				&row.Difficulty,
				&row.OverallScore,
				&row.QualityScore,
				&row.AnalysisStatus,
				&row.NeedsReview,
				&row.IngredientCount,
				&row.UpdatedAt,
				&row.ComputedAt,
			) == nil {
				recipeRows = append(recipeRows, row)
			}
		}
	}

	topRecipes := []ScoreRow{}
	if rows, err := pool.Query(r.Context(), `
		SELECT
			r.id,
			r.name,
			r.source,
			COALESCE(r.cuisine, ''),
			`+overallExpr+`,
			r.quality_score::float8,
			r.updated_at::text
		FROM recipes r
		`+joinQA+`
		ORDER BY `+overallExpr+` DESC, r.updated_at DESC
		LIMIT 10`); err == nil {
		defer rows.Close()
		for rows.Next() {
			var row ScoreRow
			if rows.Scan(&row.RecipeID, &row.Name, &row.Source, &row.Cuisine, &row.OverallScore, &row.QualityScore, &row.UpdatedAt) == nil {
				topRecipes = append(topRecipes, row)
			}
		}
	}

	lowRecipes := []ScoreRow{}
	if rows, err := pool.Query(r.Context(), `
		SELECT
			r.id,
			r.name,
			r.source,
			COALESCE(r.cuisine, ''),
			`+overallExpr+`,
			r.quality_score::float8,
			r.updated_at::text
		FROM recipes r
		`+joinQA+`
		ORDER BY `+overallExpr+` ASC, r.updated_at DESC
		LIMIT 10`); err == nil {
		defer rows.Close()
		for rows.Next() {
			var row ScoreRow
			if rows.Scan(&row.RecipeID, &row.Name, &row.Source, &row.Cuisine, &row.OverallScore, &row.QualityScore, &row.UpdatedAt) == nil {
				lowRecipes = append(lowRecipes, row)
			}
		}
	}

	byCuisine := []GroupAverageRow{}
	if rows, err := pool.Query(r.Context(), `
		SELECT
			COALESCE(NULLIF(TRIM(r.cuisine), ''), 'unknown') AS label,
			COUNT(*)::int,
			COALESCE(AVG(`+overallExpr+`), 0)::float8,
			COALESCE(AVG(r.quality_score::float8), 0)::float8
		FROM recipes r
		`+joinQA+`
		GROUP BY label
		ORDER BY COUNT(*) DESC, label ASC
		LIMIT 12`); err == nil {
		defer rows.Close()
		for rows.Next() {
			var row GroupAverageRow
			if rows.Scan(&row.Label, &row.RecipeCount, &row.AverageOverall, &row.AverageQuality) == nil {
				byCuisine = append(byCuisine, row)
			}
		}
	}

	bySource := []GroupAverageRow{}
	if rows, err := pool.Query(r.Context(), `
		SELECT
			COALESCE(NULLIF(TRIM(r.source), ''), 'unknown') AS label,
			COUNT(*)::int,
			COALESCE(AVG(`+overallExpr+`), 0)::float8,
			COALESCE(AVG(r.quality_score::float8), 0)::float8
		FROM recipes r
		`+joinQA+`
		GROUP BY label
		ORDER BY COUNT(*) DESC, label ASC
		LIMIT 12`); err == nil {
		defer rows.Close()
		for rows.Next() {
			var row GroupAverageRow
			if rows.Scan(&row.Label, &row.RecipeCount, &row.AverageOverall, &row.AverageQuality) == nil {
				bySource = append(bySource, row)
			}
		}
	}

	staleQueue := []StaleRow{}
	if rows, err := pool.Query(r.Context(), `
		SELECT
			r.id,
			r.name,
			r.source,
			COALESCE(r.cuisine, ''),
			`+statusExpr+`,
			r.updated_at::text,
			`+computedAtExpr+`
		FROM recipes r
		`+joinQA+`
		WHERE `+needsReviewExpr+`
		ORDER BY r.updated_at DESC
		LIMIT 25`); err == nil {
		defer rows.Close()
		for rows.Next() {
			var row StaleRow
			if rows.Scan(&row.RecipeID, &row.Name, &row.Source, &row.Cuisine, &row.AnalysisStatus, &row.RecipeUpdatedAt, &row.ComputedAt) == nil {
				staleQueue = append(staleQueue, row)
			}
		}
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"overview":     overview,
		"topRecipes":   topRecipes,
		"lowRecipes":   lowRecipes,
		"byCuisine":    byCuisine,
		"bySource":     bySource,
		"staleQueue":   staleQueue,
		"recipes":      recipeRows,
		"totalRecipes": total,
		"page":         page,
		"pageSize":     pageSize,
		"query":        query,
		"status":       status,
		"source":       source,
		"sort":         sortBy,
		"minQuality":   minQuality,
		"maxQuality":   maxQuality,
		"needsReview":  needsReview,
	})
}
