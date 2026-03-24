package ingredients

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"recipes/pkg/middleware"
	"recipes/pkg/response"
	"recipes/pkg/storage/postgres"

	"github.com/jackc/pgx/v5/pgconn"
)

type SuggestIngredientRequest struct {
	Name                 string  `json:"name"`
	SuggestedCanonicalID *string `json:"suggestedCanonicalId,omitempty"`
	Note                 string  `json:"note,omitempty"`
}

type ResolveCandidateRequest struct {
	Action        string  `json:"action"`
	CanonicalID   *string `json:"canonicalId,omitempty"`
	CanonicalName *string `json:"canonicalName,omitempty"`
	Alias         *string `json:"alias,omitempty"`
	Note          string  `json:"note,omitempty"`
}

type VoteCandidateRequest struct {
	Vote int `json:"vote"`
}

func SuggestIngredient(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	var req SuggestIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON body")
		return
	}

	normalized := NormalizeName(req.Name)
	if normalized == "" {
		response.BadRequest(w, "Ingredient name is required")
		return
	}

	pool := storage.Pool()
	match, found, err := ResolveIngredient(r.Context(), pool, normalized)
	if err != nil {
		response.InternalError(w, "Failed to process ingredient")
		return
	}

	if found {
		response.WriteJSON(w, http.StatusOK, map[string]any{
			"status": "matched",
			"match":  match,
		})
		return
	}

	candidateID, status, err := QueueCandidate(r.Context(), CandidateInput{
		RawName:              req.Name,
		NormalizedName:       normalized,
		Source:               "user",
		Confidence:           0,
		SuggestedCanonicalID: req.SuggestedCanonicalID,
		ProposedByUserID:     &principal.UserID,
		ResolutionNote:       req.Note,
	})
	if err != nil {
		response.InternalError(w, "Failed to queue ingredient suggestion")
		return
	}

	response.WriteJSON(w, http.StatusCreated, map[string]any{
		"status":      status,
		"candidateId": candidateID,
		"normalized":  normalized,
	})
}

func ListCandidates(w http.ResponseWriter, r *http.Request) {
	status := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("status")))
	if status == "" {
		status = "pending"
	}

	pool := storage.Pool()
	rows, err := pool.Query(r.Context(), `
		SELECT c.id, c.raw_name, c.normalized_name, c.source, c.status,
		       c.confidence, c.created_at, c.resolution_note,
		       i.canonical_name,
		       COALESCE(SUM(v.vote), 0) AS vote_score
		FROM ingredient_candidates c
		LEFT JOIN ingredients i ON i.id = c.suggested_canonical_id
		LEFT JOIN ingredient_candidate_votes v ON v.candidate_id = c.id
		WHERE c.status = $1
		GROUP BY c.id, i.canonical_name
		ORDER BY vote_score DESC, c.created_at ASC`, status)
	if err != nil {
		response.InternalError(w, "Failed to fetch candidates")
		return
	}
	defer rows.Close()

	type candidateRow struct {
		ID                 string    `json:"id"`
		RawName            string    `json:"rawName"`
		NormalizedName     string    `json:"normalizedName"`
		Source             string    `json:"source"`
		Status             string    `json:"status"`
		Confidence         float64   `json:"confidence"`
		CreatedAt          time.Time `json:"createdAt"`
		ResolutionNote     *string   `json:"resolutionNote,omitempty"`
		SuggestedCanonical *string   `json:"suggestedCanonical,omitempty"`
		VoteScore          int       `json:"voteScore"`
	}

	items := []candidateRow{}
	for rows.Next() {
		var row candidateRow
		if err := rows.Scan(&row.ID, &row.RawName, &row.NormalizedName, &row.Source, &row.Status, &row.Confidence, &row.CreatedAt, &row.ResolutionNote, &row.SuggestedCanonical, &row.VoteScore); err != nil {
			continue
		}
		items = append(items, row)
	}

	response.WriteJSON(w, http.StatusOK, items)
}

