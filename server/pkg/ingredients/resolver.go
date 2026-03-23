package ingredients

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ResolvedIngredient struct {
	IngredientID  string
	CanonicalName string
	Confidence    float64
	MatchType     string
	Created       bool
}

// ResolveOrCreateForLLM maps an ingredient to canonical data.
// If no safe match exists, it creates a canonical ingredient and logs a governance candidate row.
func ResolveOrCreateForLLM(ctx context.Context, pool *pgxpool.Pool, rawName string) (ResolvedIngredient, error) {
	normalized := NormalizeName(rawName)
	if normalized == "" {
		return ResolvedIngredient{}, nil
	}

	match, found, err := ResolveIngredient(ctx, pool, normalized)
	if err != nil {
		return ResolvedIngredient{}, err
	}
	if found {
		return ResolvedIngredient{
			IngredientID:  match.IngredientID,
			CanonicalName: match.CanonicalName,
			Confidence:    match.Confidence,
			MatchType:     match.MatchType,
			Created:       false,
		}, nil
	}

	var ingredientID string
	err = pool.QueryRow(ctx, `
		INSERT INTO ingredients (canonical_name)
		VALUES ($1)
		ON CONFLICT (canonical_name) DO UPDATE SET canonical_name = EXCLUDED.canonical_name
		RETURNING id`, normalized).Scan(&ingredientID)
	if err != nil {
		return ResolvedIngredient{}, err
	}

	_, _ = pool.Exec(ctx, `
		INSERT INTO ingredient_aliases (ingredient_id, alias)
		VALUES ($1, $2)
		ON CONFLICT (alias) DO NOTHING`, ingredientID, normalized)

	_, _, _ = QueueCandidate(ctx, CandidateInput{
		RawName:              strings.TrimSpace(rawName),
		NormalizedName:       normalized,
		Source:               "llm",
		Status:               "approved_canonical",
		Confidence:           0.5,
		SuggestedCanonicalID: &ingredientID,
		ResolutionNote:       "Auto-created canonical ingredient from LLM recipe generation",
	})

	triggerIngredientJudgeEnrichment(pool, ingredientID, normalized)

	return ResolvedIngredient{
		IngredientID:  ingredientID,
		CanonicalName: normalized,
		Confidence:    0.5,
		MatchType:     "created",
		Created:       true,
	}, nil
}
