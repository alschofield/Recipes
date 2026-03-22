package search

import (
	"net/http"
	"strconv"
	"strings"

	"recipes/pkg/response"
	"recipes/pkg/storage/postgres"
)

type RecipeCatalogItem struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description,omitempty"`
	Source       string  `json:"source"`
	Difficulty   string  `json:"difficulty"`
	Cuisine      string  `json:"cuisine,omitempty"`
	TotalMinutes int     `json:"totalMinutes"`
	Servings     int     `json:"servings"`
	QualityScore float64 `json:"qualityScore"`
	UpdatedAt    string  `json:"updatedAt"`
}

func GetRecipeCatalog(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	source := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("source")))
	if source == "" {
		source = "all"
	}
	sortBy := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("sort")))
	if sortBy == "" {
		sortBy = "updated_desc"
	}

	page := 1
	if raw := strings.TrimSpace(r.URL.Query().Get("page")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			page = parsed
		}
	}
	pageSize := 24
	if raw := strings.TrimSpace(r.URL.Query().Get("pageSize")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}
	offset := (page - 1) * pageSize

	pool := storage.Pool()
	whereSQL := `
		($1 = '' OR r.name ILIKE '%' || $1 || '%' OR COALESCE(r.cuisine, '') ILIKE '%' || $1 || '%')
		AND ($2 = 'all' OR LOWER(r.source) = $2)`

	var total int
	if err := pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM recipes r WHERE `+whereSQL, q, source).Scan(&total); err != nil {
		response.InternalError(w, "Failed to count recipes")
		return
	}

	orderBy := "r.updated_at DESC"
	switch sortBy {
	case "quality_desc":
		orderBy = "r.quality_score DESC, r.updated_at DESC"
	case "quality_asc":
		orderBy = "r.quality_score ASC, r.updated_at DESC"
	case "name_asc":
		orderBy = "r.name ASC"
	case "name_desc":
		orderBy = "r.name DESC"
	case "updated_asc":
		orderBy = "r.updated_at ASC"
	}

	items := []RecipeCatalogItem{}
	rows, err := pool.Query(r.Context(), `
		SELECT
			r.id,
			r.name,
			COALESCE(r.description, ''),
			r.source,
			r.difficulty,
			COALESCE(r.cuisine, ''),
			COALESCE(r.prep_minutes, 0) + COALESCE(r.cook_minutes, 0) AS total_minutes,
			COALESCE(r.servings, 0),
			r.quality_score::float8,
			r.updated_at::text
		FROM recipes r
		WHERE `+whereSQL+`
		ORDER BY `+orderBy+`
		LIMIT $3 OFFSET $4`, q, source, pageSize, offset)
	if err != nil {
		response.InternalError(w, "Failed to fetch recipes")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item RecipeCatalogItem
		if rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Source,
			&item.Difficulty,
			&item.Cuisine,
			&item.TotalMinutes,
			&item.Servings,
			&item.QualityScore,
			&item.UpdatedAt,
		) == nil {
			items = append(items, item)
		}
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"page":     page,
		"pageSize": pageSize,
		"total":    total,
		"query":    q,
		"source":   source,
		"sort":     sortBy,
		"items":    items,
	})
}
