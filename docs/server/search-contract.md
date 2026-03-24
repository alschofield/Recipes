# Search Contract (`strict` vs `inclusive`)

Status: v1 locked baseline (update via checklist preflight item)

## Endpoint

- `POST /recipes/search`

## Request Body

```json
{
  "ingredients": ["chicken", "rice", "garlic"],
  "mode": "strict",
  "complex": false,
  "dbOnly": false,
  "filters": {
    "maxPrepMinutes": 45,
    "servings": 2,
    "cuisine": ["asian"],
    "dietary": ["high-protein"],
    "difficulty": ["easy"]
  },
  "pagination": {
    "page": 1,
    "pageSize": 20
  }
}
```

## Mode Definitions

- `strict`: return recipes where **all required recipe ingredients** are present in the user input (after ingredient normalization).
- `inclusive`: return recipes where input ingredients are a subset match (recipe may require extras); include missing ingredient list.
- `complex=true`: optional client hint that routes LLM fallback to the complex prompt profile when fallback runs.
- If 10 or more normalized ingredients are provided, backend treats the request as complex even when `complex` is omitted.
- `dbOnly=true`: skip LLM fallback and return only database-backed results.
- Strict-mode generated policy is controlled by `LLM_STRICT_GENERATED_POLICY`:
  - `none` (default): generated recipes with missing required ingredients are excluded.
  - `degrade_inclusive`: strict requests may include generated results that miss required ingredients.

## Ingredient Normalization Rules

- Lowercase + trim whitespace.
- Map aliases to canonical names (e.g., `scallion` -> `green onion`).
- Singularize where feasible (`tomatoes` -> `tomato`).
- Ignore duplicate user inputs.

## Ranking Rules (in order)

1. Higher `matchPercent` first.
2. Fewer `missingIngredients` first.
3. Higher `recipeQualityScore` first.
4. Lower `prepMinutes` first.
5. Newer `updatedAt` first.

## Blend Policy (DB + LLM)

- Database results are ranked first using the rules above.
- Fallback-generated results are blended with deterministic interleave.
- Minimum generated insertions are controlled by `SEARCH_BLEND_MIN_GENERATED` (default `1`).
- Maximum generated share is controlled by `SEARCH_BLEND_MAX_GENERATED_SHARE` (default `0.40`).
- `SEARCH_BLEND_SEED` provides deterministic seed salt for stable ordering across pagination.

## Response Body

```json
{
  "mode": "inclusive",
  "query": {
    "ingredients": ["chicken", "rice", "garlic"]
  },
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "total": 142
  },
  "results": [
    {
      "id": "recipe_123",
      "name": "Garlic Chicken Rice Bowl",
      "source": "database",
      "blendSlot": 1,
      "rankingReason": "db_ranked",
      "matchPercent": 0.75,
      "matchedIngredients": ["chicken", "rice", "garlic"],
      "missingIngredients": ["soy sauce"],
      "optionalSubstitutions": [
        { "missing": "soy sauce", "substitutes": ["tamari", "coconut aminos"] }
      ],
      "prepMinutes": 25,
      "difficulty": "easy"
    }
  ]
}
```

## Error Cases

- `400`: invalid payload or empty `ingredients`.
- `422`: invalid `mode`.
- `500`: internal search failure.

## Testing Expectations

- Unit tests for normalization and ranking tie-breakers.
- Integration tests for strict vs inclusive output differences.
- Contract tests that validate response schema and required fields.