func ResolveCandidate(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	candidateID := r.PathValue("candidateid")
	if candidateID == "" {
		response.BadRequest(w, "Candidate ID is required")
		return
	}

	var req ResolveCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON body")
		return
	}

	action := strings.ToLower(strings.TrimSpace(req.Action))
	pool := storage.Pool()

	var normalizedName string
	err := pool.QueryRow(r.Context(), `SELECT normalized_name FROM ingredient_candidates WHERE id = $1`, candidateID).Scan(&normalizedName)
	if err != nil {
		response.NotFound(w, "Candidate not found")
		return
	}

	switch action {
	case "approve_alias":
		canonicalID, err := resolveCanonicalID(r.Context(), req.CanonicalID, req.CanonicalName)
		if err != nil {
			response.BadRequest(w, err.Error())
			return
		}

		_, err = pool.Exec(r.Context(), `INSERT INTO ingredient_aliases (ingredient_id, alias) VALUES ($1, $2) ON CONFLICT (alias) DO NOTHING`, canonicalID, normalizedName)
		if err != nil {
			response.InternalError(w, "Failed to add alias")
			return
		}

		_, err = pool.Exec(r.Context(), `
			UPDATE ingredient_candidates
			SET status='approved_alias', suggested_canonical_id=$1, resolved_by_user_id=$2, resolved_at=NOW(), resolution_note=$3
			WHERE id=$4`, canonicalID, principal.UserID, req.Note, candidateID)
		if err != nil {
			response.InternalError(w, "Failed to update candidate")
			return
		}

	case "approve_canonical":
		canonicalName := normalizedName
		if req.CanonicalName != nil && strings.TrimSpace(*req.CanonicalName) != "" {
			canonicalName = NormalizeName(*req.CanonicalName)
		}

		var canonicalID string
		err := pool.QueryRow(r.Context(), `
			INSERT INTO ingredients (canonical_name)
			VALUES ($1)
			ON CONFLICT (canonical_name) DO UPDATE SET canonical_name = EXCLUDED.canonical_name
			RETURNING id`, canonicalName).Scan(&canonicalID)
		if err != nil {
			response.InternalError(w, "Failed to create canonical ingredient")
			return
		}

		alias := normalizedName
		if req.Alias != nil && strings.TrimSpace(*req.Alias) != "" {
			alias = NormalizeName(*req.Alias)
		}
		if alias != "" {
			_, _ = pool.Exec(r.Context(), `INSERT INTO ingredient_aliases (ingredient_id, alias) VALUES ($1, $2) ON CONFLICT (alias) DO NOTHING`, canonicalID, alias)
		}

		_, err = pool.Exec(r.Context(), `
			UPDATE ingredient_candidates
			SET status='approved_canonical', suggested_canonical_id=$1, resolved_by_user_id=$2, resolved_at=NOW(), resolution_note=$3
			WHERE id=$4`, canonicalID, principal.UserID, req.Note, candidateID)
		if err != nil {
			response.InternalError(w, "Failed to update candidate")
			return
		}

	case "reject":
		_, err := pool.Exec(r.Context(), `
			UPDATE ingredient_candidates
			SET status='rejected', resolved_by_user_id=$1, resolved_at=NOW(), resolution_note=$2
			WHERE id=$3`, principal.UserID, req.Note, candidateID)
		if err != nil {
			response.InternalError(w, "Failed to reject candidate")
			return
		}

	default:
		response.BadRequest(w, "action must be approve_alias, approve_canonical, or reject")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

func VoteCandidate(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	candidateID := r.PathValue("candidateid")
	if candidateID == "" {
		response.BadRequest(w, "Candidate ID is required")
		return
	}

	var req VoteCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON body")
		return
	}
	if req.Vote != 1 && req.Vote != -1 {
		response.BadRequest(w, "vote must be 1 or -1")
		return
	}

	pool := storage.Pool()
	_, err := pool.Exec(r.Context(), `
		INSERT INTO ingredient_candidate_votes (candidate_id, user_id, vote)
		VALUES ($1, $2, $3)
		ON CONFLICT (candidate_id, user_id)
		DO UPDATE SET vote = EXCLUDED.vote, updated_at = NOW()`, candidateID, principal.UserID, req.Vote)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			response.NotFound(w, "Candidate not found")
			return
		}
		response.InternalError(w, "Failed to save vote")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "recorded"})
}

