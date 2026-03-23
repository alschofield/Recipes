# Data Patterns from `datasets/raw/server-lib`

This file summarizes practical priors derived from local project datasets for judge-model calibration.

Generated artifact:

- `llm/judge/data-priors.summary.json`
- Script: `llm/judge/scripts/analyze_server_lib_priors.py`

## Key Findings

### Ingredient seed quality patterns

- Source: `datasets/derived/server-lib/canonical_ingredient_seed_v1.csv`
- Rows: `1769`
- Analysis status counts:
  - `enriched`: `1010`
  - `pending`: `759`

Quality trends by `source_coverage` bucket:

- Coverage `2`: mean quality around `0.288`
- Coverage `3`: mean quality around `0.355`
- Coverage `4`: mean quality around `0.506`
- Coverage `5`: mean quality around `0.785`

Interpretation: multi-source coverage is a strong prior for confidence and quality.

### Recipe ingredient-count patterns

- Source: `datasets/raw/server-lib/kaggle ingredients dataset.zip` (`train.json`)
- Recipes: `39,774`
- Ingredients per recipe:
  - p50: `10`
  - p90: `17`
  - max: `65`

Interpretation: the current complex-mode trigger at `>=10` ingredients aligns with median real-world recipe complexity.

## Recommended Judge Calibration Rules

1. Use stricter auto-approve thresholds for low-coverage ingredients.
2. Route low-confidence low-coverage classifications to `review_required`.
3. Use ingredient-count priors when scoring recipe complexity and coherence.
4. Keep deterministic scoring path as fallback regardless of judge availability.
