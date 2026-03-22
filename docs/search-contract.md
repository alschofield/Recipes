# Search Contract (`strict` vs `inclusive`)

Status: v1 locked baseline (update via checklist preflight item)

## Endpoint

- `POST /recipe/search`

## Request Body

```json
{
  "ingredients": ["chicken", "rice", "garlic"],
  "mode": "strict",
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
- `dbOnly=true`: skip LLM fallback and return only database-backed results.

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