func IngredientMetrics(w http.ResponseWriter, r *http.Request) {
	pool := storage.Pool()

	var canonicalCount, aliasCount, pendingCount, resolvedCount int
	_ = pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM ingredients`).Scan(&canonicalCount)
	_ = pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM ingredient_aliases`).Scan(&aliasCount)
	_ = pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM ingredient_candidates WHERE status='pending'`).Scan(&pendingCount)
	_ = pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM ingredient_candidates WHERE status IN ('approved_alias', 'approved_canonical')`).Scan(&resolvedCount)

	var avgPendingAgeHours float64
	_ = pool.QueryRow(r.Context(), `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (NOW() - created_at)) / 3600), 0)
		FROM ingredient_candidates
		WHERE status = 'pending'`).Scan(&avgPendingAgeHours)

	slaHours := ingredientCandidateSLAHours()
	var pendingOverSLA int
	_ = pool.QueryRow(r.Context(), `
		SELECT COUNT(*)
		FROM ingredient_candidates
		WHERE status = 'pending'
		  AND EXTRACT(EPOCH FROM (NOW() - created_at)) / 3600 > $1`, slaHours).Scan(&pendingOverSLA)

	var oldestPendingAgeHours float64
	_ = pool.QueryRow(r.Context(), `
		SELECT COALESCE(MAX(EXTRACT(EPOCH FROM (NOW() - created_at)) / 3600), 0)
		FROM ingredient_candidates
		WHERE status = 'pending'`).Scan(&oldestPendingAgeHours)

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"canonicalIngredients":  canonicalCount,
		"aliases":               aliasCount,
		"pendingCandidates":     pendingCount,
		"resolvedCandidates":    resolvedCount,
		"avgPendingAgeHours":    avgPendingAgeHours,
		"oldestPendingAgeHours": oldestPendingAgeHours,
		"pendingOverSLA":        pendingOverSLA,
		"slaHours":              slaHours,
	})
}

func ingredientCandidateSLAHours() float64 {
	raw := strings.TrimSpace(os.Getenv("INGREDIENT_CANDIDATE_SLA_HOURS"))
	if raw == "" {
		return 48
	}

	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v <= 0 {
		return 48
	}

	return v
}

