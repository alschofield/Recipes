package ingredients

import (
	"context"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultFuzzyThreshold = 0.92

type MatchResult struct {
	IngredientID  string
	CanonicalName string
	Confidence    float64
	MatchType     string
}

func ResolveIngredient(ctx context.Context, pool *pgxpool.Pool, rawName string) (MatchResult, bool, error) {
	normalized := NormalizeName(rawName)
	if normalized == "" {
		return MatchResult{}, false, nil
	}

	result, ok, err := exactMatch(ctx, pool, normalized)
	if err != nil {
		return MatchResult{}, false, err
	}
	if ok {
		return result, true, nil
	}

	fuzzy, ok, err := fuzzyMatch(ctx, pool, normalized)
	if err != nil {
		return MatchResult{}, false, err
	}
	if ok {
		return fuzzy, true, nil
	}

	return MatchResult{}, false, nil
}

func NormalizeName(raw string) string {
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

func fuzzyThreshold() float64 {
	raw := strings.TrimSpace(os.Getenv("INGREDIENT_FUZZY_THRESHOLD"))
	if raw == "" {
		return defaultFuzzyThreshold
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v <= 0 || v > 1 {
		return defaultFuzzyThreshold
	}
	return v
}

func exactMatch(ctx context.Context, pool *pgxpool.Pool, normalized string) (MatchResult, bool, error) {
	var id string
	var canonical string
	var matchType string
	err := pool.QueryRow(ctx, `
		SELECT i.id, i.canonical_name,
		CASE WHEN LOWER(i.canonical_name) = $1 THEN 'canonical' ELSE 'alias' END as match_type
		FROM ingredients i
		LEFT JOIN ingredient_aliases ia ON ia.ingredient_id = i.id
		WHERE LOWER(i.canonical_name) = $1 OR LOWER(ia.alias) = $1
		ORDER BY match_type ASC
		LIMIT 1`, normalized).Scan(&id, &canonical, &matchType)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return MatchResult{}, false, nil
		}
		return MatchResult{}, false, err
	}

	return MatchResult{IngredientID: id, CanonicalName: canonical, Confidence: 1, MatchType: matchType}, true, nil
}

func fuzzyMatch(ctx context.Context, pool *pgxpool.Pool, normalized string) (MatchResult, bool, error) {
	rows, err := pool.Query(ctx, `
		SELECT i.id, i.canonical_name, LOWER(i.canonical_name) AS candidate_name
		FROM ingredients i
		UNION
		SELECT i.id, i.canonical_name, LOWER(ia.alias) AS candidate_name
		FROM ingredient_aliases ia
		JOIN ingredients i ON i.id = ia.ingredient_id`)
	if err != nil {
		return MatchResult{}, false, err
	}
	defer rows.Close()

	type candidate struct {
		id        string
		canonical string
		name      string
		score     float64
	}

	best := candidate{score: -1}
	second := candidate{score: -1}
	for rows.Next() {
		var c candidate
		if err := rows.Scan(&c.id, &c.canonical, &c.name); err != nil {
			continue
		}
		c.score = similarityScore(normalized, c.name)
		if c.score > best.score {
			second = best
			best = c
		} else if c.score > second.score {
			second = c
		}
	}

	threshold := fuzzyThreshold()
	if best.score < threshold {
		return MatchResult{}, false, nil
	}

	if second.score >= 0 && best.score-second.score < 0.05 {
		return MatchResult{}, false, nil
	}

	return MatchResult{
		IngredientID:  best.id,
		CanonicalName: best.canonical,
		Confidence:    best.score,
		MatchType:     "fuzzy",
	}, true, nil
}

func similarityScore(a, b string) float64 {
	if a == b {
		return 1
	}
	lev := levenshteinDistance(a, b)
	maxLen := max(len(a), len(b))
	if maxLen == 0 {
		return 1
	}
	levScore := 1 - float64(lev)/float64(maxLen)
	if levScore < 0 {
		levScore = 0
	}

	tokensA := tokenSet(a)
	tokensB := tokenSet(b)
	intersections := 0
	for token := range tokensA {
		if _, ok := tokensB[token]; ok {
			intersections++
		}
	}
	tokenScore := float64(intersections) / float64(max(len(tokensA), len(tokensB)))

	return math.Round((0.75*levScore+0.25*tokenScore)*1000) / 1000
}

func tokenSet(s string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, token := range strings.Fields(s) {
		set[token] = struct{}{}
	}
	return set
}

func levenshteinDistance(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	da := make([][]int, len(a)+1)
	for i := range da {
		da[i] = make([]int, len(b)+1)
		da[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		da[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			da[i][j] = min(
				da[i-1][j]+1,
				da[i][j-1]+1,
				da[i-1][j-1]+cost,
			)
		}
	}

	return da[len(a)][len(b)]
}

func singularizeToken(token string) string {
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

func min(values ...int) int {
	sort.Ints(values)
	return values[0]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
