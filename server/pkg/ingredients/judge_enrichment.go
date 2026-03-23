package ingredients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"recipes/pkg/llmjudge"

	"github.com/jackc/pgx/v5/pgxpool"
)

func triggerIngredientJudgeEnrichment(pool *pgxpool.Pool, ingredientID, canonicalName string) {
	if !llmjudge.Enabled() {
		return
	}
	if strings.TrimSpace(ingredientID) == "" || strings.TrimSpace(canonicalName) == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		result, err := llmjudge.JudgeIngredientMetadata(ctx, canonicalName)
		if err != nil {
			log.Printf("ingredient_judge event=failed ingredient_id=%s name=%q err=%q", ingredientID, canonicalName, err.Error())
			_, _ = pool.Exec(ctx, `
				UPDATE ingredients
				SET analysis_status='review_required',
				    analysis_notes=$2,
				    last_analyzed_at=NOW()
				WHERE id=$1`, ingredientID, fmt.Sprintf("judge failed: %s", err.Error()))
			return
		}

		if result.Confidence < llmjudge.MinConfidence() {
			_, _ = pool.Exec(ctx, `
				UPDATE ingredients
				SET analysis_status='review_required',
				    analysis_notes=$2,
				    last_analyzed_at=NOW()
				WHERE id=$1`, ingredientID, fmt.Sprintf("judge confidence low: %.3f (%s)", result.Confidence, result.Evidence))
			return
		}

		metadataJSON, _ := json.Marshal(map[string]any{
			"judge": map[string]any{
				"allergenHints": result.AllergenHints,
				"riskHints":     result.RiskHints,
				"confidence":    result.Confidence,
				"evidence":      result.Evidence,
				"updatedAt":     time.Now().UTC().Format(time.RFC3339),
			},
		})

		coverage := int(result.Confidence * 100)
		if coverage < 0 {
			coverage = 0
		}
		if coverage > 100 {
			coverage = 100
		}

		quality := result.Confidence
		if quality < 0 {
			quality = 0
		}
		if quality > 1 {
			quality = 1
		}

		_, err = pool.Exec(ctx, `
			UPDATE ingredients
			SET category = COALESCE(NULLIF($2, ''), category),
			    source_coverage = GREATEST(source_coverage, $3),
			    quality_score = GREATEST(quality_score, $4),
			    analysis_status='enriched',
			    analysis_notes=$5,
			    last_analyzed_at=NOW(),
			    metadata = COALESCE(metadata, '{}'::jsonb) || $6::jsonb
			WHERE id=$1`,
			ingredientID,
			strings.ToLower(strings.TrimSpace(result.Category)),
			coverage,
			quality,
			result.Evidence,
			string(metadataJSON),
		)
		if err != nil {
			log.Printf("ingredient_judge event=update_failed ingredient_id=%s err=%q", ingredientID, err.Error())
			return
		}

		for _, alias := range result.AliasSuggestions {
			alias = NormalizeName(alias)
			if alias == "" || alias == canonicalName {
				continue
			}
			_, _ = pool.Exec(ctx, `
				INSERT INTO ingredient_aliases (ingredient_id, alias)
				VALUES ($1, $2)
				ON CONFLICT (alias) DO NOTHING`, ingredientID, alias)
		}
	}()
}