func ListIngredientCatalog(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	status := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("status")))
	if status == "" {
		status = "all"
	}
	sortBy := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("sort")))
	if sortBy == "" {
		sortBy = "quality_desc"
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

	minCoverage := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("minCoverage")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 {
			minCoverage = parsed
		}
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

	pool := storage.Pool()

	type enrichmentAvailability struct {
		HasCategory       bool
		HasNaturalSource  bool
		HasMoleculeCount  bool
		HasCoverage       bool
		HasQuality        bool
		HasAnalysisStatus bool
		HasAnalysisNotes  bool
		HasAnalyzedAt     bool
	}
	availability := enrichmentAvailability{}
	_ = pool.QueryRow(r.Context(), `
		SELECT
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'ingredients' AND column_name = 'category'),
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'ingredients' AND column_name = 'natural_source'),
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'ingredients' AND column_name = 'flavour_molecule_count'),
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'ingredients' AND column_name = 'source_coverage'),
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'ingredients' AND column_name = 'quality_score'),
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'ingredients' AND column_name = 'analysis_status'),
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'ingredients' AND column_name = 'analysis_notes'),
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'ingredients' AND column_name = 'last_analyzed_at')`).Scan(
		&availability.HasCategory,
		&availability.HasNaturalSource,
		&availability.HasMoleculeCount,
		&availability.HasCoverage,
		&availability.HasQuality,
		&availability.HasAnalysisStatus,
		&availability.HasAnalysisNotes,
		&availability.HasAnalyzedAt,
	)

	whereClauses := []string{"($1 = '' OR i.canonical_name ILIKE '%' || $1 || '%')"}
	args := []any{q}

	if availability.HasAnalysisStatus {
		whereClauses = append(whereClauses, "($2 = 'all' OR i.analysis_status = $2)")
		args = append(args, status)
	} else {
		args = append(args, "all")
	}

	if availability.HasQuality {
		whereClauses = append(whereClauses, "i.quality_score::float8 >= $3", "i.quality_score::float8 <= $4")
		args = append(args, minQuality, maxQuality)
	} else {
		args = append(args, 0.0, 1.0)
	}

	if availability.HasCoverage {
		whereClauses = append(whereClauses, "i.source_coverage >= $5")
		args = append(args, minCoverage)
	} else {
		args = append(args, 0)
	}

	if availability.HasAnalysisStatus {
		whereClauses = append(whereClauses, "($6 = FALSE OR i.analysis_status IN ('pending', 'review_required'))")
		args = append(args, needsReview)
	} else {
		args = append(args, false)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := "SELECT COUNT(*) FROM ingredients i WHERE " + whereSQL
	var total int
	if err := pool.QueryRow(r.Context(), countQuery, args...).Scan(&total); err != nil {
		response.InternalError(w, "Failed to count ingredients")
		return
	}

	orderBy := "i.canonical_name ASC"
	if availability.HasQuality && availability.HasCoverage {
		orderBy = "i.quality_score DESC, i.source_coverage DESC, i.canonical_name ASC"
	}
	switch sortBy {
	case "quality_asc":
		if availability.HasQuality && availability.HasCoverage {
			orderBy = "i.quality_score ASC, i.source_coverage ASC, i.canonical_name ASC"
		}
	case "name_asc":
		orderBy = "i.canonical_name ASC"
	case "name_desc":
		orderBy = "i.canonical_name DESC"
	case "coverage_desc":
		if availability.HasCoverage {
			orderBy = "i.source_coverage DESC, i.canonical_name ASC"
		}
	case "coverage_asc":
		if availability.HasCoverage {
			orderBy = "i.source_coverage ASC, i.canonical_name ASC"
		}
	}

	categoryExpr := "''"
	naturalSourceExpr := "''"
	moleculeExpr := "0"
	coverageExpr := "0"
	qualityExpr := "0"
	analysisStatusExpr := "'pending'"
	analysisNotesExpr := "''"
	analyzedAtExpr := "NULL"

	if availability.HasCategory {
		categoryExpr = "COALESCE(i.category, '')"
	}
	if availability.HasNaturalSource {
		naturalSourceExpr = "COALESCE(i.natural_source, '')"
	}
	if availability.HasMoleculeCount {
		moleculeExpr = "COALESCE(i.flavour_molecule_count, 0)"
	}
	if availability.HasCoverage {
		coverageExpr = "COALESCE(i.source_coverage, 0)"
	}
	if availability.HasQuality {
		qualityExpr = "COALESCE(i.quality_score::float8, 0)"
	}
	if availability.HasAnalysisStatus {
		analysisStatusExpr = "COALESCE(i.analysis_status, 'pending')"
	}
	if availability.HasAnalysisNotes {
		analysisNotesExpr = "COALESCE(i.analysis_notes, '')"
	}
	if availability.HasAnalyzedAt {
		analyzedAtExpr = "i.last_analyzed_at"
	}

	listQuery := `
		SELECT
			i.id,
			i.canonical_name,
			` + categoryExpr + `,
			` + naturalSourceExpr + `,
			` + moleculeExpr + `,
			` + coverageExpr + `,
			` + qualityExpr + `,
			` + analysisStatusExpr + `,
			` + analysisNotesExpr + `,
			` + analyzedAtExpr + `,
			COALESCE(alias_counts.alias_count, 0)
		FROM ingredients i
		LEFT JOIN (
			SELECT ingredient_id, COUNT(*) AS alias_count
			FROM ingredient_aliases
			GROUP BY ingredient_id
		) alias_counts ON alias_counts.ingredient_id = i.id
		WHERE ` + whereSQL + `
		ORDER BY ` + orderBy + `
		LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)

	args = append(args, pageSize, offset)
	rows, err := pool.Query(r.Context(), listQuery, args...)
	if err != nil {
		response.InternalError(w, "Failed to fetch ingredient catalog")
		return
	}
	defer rows.Close()

	type ingredientCatalogItem struct {
		ID                   string     `json:"id"`
		CanonicalName        string     `json:"canonicalName"`
		Category             string     `json:"category,omitempty"`
		NaturalSource        string     `json:"naturalSource,omitempty"`
		FlavourMoleculeCount int        `json:"flavourMoleculeCount"`
		SourceCoverage       int        `json:"sourceCoverage"`
		QualityScore         float64    `json:"qualityScore"`
		AnalysisStatus       string     `json:"analysisStatus"`
		AnalysisNotes        string     `json:"analysisNotes,omitempty"`
		LastAnalyzedAt       *time.Time `json:"lastAnalyzedAt,omitempty"`
		AliasCount           int        `json:"aliasCount"`
	}

	items := []ingredientCatalogItem{}
	for rows.Next() {
		var item ingredientCatalogItem
		if err := rows.Scan(
			&item.ID,
			&item.CanonicalName,
			&item.Category,
			&item.NaturalSource,
			&item.FlavourMoleculeCount,
			&item.SourceCoverage,
			&item.QualityScore,
			&item.AnalysisStatus,
			&item.AnalysisNotes,
			&item.LastAnalyzedAt,
			&item.AliasCount,
		); err != nil {
			continue
		}
		items = append(items, item)
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"page":        page,
		"pageSize":    pageSize,
		"total":       total,
		"status":      status,
		"query":       q,
		"sort":        sortBy,
		"minQuality":  minQuality,
		"maxQuality":  maxQuality,
		"minCoverage": minCoverage,
		"needsReview": needsReview,
		"items":       items,
	})
}

func GetIngredientDetail(w http.ResponseWriter, r *http.Request) {
	ingredientID := strings.TrimSpace(r.PathValue("ingredientid"))
	if ingredientID == "" {
		response.BadRequest(w, "Ingredient ID is required")
		return
	}

	pool := storage.Pool()

	hasCategory := false
	hasNaturalSource := false
	hasMoleculeCount := false
	hasCoverage := false
	hasQuality := false
	hasAnalysisStatus := false
	hasAnalysisNotes := false
	_ = pool.QueryRow(r.Context(), `
		SELECT
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'ingredients' AND column_name = 'category'),
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'ingredients' AND column_name = 'natural_source'),
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'ingredients' AND column_name = 'flavour_molecule_count'),
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'ingredients' AND column_name = 'source_coverage'),
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'ingredients' AND column_name = 'quality_score'),
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'ingredients' AND column_name = 'analysis_status'),
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'ingredients' AND column_name = 'analysis_notes')`).Scan(
		&hasCategory,
		&hasNaturalSource,
		&hasMoleculeCount,
		&hasCoverage,
		&hasQuality,
		&hasAnalysisStatus,
		&hasAnalysisNotes,
	)

	categoryExpr := "''"
	naturalSourceExpr := "''"
	moleculeExpr := "0"
	coverageExpr := "0"
	qualityExpr := "0"
	analysisStatusExpr := "'pending'"
	analysisNotesExpr := "''"
	if hasCategory {
		categoryExpr = "COALESCE(i.category, '')"
	}
	if hasNaturalSource {
		naturalSourceExpr = "COALESCE(i.natural_source, '')"
	}
	if hasMoleculeCount {
		moleculeExpr = "COALESCE(i.flavour_molecule_count, 0)"
	}
	if hasCoverage {
		coverageExpr = "COALESCE(i.source_coverage, 0)"
	}
	if hasQuality {
		qualityExpr = "COALESCE(i.quality_score::float8, 0)"
	}
	if hasAnalysisStatus {
		analysisStatusExpr = "COALESCE(i.analysis_status, 'pending')"
	}
	if hasAnalysisNotes {
		analysisNotesExpr = "COALESCE(i.analysis_notes, '')"
	}

	type ingredientDetail struct {
		ID                   string   `json:"id"`
		CanonicalName        string   `json:"canonicalName"`
		Category             string   `json:"category,omitempty"`
		NaturalSource        string   `json:"naturalSource,omitempty"`
		FlavourMoleculeCount int      `json:"flavourMoleculeCount"`
		SourceCoverage       int      `json:"sourceCoverage"`
		QualityScore         float64  `json:"qualityScore"`
		AnalysisStatus       string   `json:"analysisStatus"`
		AnalysisNotes        string   `json:"analysisNotes,omitempty"`
		Aliases              []string `json:"aliases"`
		RecipeCount          int      `json:"recipeCount"`
	}

	var out ingredientDetail
	err := pool.QueryRow(r.Context(), `
		SELECT
			i.id,
			i.canonical_name,
			`+categoryExpr+`,
			`+naturalSourceExpr+`,
			`+moleculeExpr+`,
			`+coverageExpr+`,
			`+qualityExpr+`,
			`+analysisStatusExpr+`,
			`+analysisNotesExpr+`,
			COALESCE(alias_data.aliases, ARRAY[]::text[]),
			COALESCE(recipe_data.recipe_count, 0)
		FROM ingredients i
		LEFT JOIN (
			SELECT ingredient_id, ARRAY_AGG(alias ORDER BY alias) AS aliases
			FROM ingredient_aliases
			GROUP BY ingredient_id
		) alias_data ON alias_data.ingredient_id = i.id
		LEFT JOIN (
			SELECT ingredient_id, COUNT(DISTINCT recipe_id) AS recipe_count
			FROM recipe_ingredients
			GROUP BY ingredient_id
		) recipe_data ON recipe_data.ingredient_id = i.id
		WHERE i.id = $1`, ingredientID,
	).Scan(
		&out.ID,
		&out.CanonicalName,
		&out.Category,
		&out.NaturalSource,
		&out.FlavourMoleculeCount,
		&out.SourceCoverage,
		&out.QualityScore,
		&out.AnalysisStatus,
		&out.AnalysisNotes,
		&out.Aliases,
		&out.RecipeCount,
	)
	if err != nil {
		response.NotFound(w, "Ingredient not found")
		return
	}

	response.WriteJSON(w, http.StatusOK, out)
}

type CandidateInput struct {
	RawName              string
	NormalizedName       string
	Source               string
	Status               string
	Confidence           float64
	SuggestedCanonicalID *string
	ProposedByUserID     *string
	ResolutionNote       string
}

func QueueCandidate(ctx context.Context, input CandidateInput) (candidateID, status string, err error) {
	pool := storage.Pool()
	status = strings.ToLower(strings.TrimSpace(input.Status))
	if status == "" {
		status = "pending"
	}
	if status != "pending" && status != "approved_alias" && status != "approved_canonical" && status != "rejected" {
		status = "pending"
	}

	resolvedAt := any(nil)
	if status != "pending" {
		resolvedAt = time.Now().UTC()
	}

	err = pool.QueryRow(ctx, `
		INSERT INTO ingredient_candidates (
			raw_name, normalized_name, source, status, confidence,
			suggested_canonical_id, proposed_by_user_id, resolution_note, resolved_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (normalized_name) WHERE (status = 'pending')
		DO UPDATE SET
			resolution_note = EXCLUDED.resolution_note,
			suggested_canonical_id = COALESCE(ingredient_candidates.suggested_canonical_id, EXCLUDED.suggested_canonical_id)
		RETURNING id, status`,
		input.RawName,
		input.NormalizedName,
		input.Source,
		status,
		input.Confidence,
		input.SuggestedCanonicalID,
		input.ProposedByUserID,
		input.ResolutionNote,
		resolvedAt,
	).Scan(&candidateID, &status)
	return
}

func resolveCanonicalID(ctx context.Context, canonicalID *string, canonicalName *string) (string, error) {
	pool := storage.Pool()
	if canonicalID != nil && strings.TrimSpace(*canonicalID) != "" {
		return strings.TrimSpace(*canonicalID), nil
	}
	if canonicalName == nil || strings.TrimSpace(*canonicalName) == "" {
		return "", errString("canonicalId or canonicalName is required")
	}

	normalized := NormalizeName(*canonicalName)
	if normalized == "" {
		return "", errString("canonicalName is invalid")
	}

	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO ingredients (canonical_name)
		VALUES ($1)
		ON CONFLICT (canonical_name) DO UPDATE SET canonical_name = EXCLUDED.canonical_name
		RETURNING id`, normalized).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

type errString string

func (e errString) Error() string { return string(e) }
