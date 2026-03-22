# DB-First + LLM Fallback Contract

Status: v1 locked baseline (update via checklist preflight item)

## Decision

- Primary source is database recipes.
- LLM is fallback only when DB results are low-confidence or empty.

## Fallback Trigger Rules

Trigger LLM generation when either is true:

1. DB returns `0` results.
2. Top DB `matchPercent` is below `0.45` and fewer than `5` results are returned.

Do not trigger fallback if user sets `dbOnly=true`.

## LLM Input Contract

- Normalized ingredient list.
- Search mode (`strict` or `inclusive`).
- Constraints: dietary tags, max prep time, cuisine, servings.
- Required output schema (JSON only).

## LLM Output Schema (Required)

```json
{
  "recipes": [
    {
      "name": "string",
      "description": "string",
      "ingredients": [
        { "name": "string", "amount": "string", "optional": false }
      ],
      "steps": ["string"],
      "prepMinutes": 30,
      "cookMinutes": 20,
      "difficulty": "easy",
      "cuisine": "string",
      "dietaryTags": ["string"],
      "servings": 2,
      "safetyNotes": ["string"]
    }
  ]
}
```

## Validation and Persistence

- Reject non-JSON or schema-invalid responses.
- Normalize ingredient names before persistence.
- Persist generated recipes with:
  - `source=llm`
  - `generationModel`
  - `generationTimestamp`
  - `promptVersion`
- Mark generated recipes as `reviewable=true`.

## Response Composition

- Return DB matches first, then generated results.
- Include `source` per recipe (`database` or `llm`).
- Include explanation metadata (`matchedIngredients`, `missingIngredients`).

## Caching Policy

- Cache by normalized ingredient hash + mode + filters.
- Default TTL: 6 hours.
- Bypass cache when `debugNoCache=true` (non-production only).

## Safety Rules

- Block unsafe cooking instructions or dangerous ingredient pairings when detected.
- Include allergy-sensitive reminders when common allergens appear.
- Never claim medical/nutritional certainty.

## Testing Expectations

- Mock provider responses for deterministic tests.
- Validate fallback trigger edge cases.
- Validate schema rejection and recovery path.
- Ensure provenance fields are always present for generated recipes.
